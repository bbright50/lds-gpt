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

	"lds-gpt/internal/libsql/generated"
	"lds-gpt/internal/libsql/schema"
)

// loadScriptures implements Phase 1: create volumes, books, chapters, and verses.
// Returns a VerseIndex for O(1) lookups and a map of volume abbreviation -> DB ID.
func (l *Loader) loadScriptures(ctx context.Context) (VerseIndex, map[string]int, error) {
	verseIndex := NewVerseIndex()
	volumeIDs := make(map[string]int, 5)

	tx, err := l.ec.Tx(ctx)
	if err != nil {
		return verseIndex, nil, fmt.Errorf("starting transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Create volumes
	for i, volAbbrev := range volumeAbbreviations {
		displayName := volumeDisplayNames[volAbbrev]
		vol, err := tx.Volume.Create().
			SetName(displayName).
			SetAbbreviation(volAbbrev).
			Save(ctx)
		if err != nil {
			return verseIndex, nil, fmt.Errorf("creating volume %s: %w", volAbbrev, err)
		}
		volumeIDs[volAbbrev] = vol.ID
		l.stats.Volumes++
		_ = i
	}

	// Create books, chapters, and verses for each volume
	for _, volAbbrev := range volumeAbbreviations {
		volID := volumeIDs[volAbbrev]
		volDir := filepath.Join(l.dataDir, volAbbrev)

		bookSlugs, err := listBookSlugs(volDir)
		if err != nil {
			return verseIndex, nil, fmt.Errorf("listing books for %s: %w", volAbbrev, err)
		}

		for _, bookSlug := range bookSlugs {
			bookDir := filepath.Join(volDir, bookSlug)
			chapterFiles, err := listChapterFiles(bookDir)
			if err != nil {
				return verseIndex, nil, fmt.Errorf("listing chapters for %s/%s: %w", volAbbrev, bookSlug, err)
			}
			if len(chapterFiles) == 0 {
				continue
			}

			// Read first chapter to get book display name
			firstChapter, err := readChapterJSON(filepath.Join(bookDir, chapterFiles[0]))
			if err != nil {
				return verseIndex, nil, fmt.Errorf("reading first chapter of %s/%s: %w", volAbbrev, bookSlug, err)
			}

			bookName := l.resolveBookName(firstChapter.Book, volAbbrev, bookSlug)
			urlPath := volAbbrev + "/" + bookSlug

			book, err := tx.Book.Create().
				SetName(bookName).
				SetSlug(bookSlug).
				SetURLPath(urlPath).
				SetVolumeID(volID).
				Save(ctx)
			if err != nil {
				return verseIndex, nil, fmt.Errorf("creating book %s/%s: %w", volAbbrev, bookSlug, err)
			}
			l.stats.Books++

			// Load chapters and verses
			for _, chFile := range chapterFiles {
				chPath := filepath.Join(bookDir, chFile)
				chJSON, err := readChapterJSON(chPath)
				if err != nil {
					return verseIndex, nil, fmt.Errorf("reading %s: %w", chPath, err)
				}

				err = l.loadChapter(ctx, tx, book.ID, volAbbrev, bookSlug, chJSON, &verseIndex)
				if err != nil {
					return verseIndex, nil, fmt.Errorf("loading chapter %s/%s/%d: %w",
						volAbbrev, bookSlug, chJSON.Chapter, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return verseIndex, nil, fmt.Errorf("committing phase 1: %w", err)
	}
	committed = true

	return verseIndex, volumeIDs, nil
}

// loadChapter creates a Chapter and its Verses.
func (l *Loader) loadChapter(
	ctx context.Context,
	tx *generated.Tx,
	bookID int,
	volume, slug string,
	ch ChapterJSON,
	verseIndex *VerseIndex,
) error {
	chapterBuilder := tx.Chapter.Create().
		SetNumber(ch.Chapter).
		SetBookID(bookID)

	if ch.Summary != "" {
		chapterBuilder.SetSummary(ch.Summary)
	}
	if ch.URL != "" {
		chapterBuilder.SetURL(ch.URL)
	}

	chapter, err := chapterBuilder.Save(ctx)
	if err != nil {
		return fmt.Errorf("creating chapter %d: %w", ch.Chapter, err)
	}
	l.stats.Chapters++

	if l.stats.Chapters%100 == 0 {
		l.logger.Info("progress", "chapters", l.stats.Chapters, "verses", l.stats.Verses)
	}

	// Create verses (bulk create for performance)
	builders := make([]*generated.VerseCreate, 0, len(ch.Verses))
	for _, v := range ch.Verses {
		abbrev := l.slugMap[volume+"/"+slug]
		reference := fmt.Sprintf("%s %d:%d", abbrev, ch.Chapter, v.Number)

		vb := tx.Verse.Create().
			SetNumber(v.Number).
			SetText(v.Text).
			SetReference(reference).
			SetChapterID(chapter.ID)

		// Extract inline footnotes (trn, or, ie)
		trnNotes, orNotes, ieNotes := extractInlineFootnotes(ch.Footnotes, v.Number)
		if len(trnNotes) > 0 {
			vb.SetTranslationNotes(trnNotes)
		}
		if len(orNotes) > 0 {
			vb.SetAlternateReadings(orNotes)
		}
		if len(ieNotes) > 0 {
			vb.SetExplanatoryNotes(ieNotes)
		}

		builders = append(builders, vb)
	}

	verses, err := tx.Verse.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return fmt.Errorf("bulk creating verses for chapter %d: %w", ch.Chapter, err)
	}

	// Index the verses
	for i, v := range verses {
		verseNum := ch.Verses[i].Number
		verseIndex.Put(volume, slug, ch.Chapter, verseNum, v.ID)
	}
	l.stats.Verses += len(verses)

	return nil
}

// resolveBookName determines the display name for a book.
func (l *Loader) resolveBookName(jsonBookName, volume, slug string) string {
	// Use the JSON book field if it's non-empty and meaningful
	if jsonBookName != "" {
		return jsonBookName
	}

	// Try the display name map
	key := volume + "/" + slug
	abbrev, ok := l.slugMap[key]
	if ok {
		return abbrev
	}

	// Fallback to slug
	return slug
}

// extractInlineFootnotes extracts trn, or, and ie footnotes for a specific verse.
func extractInlineFootnotes(
	footnotes map[string]FootnoteJSON,
	verseNum int,
) ([]schema.TranslationNote, []schema.AlternateReading, []schema.ExplanatoryNote) {
	var trn []schema.TranslationNote
	var or []schema.AlternateReading
	var ie []schema.ExplanatoryNote

	prefix := strconv.Itoa(verseNum)

	for key, fn := range footnotes {
		// Key must start with the verse number
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		// Extract marker (the letter after the verse number)
		marker := strings.TrimPrefix(key, prefix)
		if marker == "" {
			continue
		}

		categories := strings.Split(fn.Category, ",")
		for _, cat := range categories {
			switch strings.TrimSpace(cat) {
			case "trn":
				trn = append(trn, schema.TranslationNote{
					Marker:     marker,
					HebrewText: fn.Text,
				})
			case "or":
				or = append(or, schema.AlternateReading{
					Marker: marker,
					Text:   fn.Text,
				})
			case "ie":
				ie = append(ie, schema.ExplanatoryNote{
					Marker: marker,
					Text:   fn.Text,
				})
			}
		}
	}

	return trn, or, ie
}

// listBookSlugs lists subdirectories (book slugs) in a volume directory, sorted.
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

// listChapterFiles lists JSON chapter files in a book directory, sorted by chapter number.
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

	// Sort by chapter number
	sort.Slice(files, func(i, j int) bool {
		ni := chapterNumFromFilename(files[i])
		nj := chapterNumFromFilename(files[j])
		return ni < nj
	})

	return files, nil
}

// chapterNumFromFilename extracts the chapter number from "1.json" -> 1.
func chapterNumFromFilename(name string) int {
	numStr := strings.TrimSuffix(name, ".json")
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return n
}

// readChapterJSON reads and parses a chapter JSON file.
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
