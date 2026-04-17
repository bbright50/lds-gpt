# lds-gpt — architectural reference

A personal Retrieval-Augmented Generation backend for Latter-day Saint scripture. Scrapes the official Church website, persists the normalised graph into FalkorDB, generates 1024-dim Bedrock Titan embeddings, and answers queries with a three-stage kNN + graph-expansion + re-rank pipeline. The frontend is a Vite/React UI that currently uses a mock API because the Go HTTP server hasn't been built yet.

## 1. Stack at a glance

| Layer | Technology | Where |
|---|---|---|
| Language (backend) | Go 1.25 | root module `lds-gpt` |
| Language (frontend) | TypeScript 5.9, React 19 | [frontend/](frontend/) |
| Graph store | FalkorDB (Redis module, Cypher, vector-index) | `docker run falkordb/falkordb:latest` |
| Typed client | [go-ormql](../go-ormql) (local fork via `replace` in [go.mod](go.mod)) | [internal/falkor/generated/](internal/falkor/generated/) |
| Raw client | `github.com/FalkorDB/falkordb-go/v2` | used for `Graph.Delete`, introspection |
| Embeddings | AWS Bedrock, `amazon.titan-embed-text-v2:0` (1024-dim) | [internal/bedrockembedding/](internal/bedrockembedding/) |
| Scraping | `github.com/PuerkitoBio/goquery` | [pkg/scraper/](pkg/scraper/) |
| Concurrency | `golang.org/x/sync/errgroup` | both Phase 6 embed workers and Stage-1 kNN |
| Tests | `testcontainers/testcontainers-go` (Docker-backed FalkorDB) | [internal/falkor/testing_helpers.go](internal/falkor/testing_helpers.go) |

## 2. End-to-end data flow

Three independent lifecycle arcs:

**(a) Scrape → JSON on disk.** Each `task scrape:*` spawns a scraper CLI that fetches HTML from churchofjesuschrist.org (caching raw HTML under [pkg/data/raw/](pkg/data/raw/)) and writes normalised JSON to [pkg/data/](pkg/data/). This runs rarely; output is treated as a data set.

**(b) Load → FalkorDB graph.** `task load` (or `task load-and-embed`) runs the 6-phase ETL in [internal/dataloader/](internal/dataloader/), reading the JSON tree and populating the graph via typed `createXxx` / `updateXxx` GraphQL mutations through the go-ormql client, with a few raw-Cypher edge-writes for pure node-to-node relationships.

**(c) Search → ranked hits.** `go run ./cmd/app` embeds a query with Bedrock, runs six parallel `xxxSimilar` kNN queries (Stage 1), expands each hit to 1-hop neighbours (Stage 2), and re-ranks (Stage 3). The frontend talks to a future HTTP server wrapping this same pipeline; today it uses a mock.

---

## 3. Scrapers

### Shared library ([pkg/scraper/](pkg/scraper/))

- [scraper.go:24-26](pkg/scraper/scraper.go) — shared HTTP client, 30s timeout.
- [scraper.go:58-104](pkg/scraper/scraper.go) — `fetchDocument(url, cacheDir)`: disk-cached fetcher keyed off `/study/scriptures/<path>.html`. Reads from cache if present, else GETs + writes. No in-memory caching; no rate limiting (CLI layer adds 50ms between requests).
- [scraper.go:287-304](pkg/scraper/scraper.go) — `WriteJSON(path, value)`: pretty-prints JSON without HTML escaping, ensures parent dirs.
- [types.go:1-26](pkg/scraper/types.go) — `Chapter`, `Verse`, `Footnote` DTOs (mirror [internal/dataloader/types.go](internal/dataloader/types.go)).
- [topical.go](pkg/scraper/topical.go) — Topical Guide + Triple Combination Index logic (shared; distinguishes via reference-type prefix).
- [bible_dict.go](pkg/scraper/bible_dict.go) — BD entry parsing (text + references array).
- [jst.go](pkg/scraper/jst.go) — JST chapter parsing (nested: chapter → entries → verses).
- [triple_index.go](pkg/scraper/triple_index.go) — Triple-index wrapper over topical.

### Per-scraper CLIs ([cmd/scrapers/](cmd/scrapers/))

| Scraper | Source | URL pattern | Output | Count |
|---|---|---|---|---|
| `scriptures` | [cmd/scrapers/scriptures/main.go](cmd/scrapers/scriptures/main.go) | `…/{volume}/{book}/{chapter}` per arg | `pkg/data/scriptures/<vol>/<slug>/<ch>.json` | 1584 chapters over 5 volumes |
| `tg` | [cmd/scrapers/tg/main.go](cmd/scrapers/tg/main.go) | Index at `/tg`, then each entry page | `pkg/data/topical-guide.json` (single file, `map[name][]TGEntryJSON`) | ~3514 |
| `bd` | [cmd/scrapers/bd/main.go](cmd/scrapers/bd/main.go) | Index at `/bd` | `pkg/data/bible-dictionary.json` (`map[name]BDEntryJSON`) | ~1276 |
| `tc` | [cmd/scrapers/tc/main.go](cmd/scrapers/tc/main.go) | Index at `/triple-index` | `pkg/data/triple-combination-index.json` | ~3059 |
| `jst` | [cmd/scrapers/jst/main.go](cmd/scrapers/jst/main.go) | Index at `/jst`, chapter-level pages only | `pkg/data/jst.json` (`[]JSTChapterJSON`) | ~100 |
| `inspect` | [cmd/scrapers/inspect/main.go](cmd/scrapers/inspect/main.go) | Genesis 1 only | stdout dump of first 3 footnotes | debug tool |

### Scripture parsing detail ([scraper.go](pkg/scraper/scraper.go))

- Book: `h1 span.dominant` or `h1#title1` (lines 132-144)
- Chapter number: digits from `p.title-number` (lines 146-159)
- Summary: `p.study-summary` text (lines 161-163)
- Verses: `p.verse` elements — `span.verse-number` + cleaned text + footnote markers from `a.study-note-ref sup.marker[data-value]` (lines 165-197)
- Footnotes: `footer.study-notes li[data-full-marker]` with `span[data-note-category]` (lines 240-285)

### Footnote category taxonomy

Seven categories come off `data-note-category`, each carrying distinct meaning:
- **cross-ref** — scripture-to-scripture reference
- **tg** — Topical Guide reference
- **bd** — Bible Dictionary reference
- **jst** — Joseph Smith Translation reference
- **trn** — translation note (inline JSON on the verse in the graph)
- **or** — alternate reading (inline JSON)
- **ie** — idiom explanation (inline JSON)

The first four become graph edges (Phase 3). The last three stay on the verse as JSON-string properties (Phase 1).

---

## 4. FalkorDB schema & client setup

### Schema ([internal/falkor/schema.graphql](internal/falkor/schema.graphql))

Nine `@node` types, seven `@relationshipProperties` types, six `@vector` fields:

**Structural**: `Volume`, `Book`, `Chapter`, `Verse`, `VerseGroup`.
**Study helps**: `TopicalGuideEntry`, `BibleDictEntry`, `IndexEntry`, `JSTPassage`.
**Relationship properties** (edge metadata): `VerseCrossRefProps`, `VerseTGRefProps`, `VerseBDRefProps`, `VerseJSTRefProps`, `TGVerseRefProps`, `BDVerseRefProps`, `IDXVerseRefProps`.
**Vector fields** (6): `VerseGroup.embedding`, `Chapter.summaryEmbedding`, `TopicalGuideEntry.embedding`, `BibleDictEntry.embedding`, `IndexEntry.embedding`, `JSTPassage.embedding` — all `[Float!]!` with dimension 1024, similarity cosine.

Every `@node` has `id: ID!` as the canonical lookup key. The fork-patched generator makes `id` optional in `CreateInput` (we supply deterministic strings like `"v/ot/gen/1/5"`, and the translator falls back to `randomUUID()` when omitted).

### Generated client ([internal/falkor/generated/](internal/falkor/generated/))

5.6k LOC generated by `gormql generate` (run via `task generate`, which builds gormql from the local fork path):
- `client_gen.go` (14 LOC) — `NewClient(drv, opts)` constructor.
- `graphmodel_gen.go` (3013 LOC) — `GraphModel` (node + relationship definitions) and `AugmentedSchemaSDL` (full resolved GraphQL SDL).
- `models_gen.go` (2440 LOC) — Go structs for nodes, inputs, where-filters, connections, mutation-response envelopes.
- `indexes_gen.go` (50 LOC) — `VectorIndexes` map and `CreateIndexes(ctx, drv)` emitting all six vector-index DDL statements.

### Client struct ([internal/falkor/client.go](internal/falkor/client.go))

Wraps two handles on one FalkorDB connection:
- `db *falkordb.FalkorDB` — underlying redis-backed conn.
- `graph *falkordb.Graph` — reached via `Client.Raw()`, used for `GRAPH.DELETE`, raw-Cypher introspection.
- `drv ormqldriver.Driver` — go-ormql driver built with `VectorIndexes: generated.VectorIndexes`.
- `gclient *ormql.Client` — the typed client, reached via `Client.GraphQL()`, constructed with `WithBatchSize(500)` so bulk mutations auto-chunk.

**Lifecycle**:
- `NewClient(Config{URL, GraphName, Logger})` — parses the URL into host/port for the go-ormql driver, opens both handles.
- `Migrate(ctx)` — delegates to `generated.CreateIndexes(ctx, drv)` which issues six idempotent `CREATE VECTOR INDEX` statements (`already indexed`-tolerant).
- `Close()` — closes driver then redis conn, 5-second timeout each.

---

## 5. Data loader ([internal/dataloader/](internal/dataloader/))

### Loader orchestrator ([loader.go](internal/dataloader/loader.go))

```go
type Loader struct {
    fc          *falkor.Client
    dataDir     string
    stats       LoadStats
    abbrevs     map[string]BookInfo
    slugMap     map[string]string
    bookNames   map[string]BookInfo
    refParser   RefParser
    logger      *slog.Logger
    embedClient bedrockembedding.Client
}
```

`Run(ctx)` sequences the six phases; `EmbedOnly(ctx)` is the Phase-6-only path. The options pattern (`WithEmbedClient(...)`) attaches Bedrock; without it, Phase 6 is skipped.

### Storage-agnostic helpers (shared across phases)

- [types.go](internal/dataloader/types.go) — `ChapterJSON` / `VerseJSON` / `FootnoteJSON` / `TGEntryJSON` / `BDEntryJSON` / `IDXEntryJSON` / `JSTChapterJSON` input shapes; `VerseIndex` (composite-key → node id); `JSTIndex` (book+chapter → passage ids); `LoadStats` (per-phase counters + mutex-guarded warnings).
- [abbreviations.go](internal/dataloader/abbreviations.go) — static maps: abbreviation → `BookInfo{Volume, Slug}`, slug → abbreviation, book name → `BookInfo`, ordered `volumeAbbreviations` list, `volumeDisplayNames`.
- [refparser.go](internal/dataloader/refparser.go) — parses strings like `"Gen. 1:1-3; Mosiah 4:2"` into `[]ParsedRef{Volume, Slug, Chapter, Verse, EndVerse}`.

Deterministic ID helpers (on [loader_scriptures.go](internal/dataloader/loader_scriptures.go) and [loader_studyhelps.go](internal/dataloader/loader_studyhelps.go)):
`volumeNodeID`, `bookNodeID`, `chapterNodeID`, `VerseNodeID`, `verseGroupNodeID`, `tgNodeID`, `bdNodeID`, `idxNodeID`, `jstNodeID`.

### Phase 1 — scriptures ([loader_scriptures.go](internal/dataloader/loader_scriptures.go))

**Order of operations** inside `loadScriptures(ctx)`:

1. Walk `pkg/data/scriptures/<vol>/<slug>/*.json` once, accumulating four row slices: `volRows`, `bookRows`, `chapterRows`, `verseRows`. Populate `VerseIndex` along the way.
2. Four typed bulk mutations in dependency order:
   - `createVolumes` — 5 nodes, no relationships.
   - `createBooks` — each row carries `volume: {connect: [{where: {id: "vol/<abbr>"}}]}`. The fork-patched translator emits a `FOREACH (conn IN coalesce(item.volume.connect, []) | MERGE (target:Volume {id: conn.where.id}) MERGE (target)-[:CONTAINS]->(n))` block — node + edge land in the same round-trip.
   - `createChapters` — same pattern against `book.connect`. Also writes a `placeholderEmbedding()` (1024 × 0.0001) to the required `summaryEmbedding` field.
   - `createVerses` — chapter connect + inline footnote JSON (`translationNotes` / `alternateReadings` / `explanatoryNotes` serialized from `trn` / `or` / `ie` footnotes via `extractInlineFootnotes`).
3. Auto-chunker inside the go-ormql client batches each call at 500 rows.

### Phase 2 — study helps ([loader_studyhelps.go](internal/dataloader/loader_studyhelps.go))

Each of TG / BD / IDX / JST runs its own loader method, all going through typed `createXxx` with `placeholderEmbedding()`:
- `loadTopicalGuide` reads `topical-guide.json` (`map[name][]TGEntryJSON`), emits `createTopicalGuideEntries`.
- `loadBibleDictionary` reads `bible-dictionary.json`, emits `createBibleDictEntries` with `text` too.
- `loadTripleCombIndex` reads `triple-combination-index.json`, emits `createIndexEntries`.
- `loadJST` reads `jst.json`, composes full JST text by joining verse texts, derives a `comprises` range when absent, emits `createJSTPassages`.

Returns four maps used by subsequent phases: `tgMap` / `bdMap` / `idxMap` (name → deterministic id) and `JSTIndex` (book+chapter → passage node ids).

`readOptionalJSON` tolerates missing files (logs a warning, returns empty) so the loader is usable against partial scrapes.

### Phase 3 — footnote edges ([loader_footnotes.go](internal/dataloader/loader_footnotes.go))

Re-walks the scripture JSON tree, parses footnote text, creates four edge kinds:
- `(Verse)-[:CROSS_REF {category, footnoteMarker, referenceText}]->(Verse)`
- `(Verse)-[:TG_FOOTNOTE {footnoteMarker}]->(TopicalGuideEntry)`
- `(Verse)-[:BD_FOOTNOTE {footnoteMarker}]->(BibleDictEntry)`
- `(Verse)-[:JST_FOOTNOTE {footnoteMarker}]->(JSTPassage)`

Parsing is pure-functional (`parseFootnoteKey`, `extractCrossRefPortion`, `extractTGTopics`, `extractBDEntries`, `extractJSTReference`, `findRefEnd`, `trimAtNextPrefix`). Per-(source,target) dedup via `map[[2]string]bool` seen-sets.

Writes use **raw Cypher** `UNWIND $rows AS r MATCH (src) MATCH (tgt) CREATE (src)-[:EDGE {...}]->(tgt)` in 500-row batches, because edges between pre-existing nodes don't benefit from the typed-create path (see §11).

### Phase 4 — study-ref edges ([loader_studyrefs.go](internal/dataloader/loader_studyrefs.go))

Nine edge kinds from TG/BD/IDX/JST into the verse graph:
- `(TG)-[:TG_VERSE_REF {phrase}]->(Verse)`
- `(TG)-[:TG_SEE_ALSO]->(TG)` (self)
- `(TG)-[:TG_BD_REF]->(BD)`
- `(BD)-[:BD_VERSE_REF]->(Verse)`
- `(IDX)-[:IDX_VERSE_REF {phrase}]->(Verse)`
- `(IDX)-[:IDX_SEE_ALSO]->(IDX)` (self)
- `(IDX)-[:IDX_TG_REF]->(TG)`, `(IDX)-[:IDX_BD_REF]->(BD)`
- `(JST)-[:COMPARES]->(Verse)` — unrolled from `compare` ranges

`writeSimplePairs` factors out the UNWIND-MATCH-MATCH-CREATE Cypher; typed `writeTGVerseRefs` / `writeBDVerseRefs` / `writeIDXVerseRefs` handle the edges that carry phrase props.

### Phase 5 — verse groups ([loader_versegroups.go](internal/dataloader/loader_versegroups.go))

Sliding-window retrieval units. `groupWindowSize = 5`, `groupStepSize = 3`.

Per chapter, sort verses by number, step through the window emitting `VerseGroup` rows with `text` (space-joined verse texts), `startVerseNumber`, `endVerseNumber`, `embedding: placeholderEmbedding()`, and — crucially — **both** `chapter: {connect: [...]}` AND `verses: {connect: [{where: {id}}, …]}` payloads. Single `createVerseGroups` typed mutation carries nodes + both edge kinds.

### Phase 6 — embeddings ([loader_embeddings.go](internal/dataloader/loader_embeddings.go))

Runs six per-label embedders sequentially; each internally fans out with `errgroup.SetLimit(embedConcurrency)` (= 8):

1. `embedVerseGroups` — `WHERE g.embedding IS NULL AND g.text <> ''`, text is `g.text`.
2. `embedChapters` — `WHERE c.summaryEmbedding IS NULL AND c.summary <> ''`, text is `c.summary`.
3. `embedTGEntries` — `embedNameWithPhrases` composing `"{name}. {phrase1}; {phrase2}; …"` (max 20 phrases) from TG_VERSE_REF edges.
4. `embedBDEntries` — text is `b.text`.
5. `embedIDXEntries` — same `embedNameWithPhrases` shape against IDX_VERSE_REF.
6. `embedJSTPassages` — text is `CASE WHEN j.summary … THEN j.text ELSE j.summary + ' ' + j.text END`.

Per row: `truncate(text, 25000)` → `embedClient.EmbedText(ctx, text) → []float64` → typed `writeEmbedding(label, id, prop, embedding)`. The latter calls `updateXxx(where: {id: $id}, update: {<prop>: $vec})` and the fork-patched translator wraps the value in `vecf32(...)` at emit. Index becomes live as each row upgrades.

Predicate `IS NULL` makes the phase idempotent — a re-run only backfills missing rows.

---

## 6. Search pipeline ([internal/falkor/](internal/falkor/))

### Entry point ([rag_search.go](internal/falkor/rag_search.go))

```go
func (c *Client) DoContextualSearch(ctx context.Context, embedding []float32, opts ...ContextSearchOption) ([]SearchResult, error)
```

Input: pre-computed `[]float32` embedding. Output: up to `kNN` (default 20) `SearchResult`s sorted by `rankScore` ascending.

### Stage 1 — parallel kNN

`runParallelSearches` spawns six `errgroup`-coordinated goroutines, each calling one typed `xxxSimilar` generated query:

| Label | Query | Fields selected |
|---|---|---|
| VerseGroup | `verseGroupsSimilar(vector, first)` | id, text, startVerseNumber, endVerseNumber |
| Chapter | `chaptersSimilar(vector, first)` | id, number, summary, url |
| TopicalGuideEntry | `topicalGuideEntriesSimilar(vector, first)` | id, name |
| BibleDictEntry | `bibleDictEntriesSimilar(vector, first)` | id, name, text |
| IndexEntry | `indexEntriesSimilar(vector, first)` | id, name |
| JSTPassage | `jSTPassagesSimilar(vector, first)` | id, book, chapter, comprises, compareRef, summary, text |

Each compiles (via go-ormql translator + fork-patched driver) to `CALL db.idx.vector.queryNodes('<Label>', '<prop>', $k, vecf32($vec)) YIELD node, score RETURN …`. `first = defaultSearchLimit = 10`, so Stage 1 produces up to 60 seeds total. The vector param goes through `float32sToAnySlice` which emits `[]interface{}` of `float64` (required by falkordb-go's ToString).

### Stage 2 — graph expansion ([graph_traversal.go](internal/falkor/graph_traversal.go))

Per seed, `(*Client).traverseEdges` dispatches on `EntityType` to one of six handlers, each issuing one typed GraphQL query with nested `@relationship` selections:

- `verse_group` → `versesConnection` → `[INCLUDES]→Verse`
- `chapter` → `versesConnection` → `[HAS_VERSE]→Verse`
- `topical_guide` → `verseRefsConnection` + `seeAlsoConnection` + `bdRefsConnection`
- `bible_dict` → `verseRefsConnection` + `seeAlsoConnection`
- `index` → `verseRefsConnection` + `seeAlsoConnection` + `tgRefsConnection` + `bdRefsConnection`
- `jst_passage` → `compareVersesConnection`
- `verse` → leaf (no outbound traversal) — prevents 2-hop runaway.

Each neighbour gets `SearchResult.Distance = seed.Distance + defaultHopPenalty(0.05)` via `assignSyntheticDistances`. `deduplicateResults(stage1, graph)` drops any graph hit whose `(EntityType, ID)` pair already appeared in Stage 1, keeping the lowest-distance copy among graph-side dupes.

### Stage 3 — re-rank ([ranking.go](internal/falkor/ranking.go))

`rankResults(combined)` computes `rankScore = distance - typeBonus`, where `typeBonus = defaultVerseBonus (0.05)` iff `EntityType == EntityVerse`, else 0. `sort.SliceStable` orders ascending. Final slice is trimmed to `kNN`.

### Result shape ([search_result.go](internal/falkor/search_result.go))

`SearchResult{EntityType, ID, Name, Text, Distance, Metadata: ResultMeta{…}}`. `ResultMeta` is a flat struct holding entity-specific fields (`StartVerseNumber`, `ChapterNumber`, `URL`, `Book`, `Chapter`, `Comprises`, `CompareRef`, `Summary`, `VerseNumber`, `Reference`); unused fields are zero and callers key off `EntityType`.

---

## 7. CLI wiring

### Dataloader ([cmd/dataloader/app.go](cmd/dataloader/app.go))

Flags: `-embed` (all 6 phases), `-embed-only` (Phase 6 only).
- Loads config (viper — `ENV`, `AWS_REGION`, `FALKORDB_URL`, `FALKORDB_GRAPH`, `DATA_DIR`).
- Opens `falkor.Client`.
- On a full load (not embed-only): `client.Raw().Delete()` to clear the graph (first-run warning tolerated), then `client.Migrate(ctx)` to create vector indexes.
- Constructs `dataloader.New(client, dataDir, logger, [WithEmbedClient(...)])`.
- Runs `Run(ctx)` or `EmbedOnly(ctx)`.

### Search demo ([cmd/app/app.go](cmd/app/app.go))

Hardcoded `"What is faith?"` query with `WithKNN(10)`. Wires `app.NewApp(client, embedClient)`, which embeds the query once and calls `client.DoContextualSearch`. Prints `(<distance>) <type> [<id>]: <text>` per result.

### App wrapper ([internal/app/app.go](internal/app/app.go))

Thin seam between the CLI (and future HTTP handlers) and the search client. `DoContextualSearch(ctx, query string, opts...)`: `embedClient.EmbedText(ctx, query) → []float64 → []float32 cast → client.DoContextualSearch`.

---

## 8. Frontend ([frontend/](frontend/))

### Stack

Vite 7.3.1 · React 19.2 · Tailwind 4.2 (via `@tailwindcss/vite`) · Vitest 4.0 · TypeScript 5.9 · ESLint 9.39 (flat config).
Dev: `npm run dev` (Vite @ :5173). Build: `tsc -b && vite build` → `dist/`.

### Component tree

```
<App api={SearchAPI}>
  <SearchBar onSearch loading />
  <SearchStatus loading error hasResults hasSearched />
  <ResultList results>
    <ResultCard result>
      <EntityIcon entityType />
      <DistanceBadge distance />
      <MetadataPanel metadata /> (toggled)
```

### Data flow

[App.tsx:18-49](frontend/src/App.tsx) — `handleSearch` wraps `api.search(...)` in an `AbortController` (cancels previous in-flight queries), updates `{loading, error, results, hasSearched}` state. No global store — `useState` only.

[SearchBar.tsx](frontend/src/components/SearchBar.tsx) — form with one input + button; no debounce, submit on Enter. Disabled when trimmed query is empty or loading is true.

[ResultCard.tsx](frontend/src/components/ResultCard.tsx) — header (icon + capitalised entityType + DistanceBadge), name, truncated text (300 chars), optional "Source" link if `metadata.url` exists, collapsible metadata panel.

[EntityIcon.tsx](frontend/src/components/EntityIcon.tsx) — Unicode emoji per entity type (📖 chapter, 📜 verse, 🏷️ topical_guide, 📓 bible_dict, 📋 index, ✏️ jst_passage, ≡ verse_group).

[DistanceBadge.tsx](frontend/src/components/DistanceBadge.tsx) — `(1 - clamp(distance, 0, 1)) * 100` as relevance percentage.

### API contract

[client.ts](frontend/src/api/client.ts) — `POST {baseUrl}/api/search` with body `{query, knn?}`; response `{results: SearchResult[]}`.

[main.tsx:9-13](frontend/src/main.tsx) — if `VITE_API_BASE_URL` is set, uses `client.ts`; otherwise swaps in [mockSearch.ts](frontend/src/api/mockSearch.ts) which returns canned fixtures after a 150ms delay. This is the fallback for the deferred HTTP server.

`SearchResult` in [types/search.ts:38-46](frontend/src/types/search.ts) matches the Go struct field-for-field (`entityType`, `id`, `name`, `text`, `distance`, `metadata`).

### Tests (8 files)

- [App.test.tsx](frontend/src/App.test.tsx) — 7 tests: search flow, loading, errors, cancellation.
- [SearchBar.test.tsx](frontend/src/components/SearchBar.test.tsx) — 6 tests: input validation, button state, submission.
- [ResultCard.test.tsx](frontend/src/components/ResultCard.test.tsx) — 6 tests: icon, type label, truncation, metadata toggle.
- Per-component tests for EntityIcon, DistanceBadge, MetadataPanel, ResultList, SearchStatus.
- [client.test.ts](frontend/src/api/client.test.ts) — 9 tests: fetch shape, POST path, request/response, abort signal.

Fixtures in [src/test/fixtures.ts](frontend/src/test/fixtures.ts) cover all 7 entity types so mock rendering is exercised end-to-end.

---

## 9. Testing infrastructure

### Backend

- **Integration tests** spin `falkordb/falkordb:latest` per test via [internal/falkor/testing_helpers.go](internal/falkor/testing_helpers.go). ~500ms per container boot; every test gets an isolated graph. `FALKOR_DEBUG=1` env var routes a slog Logger to the driver and prints every Cypher statement.
- **Unit tests** (no container): [ranking_test.go](internal/falkor/ranking_test.go) — pure `rankResults` / `deduplicate` / `assignSyntheticDistances`; [refparser_test.go](internal/dataloader/refparser_test.go); [loader_embeddings_test.go](internal/dataloader/loader_embeddings_test.go) `TestComposeTextFromPhrases`.
- **Bedrock mocking**: `go.uber.org/mock` generates [internal/bedrockembedding/mocks/](internal/bedrockembedding/mocks/). Phase 6 tests use `MockClient.EXPECT().EmbedText(...)` to avoid real AWS calls.
- **Regression pins**: [TestNestedConnect_CreatesStubForMissingTarget](internal/falkor/schema_test.go) documents the FOREACH+MERGE stub-creation caveat; [TestSpike_VectorIndexOnPlainList_VsVecf32](internal/falkor/vector_index_spike_test.go) pins FalkorDB's strict vector-index behavior that lets us use plain-list placeholders safely.

### Frontend

Vitest + jsdom; test files live next to their source (`*.test.tsx`). Mock API is exercised in integration tests via `App.test.tsx`.

---

## 10. The go-ormql fork ([../go-ormql](../go-ormql))

Registered via `replace github.com/tab58/go-ormql => ../go-ormql` in [go.mod](go.mod). Patches:

| Area | File | Fix |
|---|---|---|
| Vector query rewrite | `pkg/driver/falkordb/falkordb.go:295` | `$rw3` → `vecf32($rw3)` in the rewritten `db.idx.vector.queryNodes` call |
| Runtime pluralization | `pkg/translate/query.go:218-234` | `findNodeByPluralName` uses `strutil.PluralLower` (go-pluralize) instead of naive +s |
| Update vecf32 wrap | `pkg/translate/mutation.go:192-205` | update-path emits `SET n.<prop> = vecf32(<param>)` when the field is `@vector` |
| Chunker scalar-list guard | `pkg/client/chunk.go:26-60` | `isRecordList` skips chunking on vector params so 1024-dim embeddings stay whole |
| Generator id passthrough | `pkg/codegen/augment.go:243` + `pkg/codegen/models.go:208` | `id: ID` emitted as optional in CreateInput; translator uses `coalesce(item.id, randomUUID())` |
| Nested connect via variables | `pkg/translate/mutation_nested.go:11-75` | schema-driven FOREACH+MERGE template replaces the AST-only walker for `connect`; runs in every createXxx, no-op when the caller omits the field |

The `task generate` target in [Taskfile.yml](Taskfile.yml) builds gormql from the local fork before regenerating.

---

## 11. Known limitations (worth remembering)

1. **FalkorDB CALL subqueries don't iterate with outer UNWIND** — they see the union of incoming rows and fire once. The fork patch uses FOREACH for per-row writes; future typed mutations that need per-row behavior inside a CALL will need the same pattern.
2. **FOREACH can't MATCH, so our nested-connect template uses MERGE** — if a caller passes a `connect` with an id that doesn't exist, MERGE materialises a bare stub node. Safe for our loader (parents always pre-exist) but unsafe for any future user-facing write endpoint that exposes `connect` directly. Pinned by [TestNestedConnect_CreatesStubForMissingTarget](internal/falkor/schema_test.go).
3. **FalkorDB transactions are single-statement** — no multi-statement atomicity, so a mid-phase-1 failure leaves partial state. Mitigated by the destructive `GRAPH.DELETE` at the start of every full load.
4. **The HTTP server doesn't exist yet** — frontend talks to a mock (`VITE_API_BASE_URL` unset). Building the HTTP wrapper is a small job: one `POST /api/search` handler that unmarshals `{query, knn}`, calls `app.DoContextualSearch`, marshals `{results: [...]}`.

---

## 12. Relationship-property fields — status and gaps

Precise accounting of what each `@relationshipProperties` type declares and what the loader actually writes. Useful when adding display-tier features.

### All relationship-property fields currently persisted

| Field | Written by | Notes |
|---|---|---|
| `VerseCrossRefProps.category` | [loader_footnotes.go](internal/dataloader/loader_footnotes.go) (`writeCrossRefs`) | Always set to `'cross-ref'` |
| `VerseCrossRefProps.footnoteMarker` | `writeCrossRefs` | e.g. `"1a"` |
| `VerseCrossRefProps.referenceText` | `writeCrossRefs` | Human-readable reference string from source HTML |
| `VerseCrossRefProps.targetEndVerseId` | `writeCrossRefs` | End verse of a range (e.g. `"Gen 1:5-7"`); empty when the ref is a single verse |
| `VerseTGRefProps.footnoteMarker` | `writeTGFootnotes` | |
| `VerseTGRefProps.tgTopicText` | `writeTGFootnotes` | Topic name as extracted from the footnote text (what `tgMap` was keyed by) |
| `VerseTGRefProps.referenceText` | `writeTGFootnotes` | Reference text from the footnote payload |
| `VerseBDRefProps.footnoteMarker` | `writeBDFootnotes` | |
| `VerseJSTRefProps.footnoteMarker` | `writeJSTFootnotes` | |
| `TGVerseRefProps.phrase` | [loader_studyrefs.go](internal/dataloader/loader_studyrefs.go) (`writeTGVerseRefs`) | From TG entry's `phrase` JSON field |
| `BDVerseRefProps.targetEndVerseId` | `writeBDVerseRefs` | End verse of a range ref in `bible-dictionary.json` |
| `IDXVerseRefProps.phrase` | `writeIDXVerseRefs` | From IDX entry's `phrase` |

None of these fields are load-bearing for retrieval (Stage 1 kNN + Stage 2 traversal work without them). They exist so a future UI can render verse-range footnotes, TG topic text, and original reference strings alongside the raw footnote marker — no additional schema/regen work needed.
