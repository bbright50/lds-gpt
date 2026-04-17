package dataloader

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Phase 3 — per-verse footnote edges. Re-reads every chapter JSON, parses
// the footnote text, and creates four kinds of relationship:
//
//   * (Verse)-[:CROSS_REF   {VerseCrossRefProps}]->(Verse)
//   * (Verse)-[:TG_FOOTNOTE {VerseTGRefProps}]->(TopicalGuideEntry)
//   * (Verse)-[:BD_FOOTNOTE {VerseBDRefProps}]->(BibleDictEntry)
//   * (Verse)-[:JST_FOOTNOTE{VerseJSTRefProps}]->(JSTPassage)
//
// Text parsing (extractCrossRefPortion, extractTGTopics, extractBDEntries,
// extractJSTReference, parseFootnoteKey) is unchanged from the LibSQL
// implementation — those are pure functions of footnote text.
//
// Writes go through raw Cypher because the generated go-ormql client would
// require full node payloads (including the @vector embedding) and we only
// need to connect existing nodes here.

const footnoteBatchSize = 500

type crossRefRow struct {
	srcID, tgtID, marker, refText string
	endID                         string // optional; empty means "no end verse"
}
type tgRefRow struct {
	srcID, tgtID, marker string
}
type bdRefRow struct {
	srcID, tgtID, marker string
}
type jstRefRow struct {
	srcID, tgtID, marker string
}

func (l *Loader) loadFootnotes(
	ctx context.Context,
	verseIndex VerseIndex,
	tgMap, bdMap map[string]string,
	jstIndex JSTIndex,
) error {
	var (
		crossRefs []crossRefRow
		tgRefs    []tgRefRow
		bdRefs    []bdRefRow
		jstRefs   []jstRefRow

		// Pair-dedupe — the source JSON sometimes repeats the same target
		// across multiple footnote markers on the same verse; we only keep
		// one edge per (source, target) per relationship type.
		seenCross = map[[2]string]bool{}
		seenTG    = map[[2]string]bool{}
		seenBD    = map[[2]string]bool{}
		seenJST   = map[[2]string]bool{}
	)

	chapterCount := 0
	for _, volAbbrev := range volumeAbbreviations {
		volDir := filepath.Join(l.dataDir, volAbbrev)
		bookSlugs, err := listBookSlugs(volDir)
		if err != nil {
			// Missing volume dir — Phase 1 already warned; skip here.
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
				l.collectChapterFootnotes(
					volAbbrev, bookSlug, chJSON,
					verseIndex, tgMap, bdMap, jstIndex,
					&crossRefs, &tgRefs, &bdRefs, &jstRefs,
					seenCross, seenTG, seenBD, seenJST,
				)
				chapterCount++
				if chapterCount%100 == 0 {
					l.logger.Info("footnotes progress",
						"chapters_processed", chapterCount,
						"cross_refs", len(crossRefs),
						"tg_refs", len(tgRefs),
					)
				}
			}
		}
	}

	if err := l.writeCrossRefs(ctx, crossRefs); err != nil {
		return err
	}
	if err := l.writeTGFootnotes(ctx, tgRefs); err != nil {
		return err
	}
	if err := l.writeBDFootnotes(ctx, bdRefs); err != nil {
		return err
	}
	if err := l.writeJSTFootnotes(ctx, jstRefs); err != nil {
		return err
	}

	l.stats.CrossRefs = len(crossRefs)
	l.stats.VerseTGRefs = len(tgRefs)
	l.stats.VerseBDRefs = len(bdRefs)
	l.stats.VerseJSTRefs = len(jstRefs)
	l.logger.Info("phase 3 totals",
		"cross_refs", l.stats.CrossRefs,
		"verse_tg_refs", l.stats.VerseTGRefs,
		"verse_bd_refs", l.stats.VerseBDRefs,
		"verse_jst_refs", l.stats.VerseJSTRefs,
	)
	return nil
}

// collectChapterFootnotes walks every footnote on one chapter and appends
// rows to the per-type slices. Pure collection — no DB I/O.
func (l *Loader) collectChapterFootnotes(
	volume, slug string,
	ch ChapterJSON,
	verseIndex VerseIndex,
	tgMap, bdMap map[string]string,
	jstIndex JSTIndex,
	crossRefs *[]crossRefRow, tgRefs *[]tgRefRow, bdRefs *[]bdRefRow, jstRefs *[]jstRefRow,
	seenCross, seenTG, seenBD, seenJST map[[2]string]bool,
) {
	for key, fn := range ch.Footnotes {
		verseNum, marker := parseFootnoteKey(key)
		if verseNum == 0 {
			continue
		}
		verseID, ok := verseIndex.Get(volume, slug, ch.Chapter, verseNum)
		if !ok {
			l.stats.Warn(fmt.Sprintf("verse not found for footnote %s in %s/%s/%d", key, volume, slug, ch.Chapter))
			continue
		}
		fullMarker := strconv.Itoa(verseNum) + marker

		for _, cat := range strings.Split(fn.Category, ",") {
			switch strings.TrimSpace(cat) {
			case "cross-ref":
				for _, r := range l.parseCrossRefRows(fn.Text, verseID, fullMarker, fn.ReferenceText, verseIndex, seenCross) {
					*crossRefs = append(*crossRefs, r)
				}
			case "tg":
				for _, r := range l.parseTGRefRows(fn.Text, verseID, fullMarker, tgMap, seenTG) {
					*tgRefs = append(*tgRefs, r)
				}
			case "bd":
				for _, r := range l.parseBDRefRows(fn.Text, verseID, fullMarker, bdMap, seenBD) {
					*bdRefs = append(*bdRefs, r)
				}
			case "jst":
				for _, r := range l.parseJSTRefRows(fn.Text, verseID, fullMarker, jstIndex, seenJST) {
					*jstRefs = append(*jstRefs, r)
				}
			case "trn", "or", "ie":
				// Handled in Phase 1 as inline JSON properties on the Verse node.
			}
		}
	}
}

func (l *Loader) parseCrossRefRows(
	text, verseID, marker, referenceText string,
	verseIndex VerseIndex,
	seen map[[2]string]bool,
) []crossRefRow {
	crossRefText := extractCrossRefPortion(text)
	if crossRefText == "" {
		return nil
	}
	result := l.refParser.Parse(crossRefText)
	var rows []crossRefRow
	for _, ref := range result.Refs {
		targetID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
		if !ok {
			l.stats.Warn(fmt.Sprintf("cross-ref target not found: %s/%s %d:%d (marker %s)",
				ref.Volume, ref.Slug, ref.Chapter, ref.Verse, marker))
			continue
		}
		pair := [2]string{verseID, targetID}
		if seen[pair] {
			continue
		}
		seen[pair] = true

		row := crossRefRow{srcID: verseID, tgtID: targetID, marker: marker, refText: referenceText}
		if ref.EndVerse > 0 {
			if endID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.EndVerse); ok {
				row.endID = endID
			}
		}
		rows = append(rows, row)
	}
	for _, e := range result.Errors {
		l.stats.Warn(fmt.Sprintf("cross-ref parse error (marker %s): %s", marker, e))
	}
	return rows
}

func (l *Loader) parseTGRefRows(
	text, verseID, marker string,
	tgMap map[string]string,
	seen map[[2]string]bool,
) []tgRefRow {
	var rows []tgRefRow
	for _, topic := range extractTGTopics(text) {
		tgID, ok := tgMap[topic]
		if !ok {
			l.stats.Warn(fmt.Sprintf("TG topic not found: %q (marker %s)", topic, marker))
			continue
		}
		pair := [2]string{verseID, tgID}
		if seen[pair] {
			continue
		}
		seen[pair] = true
		rows = append(rows, tgRefRow{srcID: verseID, tgtID: tgID, marker: marker})
	}
	return rows
}

func (l *Loader) parseBDRefRows(
	text, verseID, marker string,
	bdMap map[string]string,
	seen map[[2]string]bool,
) []bdRefRow {
	var rows []bdRefRow
	for _, entry := range extractBDEntries(text) {
		bdID, ok := bdMap[entry]
		if !ok {
			l.stats.Warn(fmt.Sprintf("BD entry not found: %q (marker %s)", entry, marker))
			continue
		}
		pair := [2]string{verseID, bdID}
		if seen[pair] {
			continue
		}
		seen[pair] = true
		rows = append(rows, bdRefRow{srcID: verseID, tgtID: bdID, marker: marker})
	}
	return rows
}

func (l *Loader) parseJSTRefRows(
	text, verseID, marker string,
	jstIndex JSTIndex,
	seen map[[2]string]bool,
) []jstRefRow {
	jstRef := extractJSTReference(text)
	if jstRef == "" {
		return nil
	}
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
	var rows []jstRefRow
	for _, entry := range entries {
		pair := [2]string{verseID, entry.nodeID}
		if seen[pair] {
			break
		}
		seen[pair] = true
		rows = append(rows, jstRefRow{srcID: verseID, tgtID: entry.nodeID, marker: marker})
		break // first match wins — matches the LibSQL behavior
	}
	return rows
}

// --- Write helpers (batched UNWIND) ---

func (l *Loader) writeCrossRefs(ctx context.Context, rows []crossRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += footnoteBatchSize {
		end := i + footnoteBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{
				"src": r.srcID, "tgt": r.tgtID, "marker": r.marker, "refText": r.refText,
			})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (s:Verse {id: r.src})
			 MATCH (t:Verse {id: r.tgt})
			 CREATE (s)-[:CROSS_REF {
			   category: 'cross-ref',
			   footnoteMarker: r.marker,
			   referenceText: r.refText
			 }]->(t)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating cross-refs: %w", err)
		}
	}
	return nil
}

func (l *Loader) writeTGFootnotes(ctx context.Context, rows []tgRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += footnoteBatchSize {
		end := i + footnoteBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{
				"src": r.srcID, "tgt": r.tgtID, "marker": r.marker,
			})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (v:Verse {id: r.src})
			 MATCH (t:TopicalGuideEntry {id: r.tgt})
			 CREATE (v)-[:TG_FOOTNOTE {footnoteMarker: r.marker}]->(t)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating TG footnotes: %w", err)
		}
	}
	return nil
}

func (l *Loader) writeBDFootnotes(ctx context.Context, rows []bdRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += footnoteBatchSize {
		end := i + footnoteBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{
				"src": r.srcID, "tgt": r.tgtID, "marker": r.marker,
			})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (v:Verse {id: r.src})
			 MATCH (b:BibleDictEntry {id: r.tgt})
			 CREATE (v)-[:BD_FOOTNOTE {footnoteMarker: r.marker}]->(b)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating BD footnotes: %w", err)
		}
	}
	return nil
}

func (l *Loader) writeJSTFootnotes(ctx context.Context, rows []jstRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += footnoteBatchSize {
		end := i + footnoteBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{
				"src": r.srcID, "tgt": r.tgtID, "marker": r.marker,
			})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (v:Verse {id: r.src})
			 MATCH (j:JSTPassage {id: r.tgt})
			 CREATE (v)-[:JST_FOOTNOTE {footnoteMarker: r.marker}]->(j)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating JST footnotes: %w", err)
		}
	}
	return nil
}

// --- Footnote text parsing (pure functions, unchanged from the LibSQL
//     implementation — they operate on strings only) ---

func parseFootnoteKey(key string) (int, string) {
	if key == "" {
		return 0, ""
	}
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
	return verseNum, key[numEnd:]
}

func extractCrossRefPortion(text string) string {
	startPrefixes := []string{"TG ", "BD ", "JST ", "HEB ", "GR ", "IE "}
	for _, p := range startPrefixes {
		if strings.HasPrefix(text, p) {
			return ""
		}
	}
	prefixes := []string{" TG ", " BD ", " JST ", " HEB ", " GR ", " IE "}
	minIdx := len(text)
	for _, p := range prefixes {
		if idx := strings.Index(text, p); idx >= 0 && idx < minIdx {
			minIdx = idx
		}
	}
	result := strings.TrimSpace(text[:minIdx])
	return strings.TrimRight(result, ".")
}

func extractTGTopics(text string) []string {
	idx := strings.Index(text, "TG ")
	if idx < 0 {
		return nil
	}
	tgText := trimAtNextPrefix(text[idx+3:])
	tgText = strings.TrimRight(tgText, ".")
	tgText = strings.TrimSpace(tgText)
	if tgText == "" {
		return nil
	}
	var topics []string
	for _, p := range strings.Split(tgText, ";") {
		p = strings.TrimSpace(p)
		p = strings.TrimRight(p, ".")
		if p != "" {
			topics = append(topics, p)
		}
	}
	return topics
}

func extractBDEntries(text string) []string {
	idx := strings.Index(text, "BD ")
	if idx < 0 {
		return nil
	}
	bdText := text[idx+3:]
	if dotIdx := strings.Index(bdText, "."); dotIdx >= 0 {
		bdText = bdText[:dotIdx]
	}
	bdText = strings.TrimSpace(bdText)
	if bdText == "" {
		return nil
	}
	return []string{bdText}
}

func extractJSTReference(text string) string {
	idx := strings.Index(text, "JST ")
	if idx < 0 {
		return ""
	}
	refText := strings.TrimSpace(text[idx+4:])
	if endIdx := findRefEnd(refText); endIdx > 0 {
		refText = refText[:endIdx]
	}
	return strings.TrimSpace(strings.TrimRight(refText, "."))
}

func findRefEnd(s string) int {
	colonIdx := strings.Index(s, ":")
	if colonIdx < 0 {
		return 0
	}
	normalized := strings.ReplaceAll(s, "–", "-")
	normalized = strings.ReplaceAll(normalized, "—", "-")
	i := colonIdx + 1
	for i < len(normalized) {
		ch := normalized[i]
		if (ch >= '0' && ch <= '9') || ch == '-' {
			i++
			continue
		}
		break
	}
	if i < len(normalized) {
		return i
	}
	return 0
}

func trimAtNextPrefix(text string) string {
	prefixes := []string{" BD ", " JST ", " HEB ", " GR ", " IE "}
	minIdx := len(text)
	for _, p := range prefixes {
		if idx := strings.Index(text, p); idx >= 0 && idx < minIdx {
			minIdx = idx
		}
	}
	return text[:minIdx]
}
