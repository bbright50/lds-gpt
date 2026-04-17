package dataloader

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Phase 5 — create sliding-window VerseGroup nodes (the primary RAG
// retrieval unit). Re-walks the scripture JSON tree to get chapters + their
// verses in reading order, then emits a VerseGroup per (start, start+window)
// slice, stepping by `groupStepSize`. Each group gets INCLUDES edges to its
// constituent verses and a HAS_GROUP edge from its chapter.
//
// The VerseGroup.embedding property is set later by Phase 6 via
// `SET g.embedding = vecf32(...)`.

const (
	groupWindowSize = 5
	groupStepSize   = 3

	verseGroupBatchSize = 500
)

type verseGroupRow struct {
	id            string
	chapterID     string
	text          string
	startVerseNum int
	endVerseNum   int
	verseIDs      []string
}

func (l *Loader) loadVerseGroups(ctx context.Context) error {
	var rows []verseGroupRow

	for _, volAbbrev := range volumeAbbreviations {
		volDir := filepath.Join(l.dataDir, volAbbrev)
		bookSlugs, err := listBookSlugs(volDir)
		if err != nil {
			// Missing volume; Phase 1 already warned.
			continue
		}
		for _, bookSlug := range bookSlugs {
			bookDir := filepath.Join(volDir, bookSlug)
			chapterFiles, err := listChapterFiles(bookDir)
			if err != nil {
				return fmt.Errorf("listing chapters for %s/%s: %w", volAbbrev, bookSlug, err)
			}
			for _, chFile := range chapterFiles {
				chPath := filepath.Join(bookDir, chFile)
				chJSON, err := readChapterJSON(chPath)
				if err != nil {
					return fmt.Errorf("reading %s: %w", chPath, err)
				}
				rows = append(rows, l.buildVerseGroupsForChapter(volAbbrev, bookSlug, chJSON)...)
			}
		}
	}

	if err := l.writeVerseGroups(ctx, rows); err != nil {
		return err
	}
	l.stats.VerseGroups += len(rows)
	return nil
}

func (l *Loader) buildVerseGroupsForChapter(vol, slug string, ch ChapterJSON) []verseGroupRow {
	if len(ch.Verses) == 0 {
		return nil
	}

	// Ensure verses are ordered by number. The scraped JSON already walks the
	// page top-down so the order is usually correct, but sort defensively —
	// the original LibSQL loader also re-sorted.
	verses := make([]VerseJSON, len(ch.Verses))
	copy(verses, ch.Verses)
	sort.Slice(verses, func(i, j int) bool { return verses[i].Number < verses[j].Number })

	chapterID := chapterNodeID(vol, slug, ch.Chapter)
	var rows []verseGroupRow

	for start := 0; start < len(verses); start += groupStepSize {
		end := start + groupWindowSize
		if end > len(verses) {
			end = len(verses)
		}
		window := verses[start:end]
		if len(window) == 0 {
			continue
		}

		parts := make([]string, 0, len(window))
		verseIDs := make([]string, 0, len(window))
		for _, v := range window {
			parts = append(parts, v.Text)
			verseIDs = append(verseIDs, VerseNodeID(vol, slug, ch.Chapter, v.Number))
		}

		startNum, endNum := window[0].Number, window[len(window)-1].Number
		rows = append(rows, verseGroupRow{
			id:            verseGroupNodeID(vol, slug, ch.Chapter, startNum, endNum),
			chapterID:     chapterID,
			text:          strings.Join(parts, " "),
			startVerseNum: startNum,
			endVerseNum:   endNum,
			verseIDs:      verseIDs,
		})

		if end >= len(verses) {
			break
		}
	}
	return rows
}

func (l *Loader) writeVerseGroups(ctx context.Context, rows []verseGroupRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += verseGroupBatchSize {
		end := i + verseGroupBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			// verseIDs has to be passed as []interface{} for falkordb-go's
			// ToString — the BuildParamsHeader only handles []interface{} and
			// []string at the slice level. Strings would also work since
			// ToString has a []string branch; explicit any makes the shape
			// match what we use for other parameters in the package.
			ids := make([]interface{}, len(r.verseIDs))
			for j, id := range r.verseIDs {
				ids[j] = id
			}
			batch = append(batch, map[string]interface{}{
				"id":        r.id,
				"chapterId": r.chapterID,
				"text":      r.text,
				"startNum":  r.startVerseNum,
				"endNum":    r.endVerseNum,
				"verseIds":  ids,
			})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (c:Chapter {id: r.chapterId})
			 CREATE (g:VerseGroup {
			   id: r.id,
			   text: r.text,
			   startVerseNumber: r.startNum,
			   endVerseNumber: r.endNum
			 })
			 CREATE (c)-[:HAS_GROUP]->(g)
			 WITH g, r
			 UNWIND r.verseIds AS vid
			 MATCH (v:Verse {id: vid})
			 CREATE (g)-[:INCLUDES]->(v)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating verse groups: %w", err)
		}
	}
	return nil
}

func verseGroupNodeID(vol, slug string, chapter, start, end int) string {
	return fmt.Sprintf("vg/%s/%s/%d/%d-%d", vol, slug, chapter, start, end)
}
