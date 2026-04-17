package dataloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Phase 1 — create Volume/Book/Chapter/Verse nodes and structural
// relationships. Writes go through the raw FalkorDB handle because
// Chapter.summaryEmbedding is an @vector non-null in the GraphQL schema and
// embeddings do not exist yet at this phase. Phase 6 backfills embeddings
// via `SET n.summaryEmbedding = vecf32(...)` — see loader_embeddings.go.
//
// IDs are deterministic strings (e.g. "v/ot/gen/1/5") so Phases 3–5 can
// build relationships without round-tripping FalkorDB for a MATCH...RETURN id.

// loadScriptures implements Phase 1.
func (l *Loader) loadScriptures(ctx context.Context) (VerseIndex, error) {
	verseIndex := NewVerseIndex()
	graph := l.fc.Raw()

	// Volumes: one round-trip for all five.
	volRows := make([]interface{}, 0, len(volumeAbbreviations))
	for _, abbrev := range volumeAbbreviations {
		volRows = append(volRows, map[string]interface{}{
			"id":   volumeNodeID(abbrev),
			"name": volumeDisplayNames[abbrev],
			"abbr": abbrev,
		})
	}
	if _, err := graph.Query(
		`UNWIND $rows AS r
		 CREATE (:Volume {id: r.id, name: r.name, abbreviation: r.abbr})`,
		map[string]interface{}{"rows": volRows}, nil,
	); err != nil {
		return verseIndex, fmt.Errorf("creating volumes: %w", err)
	}
	l.stats.Volumes = len(volRows)

	for _, volAbbrev := range volumeAbbreviations {
		volDir := filepath.Join(l.dataDir, volAbbrev)
		if _, err := os.Stat(volDir); os.IsNotExist(err) {
			l.stats.Warn(fmt.Sprintf("volume directory %s not present under %s; skipping", volAbbrev, l.dataDir))
			continue
		}
		bookSlugs, err := listBookSlugs(volDir)
		if err != nil {
			return verseIndex, fmt.Errorf("listing books for %s: %w", volAbbrev, err)
		}

		for _, bookSlug := range bookSlugs {
			bookDir := filepath.Join(volDir, bookSlug)
			chapterFiles, err := listChapterFiles(bookDir)
			if err != nil {
				return verseIndex, fmt.Errorf("listing chapters for %s/%s: %w", volAbbrev, bookSlug, err)
			}
			if len(chapterFiles) == 0 {
				continue
			}

			firstChapter, err := readChapterJSON(filepath.Join(bookDir, chapterFiles[0]))
			if err != nil {
				return verseIndex, fmt.Errorf("reading first chapter of %s/%s: %w", volAbbrev, bookSlug, err)
			}

			bookName := l.resolveBookName(firstChapter.Book, volAbbrev, bookSlug)
			if err := l.createBook(ctx, volAbbrev, bookSlug, bookName); err != nil {
				return verseIndex, err
			}
			l.stats.Books++

			for _, chFile := range chapterFiles {
				chPath := filepath.Join(bookDir, chFile)
				chJSON, err := readChapterJSON(chPath)
				if err != nil {
					return verseIndex, fmt.Errorf("reading %s: %w", chPath, err)
				}
				if err := l.loadChapter(ctx, volAbbrev, bookSlug, chJSON, &verseIndex); err != nil {
					return verseIndex, fmt.Errorf("loading chapter %s/%s/%d: %w",
						volAbbrev, bookSlug, chJSON.Chapter, err)
				}
			}
		}
	}

	return verseIndex, nil
}

func (l *Loader) createBook(ctx context.Context, volAbbrev, bookSlug, bookName string) error {
	_, err := l.fc.Raw().Query(
		`MATCH (vol:Volume {id: $volId})
		 CREATE (b:Book {id: $id, name: $name, slug: $slug, urlPath: $url})
		 CREATE (vol)-[:CONTAINS]->(b)`,
		map[string]interface{}{
			"volId": volumeNodeID(volAbbrev),
			"id":    bookNodeID(volAbbrev, bookSlug),
			"name":  bookName,
			"slug":  bookSlug,
			"url":   volAbbrev + "/" + bookSlug,
		}, nil,
	)
	_ = ctx
	if err != nil {
		return fmt.Errorf("creating book %s/%s: %w", volAbbrev, bookSlug, err)
	}
	return nil
}

func (l *Loader) loadChapter(
	ctx context.Context,
	volume, slug string,
	ch ChapterJSON,
	verseIndex *VerseIndex,
) error {
	chapterID := chapterNodeID(volume, slug, ch.Chapter)

	// Create the chapter node + CONTAINS edge from its book. Without a
	// summaryEmbedding yet, Phase 6 fills it in via `SET`.
	chapterProps := map[string]interface{}{
		"id":     chapterID,
		"number": ch.Chapter,
		"bookId": bookNodeID(volume, slug),
	}
	q := `MATCH (b:Book {id: $bookId})
	       CREATE (c:Chapter {id: $id, number: $number`
	if ch.Summary != "" {
		chapterProps["summary"] = ch.Summary
		q += `, summary: $summary`
	}
	if ch.URL != "" {
		chapterProps["url"] = ch.URL
		q += `, url: $url`
	}
	q += `})
	      CREATE (b)-[:CONTAINS]->(c)`
	if _, err := l.fc.Raw().Query(q, chapterProps, nil); err != nil {
		return fmt.Errorf("creating chapter %d: %w", ch.Chapter, err)
	}
	l.stats.Chapters++
	if l.stats.Chapters%100 == 0 {
		l.logger.Info("progress", "chapters", l.stats.Chapters, "verses", l.stats.Verses)
	}

	// Bulk-create verses for this chapter in one UNWIND round-trip.
	abbrev := l.slugMap[volume+"/"+slug]
	verseRows := make([]interface{}, 0, len(ch.Verses))
	for _, v := range ch.Verses {
		id := VerseNodeID(volume, slug, ch.Chapter, v.Number)
		row := map[string]interface{}{
			"id":     id,
			"number": v.Number,
			"text":   v.Text,
			"ref":    fmt.Sprintf("%s %d:%d", abbrev, ch.Chapter, v.Number),
		}
		trn, or, ie := extractInlineFootnotes(ch.Footnotes, v.Number)
		if len(trn) > 0 {
			if b, err := json.Marshal(trn); err == nil {
				row["translationNotes"] = string(b)
			}
		}
		if len(or) > 0 {
			if b, err := json.Marshal(or); err == nil {
				row["alternateReadings"] = string(b)
			}
		}
		if len(ie) > 0 {
			if b, err := json.Marshal(ie); err == nil {
				row["explanatoryNotes"] = string(b)
			}
		}
		verseRows = append(verseRows, row)
		verseIndex.Put(volume, slug, ch.Chapter, v.Number, id)
	}

	if _, err := l.fc.Raw().Query(
		`MATCH (c:Chapter {id: $chapterId})
		 UNWIND $rows AS r
		 CREATE (v:Verse {
		   id: r.id,
		   number: r.number,
		   text: r.text,
		   reference: r.ref,
		   translationNotes:  coalesce(r.translationNotes,  ''),
		   alternateReadings: coalesce(r.alternateReadings, ''),
		   explanatoryNotes:  coalesce(r.explanatoryNotes,  '')
		 })
		 CREATE (c)-[:HAS_VERSE]->(v)`,
		map[string]interface{}{"chapterId": chapterID, "rows": verseRows},
		nil,
	); err != nil {
		return fmt.Errorf("bulk creating verses for chapter %d: %w", ch.Chapter, err)
	}

	l.stats.Verses += len(verseRows)
	_ = ctx
	return nil
}

// InlineFootnote types (translation / alternate-reading / explanatory) —
// JSON-serialized onto the Verse node. The Ent schema had three dedicated
// Go types for these; in FalkorDB they live as JSON strings on a single
// property per kind, so the storage layer stays simple.
type inlineTrnNote struct {
	Marker     string `json:"marker"`
	HebrewText string `json:"hebrew_text,omitempty"`
}
type inlineOrNote struct {
	Marker string `json:"marker"`
	Text   string `json:"text,omitempty"`
}
type inlineIeNote struct {
	Marker string `json:"marker"`
	Text   string `json:"text,omitempty"`
}

func extractInlineFootnotes(
	footnotes map[string]FootnoteJSON,
	verseNum int,
) ([]inlineTrnNote, []inlineOrNote, []inlineIeNote) {
	var trn []inlineTrnNote
	var or []inlineOrNote
	var ie []inlineIeNote

	prefix := strconv.Itoa(verseNum)
	for key, fn := range footnotes {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		marker := strings.TrimPrefix(key, prefix)
		if marker == "" {
			continue
		}
		for _, cat := range strings.Split(fn.Category, ",") {
			switch strings.TrimSpace(cat) {
			case "trn":
				trn = append(trn, inlineTrnNote{Marker: marker, HebrewText: fn.Text})
			case "or":
				or = append(or, inlineOrNote{Marker: marker, Text: fn.Text})
			case "ie":
				ie = append(ie, inlineIeNote{Marker: marker, Text: fn.Text})
			}
		}
	}
	return trn, or, ie
}

// --- Filesystem helpers (unchanged from the LibSQL implementation) ---

func (l *Loader) resolveBookName(jsonBookName, volume, slug string) string {
	if jsonBookName != "" {
		return jsonBookName
	}
	if abbrev, ok := l.slugMap[volume+"/"+slug]; ok {
		return abbrev
	}
	return slug
}

func listBookSlugs(volDir string) ([]string, error) {
	entries, err := os.ReadDir(volDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", volDir, err)
	}
	var slugs []string
	for _, e := range entries {
		if e.IsDir() {
			slugs = append(slugs, e.Name())
		}
	}
	sort.Strings(slugs)
	return slugs, nil
}

func listChapterFiles(bookDir string) ([]string, error) {
	entries, err := os.ReadDir(bookDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", bookDir, err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, e.Name())
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return chapterNumFromFilename(files[i]) < chapterNumFromFilename(files[j])
	})
	return files, nil
}

func chapterNumFromFilename(name string) int {
	n, err := strconv.Atoi(strings.TrimSuffix(name, ".json"))
	if err != nil {
		return 0
	}
	return n
}

func readChapterJSON(path string) (ChapterJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ChapterJSON{}, fmt.Errorf("reading file %s: %w", path, err)
	}
	var ch ChapterJSON
	if err := json.Unmarshal(data, &ch); err != nil {
		return ChapterJSON{}, fmt.Errorf("parsing JSON %s: %w", path, err)
	}
	return ch, nil
}

// --- Deterministic FalkorDB node IDs ---

func volumeNodeID(abbrev string) string            { return "vol/" + abbrev }
func bookNodeID(volAbbrev, slug string) string     { return "book/" + volAbbrev + "/" + slug }
func chapterNodeID(vol, slug string, ch int) string { return "ch/" + vol + "/" + slug + "/" + strconv.Itoa(ch) }
