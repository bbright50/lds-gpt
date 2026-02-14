package dataloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lds-gpt/internal/libsql/generated"
)

// loadStudyHelps implements Phase 2: create TG, BD, IDX, and JST entity rows.
// Returns name->ID maps for each entity type and a JSTIndex.
func (l *Loader) loadStudyHelps(ctx context.Context) (
	tgMap map[string]int,
	bdMap map[string]int,
	idxMap map[string]int,
	jstIndex JSTIndex,
	err error,
) {
	tx, err := l.ec.Tx(ctx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("starting transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	tgMap, err = l.loadTopicalGuide(ctx, tx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading topical guide: %w", err)
	}

	bdMap, err = l.loadBibleDictionary(ctx, tx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading bible dictionary: %w", err)
	}

	idxMap, err = l.loadTripleCombIndex(ctx, tx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading triple combination index: %w", err)
	}

	jstIndex, err = l.loadJST(ctx, tx)
	if err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("loading JST: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, nil, JSTIndex{}, fmt.Errorf("committing phase 2: %w", err)
	}
	committed = true

	return tgMap, bdMap, idxMap, jstIndex, nil
}

// loadTopicalGuide loads TopicalGuideEntry rows.
func (l *Loader) loadTopicalGuide(ctx context.Context, tx *generated.Tx) (map[string]int, error) {
	path := filepath.Join(l.dataDir, "topical-guide.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var tgData map[string][]TGEntryJSON
	if err := json.Unmarshal(data, &tgData); err != nil {
		return nil, fmt.Errorf("parsing topical guide JSON: %w", err)
	}

	tgMap := make(map[string]int, len(tgData))

	// Collect all topic names for bulk creation
	names := make([]string, 0, len(tgData))
	for name := range tgData {
		names = append(names, name)
	}

	// Bulk create in batches to avoid SQLite limits
	batchSize := 500
	for i := 0; i < len(names); i += batchSize {
		end := i + batchSize
		if end > len(names) {
			end = len(names)
		}
		batch := names[i:end]

		builders := make([]*generated.TopicalGuideEntryCreate, len(batch))
		for j, name := range batch {
			builders[j] = tx.TopicalGuideEntry.Create().
				SetName(name)
		}

		entries, err := tx.TopicalGuideEntry.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("bulk creating TG entries: %w", err)
		}

		for _, e := range entries {
			tgMap[e.Name] = e.ID
		}
		l.stats.TGEntries += len(entries)
	}

	l.logger.Info("loaded topical guide entries", "count", l.stats.TGEntries)
	return tgMap, nil
}

// loadBibleDictionary loads BibleDictEntry rows.
func (l *Loader) loadBibleDictionary(ctx context.Context, tx *generated.Tx) (map[string]int, error) {
	path := filepath.Join(l.dataDir, "bible-dictionary.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var bdData map[string]BDEntryJSON
	if err := json.Unmarshal(data, &bdData); err != nil {
		return nil, fmt.Errorf("parsing bible dictionary JSON: %w", err)
	}

	bdMap := make(map[string]int, len(bdData))

	// Collect entries for bulk creation
	type bdItem struct {
		name string
		text string
	}
	items := make([]bdItem, 0, len(bdData))
	for name, entry := range bdData {
		items = append(items, bdItem{name: name, text: entry.Text})
	}

	batchSize := 500
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		builders := make([]*generated.BibleDictEntryCreate, len(batch))
		for j, item := range batch {
			builders[j] = tx.BibleDictEntry.Create().
				SetName(item.name).
				SetText(item.text)
		}

		entries, err := tx.BibleDictEntry.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("bulk creating BD entries: %w", err)
		}

		for _, e := range entries {
			bdMap[e.Name] = e.ID
		}
		l.stats.BDEntries += len(entries)
	}

	l.logger.Info("loaded bible dictionary entries", "count", l.stats.BDEntries)
	return bdMap, nil
}

// loadTripleCombIndex loads IndexEntry rows.
func (l *Loader) loadTripleCombIndex(ctx context.Context, tx *generated.Tx) (map[string]int, error) {
	path := filepath.Join(l.dataDir, "triple-combination-index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var idxData map[string][]IDXEntryJSON
	if err := json.Unmarshal(data, &idxData); err != nil {
		return nil, fmt.Errorf("parsing triple combination index JSON: %w", err)
	}

	idxMap := make(map[string]int, len(idxData))

	names := make([]string, 0, len(idxData))
	for name := range idxData {
		names = append(names, name)
	}

	batchSize := 500
	for i := 0; i < len(names); i += batchSize {
		end := i + batchSize
		if end > len(names) {
			end = len(names)
		}
		batch := names[i:end]

		builders := make([]*generated.IndexEntryCreate, len(batch))
		for j, name := range batch {
			builders[j] = tx.IndexEntry.Create().
				SetName(name)
		}

		entries, err := tx.IndexEntry.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("bulk creating IDX entries: %w", err)
		}

		for _, e := range entries {
			idxMap[e.Name] = e.ID
		}
		l.stats.IDXEntries += len(entries)
	}

	l.logger.Info("loaded index entries", "count", l.stats.IDXEntries)
	return idxMap, nil
}

// loadJST loads JSTPassage rows and builds a JSTIndex.
func (l *Loader) loadJST(ctx context.Context, tx *generated.Tx) (JSTIndex, error) {
	path := filepath.Join(l.dataDir, "jst.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return JSTIndex{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var jstData []JSTChapterJSON
	if err := json.Unmarshal(data, &jstData); err != nil {
		return JSTIndex{}, fmt.Errorf("parsing JST JSON: %w", err)
	}

	jstIndex := NewJSTIndex()

	for _, ch := range jstData {
		for _, entry := range ch.Entries {
			// Concatenate verse texts
			var textParts []string
			for _, v := range entry.Verses {
				textParts = append(textParts, v.Text)
			}
			fullText := strings.Join(textParts, " ")

			if fullText == "" {
				l.stats.Warn(fmt.Sprintf("JST entry %s %s:%s has no verse text", ch.Book, ch.Chapter, entry.Comprises))
				continue
			}

			comprises := entry.Comprises
			if comprises == "" {
				// Some JST entries (e.g. Exodus 34) have empty comprises; derive from verses
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

			passage, err := tx.JSTPassage.Create().
				SetBook(ch.Book).
				SetChapter(ch.Chapter).
				SetComprises(comprises).
				SetCompareRef(entry.Compare).
				SetSummary(entry.Summary).
				SetText(fullText).
				Save(ctx)
			if err != nil {
				return JSTIndex{}, fmt.Errorf("creating JST passage %s %s: %w", ch.Book, ch.Chapter, err)
			}

			// Index by book slug + chapter
			bookSlug := l.jstBookToSlug(ch.Book)
			if bookSlug != "" {
				jstIndex.Put(bookSlug, ch.Chapter, entry.Comprises, passage.ID)
			}

			l.stats.JSTPassages++
		}
	}

	l.logger.Info("loaded JST passages", "count", l.stats.JSTPassages)
	return jstIndex, nil
}

// jstBookToSlug converts a JST full book name (e.g. "1 Samuel") to its slug.
func (l *Loader) jstBookToSlug(book string) string {
	info, ok := l.bookNames[book]
	if ok {
		return info.Slug
	}
	l.stats.Warn(fmt.Sprintf("unknown JST book name: %q", book))
	return ""
}
