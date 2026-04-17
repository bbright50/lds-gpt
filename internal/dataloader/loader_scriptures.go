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
// relationships via typed `createXxx` mutations. Each row's map includes a
// `connect` payload for its parent edge (e.g. Book.volume →
// `{connect: [{where: {id: "vol/ot"}}]}`). The fork-patched translator
// emits a runtime `UNWIND coalesce(item.<field>.connect, [])` block per
// relationship so node creation and edge wiring land in one round-trip.
// Auto-chunker handles sub-batching at 500/batch.
//
// Chapter.summaryEmbedding is a required @vector field. Phase 1 writes a
// fractional-zeros placeholder at create time; Phase 6 upgrades via
// typed updateXxx + fork-patched vecf32 auto-wrap. FalkorDB's vector index
// ignores plain-list property values (per spike), so placeholders don't
// pollute kNN.
//
// IDs are deterministic strings supplied by the caller ("v/ot/gen/1/5");
// the fork-patched generator accepts `id` as optional CreateInput and
// `SET n.id = coalesce(item.id, randomUUID())` preserves that default for
// anyone who omits it.

// loadScriptures implements Phase 1.
func (l *Loader) loadScriptures(ctx context.Context) (VerseIndex, error) {
	verseIndex := NewVerseIndex()

	// One walk over the data directory to accumulate row slices for each
	// node type. Four typed bulk mutations at the end.
	volRows := make([]any, 0, len(volumeAbbreviations))
	for _, abbrev := range volumeAbbreviations {
		volRows = append(volRows, map[string]any{
			"id":   volumeNodeID(abbrev),
			"name":         volumeDisplayNames[abbrev],
			"abbreviation": abbrev,
		})
	}

	var (
		bookRows    []any
		chapterRows []any
		verseRows   []any
	)

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

			bookRows = append(bookRows, map[string]any{
				"id": bookNodeID(volAbbrev, bookSlug),
				"name":       bookName,
				"slug":       bookSlug,
				"urlPath":    volAbbrev + "/" + bookSlug,
				"volume":     connectById(volumeNodeID(volAbbrev)),
			})

			for _, chFile := range chapterFiles {
				chPath := filepath.Join(bookDir, chFile)
				chJSON, err := readChapterJSON(chPath)
				if err != nil {
					return verseIndex, fmt.Errorf("reading %s: %w", chPath, err)
				}
				chapterRows = append(chapterRows, buildChapterRow(volAbbrev, bookSlug, chJSON))

				abbrev := l.slugMap[volAbbrev+"/"+bookSlug]
				for _, v := range chJSON.Verses {
					id := VerseNodeID(volAbbrev, bookSlug, chJSON.Chapter, v.Number)
					verseRows = append(verseRows, buildVerseRow(id, abbrev, chJSON, v))
					verseIndex.Put(volAbbrev, bookSlug, chJSON.Chapter, v.Number, id)
				}
			}
		}
	}

	// Typed bulk creates. Each row already carries its parent-edge
	// `connect` payload (see connectById); the fork-patched translator
	// unrolls that per-item at query time so node + edge land in the same
	// round-trip. Auto-chunker handles sub-batching at 500/batch.
	if err := l.createVolumes(ctx, volRows); err != nil {
		return verseIndex, err
	}
	l.stats.Volumes = len(volRows)

	if err := l.createBooks(ctx, bookRows); err != nil {
		return verseIndex, err
	}
	l.stats.Books = len(bookRows)

	if err := l.createChapters(ctx, chapterRows); err != nil {
		return verseIndex, err
	}
	l.stats.Chapters = len(chapterRows)

	if err := l.createVerses(ctx, verseRows); err != nil {
		return verseIndex, err
	}
	l.stats.Verses = len(verseRows)

	l.logger.Info("phase 1 totals",
		"volumes", l.stats.Volumes, "books", l.stats.Books,
		"chapters", l.stats.Chapters, "verses", l.stats.Verses,
	)
	return verseIndex, nil
}

// --- Row builders (no DB I/O) ---

func buildChapterRow(volAbbrev, bookSlug string, ch ChapterJSON) map[string]any {
	row := map[string]any{
		"id":       chapterNodeID(volAbbrev, bookSlug, ch.Chapter),
		"number":           ch.Chapter,
		"summaryEmbedding": placeholderEmbedding(),
		"book":             connectById(bookNodeID(volAbbrev, bookSlug)),
	}
	if ch.Summary != "" {
		row["summary"] = ch.Summary
	}
	if ch.URL != "" {
		row["url"] = ch.URL
	}
	return row
}

func buildVerseRow(id, abbrev string, ch ChapterJSON, v VerseJSON) map[string]any {
	row := map[string]any{
		"id": id,
		"number":     v.Number,
		"text":       v.Text,
		"reference":  fmt.Sprintf("%s %d:%d", abbrev, ch.Chapter, v.Number),
		"chapter":    connectById(chapterNodeID(extractVolumeFromID(id), extractBookSlugFromID(id), ch.Chapter)),
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
	return row
}

// connectById builds a `{ connect: [{ where: { id: <id> } }] }`
// payload for every `XxxYyyFieldInput` relationship field on a CreateInput.
// All of our structural edges at Phase 1 go from child → parent via this
// shape (Book.volume, Chapter.book, Verse.chapter).
func connectById(id string) map[string]any {
	return map[string]any{
		"connect": []any{
			map[string]any{"where": map[string]any{"id": id}},
		},
	}
}

// Verse.id decoders. The row builder has the verse id (e.g.
// "v/ot/gen/1/5") in hand but not the separate (vol, slug) components, so we
// pull them back out of the id string when building the chapter connect.
func extractVolumeFromID(verseID string) string {
	parts := strings.Split(verseID, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
func extractBookSlugFromID(verseID string) string {
	parts := strings.Split(verseID, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// --- Typed bulk mutation helpers ---

func (l *Loader) createVolumes(ctx context.Context, rows []any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := l.fc.GraphQL().Execute(ctx, `
		mutation ($input: [VolumeCreateInput!]!) {
		  createVolumes(input: $input) { volumes { id } }
		}`, map[string]any{"input": rows})
	if err != nil {
		return fmt.Errorf("createVolumes: %w", err)
	}
	return nil
}

func (l *Loader) createBooks(ctx context.Context, rows []any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := l.fc.GraphQL().Execute(ctx, `
		mutation ($input: [BookCreateInput!]!) {
		  createBooks(input: $input) { books { id } }
		}`, map[string]any{"input": rows})
	if err != nil {
		return fmt.Errorf("createBooks: %w", err)
	}
	return nil
}

func (l *Loader) createChapters(ctx context.Context, rows []any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := l.fc.GraphQL().Execute(ctx, `
		mutation ($input: [ChapterCreateInput!]!) {
		  createChapters(input: $input) { chapters { id } }
		}`, map[string]any{"input": rows})
	if err != nil {
		return fmt.Errorf("createChapters: %w", err)
	}
	return nil
}

func (l *Loader) createVerses(ctx context.Context, rows []any) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := l.fc.GraphQL().Execute(ctx, `
		mutation ($input: [VerseCreateInput!]!) {
		  createVerses(input: $input) { verses { id } }
		}`, map[string]any{"input": rows})
	if err != nil {
		return fmt.Errorf("createVerses: %w", err)
	}
	return nil
}

// InlineFootnote types (translation / alternate-reading / explanatory) —
// JSON-serialized onto the Verse node.
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

func volumeNodeID(abbrev string) string             { return "vol/" + abbrev }
func bookNodeID(volAbbrev, slug string) string      { return "book/" + volAbbrev + "/" + slug }
func chapterNodeID(vol, slug string, ch int) string { return "ch/" + vol + "/" + slug + "/" + strconv.Itoa(ch) }
