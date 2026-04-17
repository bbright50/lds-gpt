# Project: lds-gpt

## 1. Overview

A personal RAG (Retrieval-Augmented Generation) backend for Latter-day Saint scripture. Scrapes the official LDS scripture website, normalizes it into a **FalkorDB** property graph with 1024-dim vector embeddings, and performs contextual semantic search across verses, chapters, Topical Guide, Bible Dictionary, Triple Combination Index, and Joseph Smith Translation passages.

The system is retrieval-only today — a generative LLM step may be added later. A React frontend exists but the HTTP server that would back it is deferred.

## 2. Technology Stack

- **Language:** Go 1.24.5 (backend), TypeScript / React 19 (frontend)
- **Graph store:** FalkorDB (Redis module, OpenCypher, BSD) with native `vecf32` vector properties and `db.idx.vector.queryNodes` similarity search
- **Typed client:** [go-ormql](https://github.com/tab58/go-ormql) — a GraphQL-schema-driven code generator that compiles GraphQL queries to a single Cypher round-trip and ships with an official FalkorDB driver
- **Raw client:** `github.com/FalkorDB/falkordb-go` — used only for DDL that go-ormql does not cover (e.g. `CREATE VECTOR INDEX`) and administrative introspection
- **Embeddings:** AWS Bedrock Runtime — `amazon.titan-embed-text-v2:0` (1024-dim)
- **Scraping:** PuerkitoBio/goquery (HTML → structured JSON)
- **Config:** spf13/viper (env vars + `.env`)
- **Concurrency:** `golang.org/x/sync/errgroup`, `alitto/pond` (rate-limited worker pool)
- **Testing:** `github.com/testcontainers/testcontainers-go` — integration tests spin up a throwaway `falkordb/falkordb` container
- **Frontend:** Vite 7, React 19, Tailwind v4, Vitest 4
- **Task runner:** go-task/Taskfile.yml

## 3. Entry Points

- `cmd/dataloader/app.go` — ETL pipeline that populates the FalkorDB graph from scraped JSON (phases 1-5) and optionally generates embeddings (phase 6).
- `cmd/app/app.go` — CLI demo that performs a single hardcoded contextual search. (An HTTP server wrapping this is not yet implemented.)
- `cmd/scrapers/scriptures/main.go` — Scrape one or more scripture chapter URLs.
- `cmd/scrapers/tg/main.go` — Scrape the Topical Guide (~3514 entries).
- `cmd/scrapers/bd/main.go` — Scrape the Bible Dictionary (~1276 entries).
- `cmd/scrapers/tc/main.go` — Scrape the Triple Combination Index (~3059 entries).
- `cmd/scrapers/jst/main.go` — Scrape the Joseph Smith Translation appendix.
- `cmd/scrapers/inspect/main.go` — Debug helper to inspect HTML structure.
- `frontend/src/main.tsx` — Vite-driven React UI; calls `POST /api/search` when `VITE_API_BASE_URL` is set, otherwise uses an in-memory mock.

## 4. Directory Structure

```
cmd/
  app/                  — CLI demo for contextual search
  dataloader/           — Main ETL entry point (+ config)
  scrapers/             — Per-source scraper entry points
internal/
  app/                  — Application composition (CLI only today)
  bedrockembedding/     — AWS Bedrock Titan embedding client
  dataloader/           — Multi-phase ETL loader
  falkor/               — FalkorDB client, migrate, search, graph traversal, ranking
    schema.graphql      — go-ormql schema (source of truth for nodes + relationships)
    generated/          — go-ormql-generated client code (do not edit)
  utils/
    rate_limiter/       — Generic worker-pool wrapper around pond
pkg/
  scraper/              — Shared scraper library (scripture, TG, BD, TC, JST)
frontend/
  src/                  — React UI (SearchBar, ResultCard, etc.)
```

## 5. Data Flow

1. **Scrape** — Each scraper command fetches HTML from churchofjesuschrist.org (cached under `pkg/data/raw/`) and writes normalized JSON to `pkg/data/` (scriptures, `tg.json`, `bd.json`, `tc.json`, `jst.json`).
2. **Load (phases 1-5)** — `cmd/dataloader` clears the target FalkorDB graph, runs `Migrate` to create the six vector indexes, and populates the graph: volumes → books → chapters → verses, study-help nodes (TG/BD/IDX/JST), footnote edges (cross-refs, TG/BD/JST footnotes) as relationships with properties, study-help edges (see-also, verse refs), and sliding-window verse groups. All writes flow through go-ormql's generated `createXxx` mutations with auto-batching.
3. **Embed (phase 6)** — Each of the 6 embeddable node labels has its text composed, truncated (25k chars), and sent to Bedrock Titan with 8-way concurrency. Returned `[]float64` is converted to `[]float32` and persisted via Cypher `SET n.embedding = vecf32($vec)` (expressed as a go-ormql `updateXxx` mutation).
4. **Search** — A query is embedded once, then `DoContextualSearch`:
   a. runs 6 parallel generated `xxxSimilar(embedding, first)` GraphQL queries (10 results each, compiled to `db.idx.vector.queryNodes`),
   b. 1-hop expands each seed across its graph relationships (VerseGroup→INCLUDES→Verse, Chapter→HAS_VERSE→Verse, TG/BD/IDX→VERSE_REF→Verse and SEE_ALSO, JST→COMPARES→Verse) via one GraphQL query per seed,
   c. assigns synthetic distances to graph hits (`seedDistance + 0.05`),
   d. deduplicates graph hits against Stage 1,
   e. re-ranks by `distance − verseBonus` (verse bonus = 0.05),
   f. trims to the requested kNN (default 20).

## 6. Design Patterns

- **Retrieval-Augmented Generation (retrieval half only)** — Vector search grounded in a structured property graph.
- **GraphQL → single Cypher round-trip** — Every `client.Execute(ctx, gqlQuery, vars)` call compiles to exactly one Cypher statement, regardless of nesting depth.
- **Options pattern** — `falkor.WithKNN(...)`, `dataloader.WithEmbedClient(...)` — opt-in features without proliferating constructors.
- **Indexed in-memory lookup** — `VerseIndex` (~45k entries) and `JSTIndex` keep cross-reference resolution O(1) during the ETL so that relationship connects do not re-round-trip the DB.
- **Rate-limited worker pool** — `rate_limiter.Embeddable[T]` embeds a pond `ResultPool` into any client that needs bounded concurrency (Bedrock: max 20).
- **Phased ETL** — Each loader phase depends only on prior-phase outputs; phase 6 (embeddings) is idempotent and separately invokable via `--embed-only`.

## 7. External Integrations

- **AWS Bedrock Runtime** — Titan Text Embeddings V2 (1024-dim). Credentials via default AWS SDK chain; region from `AWS_REGION`.
- **churchofjesuschrist.org** — Source HTML for scriptures and study helps. Read-only; scraping is cached locally to avoid repeat fetches.
- **FalkorDB** — Runs as a Redis module. Local dev: `docker run -p 6379:6379 falkordb/falkordb:latest`. Requires FalkorDB 4.2+ for `@vector` support in go-ormql.

## 8. Build & Run

```bash
# One-time setup: scrape everything (takes a while, cached to pkg/data/raw/)
task scrape:all          # All standard works (OT, NT, BofM, D&C, PoGP — 1584 chapters)
task scrape:tg           # Topical Guide
task scrape:bd           # Bible Dictionary
task scrape:tc           # Triple Combination Index
task scrape:jst          # Joseph Smith Translation

# Run FalkorDB locally
docker run -d -p 6379:6379 --name falkor falkordb/falkordb:latest

# Build the graph (requires DATA_DIR + FALKORDB_URL + FALKORDB_GRAPH env vars)
task load                # Phases 1-5
task load-and-embed      # Phases 1-6 (requires AWS credentials)
task embed               # Phase 6 only — re-embed against existing graph

# Regenerate go-ormql client from schema.graphql
task generate

# Run the demo CLI search
go run ./cmd/app

# Frontend (mock-backed until the HTTP server exists)
cd frontend && npm install && npm run dev
cd frontend && npm test
```

Required env vars (loaded from `.env` or the process environment):
- `ENV` — `development` or `production`
- `FALKORDB_URL` — Redis URL for the FalkorDB instance (e.g. `redis://localhost:6379`)
- `FALKORDB_GRAPH` — graph name to use within that instance (e.g. `lds-gpt`)
- `DATA_DIR` — path to the scraped-JSON tree (e.g. `./pkg/data/scriptures`)
- `AWS_REGION` — for Bedrock calls

## 9. Domain Specifications

- [Scraper](./scraper/scraper.md) — HTML ingestion for scriptures and study helps
- [Dataloader](./dataloader/dataloader.md) — Multi-phase ETL into the knowledge graph
- [Database & Schema](./database/database.md) — FalkorDB client, GraphQL schema, vector indexes
- [Embeddings](./embeddings/embeddings.md) — Bedrock Titan client and float32 encoding
- [Contextual Search](./search/search.md) — Parallel kNN + graph expansion + ranking
- [App CLI](./app/app.md) — Wiring of client + embedding + search
- [Frontend](./frontend/frontend.md) — React UI (awaiting HTTP backend)

## 10. Domain Relationships

- App CLI → Database & Schema, Embeddings, Contextual Search
- Contextual Search → Database & Schema (generated GraphQL client for both kNN and graph traversal)
- Embeddings → (no internal dependencies; wraps AWS Bedrock)
- Dataloader → Database & Schema, Embeddings (phase 6 only), Scraper (reads its JSON output)
- Scraper → (no internal dependencies; produces JSON consumed by Dataloader)
- Frontend → (contract-only dependency on a future HTTP server around Contextual Search)
