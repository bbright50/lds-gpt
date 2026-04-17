package dataloader

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Phase 5 — create sliding-window VerseGroup nodes (the primary RAG
// retrieval unit). Typed `createVerseGroups` carries both edge kinds as
// nested connect payloads:
//   * chapter: single connect — Chapter-[:HAS_GROUP]->VerseGroup (IN direction)
//   * verses:  list connect   — VerseGroup-[:INCLUDES]->Verse    (OUT direction)
// The fork-patched translator unrolls both via runtime UNWIND, so one
// typed mutation per sub-batch covers nodes + edges.
//
// VerseGroup.embedding is a required @vector field; Phase 6 upgrades the
// placeholder via typed updateXxx + fork-patched vecf32 auto-wrap.

const (
	groupWindowSize = 5
	groupStepSize   = 3
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
			continue // Phase 1 already warned about missing volume dirs
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

	if len(rows) == 0 {
		return nil
	}

	// Typed bulk create carrying both edge kinds as nested connect payloads.
	// `chapter` is a single connect (each VerseGroup belongs to exactly one
	// Chapter, in IN direction). `verses` is a list connect (each VerseGroup
	// includes multiple Verses via INCLUDES in OUT direction).
	nodeRows := make([]any, 0, len(rows))
	for _, r := range rows {
		verseConnects := make([]any, 0, len(r.verseIDs))
		for _, vid := range r.verseIDs {
			verseConnects = append(verseConnects, map[string]any{"where": map[string]any{"id": vid}})
		}
		nodeRows = append(nodeRows, map[string]any{
			"id":               r.id,
			"text":             r.text,
			"startVerseNumber": r.startVerseNum,
			"endVerseNumber":   r.endVerseNum,
			"embedding":        placeholderEmbedding(),
			"chapter":          connectById(r.chapterID),
			"verses":           map[string]any{"connect": verseConnects},
		})
	}
	if _, err := l.fc.GraphQL().Execute(ctx, `
		mutation ($input: [VerseGroupCreateInput!]!) {
		  createVerseGroups(input: $input) { verseGroups { id } }
		}`, map[string]any{"input": nodeRows}); err != nil {
		return fmt.Errorf("createVerseGroups: %w", err)
	}

	l.stats.VerseGroups += len(rows)
	return nil
}

func (l *Loader) buildVerseGroupsForChapter(vol, slug string, ch ChapterJSON) []verseGroupRow {
	if len(ch.Verses) == 0 {
		return nil
	}

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

func verseGroupNodeID(vol, slug string, chapter, start, end int) string {
	return fmt.Sprintf("vg/%s/%s/%d/%d-%d", vol, slug, chapter, start, end)
}
