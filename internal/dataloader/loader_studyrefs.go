package dataloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"lds-gpt/internal/libsql/generated"
)

// loadStudyRefs implements Phase 4: create edges from study help entities
// (TG, BD, IDX, JST) to verses and to each other.
func (l *Loader) loadStudyRefs(
	ctx context.Context,
	verseIndex VerseIndex,
	tgMap map[string]int,
	bdMap map[string]int,
	idxMap map[string]int,
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

	if err := l.loadTGRefs(ctx, tx, verseIndex, tgMap, bdMap); err != nil {
		return fmt.Errorf("loading TG refs: %w", err)
	}

	if err := l.loadBDRefs(ctx, tx, verseIndex, bdMap); err != nil {
		return fmt.Errorf("loading BD refs: %w", err)
	}

	if err := l.loadIDXRefs(ctx, tx, verseIndex, idxMap, tgMap, bdMap); err != nil {
		return fmt.Errorf("loading IDX refs: %w", err)
	}

	if err := l.loadJSTCompares(ctx, tx, verseIndex, jstIndex); err != nil {
		return fmt.Errorf("loading JST compares: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing phase 4: %w", err)
	}
	committed = true

	return nil
}

// loadTGRefs creates TG -> verse refs, TG -> TG see-also, and TG -> BD edges.
func (l *Loader) loadTGRefs(
	ctx context.Context,
	tx *generated.Tx,
	verseIndex VerseIndex,
	tgMap map[string]int,
	bdMap map[string]int,
) error {
	data, err := os.ReadFile(filepath.Join(l.dataDir, "topical-guide.json"))
	if err != nil {
		return fmt.Errorf("reading topical guide: %w", err)
	}

	var tgData map[string][]TGEntryJSON
	if err := json.Unmarshal(data, &tgData); err != nil {
		return fmt.Errorf("parsing topical guide: %w", err)
	}

	var verseRefBuilders []*generated.TGVerseRefCreate
	batchSize := 500
	seenTGVerse := make(map[[2]int]bool)

	for topicName, entries := range tgData {
		tgID, ok := tgMap[topicName]
		if !ok {
			continue
		}

		for _, entry := range entries {
			switch entry.Reference {
			case "TG":
				// See-also: TG -> TG
				targetID, ok := tgMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("TG see-also target not found: %q (from %q)", entry.Key, topicName))
					continue
				}
				if err := tx.TopicalGuideEntry.UpdateOneID(tgID).AddSeeAlsoIDs(targetID).Exec(ctx); err != nil {
					return fmt.Errorf("adding TG see-also %q -> %q: %w", topicName, entry.Key, err)
				}
				l.stats.TGSeeAlso++

			case "BD":
				// TG -> BD reference
				bdID, ok := bdMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("TG->BD target not found: %q (from %q)", entry.Key, topicName))
					continue
				}
				if err := tx.TopicalGuideEntry.UpdateOneID(tgID).AddBdRefIDs(bdID).Exec(ctx); err != nil {
					return fmt.Errorf("adding TG->BD %q -> %q: %w", topicName, entry.Key, err)
				}
				l.stats.TGBDRefs++

			default:
				// Scripture reference
				builders := l.buildTGVerseRefs(entry, tgID, verseIndex, seenTGVerse)
				verseRefBuilders = append(verseRefBuilders, builders...)
			}
		}
	}

	// Bulk create TGVerseRef rows
	for i := 0; i < len(verseRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(verseRefBuilders) {
			end = len(verseRefBuilders)
		}
		if err := tx.TGVerseRef.CreateBulk(verseRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating TG verse refs: %w", err)
		}
	}
	l.stats.TGVerseRefs += len(verseRefBuilders)

	return nil
}

// buildTGVerseRefs creates TGVerseRef builders from a TG entry's scripture reference.
func (l *Loader) buildTGVerseRefs(entry TGEntryJSON, tgID int, verseIndex VerseIndex, seen map[[2]int]bool) []*generated.TGVerseRefCreate {
	result := l.refParser.Parse(entry.Reference)
	var builders []*generated.TGVerseRefCreate

	for _, ref := range result.Refs {
		verseID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
		if !ok {
			l.stats.Warn(fmt.Sprintf("TG verse ref not found: %s/%s %d:%d", ref.Volume, ref.Slug, ref.Chapter, ref.Verse))
			continue
		}

		pair := [2]int{tgID, verseID}
		if seen[pair] {
			continue
		}
		seen[pair] = true

		b := l.ec.TGVerseRef.Create().
			SetTgEntryID(tgID).
			SetVerseID(verseID)

		if entry.Phrase != "" {
			b.SetPhrase(entry.Phrase)
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
		l.stats.Warn(fmt.Sprintf("TG verse ref parse error: %s", e))
	}

	return builders
}

// loadBDRefs creates BD -> verse refs from the references array.
func (l *Loader) loadBDRefs(
	ctx context.Context,
	tx *generated.Tx,
	verseIndex VerseIndex,
	bdMap map[string]int,
) error {
	data, err := os.ReadFile(filepath.Join(l.dataDir, "bible-dictionary.json"))
	if err != nil {
		return fmt.Errorf("reading bible dictionary: %w", err)
	}

	var bdData map[string]BDEntryJSON
	if err := json.Unmarshal(data, &bdData); err != nil {
		return fmt.Errorf("parsing bible dictionary: %w", err)
	}

	var verseRefBuilders []*generated.BDVerseRefCreate
	batchSize := 500
	seenBDVerse := make(map[[2]int]bool)

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

				pair := [2]int{bdID, verseID}
				if seenBDVerse[pair] {
					continue
				}
				seenBDVerse[pair] = true

				b := l.ec.BDVerseRef.Create().
					SetBdEntryID(bdID).
					SetVerseID(verseID)

				if ref.EndVerse > 0 {
					endID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.EndVerse)
					if ok {
						b.SetTargetEndVerseID(endID)
					}
				}

				verseRefBuilders = append(verseRefBuilders, b)
			}

			for _, e := range result.Errors {
				l.stats.Warn(fmt.Sprintf("BD verse ref parse error (from %q): %s", entryName, e))
			}
		}
	}

	// Bulk create BDVerseRef rows
	for i := 0; i < len(verseRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(verseRefBuilders) {
			end = len(verseRefBuilders)
		}
		if err := tx.BDVerseRef.CreateBulk(verseRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating BD verse refs: %w", err)
		}
	}
	l.stats.BDVerseRefs += len(verseRefBuilders)

	return nil
}

// loadIDXRefs creates IDX -> verse refs, IDX -> IDX see-also, IDX -> TG, IDX -> BD edges.
func (l *Loader) loadIDXRefs(
	ctx context.Context,
	tx *generated.Tx,
	verseIndex VerseIndex,
	idxMap map[string]int,
	tgMap map[string]int,
	bdMap map[string]int,
) error {
	data, err := os.ReadFile(filepath.Join(l.dataDir, "triple-combination-index.json"))
	if err != nil {
		return fmt.Errorf("reading triple combination index: %w", err)
	}

	var idxData map[string][]IDXEntryJSON
	if err := json.Unmarshal(data, &idxData); err != nil {
		return fmt.Errorf("parsing triple combination index: %w", err)
	}

	var verseRefBuilders []*generated.IDXVerseRefCreate
	batchSize := 500
	seenIDXVerse := make(map[[2]int]bool)

	for entryName, entries := range idxData {
		idxID, ok := idxMap[entryName]
		if !ok {
			continue
		}

		for _, entry := range entries {
			switch entry.Reference {
			case "IDX":
				// See-also: IDX -> IDX
				targetID, ok := idxMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("IDX see-also target not found: %q (from %q)", entry.Key, entryName))
					continue
				}
				if err := tx.IndexEntry.UpdateOneID(idxID).AddSeeAlsoIDs(targetID).Exec(ctx); err != nil {
					return fmt.Errorf("adding IDX see-also %q -> %q: %w", entryName, entry.Key, err)
				}
				l.stats.IDXSeeAlso++

			case "TG":
				// IDX -> TG reference
				tgID, ok := tgMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("IDX->TG target not found: %q (from %q)", entry.Key, entryName))
					continue
				}
				if err := tx.IndexEntry.UpdateOneID(idxID).AddTgRefIDs(tgID).Exec(ctx); err != nil {
					return fmt.Errorf("adding IDX->TG %q -> %q: %w", entryName, entry.Key, err)
				}
				l.stats.IDXTGRefs++

			case "BD":
				// IDX -> BD reference
				bdID, ok := bdMap[entry.Key]
				if !ok {
					l.stats.Warn(fmt.Sprintf("IDX->BD target not found: %q (from %q)", entry.Key, entryName))
					continue
				}
				if err := tx.IndexEntry.UpdateOneID(idxID).AddBdRefIDs(bdID).Exec(ctx); err != nil {
					return fmt.Errorf("adding IDX->BD %q -> %q: %w", entryName, entry.Key, err)
				}
				l.stats.IDXBDRefs++

			default:
				// Scripture reference
				result := l.refParser.Parse(entry.Reference)
				for _, ref := range result.Refs {
					verseID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
					if !ok {
						l.stats.Warn(fmt.Sprintf("IDX verse ref not found: %s/%s %d:%d (from %q)",
							ref.Volume, ref.Slug, ref.Chapter, ref.Verse, entryName))
						continue
					}

					pair := [2]int{idxID, verseID}
					if seenIDXVerse[pair] {
						continue
					}
					seenIDXVerse[pair] = true

					b := l.ec.IDXVerseRef.Create().
						SetIndexEntryID(idxID).
						SetVerseID(verseID)

					if entry.Phrase != "" {
						b.SetPhrase(entry.Phrase)
					}

					if ref.EndVerse > 0 {
						endID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.EndVerse)
						if ok {
							b.SetTargetEndVerseID(endID)
						}
					}

					verseRefBuilders = append(verseRefBuilders, b)
				}

				for _, e := range result.Errors {
					l.stats.Warn(fmt.Sprintf("IDX verse ref parse error (from %q): %s", entryName, e))
				}
			}
		}
	}

	// Bulk create IDXVerseRef rows
	for i := 0; i < len(verseRefBuilders); i += batchSize {
		end := i + batchSize
		if end > len(verseRefBuilders) {
			end = len(verseRefBuilders)
		}
		if err := tx.IDXVerseRef.CreateBulk(verseRefBuilders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating IDX verse refs: %w", err)
		}
	}
	l.stats.IDXVerseRefs += len(verseRefBuilders)

	return nil
}

// loadJSTCompares creates JST -> Verse compare edges.
func (l *Loader) loadJSTCompares(
	ctx context.Context,
	tx *generated.Tx,
	verseIndex VerseIndex,
	jstIndex JSTIndex,
) error {
	data, err := os.ReadFile(filepath.Join(l.dataDir, "jst.json"))
	if err != nil {
		return fmt.Errorf("reading JST: %w", err)
	}

	var jstData []JSTChapterJSON
	if err := json.Unmarshal(data, &jstData); err != nil {
		return fmt.Errorf("parsing JST: %w", err)
	}

	for _, ch := range jstData {
		bookSlug := l.jstBookToSlug(ch.Book)
		if bookSlug == "" {
			continue
		}

		for _, entry := range ch.Entries {
			if entry.Compare == "" {
				continue
			}

			// Find the JST passage DB ID
			jstEntries := jstIndex.Get(bookSlug, ch.Chapter)
			var passageID int
			for _, je := range jstEntries {
				if je.comprises == entry.Comprises {
					passageID = je.dbID
					break
				}
			}
			if passageID == 0 {
				l.stats.Warn(fmt.Sprintf("JST passage not found for compare: %s %s:%s", ch.Book, ch.Chapter, entry.Comprises))
				continue
			}

			// Parse the compare reference to get verse IDs
			result := l.refParser.Parse(entry.Compare)
			var verseIDs []int

			for _, ref := range result.Refs {
				vID, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, ref.Verse)
				if !ok {
					l.stats.Warn(fmt.Sprintf("JST compare verse not found: %s/%s %d:%d",
						ref.Volume, ref.Slug, ref.Chapter, ref.Verse))
					continue
				}
				verseIDs = append(verseIDs, vID)

				// If range, add all verses in between
				if ref.EndVerse > 0 {
					for v := ref.Verse + 1; v <= ref.EndVerse; v++ {
						id, ok := verseIndex.Get(ref.Volume, ref.Slug, ref.Chapter, v)
						if ok {
							verseIDs = append(verseIDs, id)
						}
					}
				}
			}

			if len(verseIDs) > 0 {
				if err := tx.JSTPassage.UpdateOneID(passageID).AddCompareVerseIDs(verseIDs...).Exec(ctx); err != nil {
					return fmt.Errorf("adding JST compare verses for %s %s: %w", ch.Book, ch.Chapter, err)
				}
				l.stats.JSTCompares += len(verseIDs)
			}

			for _, e := range result.Errors {
				l.stats.Warn(fmt.Sprintf("JST compare parse error (%s %s): %s", ch.Book, ch.Chapter, e))
			}
		}
	}

	return nil
}

