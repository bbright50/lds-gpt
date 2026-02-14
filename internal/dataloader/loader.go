package dataloader

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"lds-gpt/internal/bedrockembedding"
	"lds-gpt/internal/libsql/generated"
)

// Loader orchestrates the ETL pipeline that populates the scripture
// knowledge graph from scraped JSON data into a LibSQL (SQLite) database.
type Loader struct {
	ec          *generated.Client
	dataDir     string
	stats       LoadStats
	abbrevs     map[string]BookInfo
	slugMap     map[string]string
	bookNames   map[string]BookInfo
	refParser   RefParser
	logger      *slog.Logger
	embedClient bedrockembedding.Client
}

// LoaderOption configures optional Loader dependencies.
type LoaderOption func(*Loader)

// WithEmbedClient attaches a Bedrock embedding client to enable Phase 6.
func WithEmbedClient(c bedrockembedding.Client) LoaderOption {
	return func(l *Loader) { l.embedClient = c }
}

// New creates a new Loader.
func New(ec *generated.Client, dataDir string, logger *slog.Logger, opts ...LoaderOption) *Loader {
	abbrevs := buildAbbreviationMap()
	l := &Loader{
		ec:        ec,
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

// Run executes all 5 loading phases in order.
func (l *Loader) Run(ctx context.Context) error {
	start := time.Now()
	l.logger.Info("starting dataloader")

	// Phase 1: Structural data (volumes, books, chapters, verses)
	l.logger.Info("phase 1: loading structural data")
	phaseStart := time.Now()

	verseIndex, volumeIDs, err := l.loadScriptures(ctx)
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

	_ = volumeIDs // used implicitly through book creation

	// Phase 2: Study help entities (TG, BD, IDX, JST)
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

	// Phase 3: Footnote edges (cross-refs, TG/BD/JST footnotes)
	l.logger.Info("phase 3: loading footnote edges")
	phaseStart = time.Now()

	err = l.loadFootnotes(ctx, verseIndex, tgMap, bdMap, jstIndex)
	if err != nil {
		return fmt.Errorf("phase 3 (footnotes): %w", err)
	}

	l.logger.Info("phase 3 complete",
		"cross_refs", l.stats.CrossRefs,
		"verse_tg_refs", l.stats.VerseTGRefs,
		"verse_bd_refs", l.stats.VerseBDRefs,
		"verse_jst_refs", l.stats.VerseJSTRefs,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	// Phase 4: Study help edges (see-also, verse refs)
	l.logger.Info("phase 4: loading study help edges")
	phaseStart = time.Now()

	err = l.loadStudyRefs(ctx, verseIndex, tgMap, bdMap, idxMap, jstIndex)
	if err != nil {
		return fmt.Errorf("phase 4 (study refs): %w", err)
	}

	l.logger.Info("phase 4 complete",
		"tg_verse_refs", l.stats.TGVerseRefs,
		"bd_verse_refs", l.stats.BDVerseRefs,
		"idx_verse_refs", l.stats.IDXVerseRefs,
		"tg_see_also", l.stats.TGSeeAlso,
		"bd_see_also", l.stats.BDSeeAlso,
		"idx_see_also", l.stats.IDXSeeAlso,
		"jst_compares", l.stats.JSTCompares,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	// Phase 5: Verse groups (sliding window)
	l.logger.Info("phase 5: creating verse groups")
	phaseStart = time.Now()

	err = l.loadVerseGroups(ctx)
	if err != nil {
		return fmt.Errorf("phase 5 (verse groups): %w", err)
	}

	l.logger.Info("phase 5 complete",
		"verse_groups", l.stats.VerseGroups,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	// Phase 6: Embeddings (only if embed client is configured)
	if l.embedClient != nil {
		if err := l.runEmbeddings(ctx); err != nil {
			return fmt.Errorf("phase 6 (embeddings): %w", err)
		}
	}

	// Summary
	l.logger.Info("dataloader complete",
		"total_duration", time.Since(start).Round(time.Millisecond),
		"warnings", len(l.stats.Warnings),
	)

	l.logWarnings()

	return nil
}

// EmbedOnly runs Phase 6 only (embedding generation) against an existing database.
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

	l.logger.Info("phase 6 complete",
		"verse_groups", l.stats.EmbVerseGroups,
		"chapters", l.stats.EmbChapters,
		"tg_entries", l.stats.EmbTGEntries,
		"bd_entries", l.stats.EmbBDEntries,
		"idx_entries", l.stats.EmbIDXEntries,
		"jst_passages", l.stats.EmbJSTPassages,
		"duration", time.Since(phaseStart).Round(time.Millisecond),
	)

	return nil
}

func (l *Loader) logWarnings() {
	if len(l.stats.Warnings) > 0 {
		l.logger.Warn("warnings during loading", "count", len(l.stats.Warnings))
		limit := min(len(l.stats.Warnings), 20)
		for _, w := range l.stats.Warnings[:limit] {
			l.logger.Warn(w)
		}
		if len(l.stats.Warnings) > 20 {
			l.logger.Warn("... and more warnings", "remaining", len(l.stats.Warnings)-20)
		}
	}
}

// Stats returns a snapshot of the current loading statistics.
func (l *Loader) Stats() *LoadStats {
	return &l.stats
}
