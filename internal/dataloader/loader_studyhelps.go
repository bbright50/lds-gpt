package dataloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Phase 2 — create TopicalGuideEntry / BibleDictEntry / IndexEntry /
// JSTPassage nodes from the four study-help JSON files. Returns name →
// id maps (used by Phases 3–4 to connect verses) and a JSTIndex keyed
// by (bookSlug, chapter).
//
// Writes flow through the typed go-ormql client; auto-chunking handles the
// per-batch loop we used to write by hand. Every @vector field gets a
// fractional-zeros placeholder — the spike
// TestSpike_VectorIndexOnPlainList_VsVecf32 proved FalkorDB's vector index is
// strict (ignores plain lists), so these placeholders don't pollute kNN.
// Phase 6 upgrades them to real vectors via typed updateXxx (fork-patched to
// wrap the @vector SET value in vecf32).

func (l *Loader) loadStudyHelps(ctx context.Context) (
	tgMap, bdMap, idxMap map[string]string,
	jstIndex JSTIndex,
	err error,
) {
	tgMap, err = l.loadTopicalGuide(ctx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading topical guide: %w", err)
	}

	bdMap, err = l.loadBibleDictionary(ctx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading bible dictionary: %w", err)
	}

	idxMap, err = l.loadTripleCombIndex(ctx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading triple combination index: %w", err)
	}

	jstIndex, err = l.loadJST(ctx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading JST: %w", err)
	}

	return tgMap, bdMap, idxMap, jstIndex, nil
}

func (l *Loader) loadTopicalGuide(ctx context.Context) (map[string]string, error) {
	tgData := map[string][]TGEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "topical-guide.json", &tgData); err != nil || !ok {
		return map[string]string{}, err
	}

	tgMap := make(map[string]string, len(tgData))
	rows := make([]any, 0, len(tgData))
	for name := range tgData {
		id := tgNodeID(name)
		tgMap[name] = id
		rows = append(rows, map[string]any{
			"id": id,
			"name":       name,
			"embedding":  placeholderEmbedding(),
		})
	}
	if len(rows) > 0 {
		if _, err := l.fc.GraphQL().Execute(ctx, `
			mutation ($input: [TopicalGuideEntryCreateInput!]!) {
			  createTopicalGuideEntries(input: $input) {
			    topicalGuideEntries { id }
			  }
			}`, map[string]any{"input": rows}); err != nil {
			return nil, fmt.Errorf("createTopicalGuideEntries: %w", err)
		}
	}
	l.stats.TGEntries += len(rows)
	l.logger.Info("loaded topical guide entries", "count", l.stats.TGEntries)
	return tgMap, nil
}

func (l *Loader) loadBibleDictionary(ctx context.Context) (map[string]string, error) {
	bdData := map[string]BDEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "bible-dictionary.json", &bdData); err != nil || !ok {
		return map[string]string{}, err
	}

	bdMap := make(map[string]string, len(bdData))
	rows := make([]any, 0, len(bdData))
	for name, entry := range bdData {
		id := bdNodeID(name)
		bdMap[name] = id
		rows = append(rows, map[string]any{
			"id": id,
			"name":       name,
			"text":       entry.Text,
			"embedding":  placeholderEmbedding(),
		})
	}
	if len(rows) > 0 {
		if _, err := l.fc.GraphQL().Execute(ctx, `
			mutation ($input: [BibleDictEntryCreateInput!]!) {
			  createBibleDictEntries(input: $input) {
			    bibleDictEntries { id }
			  }
			}`, map[string]any{"input": rows}); err != nil {
			return nil, fmt.Errorf("createBibleDictEntries: %w", err)
		}
	}
	l.stats.BDEntries += len(rows)
	l.logger.Info("loaded bible dictionary entries", "count", l.stats.BDEntries)
	return bdMap, nil
}

func (l *Loader) loadTripleCombIndex(ctx context.Context) (map[string]string, error) {
	idxData := map[string][]IDXEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "triple-combination-index.json", &idxData); err != nil || !ok {
		return map[string]string{}, err
	}

	idxMap := make(map[string]string, len(idxData))
	rows := make([]any, 0, len(idxData))
	for name := range idxData {
		id := idxNodeID(name)
		idxMap[name] = id
		rows = append(rows, map[string]any{
			"id": id,
			"name":       name,
			"embedding":  placeholderEmbedding(),
		})
	}
	if len(rows) > 0 {
		if _, err := l.fc.GraphQL().Execute(ctx, `
			mutation ($input: [IndexEntryCreateInput!]!) {
			  createIndexEntries(input: $input) {
			    indexEntries { id }
			  }
			}`, map[string]any{"input": rows}); err != nil {
			return nil, fmt.Errorf("createIndexEntries: %w", err)
		}
	}
	l.stats.IDXEntries += len(rows)
	l.logger.Info("loaded index entries", "count", l.stats.IDXEntries)
	return idxMap, nil
}

func (l *Loader) loadJST(ctx context.Context) (JSTIndex, error) {
	jstIndex := NewJSTIndex()
	var jstData []JSTChapterJSON
	if ok, err := l.readOptionalJSON(ctx, "jst.json", &jstData); err != nil || !ok {
		return jstIndex, err
	}

	rows := make([]any, 0, 128)
	for _, ch := range jstData {
		for _, entry := range ch.Entries {
			var parts []string
			for _, v := range entry.Verses {
				parts = append(parts, v.Text)
			}
			fullText := strings.Join(parts, " ")
			if fullText == "" {
				l.stats.Warn(fmt.Sprintf("JST entry %s %s:%s has no verse text", ch.Book, ch.Chapter, entry.Comprises))
				continue
			}

			comprises := entry.Comprises
			if comprises == "" {
				if len(entry.Verses) > 0 {
					comprises = strconv.Itoa(entry.Verses[0].Number)
					if len(entry.Verses) > 1 {
						comprises += "-" + strconv.Itoa(entry.Verses[len(entry.Verses)-1].Number)
					}
				} else {
					comprises = "unknown"
				}
				l.stats.Warn(fmt.Sprintf("JST entry %s %s has empty comprises, using %q", ch.Book, ch.Chapter, comprises))
			}

			bookSlug := l.jstBookToSlug(ch.Book)
			id := jstNodeID(bookSlug, ch.Chapter, comprises)
			rows = append(rows, map[string]any{
				"id": id,
				"book":       ch.Book,
				"chapter":    ch.Chapter,
				"comprises":  comprises,
				"compareRef": entry.Compare,
				"summary":    entry.Summary,
				"text":       fullText,
				"embedding":  placeholderEmbedding(),
			})

			if bookSlug != "" {
				jstIndex.Put(bookSlug, ch.Chapter, entry.Comprises, id)
			}
		}
	}

	if len(rows) > 0 {
		if _, err := l.fc.GraphQL().Execute(ctx, `
			mutation ($input: [JSTPassageCreateInput!]!) {
			  createJSTPassages(input: $input) {
			    jSTPassages { id }
			  }
			}`, map[string]any{"input": rows}); err != nil {
			return jstIndex, fmt.Errorf("createJSTPassages: %w", err)
		}
	}
	l.stats.JSTPassages += len(rows)
	l.logger.Info("loaded JST passages", "count", l.stats.JSTPassages)
	return jstIndex, nil
}

// readOptionalJSON reads <dataDir>/<file> and unmarshals it into target.
// A missing file is treated as "no data" — returns (false, nil) with a
// warning so a developer can run Phase 2 incrementally before the study
// helps have been scraped. Any other error is propagated.
func (l *Loader) readOptionalJSON(ctx context.Context, file string, target any) (bool, error) {
	_ = ctx
	path := filepath.Join(l.dataDir, file)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			l.stats.Warn(fmt.Sprintf("%s not present in %s; skipping", file, l.dataDir))
			return false, nil
		}
		return false, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("parsing %s: %w", path, err)
	}
	return true, nil
}

func (l *Loader) jstBookToSlug(book string) string {
	if info, ok := l.bookNames[book]; ok {
		return info.Slug
	}
	l.stats.Warn(fmt.Sprintf("unknown JST book name: %q", book))
	return ""
}

// Deterministic IDs. Names can contain spaces / punctuation; we store them
// verbatim as Cypher string property values (passed as params, not
// interpolated), so no escaping is required.
func tgNodeID(name string) string  { return "tg/" + name }
func bdNodeID(name string) string  { return "bd/" + name }
func idxNodeID(name string) string { return "idx/" + name }
func jstNodeID(bookSlug, chapter, comprises string) string {
	return "jst/" + bookSlug + "/" + chapter + "/" + comprises
}

// placeholderEmbedding returns a 1024-length slice of small fractional
// values as an []any (float64). Used wherever the schema's @vector field
// requires a non-null value at create time but Phase 6 hasn't computed the
// real embedding yet. Fractional values (0.0001) avoid falkordb-go's
// ToString rendering as LIST<INTEGER> and, per spike, FalkorDB's vector
// index ignores plain-list values — so the placeholder does not pollute
// kNN until Phase 6 upgrades it via `SET n.<prop> = vecf32(...)`.
func placeholderEmbedding() []any {
	out := make([]any, 1024)
	for i := range out {
		out[i] = 0.0001
	}
	return out
}
