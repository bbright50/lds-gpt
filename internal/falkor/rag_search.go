package falkor

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	ormql "github.com/tab58/go-ormql/pkg/client"
)

type contextSearchOptions struct {
	kNN int
}

type ContextSearchOption func(*contextSearchOptions)

// WithKNN caps the final result-set size after Stage-3 ranking (default 20).
func WithKNN(kNN int) ContextSearchOption {
	return func(o *contextSearchOptions) { o.kNN = kNN }
}

// DoContextualSearch runs the three-stage retrieval pipeline:
//
//  1. Parallel vector similarity kNN across the six embeddable node labels
//     (VerseGroup, Chapter.summaryEmbedding, TopicalGuideEntry,
//     BibleDictEntry, IndexEntry, JSTPassage). Each label fires a typed
//     `xxxSimilar(vector:, first:)` GraphQL query through the go-ormql
//     generated client; the fork-patched FalkorDB driver rewrites the
//     translator output to `CALL db.idx.vector.queryNodes(..., vecf32(...))`.
//  2. One-hop graph expansion from each Stage 1 seed, synthetic-distance
//     bumped and deduplicated against Stage 1.
//  3. Heuristic re-ranking (`rankScore = distance - verseBonus`) and trim
//     to kNN.
//
// Embedding is `[]float32` — FalkorDB's `vecf32()` takes floats directly.
func (c *Client) DoContextualSearch(
	ctx context.Context,
	embedding []float32,
	options ...ContextSearchOption,
) ([]SearchResult, error) {
	opts := &contextSearchOptions{kNN: defaultKNN}
	for _, opt := range options {
		opt(opts)
	}
	if len(embedding) == 0 {
		return nil, fmt.Errorf("falkor: embedding must not be empty")
	}
	if opts.kNN <= 0 {
		return nil, fmt.Errorf("falkor: kNN must be positive, got %d", opts.kNN)
	}

	vecParam := float32sToAnySlice(embedding)

	stage1, err := c.runParallelSearches(ctx, vecParam)
	if err != nil {
		return nil, fmt.Errorf("falkor: contextual search: %w", err)
	}

	graphResults, err := c.graphExpandAndDedup(ctx, stage1)
	if err != nil {
		return nil, fmt.Errorf("falkor: graph expansion: %w", err)
	}

	combined := make([]SearchResult, 0, len(stage1)+len(graphResults))
	combined = append(combined, stage1...)
	combined = append(combined, graphResults...)
	ranked := rankResults(combined)
	if len(ranked) > opts.kNN {
		ranked = ranked[:opts.kNN]
	}
	return ranked, nil
}

// --- Stage 1: six parallel kNN queries via the typed go-ormql client ---

type searchFn func(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error)

func (c *Client) runParallelSearches(ctx context.Context, vec []any) ([]SearchResult, error) {
	searches := []searchFn{
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
		fn := fn
		g.Go(func() error {
			rs, err := fn(gctx, c.GraphQL(), vec, defaultSearchLimit)
			if err != nil {
				return err
			}
			resultsCh <- rs
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(resultsCh)

	stage1 := make([]SearchResult, 0, len(searches)*defaultSearchLimit)
	for batch := range resultsCh {
		stage1 = append(stage1, batch...)
	}
	return stage1, nil
}

// similarHit is the decode shape for the six xxxSimilar queries — the
// generator produces a `{ score: Float!, node: <NodeType>! }` envelope.
type similarHit[T any] struct {
	Score float64 `json:"score"`
	Node  T       `json:"node"`
}

type verseGroupHit struct {
	Id       string `json:"id"`
	Text             string `json:"text"`
	StartVerseNumber int    `json:"startVerseNumber"`
	EndVerseNumber   int    `json:"endVerseNumber"`
}

func searchVerseGroups(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error) {
	var out struct {
		Hits []similarHit[verseGroupHit] `json:"verseGroupsSimilar"`
	}
	if err := execQuery(ctx, gc, `
		query ($vec: [Float!]!, $first: Int) {
		  verseGroupsSimilar(vector: $vec, first: $first) {
		    score
		    node { id, text, startVerseNumber, endVerseNumber }
		  }
		}`, map[string]any{"vec": vec, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("searchVerseGroups: %w", err)
	}
	results := make([]SearchResult, 0, len(out.Hits))
	for _, h := range out.Hits {
		results = append(results, SearchResult{
			EntityType: EntityVerseGroup,
			ID:         h.Node.Id,
			Text:       h.Node.Text,
			Distance:   h.Score,
			Metadata: ResultMeta{
				StartVerseNumber: h.Node.StartVerseNumber,
				EndVerseNumber:   h.Node.EndVerseNumber,
			},
		})
	}
	return results, nil
}

type chapterHit struct {
	Id string `json:"id"`
	Number     int    `json:"number"`
	Summary    string `json:"summary"`
	URL        string `json:"url"`
}

func searchChapters(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error) {
	var out struct {
		Hits []similarHit[chapterHit] `json:"chaptersSimilar"`
	}
	if err := execQuery(ctx, gc, `
		query ($vec: [Float!]!, $first: Int) {
		  chaptersSimilar(vector: $vec, first: $first) {
		    score
		    node { id, number, summary, url }
		  }
		}`, map[string]any{"vec": vec, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("searchChapters: %w", err)
	}
	results := make([]SearchResult, 0, len(out.Hits))
	for _, h := range out.Hits {
		results = append(results, SearchResult{
			EntityType: EntityChapter,
			ID:         h.Node.Id,
			Text:       h.Node.Summary,
			Distance:   h.Score,
			Metadata: ResultMeta{
				ChapterNumber: h.Node.Number,
				URL:           h.Node.URL,
			},
		})
	}
	return results, nil
}

type tgHit struct {
	Id string `json:"id"`
	Name       string `json:"name"`
}

func searchTopicalGuide(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error) {
	var out struct {
		Hits []similarHit[tgHit] `json:"topicalGuideEntriesSimilar"`
	}
	if err := execQuery(ctx, gc, `
		query ($vec: [Float!]!, $first: Int) {
		  topicalGuideEntriesSimilar(vector: $vec, first: $first) {
		    score
		    node { id, name }
		  }
		}`, map[string]any{"vec": vec, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("searchTopicalGuide: %w", err)
	}
	results := make([]SearchResult, 0, len(out.Hits))
	for _, h := range out.Hits {
		results = append(results, SearchResult{
			EntityType: EntityTopicalGuide,
			ID:         h.Node.Id,
			Name:       h.Node.Name,
			Distance:   h.Score,
		})
	}
	return results, nil
}

type bdHit struct {
	Id string `json:"id"`
	Name       string `json:"name"`
	Text       string `json:"text"`
}

func searchBibleDict(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error) {
	var out struct {
		Hits []similarHit[bdHit] `json:"bibleDictEntriesSimilar"`
	}
	if err := execQuery(ctx, gc, `
		query ($vec: [Float!]!, $first: Int) {
		  bibleDictEntriesSimilar(vector: $vec, first: $first) {
		    score
		    node { id, name, text }
		  }
		}`, map[string]any{"vec": vec, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("searchBibleDict: %w", err)
	}
	results := make([]SearchResult, 0, len(out.Hits))
	for _, h := range out.Hits {
		results = append(results, SearchResult{
			EntityType: EntityBibleDict,
			ID:         h.Node.Id,
			Name:       h.Node.Name,
			Text:       h.Node.Text,
			Distance:   h.Score,
		})
	}
	return results, nil
}

type idxHit struct {
	Id string `json:"id"`
	Name       string `json:"name"`
}

func searchIndex(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error) {
	var out struct {
		Hits []similarHit[idxHit] `json:"indexEntriesSimilar"`
	}
	if err := execQuery(ctx, gc, `
		query ($vec: [Float!]!, $first: Int) {
		  indexEntriesSimilar(vector: $vec, first: $first) {
		    score
		    node { id, name }
		  }
		}`, map[string]any{"vec": vec, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("searchIndex: %w", err)
	}
	results := make([]SearchResult, 0, len(out.Hits))
	for _, h := range out.Hits {
		results = append(results, SearchResult{
			EntityType: EntityIndex,
			ID:         h.Node.Id,
			Name:       h.Node.Name,
			Distance:   h.Score,
		})
	}
	return results, nil
}

type jstHit struct {
	Id string `json:"id"`
	Book       string `json:"book"`
	Chapter    string `json:"chapter"`
	Comprises  string `json:"comprises"`
	CompareRef string `json:"compareRef"`
	Summary    string `json:"summary"`
	Text       string `json:"text"`
}

func searchJSTPassages(ctx context.Context, gc *ormql.Client, vec []any, limit int) ([]SearchResult, error) {
	var out struct {
		Hits []similarHit[jstHit] `json:"jSTPassagesSimilar"` // generator camelCase: "JST" → "jST"
	}
	if err := execQuery(ctx, gc, `
		query ($vec: [Float!]!, $first: Int) {
		  jSTPassagesSimilar(vector: $vec, first: $first) {
		    score
		    node { id, book, chapter, comprises, compareRef, summary, text }
		  }
		}`, map[string]any{"vec": vec, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("searchJSTPassages: %w", err)
	}
	results := make([]SearchResult, 0, len(out.Hits))
	for _, h := range out.Hits {
		results = append(results, SearchResult{
			EntityType: EntityJSTPassage,
			ID:         h.Node.Id,
			Text:       h.Node.Text,
			Distance:   h.Score,
			Metadata: ResultMeta{
				Book:       h.Node.Book,
				Chapter:    h.Node.Chapter,
				Comprises:  h.Node.Comprises,
				CompareRef: h.Node.CompareRef,
				Summary:    h.Node.Summary,
			},
		})
	}
	return results, nil
}

// --- Helpers ---

func float32sToAnySlice(v []float32) []interface{} {
	out := make([]interface{}, len(v))
	for i, x := range v {
		// Pass as float64 — falkordb-go's ToString serializes via
		// strconv.FormatFloat('f', -1, 64) which correctly preserves decimal
		// points for non-integer values.
		out[i] = float64(x)
	}
	return out
}
