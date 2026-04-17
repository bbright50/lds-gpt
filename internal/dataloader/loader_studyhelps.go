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
// JSTPassage nodes from the four study-help JSON files. Returns name → node-ID
// maps (used by Phases 3–4 to connect verses) and a JSTIndex keyed by
// (bookSlug, chapter). Writes go through the raw FalkorDB handle because
// each of these node types has a required @vector `embedding` field that
// Phase 6 backfills via `SET n.embedding = vecf32(...)`.

const studyHelpBatchSize = 500

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

	names := make([]string, 0, len(tgData))
	for name := range tgData {
		names = append(names, name)
	}

	tgMap := make(map[string]string, len(names))
	for i := 0; i < len(names); i += studyHelpBatchSize {
		end := i + studyHelpBatchSize
		if end > len(names) {
			end = len(names)
		}
		rows := make([]interface{}, 0, end-i)
		for _, name := range names[i:end] {
			id := tgNodeID(name)
			tgMap[name] = id
			rows = append(rows, map[string]interface{}{"id": id, "name": name})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 CREATE (:TopicalGuideEntry {id: r.id, name: r.name})`,
			map[string]interface{}{"rows": rows}, nil,
		); err != nil {
			return nil, fmt.Errorf("bulk creating TG entries: %w", err)
		}
		l.stats.TGEntries += end - i
	}

	l.logger.Info("loaded topical guide entries", "count", l.stats.TGEntries)
	return tgMap, nil
}

func (l *Loader) loadBibleDictionary(ctx context.Context) (map[string]string, error) {
	bdData := map[string]BDEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "bible-dictionary.json", &bdData); err != nil || !ok {
		return map[string]string{}, err
	}

	type item struct{ name, text string }
	items := make([]item, 0, len(bdData))
	for name, entry := range bdData {
		items = append(items, item{name: name, text: entry.Text})
	}

	bdMap := make(map[string]string, len(items))
	for i := 0; i < len(items); i += studyHelpBatchSize {
		end := i + studyHelpBatchSize
		if end > len(items) {
			end = len(items)
		}
		rows := make([]interface{}, 0, end-i)
		for _, it := range items[i:end] {
			id := bdNodeID(it.name)
			bdMap[it.name] = id
			rows = append(rows, map[string]interface{}{"id": id, "name": it.name, "text": it.text})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 CREATE (:BibleDictEntry {id: r.id, name: r.name, text: r.text})`,
			map[string]interface{}{"rows": rows}, nil,
		); err != nil {
			return nil, fmt.Errorf("bulk creating BD entries: %w", err)
		}
		l.stats.BDEntries += end - i
	}

	l.logger.Info("loaded bible dictionary entries", "count", l.stats.BDEntries)
	return bdMap, nil
}

func (l *Loader) loadTripleCombIndex(ctx context.Context) (map[string]string, error) {
	idxData := map[string][]IDXEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "triple-combination-index.json", &idxData); err != nil || !ok {
		return map[string]string{}, err
	}

	names := make([]string, 0, len(idxData))
	for name := range idxData {
		names = append(names, name)
	}

	idxMap := make(map[string]string, len(names))
	for i := 0; i < len(names); i += studyHelpBatchSize {
		end := i + studyHelpBatchSize
		if end > len(names) {
			end = len(names)
		}
		rows := make([]interface{}, 0, end-i)
		for _, name := range names[i:end] {
			id := idxNodeID(name)
			idxMap[name] = id
			rows = append(rows, map[string]interface{}{"id": id, "name": name})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 CREATE (:IndexEntry {id: r.id, name: r.name})`,
			map[string]interface{}{"rows": rows}, nil,
		); err != nil {
			return nil, fmt.Errorf("bulk creating IDX entries: %w", err)
		}
		l.stats.IDXEntries += end - i
	}

	l.logger.Info("loaded index entries", "count", l.stats.IDXEntries)
	return idxMap, nil
}

func (l *Loader) loadJST(ctx context.Context) (JSTIndex, error) {
	jstIndex := NewJSTIndex()
	var jstData []JSTChapterJSON
	if ok, err := l.readOptionalJSON(ctx, "jst.json", &jstData); err != nil || !ok {
		return jstIndex, err
	}

	type passageRow struct {
		id, book, chapter, comprises, compareRef, summary, text string
	}
	var rows []passageRow

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
			rows = append(rows, passageRow{
				id:         id,
				book:       ch.Book,
				chapter:    ch.Chapter,
				comprises:  comprises,
				compareRef: entry.Compare,
				summary:    entry.Summary,
				text:       fullText,
			})

			if bookSlug != "" {
				jstIndex.Put(bookSlug, ch.Chapter, entry.Comprises, id)
			}
		}
	}

	for i := 0; i < len(rows); i += studyHelpBatchSize {
		end := i + studyHelpBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{
				"id":         r.id,
				"book":       r.book,
				"chapter":    r.chapter,
				"comprises":  r.comprises,
				"compareRef": r.compareRef,
				"summary":    r.summary,
				"text":       r.text,
			})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 CREATE (:JSTPassage {
			   id: r.id, book: r.book, chapter: r.chapter, comprises: r.comprises,
			   compareRef: r.compareRef, summary: r.summary, text: r.text
			 })`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return jstIndex, fmt.Errorf("bulk creating JST passages: %w", err)
		}
		l.stats.JSTPassages += end - i
	}

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
