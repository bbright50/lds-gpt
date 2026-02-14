package dataloader

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"lds-gpt/internal/libsql/generated"
	"lds-gpt/internal/libsql/generated/bibledictentry"
	"lds-gpt/internal/libsql/generated/chapter"
	"lds-gpt/internal/libsql/generated/idxverseref"
	"lds-gpt/internal/libsql/generated/indexentry"
	"lds-gpt/internal/libsql/generated/jstpassage"
	"lds-gpt/internal/libsql/generated/tgverseref"
	"lds-gpt/internal/libsql/generated/topicalguideentry"
	"lds-gpt/internal/libsql/generated/versegroup"
	"lds-gpt/internal/utils/vec"

	"golang.org/x/sync/errgroup"
)

const (
	embedConcurrency = 8
	progressInterval = 500
	maxTextLen       = 25000
	maxPhrases       = 20
)

// embedAll runs embedding generation for all 6 entity types sequentially.
func (l *Loader) embedAll(ctx context.Context) error {
	type embedFunc struct {
		name string
		fn   func(context.Context) (int, error)
		stat *int
	}

	funcs := []embedFunc{
		{"verse_groups", l.embedVerseGroups, &l.stats.EmbVerseGroups},
		{"chapters", l.embedChapters, &l.stats.EmbChapters},
		{"tg_entries", l.embedTGEntries, &l.stats.EmbTGEntries},
		{"bd_entries", l.embedBDEntries, &l.stats.EmbBDEntries},
		{"idx_entries", l.embedIDXEntries, &l.stats.EmbIDXEntries},
		{"jst_passages", l.embedJSTPassages, &l.stats.EmbJSTPassages},
	}

	for _, ef := range funcs {
		count, err := ef.fn(ctx)
		if err != nil {
			return fmt.Errorf("embedding %s: %w", ef.name, err)
		}
		*ef.stat = count
		l.logger.Info("embedded entity type", "type", ef.name, "count", count)
	}

	return nil
}

func (l *Loader) embedVerseGroups(ctx context.Context) (int, error) {
	rows, err := l.ec.VerseGroup.Query().
		Where(versegroup.EmbeddingIsNil()).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying verse groups: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	l.logger.Info("embedding verse groups", "total", len(rows))

	var completed atomic.Int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, row := range rows {
		g.Go(func() error {
			text := truncateText(row.Text)
			emb, err := l.embedClient.EmbedText(gCtx, text)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embed verse_group %d: %v", row.ID, err))
				return nil
			}

			if err := l.ec.VerseGroup.UpdateOneID(row.ID).
				SetEmbedding(vec.Float64sToFloat32Bytes(emb)).
				Exec(gCtx); err != nil {
				l.stats.Warn(fmt.Sprintf("save verse_group embedding %d: %v", row.ID, err))
				return nil
			}

			n := completed.Add(1)
			if n%progressInterval == 0 {
				l.logger.Info("verse group embedding progress", "completed", n, "total", len(rows))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return int(completed.Load()), nil
}

func (l *Loader) embedChapters(ctx context.Context) (int, error) {
	rows, err := l.ec.Chapter.Query().
		Where(
			chapter.SummaryEmbeddingIsNil(),
			chapter.SummaryNEQ(""),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying chapters: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	l.logger.Info("embedding chapters", "total", len(rows))

	var completed atomic.Int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, row := range rows {
		g.Go(func() error {
			if row.Summary == "" {
				return nil
			}

			emb, err := l.embedClient.EmbedText(gCtx, row.Summary)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embed chapter %d: %v", row.ID, err))
				return nil
			}

			if err := l.ec.Chapter.UpdateOneID(row.ID).
				SetSummaryEmbedding(vec.Float64sToFloat32Bytes(emb)).
				Exec(gCtx); err != nil {
				l.stats.Warn(fmt.Sprintf("save chapter embedding %d: %v", row.ID, err))
				return nil
			}

			n := completed.Add(1)
			if n%progressInterval == 0 {
				l.logger.Info("chapter embedding progress", "completed", n, "total", len(rows))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return int(completed.Load()), nil
}

func (l *Loader) embedTGEntries(ctx context.Context) (int, error) {
	rows, err := l.ec.TopicalGuideEntry.Query().
		Where(topicalguideentry.EmbeddingIsNil()).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying TG entries: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	l.logger.Info("embedding TG entries", "total", len(rows))

	var completed atomic.Int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, row := range rows {
		g.Go(func() error {
			text, err := l.composeTGEmbedText(gCtx, row.ID, row.Name)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("compose TG text %d: %v", row.ID, err))
				text = row.Name
			}

			emb, err := l.embedClient.EmbedText(gCtx, text)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embed tg_entry %d: %v", row.ID, err))
				return nil
			}

			if err := l.ec.TopicalGuideEntry.UpdateOneID(row.ID).
				SetEmbedding(vec.Float64sToFloat32Bytes(emb)).
				Exec(gCtx); err != nil {
				l.stats.Warn(fmt.Sprintf("save tg_entry embedding %d: %v", row.ID, err))
				return nil
			}

			n := completed.Add(1)
			if n%progressInterval == 0 {
				l.logger.Info("TG entry embedding progress", "completed", n, "total", len(rows))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return int(completed.Load()), nil
}

func (l *Loader) embedBDEntries(ctx context.Context) (int, error) {
	rows, err := l.ec.BibleDictEntry.Query().
		Where(bibledictentry.EmbeddingIsNil()).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying BD entries: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	l.logger.Info("embedding BD entries", "total", len(rows))

	var completed atomic.Int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, row := range rows {
		g.Go(func() error {
			text := truncateText(row.Text)
			emb, err := l.embedClient.EmbedText(gCtx, text)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embed bd_entry %d: %v", row.ID, err))
				return nil
			}

			if err := l.ec.BibleDictEntry.UpdateOneID(row.ID).
				SetEmbedding(vec.Float64sToFloat32Bytes(emb)).
				Exec(gCtx); err != nil {
				l.stats.Warn(fmt.Sprintf("save bd_entry embedding %d: %v", row.ID, err))
				return nil
			}

			n := completed.Add(1)
			if n%progressInterval == 0 {
				l.logger.Info("BD entry embedding progress", "completed", n, "total", len(rows))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return int(completed.Load()), nil
}

func (l *Loader) embedIDXEntries(ctx context.Context) (int, error) {
	rows, err := l.ec.IndexEntry.Query().
		Where(indexentry.EmbeddingIsNil()).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying IDX entries: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	l.logger.Info("embedding IDX entries", "total", len(rows))

	var completed atomic.Int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, row := range rows {
		g.Go(func() error {
			text, err := l.composeIDXEmbedText(gCtx, row.ID, row.Name)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("compose IDX text %d: %v", row.ID, err))
				text = row.Name
			}

			emb, err := l.embedClient.EmbedText(gCtx, text)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embed idx_entry %d: %v", row.ID, err))
				return nil
			}

			if err := l.ec.IndexEntry.UpdateOneID(row.ID).
				SetEmbedding(vec.Float64sToFloat32Bytes(emb)).
				Exec(gCtx); err != nil {
				l.stats.Warn(fmt.Sprintf("save idx_entry embedding %d: %v", row.ID, err))
				return nil
			}

			n := completed.Add(1)
			if n%progressInterval == 0 {
				l.logger.Info("IDX entry embedding progress", "completed", n, "total", len(rows))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return int(completed.Load()), nil
}

func (l *Loader) embedJSTPassages(ctx context.Context) (int, error) {
	rows, err := l.ec.JSTPassage.Query().
		Where(jstpassage.EmbeddingIsNil()).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying JST passages: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	l.logger.Info("embedding JST passages", "total", len(rows))

	var completed atomic.Int64
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(embedConcurrency)

	for _, row := range rows {
		g.Go(func() error {
			text := row.Text
			if row.Summary != "" {
				text = row.Summary + " " + text
			}
			text = truncateText(text)

			emb, err := l.embedClient.EmbedText(gCtx, text)
			if err != nil {
				l.stats.Warn(fmt.Sprintf("embed jst_passage %d: %v", row.ID, err))
				return nil
			}

			if err := l.ec.JSTPassage.UpdateOneID(row.ID).
				SetEmbedding(vec.Float64sToFloat32Bytes(emb)).
				Exec(gCtx); err != nil {
				l.stats.Warn(fmt.Sprintf("save jst_passage embedding %d: %v", row.ID, err))
				return nil
			}

			n := completed.Add(1)
			if n%progressInterval == 0 {
				l.logger.Info("JST passage embedding progress", "completed", n, "total", len(rows))
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	return int(completed.Load()), nil
}

// composeTGEmbedText builds "{name}. {phrase1}; {phrase2}; ..." for a TG entry.
func (l *Loader) composeTGEmbedText(ctx context.Context, entryID int, name string) (string, error) {
	refs, err := l.ec.TGVerseRef.Query().
		Where(tgverseref.TgEntryID(entryID)).
		All(ctx)
	if err != nil {
		return "", fmt.Errorf("querying TG verse refs: %w", err)
	}

	phrases := collectPhrases(refs, func(r *generated.TGVerseRef) string { return r.Phrase })
	if len(phrases) == 0 {
		return name, nil
	}

	return name + ". " + strings.Join(phrases, "; "), nil
}

// composeIDXEmbedText builds "{name}. {phrase1}; {phrase2}; ..." for an IDX entry.
func (l *Loader) composeIDXEmbedText(ctx context.Context, entryID int, name string) (string, error) {
	refs, err := l.ec.IDXVerseRef.Query().
		Where(idxverseref.IndexEntryID(entryID)).
		All(ctx)
	if err != nil {
		return "", fmt.Errorf("querying IDX verse refs: %w", err)
	}

	phrases := collectPhrases(refs, func(r *generated.IDXVerseRef) string { return r.Phrase })
	if len(phrases) == 0 {
		return name, nil
	}

	return name + ". " + strings.Join(phrases, "; "), nil
}

// collectPhrases extracts unique, non-empty phrases from ref rows, up to maxPhrases.
func collectPhrases[T any](refs []T, getPhrase func(T) string) []string {
	seen := make(map[string]struct{}, maxPhrases)
	phrases := make([]string, 0, maxPhrases)

	for _, ref := range refs {
		if len(phrases) >= maxPhrases {
			break
		}
		p := strings.TrimSpace(getPhrase(ref))
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		phrases = append(phrases, p)
	}

	return phrases
}

// truncateText limits text to maxTextLen characters to stay within Titan's token limit.
func truncateText(s string) string {
	if len(s) <= maxTextLen {
		return s
	}
	return s[:maxTextLen]
}
