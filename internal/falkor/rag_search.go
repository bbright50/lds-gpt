package falkor

import (
	"context"
	"fmt"

	"github.com/FalkorDB/falkordb-go/v2"
	"golang.org/x/sync/errgroup"
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
//     (VerseGroup, Chapter.summaryEmbedding, TopicalGuideEntry, BibleDictEntry,
//     IndexEntry, JSTPassage).
//  2. One-hop graph expansion from each Stage 1 seed, with a synthetic
//     distance bump (defaultHopPenalty) and deduplication against Stage 1.
//  3. Heuristic re-ranking (`rankScore = distance - verseBonus`) and trim to
//     kNN.
//
// The embedding is `[]float32` — FalkorDB's `vecf32()` constructor accepts
// floats directly so the byte-packing used by the prior LibSQL client is
// gone. Stage 1 flows through `Client.Raw()` because go-ormql's @vector
// query rewrite has a known bug (missing vecf32() wrap on the vector param);
// Stage 2 flows through `Client.GraphQL()` because graph traversal is the
// shape go-ormql handles well.
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

// --- Stage 1: six parallel kNN queries via raw Cypher ---

type searchFn func(ctx context.Context, g *falkordb.Graph, vecParam []interface{}, limit int) ([]SearchResult, error)

func (c *Client) runParallelSearches(ctx context.Context, vecParam []interface{}) ([]SearchResult, error) {
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
			rs, err := fn(gctx, c.Raw(), vecParam, defaultSearchLimit)
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

func searchVerseGroups(ctx context.Context, g *falkordb.Graph, vec []interface{}, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('VerseGroup', 'embedding', $k, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, node.text AS text,
		        node.startVerseNumber AS startNum, node.endVerseNumber AS endNum,
		        score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": vec, "k": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("searchVerseGroups: %w", err)
	}
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		text, _ := rec.Get("text")
		startNum, _ := rec.Get("startNum")
		endNum, _ := rec.Get("endNum")
		score, _ := rec.Get("score")
		out = append(out, SearchResult{
			EntityType: EntityVerseGroup,
			ID:         asString(id),
			Text:       asString(text),
			Distance:   asFloat(score),
			Metadata: ResultMeta{
				StartVerseNumber: asInt(startNum),
				EndVerseNumber:   asInt(endNum),
			},
		})
	}
	return out, nil
}

func searchChapters(ctx context.Context, g *falkordb.Graph, vec []interface{}, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('Chapter', 'summaryEmbedding', $k, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, node.number AS number, node.summary AS summary, node.url AS url, score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": vec, "k": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("searchChapters: %w", err)
	}
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		number, _ := rec.Get("number")
		summary, _ := rec.Get("summary")
		url, _ := rec.Get("url")
		score, _ := rec.Get("score")
		out = append(out, SearchResult{
			EntityType: EntityChapter,
			ID:         asString(id),
			Text:       asString(summary),
			Distance:   asFloat(score),
			Metadata: ResultMeta{
				ChapterNumber: asInt(number),
				URL:           asString(url),
			},
		})
	}
	return out, nil
}

func searchTopicalGuide(ctx context.Context, g *falkordb.Graph, vec []interface{}, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('TopicalGuideEntry', 'embedding', $k, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, node.name AS name, score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": vec, "k": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("searchTopicalGuide: %w", err)
	}
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		name, _ := rec.Get("name")
		score, _ := rec.Get("score")
		out = append(out, SearchResult{
			EntityType: EntityTopicalGuide,
			ID:         asString(id),
			Name:       asString(name),
			Distance:   asFloat(score),
		})
	}
	return out, nil
}

func searchBibleDict(ctx context.Context, g *falkordb.Graph, vec []interface{}, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('BibleDictEntry', 'embedding', $k, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, node.name AS name, node.text AS text, score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": vec, "k": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("searchBibleDict: %w", err)
	}
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		name, _ := rec.Get("name")
		text, _ := rec.Get("text")
		score, _ := rec.Get("score")
		out = append(out, SearchResult{
			EntityType: EntityBibleDict,
			ID:         asString(id),
			Name:       asString(name),
			Text:       asString(text),
			Distance:   asFloat(score),
		})
	}
	return out, nil
}

func searchIndex(ctx context.Context, g *falkordb.Graph, vec []interface{}, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('IndexEntry', 'embedding', $k, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, node.name AS name, score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": vec, "k": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("searchIndex: %w", err)
	}
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		name, _ := rec.Get("name")
		score, _ := rec.Get("score")
		out = append(out, SearchResult{
			EntityType: EntityIndex,
			ID:         asString(id),
			Name:       asString(name),
			Distance:   asFloat(score),
		})
	}
	return out, nil
}

func searchJSTPassages(ctx context.Context, g *falkordb.Graph, vec []interface{}, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('JSTPassage', 'embedding', $k, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, node.book AS book, node.chapter AS chapter,
		        node.comprises AS comprises, node.compareRef AS compareRef,
		        node.summary AS summary, node.text AS text,
		        score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": vec, "k": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("searchJSTPassages: %w", err)
	}
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		book, _ := rec.Get("book")
		chapter, _ := rec.Get("chapter")
		comprises, _ := rec.Get("comprises")
		compareRef, _ := rec.Get("compareRef")
		summary, _ := rec.Get("summary")
		text, _ := rec.Get("text")
		score, _ := rec.Get("score")
		out = append(out, SearchResult{
			EntityType: EntityJSTPassage,
			ID:         asString(id),
			Text:       asString(text),
			Distance:   asFloat(score),
			Metadata: ResultMeta{
				Book:       asString(book),
				Chapter:    asString(chapter),
				Comprises:  asString(comprises),
				CompareRef: asString(compareRef),
				Summary:    asString(summary),
			},
		})
	}
	return out, nil
}

// --- Helpers ---

func float32sToAnySlice(v []float32) []interface{} {
	out := make([]interface{}, len(v))
	for i, x := range v {
		// Pass as float64 — falkordb-go's ToString serializes via
		// strconv.FormatFloat('f', -1, 64) which correctly preserves decimal
		// points for non-integer values. See Phase-B writeup in search.md.
		out[i] = float64(x)
	}
	return out
}

// asString / asInt / asFloat pull typed values out of FalkorDB's
// map[string]any Records with forgiving fallbacks (the driver returns nil
// for missing props).
func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func asInt(v any) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}
func asFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	}
	return 0
}
