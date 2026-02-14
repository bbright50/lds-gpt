package dataloader

import (
	"context"
	"fmt"
	"strings"

	"lds-gpt/internal/libsql/generated"
)

const (
	groupWindowSize = 5
	groupStepSize   = 3
)

// loadVerseGroups implements Phase 5: create sliding-window verse groups
// for semantic search embeddings.
func (l *Loader) loadVerseGroups(ctx context.Context) error {
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

	// Query all chapters with their verses ordered by verse number
	chapters, err := l.ec.Chapter.Query().
		WithVerses().
		All(ctx)
	if err != nil {
		return fmt.Errorf("querying chapters: %w", err)
	}

	var builders []*generated.VerseGroupCreate
	batchSize := 500

	for _, ch := range chapters {
		verses := ch.Edges.Verses
		if len(verses) == 0 {
			continue
		}

		// Sort verses by number (Ent may not guarantee order)
		sortedVerses := make([]*generated.Verse, len(verses))
		copy(sortedVerses, verses)
		sortVersesByNumber(sortedVerses)

		// Sliding window
		for start := 0; start < len(sortedVerses); start += groupStepSize {
			end := start + groupWindowSize
			if end > len(sortedVerses) {
				end = len(sortedVerses)
			}

			window := sortedVerses[start:end]
			if len(window) == 0 {
				continue
			}

			// Concatenate verse texts
			var textParts []string
			var verseIDs []int
			for _, v := range window {
				textParts = append(textParts, v.Text)
				verseIDs = append(verseIDs, v.ID)
			}
			text := strings.Join(textParts, " ")

			startVerseNum := window[0].Number
			endVerseNum := window[len(window)-1].Number

			b := l.ec.VerseGroup.Create().
				SetText(text).
				SetStartVerseNumber(startVerseNum).
				SetEndVerseNumber(endVerseNum).
				SetChapterID(ch.ID).
				AddVerseIDs(verseIDs...)

			builders = append(builders, b)

			// If window reached the end, stop sliding
			if end >= len(sortedVerses) {
				break
			}
		}
	}

	// Bulk create in batches
	for i := 0; i < len(builders); i += batchSize {
		end := i + batchSize
		if end > len(builders) {
			end = len(builders)
		}
		if err := tx.VerseGroup.CreateBulk(builders[i:end]...).Exec(ctx); err != nil {
			return fmt.Errorf("creating verse groups: %w", err)
		}
	}
	l.stats.VerseGroups += len(builders)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing phase 5: %w", err)
	}
	committed = true

	return nil
}

// sortVersesByNumber sorts verses by their Number field.
func sortVersesByNumber(verses []*generated.Verse) {
	// Simple insertion sort - chapters have at most ~176 verses (Psalm 119)
	for i := 1; i < len(verses); i++ {
		key := verses[i]
		j := i - 1
		for j >= 0 && verses[j].Number > key.Number {
			verses[j+1] = verses[j]
			j--
		}
		verses[j+1] = key
	}
}
