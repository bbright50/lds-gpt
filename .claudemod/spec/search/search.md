# Specification: Contextual Search

## 1. Purpose

The retrieval half of the RAG system. Given a pre-computed query embedding, returns a ranked list of the most relevant nodes across the seven entity types — blending dense vector similarity with 1-hop graph expansion so that a hit on a TG topic or JST passage can pull in the verses it references. This is the only domain that combines the generated go-ormql `@vector` similarity queries with graph-relationship traversal, and the blending logic lives entirely here.

## 2. Key Components

- `internal/falkor/rag_search.go` — `DoContextualSearch` entry point; six per-label `search*` functions issue parallel `CALL db.idx.vector.queryNodes(..., vecf32(...))` Cypher statements through the raw `*falkordb.Graph` handle. Stage 1 deliberately bypasses the go-ormql generated client because go-ormql's FalkorDB driver has a bug — its `@vector` rewrite emits `CALL db.idx.vector.queryNodes($rw0, $rw1, $rw2, $rw3)` without wrapping `$rw3` in `vecf32(...)`, and FalkorDB ≥ 4.18 rejects the unwrapped list. Verified in Phase B by `TestVectorSimilarity_ViaRawClient`; the skipped test `TestGeneratedClient_VectorQuery_KnownBug` pins the regression.
- `internal/falkor/graph_traversal.go` — `traverseEdges` dispatch + six per-entity traversal functions (`traverseVerseGroup`, `traverseChapter`, `traverseTopicalGuide`, `traverseBibleDict`, `traverseIndex`, `traverseJSTPassage`) and a `versesToResults` helper. Each traversal runs one raw Cypher `MATCH` against `Client.Raw()`. The Phase B smoke-test (`TestGraphTraversal_ViaGeneratedClient`) proved go-ormql's `Execute` path works for this shape, but once Stage 1 had to go raw for the `vecf32(...)` bug, running Stage 2 through go-ormql reduced the typed client's value to `Migrate()` alone while costing us six nested-decode structs. Raw is simpler and consistent with Stage 1. Whether to keep go-ormql at all is revisited in Phase F.
- `internal/falkor/ranking.go` — `assignSyntheticDistances`, `deduplicateResults`, `rankResults`. Pure functions; easy to unit-test and reused unchanged from the prior LibSQL implementation.
- `internal/falkor/search_result.go` — `SearchResult`, `ResultMeta`, `EntityType` constants, and the ranking tuning constants: `defaultHopPenalty = 0.05`, `defaultVerseBonus = 0.05`, `defaultGraphLimit = 5`. `SortByDistance` helper.

## 3. Data Models

- **`EntityType`** — String enum: `verse_group`, `chapter`, `topical_guide`, `bible_dict`, `index`, `jst_passage`, `verse`. Matches the frontend's `EntityType` union.
- **`SearchResult`** — `{EntityType, ID, Name, Text, Distance, Metadata}`. `Distance` is the cosine distance returned directly by FalkorDB's `db.idx.vector.queryNodes` when the index is created with `similarityFunction: 'cosine'` (0.0 for an identical vector, 1.0 for orthogonal — verified in Phase A against the `score` column), or a synthetic seed-distance + hop-penalty from Stage 2.
- **`ResultMeta`** — Flat struct with all entity-type-specific fields (`StartVerseNumber`, `ChapterNumber`, `URL`, `Book`, `Chapter`, `Comprises`, `CompareRef`, `Summary`, `VerseNumber`, `Reference`). Unused fields remain zero; callers key off `EntityType`.
- **Tuning constants** — `defaultSearchLimit = 10` (per-label Stage 1 kNN), `defaultGraphLimit = 5` (per-edge Stage 2 limit), `defaultHopPenalty = 0.05`, `defaultVerseBonus = 0.05`. With 6 labels at 10 hits each, Stage 1 produces up to 60 results; Stage 2 can add up to ~30 more per seed.

## 4. Interfaces

- **`(*Client).DoContextualSearch(ctx, embedding []float32, opts...) ([]SearchResult, error)`** — The only exported entry point. `embedding` is a `[]float32` slice (no byte-packing required — FalkorDB's `vecf32` accepts floats directly). Options: `WithKNN(n)` (default 20). Empty embedding or non-positive `kNN` returns an error.
- **Three-stage pipeline**
  - *Stage 1* — `runParallelSearches` runs the six `searchXxx` functions concurrently via `errgroup`; each issues `CALL db.idx.vector.queryNodes('<Label>', '<embeddingProp>', 10, vecf32($q)) YIELD node, score RETURN ...` through the raw `*falkordb.Graph` handle (`Client.Raw()`). The `$q` parameter is passed as `[]interface{}` of `float64` values so falkordb-go's `BuildParamsHeader` serializes each element with a decimal point — bare `[]float64` is not a supported `ToString` input, and integer-like values are stringified without decimals, making FalkorDB type the list as `LIST<INTEGER>` and reject the call.
  - *Stage 2* — `graphExpandAndDedup` iterates Stage 1 seeds, calls `traverseEdges` (one GraphQL query per seed through `Client.GraphQL().Execute`; go-ormql translates nested `@relationship` selections into a single Cypher round-trip), applies `assignSyntheticDistances`, then `deduplicateResults` drops any graph hit whose `(EntityType, ID)` already appears in Stage 1 and keeps the lowest-distance copy among graph duplicates.
  - *Stage 3* — `rankResults` computes `rankScore = distance - verseBonus` (verse bonus only on `EntityVerse`) and `sort.SliceStable` by ascending score. Slice is trimmed to `kNN`.
- **Graph traversal edges** — Exhaustive list:
  - `verse_group` → `INCLUDES` → Verse
  - `chapter` → `HAS_VERSE` → Verse
  - `topical_guide` → `VERSE_REF` → Verse (with `TGVerseRefProps`), `SEE_ALSO` → TopicalGuideEntry, `BD_REF` → BibleDictEntry
  - `bible_dict` → `VERSE_REF` → Verse (with `BDVerseRefProps`), `SEE_ALSO` → BibleDictEntry
  - `index` → `VERSE_REF` → Verse (with `IDXVerseRefProps`), `SEE_ALSO` → IndexEntry, `TG_REF` → TopicalGuideEntry, `BD_REF` → BibleDictEntry
  - `jst_passage` → `COMPARES` → Verse
  - `verse` — leaf; returns no neighbors (prevents runaway 2-hop blow-up).

## 5. Dependencies

- **Depends on:** Database & Schema (`falkor.Client` providing `GraphQL()` for generated kNN + traversal queries, plus the six vector indexes created by `Migrate`).
- **Depended on by:** App CLI (`internal/app.App.DoContextualSearch` is a thin wrapper that embeds the query first).

## 6. Acceptance Criteria

- With a seeded test graph (via `testing_helpers.StartFalkorContainer`), `DoContextualSearch(WithKNN(n))` returns exactly `n` results when the graph has ≥ `n` embedded nodes, sorted by ascending `rankScore`.
- Verses reachable from a Stage 1 hit appear in the final list and carry a distance strictly greater than the seed's distance (by at least `defaultHopPenalty`).
- A result that appears in Stage 1 never re-appears as a Stage 2 duplicate.
- Among Stage 2 duplicates pointing at the same `(EntityType, ID)`, the one with the lowest synthetic distance wins.
- Verse results outrank non-verse results with the same raw distance because of `defaultVerseBonus = 0.05`.
- Stage 1 searches run in parallel: a failure in one label cancels the `errgroup` context and returns the underlying error wrapped with `"falkor: contextual search"`.
- `EntityVerse` seeds are no-ops for graph expansion (leaf nodes) — they pass through Stage 1 untouched.

## 7. Edge Cases

- **Zero-length embedding** — Returns `"falkor: embedding must not be empty"` without touching the graph.
- **Fewer embedded nodes than `defaultSearchLimit`** — Per-label queries return what is available; the final slice may contain fewer than `kNN` results.
- **Entirely empty graph** — Stage 1 returns `nil`, Stage 2 has nothing to expand, the function returns an empty slice (not an error).
- **Graph traversal over a seed whose relationships are empty** — Handled silently; the seed contributes nothing to Stage 2. Not an error.
- **Result with `Text == ""`** — Permitted (TG/IDX entries intentionally omit a `Text` field because their displayable content lives on `Name` and reachable verses). Frontend renders on `Name` when `Text` is empty.
- **The `kNN` option is larger than the combined pool** — All available results are returned; no padding.
- **Ranking ties** — `sort.SliceStable` preserves input order, so ties resolve by Stage 1 entity order (as returned by the errgroup channel) which is nondeterministic across runs.
- **FalkorDB `score` semantics under `similarityFunction: 'cosine'`** — The `score` column returned by `db.idx.vector.queryNodes` is cosine *distance*, not similarity (verified in Phase A: identical vectors yield 0, orthogonal yield 1). Downstream ranking math is therefore unchanged from the prior LibSQL `vector_distance_cos` implementation — no conversion needed.
