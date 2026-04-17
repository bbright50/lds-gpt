# Specification: Database & Schema

## 1. Purpose

Defines and owns the FalkorDB property graph and vector store. Wraps a FalkorDB connection behind two clients — a go-ormql generated client for typed CRUD + graph traversal + `@vector` similarity queries, and a raw `falkordb-go` handle for DDL that go-ormql does not cover (most importantly `CREATE VECTOR INDEX` and administrative introspection). This domain is the single source of truth for the data model — all other domains either consume the generated go-ormql code or issue raw Cypher against the same graph.

## 2. Key Components

- `internal/falkor/schema.graphql` — go-ormql schema source of truth (nine `@node` types + six `@relationshipProperties` types + six `@vector` fields). Edits here require re-running `task generate`.
- `internal/falkor/generated/` — go-ormql-generated client code. Never hand-edited. Exports a typed client, model structs, and a `VectorIndexes` descriptor passed to the driver at connect time.
- `internal/falkor/client.go` — `Client` struct owning the go-ormql generated client and the raw `*falkordb.FalkorDB` connection. `NewClient` parses the DSN, opens the graph, and attaches the generated `VectorIndexes` so the driver can rewrite vector queries to `db.idx.vector.queryNodes`.
- `internal/falkor/migrate.go` — `Migrate` idempotently issues `CREATE VECTOR INDEX` for the six embedding fields and any other schema-level DDL that go-ormql does not emit on its own.
- `internal/falkor/testing_helpers.go` — Testcontainer-backed fixture: `StartFalkorContainer(t)` spins up a throwaway `falkordb/falkordb` container, returns a connected `*Client`, and registers `t.Cleanup` to tear it down.

## 3. Data Models

Nine `@node` types + six `@relationshipProperties` types (the six through-entities collapse into edge-property types rather than separate tables).

**Structural (hierarchical)**
- `Volume` — `name`, `abbreviation` (both unique). Relationship: `-[:CONTAINS]->Book`.
- `Book` — `name`, `slug`, `urlPath`. Relationships: `<-[:CONTAINS]-Volume`, `-[:CONTAINS]->Chapter`.
- `Chapter` — `number`, `summary?`, `summaryEmbedding? @vector(dimensions: 1024, similarity: "cosine")`, `url?`. Relationships: `<-[:CONTAINS]-Book`, `-[:HAS_VERSE]->Verse`, `-[:HAS_GROUP]->VerseGroup`.
- `Verse` — `number`, `text`, `reference` (e.g. `"1 Ne. 1:1"`), `translationNotes` (JSON), `alternateReadings` (JSON), `explanatoryNotes` (JSON). Relationships: `<-[:HAS_VERSE]-Chapter`; outgoing `-[:CROSS_REF {VerseCrossRefProps}]->Verse`; incoming study-help refs from TG/BD/IDX via `-[:VERSE_REF {…Props}]->`; `<-[:COMPARES]-JSTPassage`; `<-[:INCLUDES]-VerseGroup`.
- `VerseGroup` — `text` (concatenated verses), `embedding? @vector(dimensions: 1024, similarity: "cosine")`, `startVerseNumber`, `endVerseNumber`. Relationships: `<-[:HAS_GROUP]-Chapter`, `-[:INCLUDES]->Verse`. Primary RAG retrieval unit.

**Study helps**
- `TopicalGuideEntry` — `name` (unique), `embedding? @vector(dimensions: 1024, similarity: "cosine")`. Relationships: `-[:SEE_ALSO]->TopicalGuideEntry` (self), `-[:BD_REF]->BibleDictEntry`, `-[:VERSE_REF {TGVerseRefProps}]->Verse`.
- `BibleDictEntry` — `name` (unique), `text`, `embedding? @vector(dimensions: 1024, similarity: "cosine")`. Relationships: `-[:SEE_ALSO]->BibleDictEntry` (self), `-[:VERSE_REF {BDVerseRefProps}]->Verse`, incoming `<-[:BD_REF]-(TopicalGuideEntry|IndexEntry)`.
- `IndexEntry` — `name` (unique), `embedding? @vector(dimensions: 1024, similarity: "cosine")`. Relationships: `-[:SEE_ALSO]->IndexEntry` (self), `-[:TG_REF]->TopicalGuideEntry`, `-[:BD_REF]->BibleDictEntry`, `-[:VERSE_REF {IDXVerseRefProps}]->Verse`.
- `JSTPassage` — `book`, `chapter`, `comprises` (verse range), `compareRef?`, `summary?`, `text`, `embedding? @vector(dimensions: 1024, similarity: "cosine")`. Relationships: `-[:COMPARES]->Verse`.

**Relationship properties** (metadata carried on edges — replaces Ent through-entities)
- `VerseCrossRefProps` — `{category, footnoteMarker, referenceText?}` — on `(Verse)-[:CROSS_REF]->(Verse)`.
- `TGVerseRefProps` — `{phrase?}` — on `(TopicalGuideEntry)-[:VERSE_REF]->(Verse)`.
- `BDVerseRefProps` — `{targetEndVerseId?}` — on `(BibleDictEntry)-[:VERSE_REF]->(Verse)`.
- `IDXVerseRefProps` — `{phrase?}` — on `(IndexEntry)-[:VERSE_REF]->(Verse)`.
- `VerseJSTRefProps` — on `(Verse)-[:JST_FOOTNOTE]->(JSTPassage)`.
- `VerseTGRefProps` — on `(Verse)-[:TG_FOOTNOTE]->(TopicalGuideEntry)`.

## 4. Interfaces

- **`NewClient(Config{URL, GraphName}) (*Client, error)`** — Opens a FalkorDB connection over the Redis protocol, selects the named graph, and wraps it with the go-ormql generated client. Errors if `URL` or `GraphName` is empty. No in-memory mode — tests must use `testing_helpers.StartFalkorContainer`.
- **`(*Client).GraphQL() *generated.Client`** — Typed GraphQL client for all node/relationship CRUD, graph traversal, and `@vector` kNN queries. Every call compiles to one Cypher round-trip.
- **`(*Client).Raw() *falkordb.FalkorDB`** — Escape hatch for DDL (`CREATE VECTOR INDEX`, `db.indexes`) and admin queries that do not fit GraphQL shapes.
- **`(*Client).Migrate(ctx) error`** — Idempotent. Creates the six `CREATE VECTOR INDEX` statements (one per `@vector` field) using `IF NOT EXISTS` semantics. Safe to re-run.
- **`(*Client).Close() error`** — Closes the underlying Redis connection. After `Close` both wrappers are unusable.
- **Vector index contract** — The six vector indexes are created on: `VerseGroup.embedding`, `Chapter.summaryEmbedding`, `TopicalGuideEntry.embedding`, `BibleDictEntry.embedding`, `IndexEntry.embedding`, `JSTPassage.embedding`. All use `dimension: 1024, similarityFunction: 'cosine'`. Search queries rely on these; queries without an index fall back to a full scan and are orders of magnitude slower.

## 5. Dependencies

- **Depends on:** `github.com/FalkorDB/falkordb-go`, `github.com/tab58/go-ormql`, `github.com/testcontainers/testcontainers-go` (test-only). No other internal packages.
- **Depended on by:** Dataloader (writes), Contextual Search (reads + graph), App CLI (wiring).

## 6. Acceptance Criteria

- Every node type has a unique display/name column that prevents silent duplication (`Volume.name`, `Volume.abbreviation`, `Book.name+slug`, `TopicalGuideEntry.name`, `BibleDictEntry.name`, `IndexEntry.name`).
- Every `@vector` field is declared with `dimensions: 1024, similarity: "cosine"` and is optional (nullable) so un-embedded nodes can coexist with embedded nodes.
- `Migrate` is idempotent — running it twice produces no error and leaves exactly six vector indexes.
- `NewClient` with `URL == ""` or `GraphName == ""` returns an error rather than producing a broken client.
- The generated client and the raw handle share the same FalkorDB connection (one Redis pool per `Client`).
- Every relationship-property type carries exactly the metadata required to avoid a second lookup (category, phrase, etc.) — verse-to-verse and verse-to-study-help relationships do not drop footnote context.
- Testcontainer fixture boots a fresh graph per test; `t.Cleanup` tears down the container.

## 6.1 Known schema gaps (follow-ups)

The first-pass port to FalkorDB kept the relationship-property types minimal so
the loader could land quickly. A few fields that existed on the prior Ent
through-entities are not yet modelled on the GraphQL `@relationshipProperties`
types, and the Phase 3/4 loaders therefore do not persist them:

- **`VerseCrossRefProps.targetEndVerseId`** — the ID of the trailing verse in
  a ranged cross-reference (e.g. `Gen. 1:1-3`). The Phase-3 collector still
  computes it into the in-memory row struct; adding the field to the schema
  (and re-running `task generate`) plus a one-line edit in
  `writeCrossRefs` will start persisting it. Needed only for accurate range
  rendering in a display-tier; not needed for retrieval.
- **`VerseTGRefProps.tgTopicText`**, **`VerseTGRefProps.referenceText`**,
  **`VerseBDRefProps.referenceText`**, **`BDVerseRefProps.targetEndVerseId`**
  — same story: present in the source JSON, dropped on the floor today,
  added with a schema edit + `task generate` + a few loader lines when the
  display layer starts consuming them.

None of these are required for Stage 1 (vector kNN) or Stage 2 (1-hop
expansion) — the migration ships without them and picks them up when a UI
needs them.

## 7. Edge Cases

- **Testcontainer start-up latency** — The first `StartFalkorContainer` call per test binary pulls the image; subsequent calls reuse the cached image. Individual tests pay only container-boot time (~1s).
- **`Close()` on a nil `*Client`** — Returns `nil` (defensive; matches `defer client.Close()` patterns on error paths).
- **Index creation on an empty graph** — Harmless; the index starts empty and is populated as nodes are written.
- **Re-generating the go-ormql client after schema edit without running `task generate`** — Call sites that reference removed fields fail to compile; this is the intended safety net.
- **Concurrent readers during loader writes** — FalkorDB serializes graph mutations per transaction; go-ormql auto-batches writes (default 50/batch) so a long load does not hold a single mega-transaction. Readers see intermediate state but never torn state within a single query.
- **Schema drift between `schema.graphql` and the live graph** — FalkorDB is schemaless at the storage layer, so a stale graph may contain orphaned labels/properties after a schema change; the dataloader is destructive by design (`GRAPH.DELETE` in phase 1) so a full rebuild is the canonical migration path.
