package libsql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	"lds-gpt/internal/libsql/generated"
)

type contextSearchOptions struct {
	kNN int
}

type ContextSearchOption func(*contextSearchOptions)

func WithKNN(kNN int) ContextSearchOption {
	return func(o *contextSearchOptions) {
		o.kNN = kNN
	}
}

// defaultSearchLimit is the number of results returned per entity type.
// With 6 entity types, this yields up to 60 total results before merging and sorting.
const defaultSearchLimit = 10

// DoContextualSearch runs parallel vector similarity searches across all 6 entity
// tables, expands results via graph traversal, and returns ranked results.
// The embedding parameter must be pre-computed packed float32 bytes (F32_BLOB).
func (c *Client) DoContextualSearch(ctx context.Context, embedding []byte, options ...ContextSearchOption) ([]SearchResult, error) {
	opts := &contextSearchOptions{
		kNN: 20,
	}
	for _, opt := range options {
		opt(opts)
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("libsql: embedding must not be empty")
	}

	if opts.kNN <= 0 {
		return nil, fmt.Errorf("libsql: kNN must be positive, got %d", opts.kNN)
	}

	// Stage 1: Parallel vector search across all entity tables.
	stage1, err := runParallelSearches(ctx, c.Sqlx(), embedding)
	if err != nil {
		return nil, fmt.Errorf("libsql: vector search: %w", err)
	}

	// Stage 2: Graph traversal + deduplication.
	graphResults, err := graphExpandAndDedup(ctx, c.Ent(), stage1)
	if err != nil {
		return nil, fmt.Errorf("libsql: graph expansion: %w", err)
	}

	// Stage 3: Combine, rank, and trim to kNN.
	combined := make([]SearchResult, 0, len(stage1)+len(graphResults))
	combined = append(combined, stage1...)
	combined = append(combined, graphResults...)

	ranked := rankResults(combined)
	if len(ranked) > opts.kNN {
		ranked = ranked[:opts.kNN]
	}

	return ranked, nil
}

// runParallelSearches executes vector similarity searches across all 6 entity
// tables concurrently and returns the combined results.
func runParallelSearches(ctx context.Context, db *sqlx.DB, embedding []byte) ([]SearchResult, error) {
	type searchFunc func(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error)

	searches := []searchFunc{
		searchVerseGroups,
		searchChapters,
		searchTopicalGuide,
		searchBibleDict,
		searchIndex,
		searchJSTPassages,
	}

	resultsCh := make(chan []SearchResult, len(searches))

	g, gctx := errgroup.WithContext(ctx)
	for _, fn := range searches {
		g.Go(func() error {
			results, err := fn(gctx, db, embedding, defaultSearchLimit)
			if err != nil {
				return err
			}
			resultsCh <- results
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("libsql: contextual search: %w", err)
	}
	close(resultsCh)

	stage1 := make([]SearchResult, 0, len(searches)*defaultSearchLimit)
	for batch := range resultsCh {
		stage1 = append(stage1, batch...)
	}

	return stage1, nil
}

// graphExpandAndDedup performs 1-hop graph traversal from each Stage 1 seed,
// assigns synthetic distances, and deduplicates against Stage 1 results.
func graphExpandAndDedup(ctx context.Context, ec *generated.Client, stage1 []SearchResult) ([]SearchResult, error) {
	var allGraphResults []SearchResult
	for _, seed := range stage1 {
		neighbors, err := traverseEdges(ctx, ec, seed, defaultGraphLimit)
		if err != nil {
			return nil, fmt.Errorf("libsql: graph traversal: %w", err)
		}
		scored := assignSyntheticDistances(neighbors, seed.Distance)
		allGraphResults = append(allGraphResults, scored...)
	}
	return deduplicateResults(stage1, allGraphResults), nil
}

// verseGroupRow is the sqlx scan target for verse_groups vector search.
type verseGroupRow struct {
	ID               int     `db:"id"`
	Text             string  `db:"text"`
	StartVerseNumber int     `db:"start_verse_number"`
	EndVerseNumber   int     `db:"end_verse_number"`
	ChapterID        int     `db:"chapter_verse_groups"`
	Distance         float64 `db:"distance"`
}

func searchVerseGroups(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error) {
	const query = `SELECT id, text, start_verse_number, end_verse_number, chapter_verse_groups,
		vector_distance_cos(embedding, ?) AS distance
		FROM verse_groups
		WHERE embedding IS NOT NULL
		ORDER BY distance
		LIMIT ?`

	var rows []verseGroupRow
	if err := db.SelectContext(ctx, &rows, query, embedding, limit); err != nil {
		return nil, fmt.Errorf("searching verse_groups: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		results[i] = SearchResult{
			EntityType: EntityVerseGroup,
			ID:         r.ID,
			Text:       r.Text,
			Distance:   r.Distance,
			Metadata: ResultMeta{
				StartVerseNumber: r.StartVerseNumber,
				EndVerseNumber:   r.EndVerseNumber,
				ChapterID:        r.ChapterID,
			},
		}
	}
	return results, nil
}

// chapterRow is the sqlx scan target for chapters vector search.
type chapterRow struct {
	ID       int     `db:"id"`
	Number   int     `db:"number"`
	Summary  *string `db:"summary"`
	URL      *string `db:"url"`
	BookID   int     `db:"book_chapters"`
	Distance float64 `db:"distance"`
}

func searchChapters(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error) {
	const query = `SELECT id, number, summary, url, book_chapters,
		vector_distance_cos(summary_embedding, ?) AS distance
		FROM chapters
		WHERE summary_embedding IS NOT NULL
		ORDER BY distance
		LIMIT ?`

	var rows []chapterRow
	if err := db.SelectContext(ctx, &rows, query, embedding, limit); err != nil {
		return nil, fmt.Errorf("searching chapters: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		text := ""
		if r.Summary != nil {
			text = *r.Summary
		}
		url := ""
		if r.URL != nil {
			url = *r.URL
		}
		results[i] = SearchResult{
			EntityType: EntityChapter,
			ID:         r.ID,
			Text:       text,
			Distance:   r.Distance,
			Metadata: ResultMeta{
				ChapterNumber: r.Number,
				URL:           url,
			},
		}
	}
	return results, nil
}

// topicalGuideRow is the sqlx scan target for topical_guide_entries vector search.
type topicalGuideRow struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Distance float64 `db:"distance"`
}

func searchTopicalGuide(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error) {
	const query = `SELECT id, name,
		vector_distance_cos(embedding, ?) AS distance
		FROM topical_guide_entries
		WHERE embedding IS NOT NULL
		ORDER BY distance
		LIMIT ?`

	var rows []topicalGuideRow
	if err := db.SelectContext(ctx, &rows, query, embedding, limit); err != nil {
		return nil, fmt.Errorf("searching topical_guide_entries: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		results[i] = SearchResult{
			EntityType: EntityTopicalGuide,
			ID:         r.ID,
			Name:       r.Name,
			Distance:   r.Distance,
		}
	}
	return results, nil
}

// bibleDictRow is the sqlx scan target for bible_dict_entries vector search.
type bibleDictRow struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Text     string  `db:"text"`
	Distance float64 `db:"distance"`
}

func searchBibleDict(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error) {
	const query = `SELECT id, name, text,
		vector_distance_cos(embedding, ?) AS distance
		FROM bible_dict_entries
		WHERE embedding IS NOT NULL
		ORDER BY distance
		LIMIT ?`

	var rows []bibleDictRow
	if err := db.SelectContext(ctx, &rows, query, embedding, limit); err != nil {
		return nil, fmt.Errorf("searching bible_dict_entries: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		results[i] = SearchResult{
			EntityType: EntityBibleDict,
			ID:         r.ID,
			Name:       r.Name,
			Text:       r.Text,
			Distance:   r.Distance,
		}
	}
	return results, nil
}

// indexRow is the sqlx scan target for index_entries vector search.
type indexRow struct {
	ID       int     `db:"id"`
	Name     string  `db:"name"`
	Distance float64 `db:"distance"`
}

func searchIndex(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error) {
	const query = `SELECT id, name,
		vector_distance_cos(embedding, ?) AS distance
		FROM index_entries
		WHERE embedding IS NOT NULL
		ORDER BY distance
		LIMIT ?`

	var rows []indexRow
	if err := db.SelectContext(ctx, &rows, query, embedding, limit); err != nil {
		return nil, fmt.Errorf("searching index_entries: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		results[i] = SearchResult{
			EntityType: EntityIndex,
			ID:         r.ID,
			Name:       r.Name,
			Distance:   r.Distance,
		}
	}
	return results, nil
}

// jstPassageRow is the sqlx scan target for jst_passages vector search.
type jstPassageRow struct {
	ID         int     `db:"id"`
	Book       string  `db:"book"`
	Chapter    string  `db:"chapter"`
	Comprises  string  `db:"comprises"`
	CompareRef *string `db:"compare_ref"`
	Summary    *string `db:"summary"`
	Text       string  `db:"text"`
	Distance   float64 `db:"distance"`
}

func searchJSTPassages(ctx context.Context, db *sqlx.DB, embedding []byte, limit int) ([]SearchResult, error) {
	const query = `SELECT id, book, chapter, comprises, compare_ref, summary, text,
		vector_distance_cos(embedding, ?) AS distance
		FROM jst_passages
		WHERE embedding IS NOT NULL
		ORDER BY distance
		LIMIT ?`

	var rows []jstPassageRow
	if err := db.SelectContext(ctx, &rows, query, embedding, limit); err != nil {
		return nil, fmt.Errorf("searching jst_passages: %w", err)
	}

	results := make([]SearchResult, len(rows))
	for i, r := range rows {
		compareRef := ""
		if r.CompareRef != nil {
			compareRef = *r.CompareRef
		}
		summary := ""
		if r.Summary != nil {
			summary = *r.Summary
		}
		results[i] = SearchResult{
			EntityType: EntityJSTPassage,
			ID:         r.ID,
			Text:       r.Text,
			Distance:   r.Distance,
			Metadata: ResultMeta{
				Book:       r.Book,
				Chapter:    r.Chapter,
				Comprises:  r.Comprises,
				CompareRef: compareRef,
				Summary:    summary,
			},
		}
	}
	return results, nil
}
