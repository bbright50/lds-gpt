package dataloader

import (
	"context"
	"fmt"
	"strconv"
)

// Phase 4 — edges from study-help entries to verses and to each other.
//
// Nine relationship kinds:
//   * TG_VERSE_REF  {phrase}            TG  → Verse
//   * TG_SEE_ALSO                       TG  → TG   (self)
//   * TG_BD_REF                         TG  → BD
//   * BD_VERSE_REF                      BD  → Verse
//   * IDX_VERSE_REF {phrase}            IDX → Verse
//   * IDX_SEE_ALSO                      IDX → IDX  (self)
//   * IDX_TG_REF                        IDX → TG
//   * IDX_BD_REF                        IDX → BD
//   * COMPARES                          JST → Verse
//
// Reads topical-guide.json / bible-dictionary.json / triple-combination-index.json
// / jst.json a second time (after Phase 2 created the nodes) and connects
// them into a graph. Writes go through raw Cypher because these are pure
// relationship creates between pre-existing nodes.

const studyRefsBatchSize = 500

type tgVerseRefRow struct {
	tgID, verseID, phrase string
}
type bdVerseRefRow struct {
	bdID, verseID string
}
type idxVerseRefRow struct {
	idxID, verseID, phrase string
}
type simplePairRow struct {
	srcID, tgtID string
}

func (l *Loader) loadStudyRefs(
	ctx context.Context,
	verseIndex VerseIndex,
	tgMap, bdMap, idxMap map[string]string,
	jstIndex JSTIndex,
) error {
	if err := l.loadTGRefs(ctx, verseIndex, tgMap, bdMap); err != nil {
		return fmt.Errorf("loading TG refs: %w", err)
	}
	if err := l.loadBDRefs(ctx, verseIndex, bdMap); err != nil {
		return fmt.Errorf("loading BD refs: %w", err)
	}
	if err := l.loadIDXRefs(ctx, verseIndex, idxMap, tgMap, bdMap); err != nil {
		return fmt.Errorf("loading IDX refs: %w", err)
	}
	if err := l.loadJSTCompares(ctx, verseIndex, jstIndex); err != nil {
		return fmt.Errorf("loading JST compares: %w", err)
	}
	return nil
}

func (l *Loader) loadTGRefs(
	ctx context.Context,
	verseIndex VerseIndex,
	tgMap, bdMap map[string]string,
) error {
	tgData := map[string][]TGEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "topical-guide.json", &tgData); err != nil || !ok {
		return err
	}

	var (
		verseRows []tgVerseRefRow
		seeAlso   []simplePairRow
		bdRefs    []simplePairRow
		seenV     = map[[2]string]bool{}
	)

	for topicName, entries := range tgData {
		tgID, ok := tgMap[topicName]
		if !ok {
			continue
		}
		for _, entry := range entries {
			switch entry.Reference {
			case "TG":
				targetID, ok := tgMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("TG see-also target not found: %q (from %q)", entry.Key, topicName))
					continue
				}
				seeAlso = append(seeAlso, simplePairRow{srcID: tgID, tgtID: targetID})
			case "BD":
				bdID, ok := bdMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("TG->BD target not found: %q (from %q)", entry.Key, topicName))
					continue
				}
				bdRefs = append(bdRefs, simplePairRow{srcID: tgID, tgtID: bdID})
			default:
				verseRows = append(verseRows, l.collectTGVerseRefs(entry, tgID, verseIndex, seenV)...)
			}
		}
	}

	if err := l.writeTGVerseRefs(ctx, verseRows); err != nil {
		return err
	}
	if err := l.writeSimplePairs(ctx, seeAlso, "TopicalGuideEntry", "TopicalGuideEntry", "TG_SEE_ALSO"); err != nil {
		return err
	}
	if err := l.writeSimplePairs(ctx, bdRefs, "TopicalGuideEntry", "BibleDictEntry", "TG_BD_REF"); err != nil {
		return err
	}

	l.stats.TGVerseRefs += len(verseRows)
	l.stats.TGSeeAlso += len(seeAlso)
	l.stats.TGBDRefs += len(bdRefs)
	return nil
}

func (l *Loader) collectTGVerseRefs(entry TGEntryJSON, tgID string, verseIndex VerseIndex, seen map[[2]string]bool) []tgVerseRefRow {
	result := l.refParser.Parse(entry.Reference)
	var rows []tgVerseRefRow
	for _, ref := range result.Refs {
		verseID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
		if !ok {
			l.stats.Warn(fmt.Sprintf("TG verse ref not found: %s/%s %d:%d", ref.Volume, ref.Slug, ref.Chapter, ref.Verse))
			continue
		}
		pair := [2]string{tgID, verseID}
		if seen[pair] {
			continue
		}
		seen[pair] = true
		rows = append(rows, tgVerseRefRow{tgID: tgID, verseID: verseID, phrase: entry.Phrase})
	}
	for _, e := range result.Errors {
		l.stats.Warn(fmt.Sprintf("TG verse ref parse error: %s", e))
	}
	return rows
}

func (l *Loader) loadBDRefs(
	ctx context.Context,
	verseIndex VerseIndex,
	bdMap map[string]string,
) error {
	bdData := map[string]BDEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "bible-dictionary.json", &bdData); err != nil || !ok {
		return err
	}

	var rows []bdVerseRefRow
	seen := map[[2]string]bool{}
	for entryName, entry := range bdData {
		bdID, ok := bdMap[entryName]
		if !ok {
			continue
		}
		for _, refStr := range entry.References {
			result := l.refParser.Parse(refStr)
			for _, ref := range result.Refs {
				verseID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
				if !ok {
					l.stats.Warn(fmt.Sprintf("BD verse ref not found: %s/%s %d:%d (from %q)",
						ref.Volume, ref.Slug, ref.Chapter, ref.Verse, entryName))
					continue
				}
				pair := [2]string{bdID, verseID}
				if seen[pair] {
					continue
				}
				seen[pair] = true
				rows = append(rows, bdVerseRefRow{bdID: bdID, verseID: verseID})
			}
			for _, e := range result.Errors {
				l.stats.Warn(fmt.Sprintf("BD verse ref parse error (from %q): %s", entryName, e))
			}
		}
	}

	if err := l.writeBDVerseRefs(ctx, rows); err != nil {
		return err
	}
	l.stats.BDVerseRefs += len(rows)
	return nil
}

func (l *Loader) loadIDXRefs(
	ctx context.Context,
	verseIndex VerseIndex,
	idxMap, tgMap, bdMap map[string]string,
) error {
	idxData := map[string][]IDXEntryJSON{}
	if ok, err := l.readOptionalJSON(ctx, "triple-combination-index.json", &idxData); err != nil || !ok {
		return err
	}

	var (
		verseRows []idxVerseRefRow
		seeAlso   []simplePairRow
		tgRefs    []simplePairRow
		bdRefs    []simplePairRow
		seenV     = map[[2]string]bool{}
	)

	for entryName, entries := range idxData {
		idxID, ok := idxMap[entryName]
		if !ok {
			continue
		}
		for _, entry := range entries {
			switch entry.Reference {
			case "IDX":
				if targetID, ok := idxMap[entry.Key]; ok {
					seeAlso = append(seeAlso, simplePairRow{srcID: idxID, tgtID: targetID})
				} else {
					l.stats.Warn(fmt.Sprintf("IDX see-also target not found: %q (from %q)", entry.Key, entryName))
				}
			case "TG":
				if tgID, ok := tgMap[entry.Key]; ok {
					tgRefs = append(tgRefs, simplePairRow{srcID: idxID, tgtID: tgID})
				} else {
					l.stats.Warn(fmt.Sprintf("IDX->TG target not found: %q (from %q)", entry.Key, entryName))
				}
			case "BD":
				if bdID, ok := bdMap[entry.Key]; ok {
					bdRefs = append(bdRefs, simplePairRow{srcID: idxID, tgtID: bdID})
				} else {
					l.stats.Warn(fmt.Sprintf("IDX->BD target not found: %q (from %q)", entry.Key, entryName))
				}
			default:
				result := l.refParser.Parse(entry.Reference)
				for _, ref := range result.Refs {
					verseID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
					if !ok {
						l.stats.Warn(fmt.Sprintf("IDX verse ref not found: %s/%s %d:%d (from %q)",
							ref.Volume, ref.Slug, ref.Chapter, ref.Verse, entryName))
						continue
					}
					pair := [2]string{idxID, verseID}
					if seenV[pair] {
						continue
					}
					seenV[pair] = true
					verseRows = append(verseRows, idxVerseRefRow{idxID: idxID, verseID: verseID, phrase: entry.Phrase})
				}
				for _, e := range result.Errors {
					l.stats.Warn(fmt.Sprintf("IDX verse ref parse error (from %q): %s", entryName, e))
				}
			}
		}
	}

	if err := l.writeIDXVerseRefs(ctx, verseRows); err != nil {
		return err
	}
	if err := l.writeSimplePairs(ctx, seeAlso, "IndexEntry", "IndexEntry", "IDX_SEE_ALSO"); err != nil {
		return err
	}
	if err := l.writeSimplePairs(ctx, tgRefs, "IndexEntry", "TopicalGuideEntry", "IDX_TG_REF"); err != nil {
		return err
	}
	if err := l.writeSimplePairs(ctx, bdRefs, "IndexEntry", "BibleDictEntry", "IDX_BD_REF"); err != nil {
		return err
	}

	l.stats.IDXVerseRefs += len(verseRows)
	l.stats.IDXSeeAlso += len(seeAlso)
	l.stats.IDXTGRefs += len(tgRefs)
	l.stats.IDXBDRefs += len(bdRefs)
	return nil
}

func (l *Loader) loadJSTCompares(
	ctx context.Context,
	verseIndex VerseIndex,
	jstIndex JSTIndex,
) error {
	var jstData []JSTChapterJSON
	if ok, err := l.readOptionalJSON(ctx, "jst.json", &jstData); err != nil || !ok {
		return err
	}

	var pairs []simplePairRow
	for _, ch := range jstData {
		bookSlug := l.jstBookToSlug(ch.Book)
		if bookSlug == "" {
			continue
		}
		for _, entry := range ch.Entries {
			if entry.Compare == "" {
				continue
			}
			var passageID string
			for _, je := range jstIndex.Get(bookSlug, ch.Chapter) {
				if je.comprises == entry.Comprises {
					passageID = je.nodeID
					break
				}
			}
			if passageID == "" {
				l.stats.Warn(fmt.Sprintf("JST passage not found for compare: %s %s:%s", ch.Book, ch.Chapter, entry.Comprises))
				continue
			}

			result := l.refParser.Parse(entry.Compare)
			for _, ref := range result.Refs {
				if vID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse); ok {
					pairs = append(pairs, simplePairRow{srcID: passageID, tgtID: vID})
				} else {
					l.stats.Warn(fmt.Sprintf("JST compare verse not found: %s/%s %d:%d",
						ref.Volume, ref.Slug, ref.Chapter, ref.Verse))
				}
				if ref.EndVerse > 0 {
					for v := ref.Verse + 1; v <= ref.EndVerse; v++ {
						if id, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, v); ok {
							pairs = append(pairs, simplePairRow{srcID: passageID, tgtID: id})
						}
					}
				}
			}
			for _, e := range result.Errors {
				l.stats.Warn(fmt.Sprintf("JST compare parse error (%s %s): %s", ch.Book, ch.Chapter, e))
			}
		}
	}

	if err := l.writeSimplePairs(ctx, pairs, "JSTPassage", "Verse", "COMPARES"); err != nil {
		return err
	}
	l.stats.JSTCompares += len(pairs)
	// Touch strconv so its existing use for verse numbers stays linked — keeps
	// the import list minimal across the package.
	_ = strconv.Atoi
	return nil
}

// --- Write helpers (batched UNWIND) ---

func (l *Loader) writeTGVerseRefs(ctx context.Context, rows []tgVerseRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += studyRefsBatchSize {
		end := i + studyRefsBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{"tg": r.tgID, "v": r.verseID, "phrase": r.phrase})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (t:TopicalGuideEntry {id: r.tg})
			 MATCH (v:Verse {id: r.v})
			 CREATE (t)-[:TG_VERSE_REF {phrase: r.phrase}]->(v)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating TG verse refs: %w", err)
		}
	}
	return nil
}

func (l *Loader) writeBDVerseRefs(ctx context.Context, rows []bdVerseRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += studyRefsBatchSize {
		end := i + studyRefsBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{"bd": r.bdID, "v": r.verseID})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (b:BibleDictEntry {id: r.bd})
			 MATCH (v:Verse {id: r.v})
			 CREATE (b)-[:BD_VERSE_REF]->(v)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating BD verse refs: %w", err)
		}
	}
	return nil
}

func (l *Loader) writeIDXVerseRefs(ctx context.Context, rows []idxVerseRefRow) error {
	_ = ctx
	for i := 0; i < len(rows); i += studyRefsBatchSize {
		end := i + studyRefsBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{"idx": r.idxID, "v": r.verseID, "phrase": r.phrase})
		}
		if _, err := l.fc.Raw().Query(
			`UNWIND $rows AS r
			 MATCH (i:IndexEntry {id: r.idx})
			 MATCH (v:Verse {id: r.v})
			 CREATE (i)-[:IDX_VERSE_REF {phrase: r.phrase}]->(v)`,
			map[string]interface{}{"rows": batch}, nil,
		); err != nil {
			return fmt.Errorf("creating IDX verse refs: %w", err)
		}
	}
	return nil
}

// writeSimplePairs creates a property-less relationship `relType` from every
// (srcLabel, src.id) to (tgtLabel, tgt.id) in rows. Used for SEE_ALSO, BD_REF,
// TG_REF, COMPARES — all edges that carry no metadata.
func (l *Loader) writeSimplePairs(ctx context.Context, rows []simplePairRow, srcLabel, tgtLabel, relType string) error {
	_ = ctx
	if len(rows) == 0 {
		return nil
	}
	query := fmt.Sprintf(
		`UNWIND $rows AS r
		 MATCH (a:%s {id: r.src})
		 MATCH (b:%s {id: r.tgt})
		 CREATE (a)-[:%s]->(b)`,
		srcLabel, tgtLabel, relType,
	)
	for i := 0; i < len(rows); i += studyRefsBatchSize {
		end := i + studyRefsBatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := make([]interface{}, 0, end-i)
		for _, r := range rows[i:end] {
			batch = append(batch, map[string]interface{}{"src": r.srcID, "tgt": r.tgtID})
		}
		if _, err := l.fc.Raw().Query(query, map[string]interface{}{"rows": batch}, nil); err != nil {
			return fmt.Errorf("creating %s edges: %w", relType, err)
		}
	}
	return nil
}
