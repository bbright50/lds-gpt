package dataloader

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"lds-gpt/internal/embedding"
	"lds-gpt/internal/falkor"
)

// Loader orchestrates the ETL pipeline that populates the scripture
// knowledge graph from scraped JSON data into a FalkorDB property graph.
type Loader struct {
	fc               *falkor.Client
	dataDir          string
	stats            LoadStats
	abbrevs          map[string]BookInfo
	slugMap          map[string]string
	bookNames        map[string]BookInfo
	refParser        RefParser
	logger           *slog.Logger
	embedClient      embedding.Client
	embedBatchSize   int
	embedConcurrency int
	embedMaxTextLen  int
}

// LoaderOption configures optional Loader dependencies.
type LoaderOption func(*Loader)

// WithEmbedClient attaches an embedding client to enable Phase 6.
func WithEmbedClient(c embedding.Client) LoaderOption {
	return func(l *Loader) { l.embedClient = c }
}

// WithEmbedBatchSize tunes how many chunks Phase 6 sends per /api/embed
// request. Zero or negative falls back to the default (32).
func WithEmbedBatchSize(n int) LoaderOption {
	return func(l *Loader) { l.embedBatchSize = n }
}

// WithEmbedConcurrency caps how many /api/embed requests fly at once. Zero
// or negative falls back to the default (1 — matches stock Ollama, which
// serialises per model unless OLLAMA_NUM_PARALLEL is raised).
func WithEmbedConcurrency(n int) LoaderOption {
	return func(l *Loader) { l.embedConcurrency = n }
}

// WithEmbedMaxTextLen overrides how many characters each chunk is truncated
// to before hitting the embed model. Zero or negative → default (2000).
func WithEmbedMaxTextLen(n int) LoaderOption {
	return func(l *Loader) { l.embedMaxTextLen = n }
}

// New creates a new Loader.
func New(fc *falkor.Client, dataDir string, logger *slog.Logger, opts ...LoaderOption) *Loader {
	abbrevs := buildAbbreviationMap()
	l := &Loader{
		fc:        fc,
		dataDir:   dataDir,
		abbrevs:   abbrevs,
		slugMap:   buildSlugToAbbrevMap(),
		bookNames: buildBookDisplayNameMap(),
		refParser: NewRefParser(abbrevs),
		logger:    logger,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Run executes all 5 loading phases in order (plus Phase 6 if an embed
// client is attached). Phases 2-5 are currently stubbed during the FalkorDB
// migration — see the individual loader_*.go files for status.
func (l *Loader) Run(ctx context.Context) error {
	start := time.Now()
	l.logger.Info("starting dataloader")

	l.logger.Info("phase 1: loading structural data")
	phaseStart := time.Now()
	verseIndex, err := l.loadScriptures(ctx)
	if err != nil {
		return fmt.Errorf("phase 1 (scriptures): %w", err)
	}
	l.logger.Info("phase 1 complete",
		"volumes", l.stats.Volumes,
		"books", l.stats.Books,
		"chapters", l.stats.Chapters,
		"verses", l.stats.Verses,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	l.logger.Info("phase 2: loading study help entities")
	phaseStart = time.Now()
	tgMap, bdMap, idxMap, jstIndex, err := l.loadStudyHelps(ctx)
	if err != nil {
		return fmt.Errorf("phase 2 (study helps): %w", err)
	}
	l.logger.Info("phase 2 complete",
		"tg_entries", l.stats.TGEntries,
		"bd_entries", l.stats.BDEntries,
		"idx_entries", l.stats.IDXEntries,
		"jst_passages", l.stats.JSTPassages,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	l.logger.Info("phase 3: loading footnote edges")
	phaseStart = time.Now()
	if err := l.loadFootnotes(ctx, verseIndex, tgMap, bdMap, jstIndex); err != nil {
		return fmt.Errorf("phase 3 (footnotes): %w", err)
	}
	l.logger.Info("phase 3 complete", "duration", time.Since(phaseStart).Round(time.Millisecond))

	l.logger.Info("phase 4: loading study help edges")
	phaseStart = time.Now()
	if err := l.loadStudyRefs(ctx, verseIndex, tgMap, bdMap, idxMap, jstIndex); err != nil {
		return fmt.Errorf("phase 4 (study refs): %w", err)
	}
	l.logger.Info("phase 4 complete", "duration", time.Since(phaseStart).Round(time.Millisecond))

	l.logger.Info("phase 5: creating verse groups")
	phaseStart = time.Now()
	if err := l.loadVerseGroups(ctx); err != nil {
		return fmt.Errorf("phase 5 (verse groups): %w", err)
	}
	l.logger.Info("phase 5 complete",
		"verse_groups", l.stats.VerseGroups,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	if l.embedClient != nil {
		if err := l.runEmbeddings(ctx); err != nil {
			return fmt.Errorf("phase 6 (embeddings): %w", err)
		}
	}

	l.logger.Info("dataloader complete",
		"total_duration", time.Since(start).Round(time.Millisecond),
		"warnings", len(l.stats.Warnings),
	)
	l.logWarnings()
	return nil
}

// EmbedOnly runs Phase 6 only (embedding generation) against an existing graph.
func (l *Loader) EmbedOnly(ctx context.Context) error {
	if l.embedClient == nil {
		return fmt.Errorf("embed client is required for embed-only mode")
	}
	start := time.Now()
	l.logger.Info("starting embed-only mode")
	if err := l.runEmbeddings(ctx); err != nil {
		return fmt.Errorf("phase 6 (embeddings): %w", err)
	}
	l.logger.Info("embed-only complete",
		"total_duration", time.Since(start).Round(time.Millisecond),
		"warnings", len(l.stats.Warnings),
	)
	l.logWarnings()
	return nil
}

func (l *Loader) runEmbeddings(ctx context.Context) error {
	l.logger.Info("phase 6: generating embeddings")
	phaseStart := time.Now()
	if err := l.embedAll(ctx); err != nil {
		return err
	}
	l.logger.Info("phase 6 complete", "duration", time.Since(phaseStart).Round(time.Millisecond))
	return nil
}

func (l *Loader) logWarnings() {
	if len(l.stats.Warnings) == 0 {
		return
	}
	l.logger.Warn("warnings during loading", "count", len(l.stats.Warnings))
	limit := len(l.stats.Warnings)
	if limit > 20 {
		limit = 20
	}
	for _, w := range l.stats.Warnings[:limit] {
		l.logger.Warn(w)
	}
	if len(l.stats.Warnings) > 20 {
		l.logger.Warn("... and more warnings", "remaining", len(l.stats.Warnings)-20)
	}
}

// Stats returns a snapshot of the current loading statistics.
func (l *Loader) Stats() *LoadStats {
	return &l.stats
}
