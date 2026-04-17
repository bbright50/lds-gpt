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
	maxPhrases              = 20   // cap per TG/IDX composition
	defaultEmbedBatchSize   = 32   // used when Loader.embedBatchSize is unset
	defaultEmbedConcurrency = 1    // stock Ollama serialises per model
	defaultEmbedMaxTextLen  = 2000 // safe for mxbai-embed-large (512 tokens)
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

// needsEmbedWhere builds a WHERE fragment that matches "not yet embedded".
//
// At node-create time every @vector slot gets populated with a plain LIST
// placeholder (the schema forces `[Float!]!` non-null). Phase 6 upgrades
// that LIST to a VECTORF32 via `SET n.<prop> = vecf32(...)`. So the ground
// truth is: LIST ⇒ still needs embedding; VECTORF32 ⇒ already done.
//
// `typeOf(x)[0] = …` would error on VECTORF32 (`Type mismatch: expected
// Map, Node, Edge, List, or Null but was Vectorf32`), which bricked `task
// embed` reruns. typeOf short-circuits that cleanly. The `IS NULL` arm
// keeps unit-test seeds that create nodes without an embedding property
// working (the real loader never takes that branch because the non-null
// schema forces the placeholder).
func needsEmbedWhere(alias, prop string) string {
	return fmt.Sprintf("(%[1]s.%[2]s IS NULL OR typeOf(%[1]s.%[2]s) = 'List')", alias, prop)
}

func (l *Loader) embedVerseGroups(ctx context.Context) (int, error) {
	targets, err := l.fetchTargets(ctx,
		fmt.Sprintf(`MATCH (g:VerseGroup)
		 WHERE %s AND g.text IS NOT NULL AND g.text <> ''
		 RETURN g.id AS id, g.text AS text`, needsEmbedWhere("g", "embedding")))
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "VerseGroup", "embedding", &l.stats.EmbVerseGroups)
}

func (l *Loader) embedChapters(ctx context.Context) (int, error) {
	targets, err := l.fetchTargets(ctx,
		fmt.Sprintf(`MATCH (c:Chapter)
		 WHERE %s AND c.summary IS NOT NULL AND c.summary <> ''
		 RETURN c.id AS id, c.summary AS text`, needsEmbedWhere("c", "summaryEmbedding")))
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "Chapter", "summaryEmbedding", &l.stats.EmbChapters)
}

func (l *Loader) embedBDEntries(ctx context.Context) (int, error) {
	targets, err := l.fetchTargets(ctx,
		fmt.Sprintf(`MATCH (b:BibleDictEntry)
		 WHERE %s AND b.text IS NOT NULL AND b.text <> ''
		 RETURN b.id AS id, b.text AS text`, needsEmbedWhere("b", "embedding")))
	if err != nil {
		return 0, err
	}
	return l.runEmbeddingWorkers(ctx, targets, "BibleDictEntry", "embedding", &l.stats.EmbBDEntries)
}

func (l *Loader) embedJSTPassages(ctx context.Context) (int, error) {
	// Prefer "summary + text" when summary is non-empty; otherwise just text.
	targets, err := l.fetchTargets(ctx,
		fmt.Sprintf(`MATCH (j:JSTPassage)
		 WHERE %s AND j.text IS NOT NULL AND j.text <> ''
		 RETURN j.id AS id,
		        CASE WHEN j.summary IS NULL OR j.summary = '' THEN j.text
		             ELSE j.summary + ' ' + j.text END AS text`, needsEmbedWhere("j", "embedding")))
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
		`MATCH (n:%s) WHERE %s
		 OPTIONAL MATCH (n)-[r:%s]->(:Verse) WHERE r.phrase IS NOT NULL AND r.phrase <> ''
		 WITH n, collect(r.phrase) AS phrases
		 RETURN n.id AS id, n.name AS name, phrases`,
		label, needsEmbedWhere("n", embeddingProp), relType,
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

// runEmbeddingWorkers embeds every target in batches through the embedding
// client's /api/embed endpoint, then issues a `SET n.<prop> = vecf32($vec)`
// per result. Batch size comes from WithEmbedBatchSize; up to
// embedConcurrency batches are in flight at once. Per-batch failures are
// logged as warnings and don't abort the phase — matches the original
// LibSQL behaviour where one flaky call couldn't sink a long load.
func (l *Loader) runEmbeddingWorkers(
	ctx context.Context,
	targets []embedTarget,
	label, embeddingProp string,
	counter *int,
) (int, error) {
	if len(targets) == 0 {
		return 0, nil
	}
	batchSize := l.embedBatchSize
	if batchSize <= 0 {
		batchSize = defaultEmbedBatchSize
	}
	concurrency := l.embedConcurrency
	if concurrency <= 0 {
		concurrency = defaultEmbedConcurrency
	}
	maxTextLen := l.embedMaxTextLen
	if maxTextLen <= 0 {
		maxTextLen = defaultEmbedMaxTextLen
	}

	var written int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i := 0; i < len(targets); i += batchSize {
		end := i + batchSize
		if end > len(targets) {
			end = len(targets)
		}
		batch := targets[i:end]
		g.Go(func() error {
			texts := make([]string, len(batch))
			for j, tgt := range batch {
				texts[j] = truncate(tgt.text, maxTextLen)
			}
			vectors, err := l.embedClient.EmbedBatch(gCtx, texts)
			if err != nil {
				// Ollama rejects an entire batch with 400 when any single item
				// exceeds the context window, and transient network errors
				// similarly take the whole batch down. Fall back to per-item
				// calls so one bad chunk doesn't cost us 63 healthy siblings.
				l.stats.Warn(fmt.Sprintf("embedding batch of %d %s failed, retrying per-item: %v", len(batch), label, err))
				for j, tgt := range batch {
					vec, itemErr := l.embedClient.EmbedText(gCtx, texts[j])
					if itemErr != nil {
						l.stats.Warn(fmt.Sprintf("embedding %s %s: %v", label, tgt.id, itemErr))
						continue
					}
					if writeErr := l.writeEmbedding(gCtx, label, tgt.id, embeddingProp, vec); writeErr != nil {
						l.stats.Warn(fmt.Sprintf("writing embedding for %s %s: %v", label, tgt.id, writeErr))
						continue
					}
					atomic.AddInt64(&written, 1)
				}
				return nil
			}
			if len(vectors) != len(batch) {
				l.stats.Warn(fmt.Sprintf(
					"embedding batch of %d %s: got %d vectors back, skipping batch",
					len(batch), label, len(vectors),
				))
				return nil
			}
			for j, tgt := range batch {
				if err := l.writeEmbedding(gCtx, label, tgt.id, embeddingProp, vectors[j]); err != nil {
					l.stats.Warn(fmt.Sprintf("writing embedding for %s %s: %v", label, tgt.id, err))
					continue
				}
				atomic.AddInt64(&written, 1)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return 0, err
	}
	*counter += int(written)
	return int(written), nil
}

// writeEmbedding runs a typed `updateXxx(where: {id: $id}, update:
// {<prop>: $vec})` mutation. The fork-patched translator wraps the @vector
// field in vecf32(...) on emit, so the property becomes a VectorF32 and
// FalkorDB's vector index picks it up.
//
// Note: `$id` and `$vec` are variable-bound at the LEAF positions inside
// literal input objects. If we instead bound `$where` and `$update` as
// whole objects, go-ormql's translator would read an empty AST at those
// positions (same limitation as nested create/connect) and emit a no-op
// Cypher statement with neither WHERE nor SET.
func (l *Loader) writeEmbedding(ctx context.Context, label, id, prop string, embedding []float64) error {
	vec := make([]any, len(embedding))
	for i, x := range embedding {
		vec[i] = x
	}

	mutName := updateMutationName(label)
	respField := updateResponseField(label)
	mutationStr := fmt.Sprintf(`
		mutation ($id: ID, $vec: [Float!]) {
		  %s(where: { id: $id }, update: { %s: $vec }) {
		    %s { id }
		  }
		}`, mutName, prop, respField)

	_, err := l.fc.GraphQL().Execute(ctx, mutationStr, map[string]any{
		"id":  id,
		"vec": vec,
	})
	return err
}

// updateMutationName / updateResponseField map a node label to the
// generator's mutation / response-field naming. go-ormql's generator
// pluralizes the response field name (and mutation field) via go-pluralize
// and camel-cases irregularly for acronym-led names (JST → jST).
func updateMutationName(label string) string {
	switch label {
	case "VerseGroup":
		return "updateVerseGroups"
	case "Chapter":
		return "updateChapters"
	case "TopicalGuideEntry":
		return "updateTopicalGuideEntries"
	case "BibleDictEntry":
		return "updateBibleDictEntries"
	case "IndexEntry":
		return "updateIndexEntries"
	case "JSTPassage":
		return "updateJSTPassages"
	}
	return ""
}

func updateResponseField(label string) string {
	switch label {
	case "VerseGroup":
		return "verseGroups"
	case "Chapter":
		return "chapters"
	case "TopicalGuideEntry":
		return "topicalGuideEntries"
	case "BibleDictEntry":
		return "bibleDictEntries"
	case "IndexEntry":
		return "indexEntries"
	case "JSTPassage":
		return "jSTPassages"
	}
	return ""
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
