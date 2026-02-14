package dataloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lds-gpt/internal/libsql/generated"
)

// loadFootnotes implements Phase 3: re-read all chapter JSONs and create
// cross-ref, TG, BD, and JST footnote junction rows.
func (l *Loader) loadFootnotes(
	ctx context.Context,
	verseIndex VerseIndex,
	tgMap map[string]int,
	bdMap map[string]int,
	jstIndex JSTIndex,
) error {
	tx, err := l.ec.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	chapterCount := 0

	for _, volAbbrev := range volumeAbbreviations {
		volDir := filepath.Join(l.dataDir, volAbbrev)
		bookSlugs, err := listBookSlugs(volDir)
		if err != nil {
			return fmt.Errorf("listing books for %s: %w", volAbbrev, err)
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

				err = l.processChapterFootnotes(ctx, tx, volAbbrev, bookSlug, chJSON, verseIndex, tgMap, bdMap, jstIndex)
				if err != nil {
					return fmt.Errorf("processing footnotes for %s/%s/%d: %w",
						volAbbrev, bookSlug, chJSON.Chapter, err)
				}

				chapterCount++
				if chapterCount%100 == 0 {
					l.logger.Info("footnotes progress",
						"chapters_processed", chapterCount,
						"cross_refs", l.stats.CrossRefs,
						"verse_tg_refs", l.stats.VerseTGRefs,
					)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing phase 3: %w", err)
	}
	committed = true

	return nil
}

// processChapterFootnotes processes all footnotes in a single chapter.
func (l *Loader) processChapterFootnotes(
	ctx context.Context,
	tx *generated.Tx,
	volume, slug string,
	ch ChapterJSON,
	verseIndex VerseIndex,
	tgMap map[string]int,
	bdMap map[string]int,
	jstIndex JSTIndex,
) error {
	// Batch builders for each type
	var crossRefBuilders []*generated.VerseCrossRefCreate
	var tgRefBuilders []*generated.VerseTGRefCreate
	var bdRefBuilders []*generated.VerseBDRefCreate
	var jstRefBuilders []*generated.VerseJSTRefCreate

	// Dedup sets to avoid unique constraint violations on (source_id, target_id)
	seenCrossRef := make(map[[2]int]bool)
	seenTGRef := make(map[[2]int]bool)
	seenBDRef := make(map[[2]int]bool)
	seenJSTRef := make(map[[2]int]bool)

	for key, fn := range ch.Footnotes {
		// Parse verse number and marker from key
		verseNum, marker := parseFootnoteKey(key)
		if verseNum == 0 {
			// Title-level or empty key footnote, skip for junction rows
			continue
		}

		// Look up source verse ID
		verseID, ok := verseIndex.Get(volume, slug, ch.Chapter, verseNum)
		if !ok {
			l.stats.Warn(fmt.Sprintf("verse not found for footnote %s in %s/%s/%d",
				key, volume, slug, ch.Chapter))
			continue
		}

		fullMarker := strconv.Itoa(verseNum) + marker
		categories := strings.Split(fn.Category, ",")

		for _, cat := range categories {
			cat = strings.TrimSpace(cat)
			switch cat {
			case "cross-ref":
				builders := l.buildCrossRefRows(fn.Text, verseID, fullMarker, fn.ReferenceText, verseIndex, seenCrossRef)
				crossRefBuilders = append(crossRefBuilders, builders...)

			case "tg":
				builders := l.buildTGRefRows(fn.Text, verseID, fullMarker, fn.ReferenceText, tgMap, seenTGRef)
				tgRefBuilders = append(tgRefBuilders, builders...)

			case "bd":
				builders := l.buildBDRefRows(fn.Text, verseID, fullMarker, fn.ReferenceText, bdMap, seenBDRef)
				bdRefBuilders = append(bdRefBuilders, builders...)

			case "jst":
				builders := l.buildJSTRefRows(fn.Text, verseID, fullMarker, jstIndex, seenJSTRef)
				jstRefBuilders = append(jstRefBuilders, builders...)

			case "trn", "or", "ie":
				// Already handled in Phase 1 as inline JSON fields
			}
		}
	}

	// Bulk insert all junction rows
	batchSize := 500

	for i := 0; i < len(crossRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(crossRefBuilders) {
			end = len(crossRefBuilders)
		}
		if err := tx.VerseCrossRef.CreateBulk(crossRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating cross-refs: %w", err)
		}
	}
	l.stats.CrossRefs += len(crossRefBuilders)

	for i := 0; i < len(tgRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(tgRefBuilders) {
			end = len(tgRefBuilders)
		}
		if err := tx.VerseTGRef.CreateBulk(tgRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating TG refs: %w", err)
		}
	}
	l.stats.VerseTGRefs += len(tgRefBuilders)

	for i := 0; i < len(bdRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(bdRefBuilders) {
			end = len(bdRefBuilders)
		}
		if err := tx.VerseBDRef.CreateBulk(bdRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating BD refs: %w", err)
		}
	}
	l.stats.VerseBDRefs += len(bdRefBuilders)

	for i := 0; i < len(jstRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(jstRefBuilders) {
			end = len(jstRefBuilders)
		}
		if err := tx.VerseJSTRef.CreateBulk(jstRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating JST refs: %w", err)
		}
	}
	l.stats.VerseJSTRefs += len(jstRefBuilders)

	return nil
}

// buildCrossRefRows creates VerseCrossRef builders from cross-ref footnote text.
func (l *Loader) buildCrossRefRows(
	text string,
	verseID int,
	marker string,
	referenceText string,
	verseIndex VerseIndex,
	seen map[[2]int]bool,
) []*generated.VerseCrossRefCreate {
	// Extract only cross-reference portions (before "TG ", "BD ", "JST " prefixes)
	crossRefText := extractCrossRefPortion(text)
	if crossRefText == "" {
		return nil
	}

	result := l.refParser.Parse(crossRefText)
	var builders []*generated.VerseCrossRefCreate

	for _, ref := range result.Refs {
		targetID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
		if !ok {
			l.stats.Warn(fmt.Sprintf("cross-ref target not found: %s/%s %d:%d (from marker %s)",
				ref.Volume, ref.Slug, ref.Chapter, ref.Verse, marker))
			continue
		}

		pair := [2]int{verseID, targetID}
		if seen[pair] {
			continue
		}
		seen[pair] = true

		b := l.ec.VerseCrossRef.Create().
			SetFootnoteMarker(marker).
			SetVerseID(verseID).
			SetCrossRefTargetID(targetID)

		if referenceText != "" {
			b.SetReferenceText(referenceText)
		}

		if ref.EndVerse > 0 {
			endID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.EndVerse)
			if ok {
				b.SetTargetEndVerseID(endID)
			}
		}

		builders = append(builders, b)
	}

	for _, e := range result.Errors {
		l.stats.Warn(fmt.Sprintf("cross-ref parse error (marker %s): %s", marker, e))
	}

	return builders
}

// buildTGRefRows creates VerseTGRef builders from TG footnote text.
func (l *Loader) buildTGRefRows(
	text string,
	verseID int,
	marker string,
	referenceText string,
	tgMap map[string]int,
	seen map[[2]int]bool,
) []*generated.VerseTGRefCreate {
	topics := extractTGTopics(text)
	var builders []*generated.VerseTGRefCreate

	for _, topic := range topics {
		tgID, ok := tgMap[topic]
		if !ok {
			l.stats.Warn(fmt.Sprintf("TG topic not found: %q (marker %s)", topic, marker))
			continue
		}

		pair := [2]int{verseID, tgID}
		if seen[pair] {
			continue
		}
		seen[pair] = true

		b := l.ec.VerseTGRef.Create().
			SetFootnoteMarker(marker).
			SetVerseID(verseID).
			SetTgEntryID(tgID).
			SetTgTopicText("TG " + topic)

		if referenceText != "" {
			b.SetReferenceText(referenceText)
		}

		builders = append(builders, b)
	}

	return builders
}

// buildBDRefRows creates VerseBDRef builders from BD footnote text.
func (l *Loader) buildBDRefRows(
	text string,
	verseID int,
	marker string,
	referenceText string,
	bdMap map[string]int,
	seen map[[2]int]bool,
) []*generated.VerseBDRefCreate {
	entries := extractBDEntries(text)
	var builders []*generated.VerseBDRefCreate

	for _, entry := range entries {
		bdID, ok := bdMap[entry]
		if !ok {
			l.stats.Warn(fmt.Sprintf("BD entry not found: %q (marker %s)", entry, marker))
			continue
		}

		pair := [2]int{verseID, bdID}
		if seen[pair] {
			continue
		}
		seen[pair] = true

		b := l.ec.VerseBDRef.Create().
			SetFootnoteMarker(marker).
			SetVerseID(verseID).
			SetBdEntryID(bdID)

		if referenceText != "" {
			b.SetReferenceText(referenceText)
		}

		builders = append(builders, b)
	}

	return builders
}

// buildJSTRefRows creates VerseJSTRef builders from JST footnote text.
func (l *Loader) buildJSTRefRows(
	text string,
	verseID int,
	marker string,
	jstIndex JSTIndex,
	seen map[[2]int]bool,
) []*generated.VerseJSTRefCreate {
	jstRef := extractJSTReference(text)
	if jstRef == "" {
		return nil
	}

	// Parse the JST reference to get book/chapter
	result := l.refParser.Parse(jstRef)
	if len(result.Refs) == 0 {
		l.stats.Warn(fmt.Sprintf("JST ref parse failed: %q (marker %s)", text, marker))
		return nil
	}

	ref := result.Refs[0]
	entries := jstIndex.Get(ref.Slug, strconv.Itoa(ref.Chapter))

	if len(entries) == 0 {
		l.stats.Warn(fmt.Sprintf("JST passage not found: %s %d (marker %s)", ref.Slug, ref.Chapter, marker))
		return nil
	}

	// Use the first matching entry (there may be multiple entries per chapter)
	var builders []*generated.VerseJSTRefCreate
	for _, entry := range entries {
		pair := [2]int{verseID, entry.dbID}
		if seen[pair] {
			break
		}
		seen[pair] = true

		b := l.ec.VerseJSTRef.Create().
			SetFootnoteMarker(marker).
			SetVerseID(verseID).
			SetJstPassageID(entry.dbID)
		builders = append(builders, b)
		break // one JST ref per footnote
	}

	return builders
}

// parseFootnoteKey extracts verse number and marker letter from a key like "11a".
// Returns (0, "") for empty or title-level keys.
func parseFootnoteKey(key string) (int, string) {
	if key == "" {
		return 0, ""
	}

	// Find where the number ends and the letter begins
	numEnd := 0
	for i, ch := range key {
		if ch >= '0' && ch <= '9' {
			numEnd = i + 1
		} else {
			break
		}
	}

	if numEnd == 0 {
		return 0, ""
	}

	verseNum, err := strconv.Atoi(key[:numEnd])
	if err != nil {
		return 0, ""
	}

	marker := key[numEnd:]
	return verseNum, marker
}

// extractCrossRefPortion extracts just the cross-reference text from a footnote,
// stopping at "TG ", "BD ", "JST ", "HEB ", "GR ", "IE " prefixes.
func extractCrossRefPortion(text string) string {
	// Find first occurrence of known prefix markers
	prefixes := []string{" TG ", " BD ", " JST ", " HEB ", " GR ", " IE "}
	minIdx := len(text)

	for _, p := range prefixes {
		idx := strings.Index(text, p)
		if idx >= 0 && idx < minIdx {
			minIdx = idx
		}
	}

	// Also check for these at the start of text
	startPrefixes := []string{"TG ", "BD ", "JST ", "HEB ", "GR ", "IE "}
	for _, p := range startPrefixes {
		if strings.HasPrefix(text, p) {
			return ""
		}
	}

	result := strings.TrimSpace(text[:minIdx])
	result = strings.TrimRight(result, ".")
	return result
}

// extractTGTopics extracts topic names from TG footnote text.
// Input: "TG Creation; God, Creator." or "D&C 1:1. TG Creation."
// Output: ["Creation", "God, Creator"]
func extractTGTopics(text string) []string {
	// Find "TG " prefix (possibly after other content)
	idx := strings.Index(text, "TG ")
	if idx < 0 {
		return nil
	}

	tgText := text[idx+3:]

	// Remove trailing period and anything after next known prefix
	tgText = trimAtNextPrefix(tgText)
	tgText = strings.TrimRight(tgText, ".")
	tgText = strings.TrimSpace(tgText)

	if tgText == "" {
		return nil
	}

	// Split on semicolons
	parts := strings.Split(tgText, ";")
	var topics []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimRight(p, ".")
		if p != "" {
			topics = append(topics, p)
		}
	}

	return topics
}

// extractBDEntries extracts entry names from BD footnote text.
// Input: "BD Money." or "BD Lost books. See also ..."
// Output: ["Money"] or ["Lost books"]
func extractBDEntries(text string) []string {
	idx := strings.Index(text, "BD ")
	if idx < 0 {
		return nil
	}

	bdText := text[idx+3:]

	// BD entry name ends at first period
	dotIdx := strings.Index(bdText, ".")
	if dotIdx >= 0 {
		bdText = bdText[:dotIdx]
	}

	bdText = strings.TrimSpace(bdText)
	if bdText == "" {
		return nil
	}

	return []string{bdText}
}

// extractJSTReference extracts the scripture reference from JST footnote text.
// Input: "JST Matt. 5:21 (Appendix)." -> "Matt. 5:21"
// Input: "JST Rom. 1:17 ... through faith" -> "Rom. 1:17"
func extractJSTReference(text string) string {
	idx := strings.Index(text, "JST ")
	if idx < 0 {
		return ""
	}

	refText := text[idx+4:]
	refText = strings.TrimSpace(refText)

	// The reference ends at the first space after chapter:verse pattern,
	// or at a parenthesis, or at specific keywords
	// Strategy: parse character by character, looking for end of reference
	endIdx := findRefEnd(refText)
	if endIdx > 0 {
		refText = refText[:endIdx]
	}

	refText = strings.TrimRight(refText, ".")
	refText = strings.TrimSpace(refText)

	return refText
}

// findRefEnd finds where the scripture reference ends in a string.
// Returns the index of the first character after the reference, or 0 to use the full string.
func findRefEnd(s string) int {
	// Look for chapter:verse pattern, then find where digits stop
	colonIdx := strings.Index(s, ":")
	if colonIdx < 0 {
		return 0
	}

	// Normalize dashes first, then scan for digits and hyphens
	normalized := strings.ReplaceAll(s, "–", "-")
	normalized = strings.ReplaceAll(normalized, "—", "-")

	// After the colon, scan for digits and dashes (for ranges like "5:21-34")
	i := colonIdx + 1
	for i < len(normalized) {
		ch := normalized[i]
		if (ch >= '0' && ch <= '9') || ch == '-' {
			i++
			continue
		}
		break
	}

	// If we stopped at a space or special char, that's the end
	if i < len(normalized) {
		return i
	}
	return 0
}

// trimAtNextPrefix trims text at the next occurrence of a known prefix.
func trimAtNextPrefix(text string) string {
	prefixes := []string{" BD ", " JST ", " HEB ", " GR ", " IE "}
	minIdx := len(text)

	for _, p := range prefixes {
		idx := strings.Index(text, p)
		if idx >= 0 && idx < minIdx {
			minIdx = idx
		}
	}

	return text[:minIdx]
}

// readStudyHelpJSON reads and returns the raw bytes of a study help JSON file.
func readStudyHelpJSON(dataDir, filename string) ([]byte, error) {
	path := filepath.Join(dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return data, nil
}
