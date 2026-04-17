package dataloader

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// Phase 6 — generate Bedrock embeddings for the six embeddable node types
// and persist them as `vecf32(...)` values on the corresponding `embedding`
// (or `summaryEmbedding`) property.
//
// One worker per node, bounded by `embedConcurrency`. The Bedrock client
// lives on the Loader (attached via WithEmbedClient); this phase is a no-op
// when no client is configured — Run() only enters Phase 6 if the client is
// present; EmbedOnly() errors out up-front.

const (
	embedConcurrency = 8
	maxTextLen       = 25000 // Titan v2 context-window safety margin
	maxPhrases       = 20    // cap per TG/IDX composition
)

type embedTarget struct {
	id   string
	text string
}

// embedAll runs the six per-label embedders sequentially. Each embedder
// internally parallelises with its own errgroup.
func (l *Loader) embedAll(ctx context.Context) error {
	if l.embedClient == nil {
		return fmt.Errorf("no embed client configured")
	}

	type phase struct {
		name string
		run  func(context.Context) (int, error)
	}
	phases := []phase{
		{"verse groups", l.embedVerseGroups},
		{"chapters", l.embedChapters},
		{"topical guide", l.embedTGEntries},
		{"bible dictionary", l.embedBDEntries},
		{"triple combination index", l.embedIDXEntries},
		{"JST passages", l.embedJSTPassages},
	}
	for _, p := range phases {
		count, err := p.run(ctx)
		if err != nil {
			return fmt.Errorf("embedding %s: %w", p.name, err)
		}
		l.logger.Info("phase 6 group complete", "group", p.name, "embedded", count)
	}
	return nil
}

// --- Per-entity-type embedders ---

func (l *Loader) embedVerseGroups(ctx context.Context) (int, error) {
	targets, err := l.fetchTargets(ctx,
		`MATCH (g:VerseGroup) WHERE g.embedding IS NULL AND g.text IS NOT NULL AND g.text <> ''
		 RETURN g.id AS id, g.text AS text`)
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "VerseGroup", "embedding", &l.stats.EmbVerseGroups)
}

func (l *Loader) embedChapters(ctx context.Context) (int, error) {
	targets, err := l.fetchTargets(ctx,
		`MATCH (c:Chapter) WHERE c.summaryEmbedding IS NULL AND c.summary IS NOT NULL AND c.summary <> ''
		 RETURN c.id AS id, c.summary AS text`)
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "Chapter", "summaryEmbedding", &l.stats.EmbChapters)
}

func (l *Loader) embedBDEntries(ctx context.Context) (int, error) {
	targets, err := l.fetchTargets(ctx,
		`MATCH (b:BibleDictEntry) WHERE b.embedding IS NULL AND b.text IS NOT NULL AND b.text <> ''
		 RETURN b.id AS id, b.text AS text`)
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "BibleDictEntry", "embedding", &l.stats.EmbBDEntries)
}

func (l *Loader) embedJSTPassages(ctx context.Context) (int, error) {
	// Prefer "summary + text" when summary is non-empty; otherwise just text.
	targets, err := l.fetchTargets(ctx,
		`MATCH (j:JSTPassage)
		 WHERE j.embedding IS NULL AND j.text IS NOT NULL AND j.text <> ''
		 RETURN j.id AS id,
		        CASE WHEN j.summary IS NULL OR j.summary = '' THEN j.text
		             ELSE j.summary + ' ' + j.text END AS text`)
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "JSTPassage", "embedding", &l.stats.EmbJSTPassages)
}

// TG and IDX have no free-text body in the schema — their embedding input
// is composed from the entry's name plus up to `maxPhrases` phrases pulled
// off their *_VERSE_REF edges. The composition is a single Cypher query
// with OPTIONAL MATCH + collect.
func (l *Loader) embedTGEntries(ctx context.Context) (int, error) {
	return l.embedNameWithPhrases(ctx,
		"TopicalGuideEntry", "TG_VERSE_REF", "embedding", &l.stats.EmbTGEntries,
	)
}

func (l *Loader) embedIDXEntries(ctx context.Context) (int, error) {
	return l.embedNameWithPhrases(ctx,
		"IndexEntry", "IDX_VERSE_REF", "embedding", &l.stats.EmbIDXEntries,
	)
}

func (l *Loader) embedNameWithPhrases(
	ctx context.Context,
	label, relType, embeddingProp string,
	counter *int,
) (int, error) {
	query := fmt.Sprintf(
		`MATCH (n:%s) WHERE n.%s IS NULL
		 OPTIONAL MATCH (n)-[r:%s]->(:Verse) WHERE r.phrase IS NOT NULL AND r.phrase <> ''
		 WITH n, collect(r.phrase) AS phrases
		 RETURN n.id AS id, n.name AS name, phrases`,
		label, embeddingProp, relType,
	)
	res, err := l.fc.Raw().Query(query, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("querying %s targets: %w", label, err)
	}

	var targets []embedTarget
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		name, _ := rec.Get("name")
		phrasesVal, _ := rec.Get("phrases")

		idStr, _ := id.(string)
		nameStr, _ := name.(string)
		if idStr == "" {
			continue
		}
		text := composeTextFromPhrases(nameStr, toStringSlice(phrasesVal))
		if text == "" {
			continue
		}
		targets = append(targets, embedTarget{id: idStr, text: text})
	}
	return l.runEmbeddingWorkers(ctx, targets, label, embeddingProp, counter)
}

// --- Shared plumbing ---

// fetchTargets runs a Cypher query with two returned columns (`id`, `text`)
// and materialises them into a slice of embedTargets.
func (l *Loader) fetchTargets(ctx context.Context, query string) ([]embedTarget, error) {
	_ = ctx
	res, err := l.fc.Raw().Query(query, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("querying targets: %w", err)
	}
	var out []embedTarget
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		text, _ := rec.Get("text")
		idStr, _ := id.(string)
		textStr, _ := text.(string)
		if idStr == "" || textStr == "" {
			continue
		}
		out = append(out, embedTarget{id: idStr, text: textStr})
	}
	return out, nil
}

// runEmbeddingWorkers embeds every target in parallel (bounded by
// embedConcurrency) and issues a `SET n.<prop> = vecf32($vec)` per result.
// Per-row failures are logged as warnings — one Bedrock failure does not
// abort the whole phase, matching the original LibSQL behavior.
func (l *Loader) runEmbeddingWorkers(
	ctx context.Context,
	targets []embedTarget,
	label, embeddingProp string,
	counter *int,
) (int, error) {
	if len(targets) == 0 {
		return 0, nil
	}
	var written int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, tgt := range targets {
		tgt := tgt
		g.Go(func() error {
			text := truncate(tgt.text, maxTextLen)
			embedding, err := l.embedClient.EmbedText(gCtx, text)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embedding %s %s: %v", label, tgt.id, err))
				return nil
			}
			if err := l.writeEmbedding(gCtx, label, tgt.id, embeddingProp, embedding); err != nil {
				l.stats.Warn(fmt.Sprintf("writing embedding for %s %s: %v", label, tgt.id, err))
				return nil
			}
			atomic.AddInt64(&written, 1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return 0, err
	}
	*counter += int(written)
	return int(written), nil
}

func (l *Loader) writeEmbedding(ctx context.Context, label, id, prop string, embedding []float64) error {
	_ = ctx
	vec := make([]interface{}, len(embedding))
	for i, x := range embedding {
		vec[i] = x
	}
	query := fmt.Sprintf(
		`MATCH (n:%s {id: $id})
		 SET n.%s = vecf32($vec)`,
		label, prop,
	)
	if _, err := l.fc.Raw().Query(query, map[string]interface{}{
		"id": id, "vec": vec,
	}, nil); err != nil {
		return err
	}
	return nil
}

// --- Text helpers ---

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func composeTextFromPhrases(name string, phrases []string) string {
	name = strings.TrimSpace(name)
	if name == "" && len(phrases) == 0 {
		return ""
	}
	if len(phrases) == 0 {
		return name
	}
	limit := len(phrases)
	if limit > maxPhrases {
		limit = maxPhrases
	}
	return name + ". " + strings.Join(phrases[:limit], "; ")
}

// toStringSlice coerces FalkorDB's `LIST<STRING>` return into []string. The
// underlying value is []interface{} of strings.
func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	xs, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if s, ok := x.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}
