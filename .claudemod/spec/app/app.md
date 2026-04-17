# Specification: App CLI

## 1. Purpose

The composition root for query-time operations. Wires a LibSQL client, a Bedrock embedding client, and the contextual-search pipeline together, then runs a single demo query. Today it is a CLI demo; it is the intended home of the eventual HTTP server that will back the frontend.

## 2. Key Components

- `cmd/app/app.go` — `main`: loads config, constructs `libsql.Client` and `bedrockembedding.Client`, calls `app.NewApp(...).DoContextualSearch(ctx, "What is faith?", libsql.WithKNN(10))`, prints `(distance) entityType [id]: text` lines.
- `cmd/app/config/config.go` — Independent config package (currently duplicates `cmd/dataloader/config` wholesale — both read the same env vars via viper).
- `internal/app/app.go` — `App` struct + `NewApp` constructor; single method `DoContextualSearch(ctx, query, opts...)` that embeds the query then calls `libsqlClient.DoContextualSearch`.
- `internal/app/app_test.go` — Unit tests using `bedrockembedding/mocks` + the in-memory seeded DB from `internal/libsql/testing_helpers.go`.

## 3. Data Models

- **`App`** — Holds `*libsql.Client` and `bedrockembedding.Client`. No per-request state.
- **Inputs** — Query string (must be non-empty) plus `libsql.ContextSearchOption` variadics (currently just `WithKNN`).
- **Output** — `[]libsql.SearchResult` sorted by rank score. Printed lines in the demo: `"(%.3f) %s [%d]: %s"` with `Distance`, `EntityType`, `ID`, `Text`.

## 4. Interfaces

- **`app.NewApp(libsqlClient, embeddingClient) *App`** — Plain constructor; no options pattern.
- **`(*App).DoContextualSearch(ctx, query, opts...) ([]SearchResult, error)`** — Pipeline: validate non-empty query → `EmbedText` → `Float64sToFloat32Bytes` → `libsqlClient.DoContextualSearch`. Errors from any step are wrapped with an `"app:"` prefix.
- **CLI exit codes** — Non-zero on config load, config validation, libsql client creation, AWS config load, or search error; `0` on success.

## 5. Dependencies

- **Depends on:** Database & Schema, Embeddings, Contextual Search, `spf13/viper` (via `cmd/app/config`), AWS SDK config.
- **Depended on by:** None (entry point).

## 6. Acceptance Criteria

- Running `go run ./cmd/app` against a loaded + embedded DB prints at least one search result line and exits 0.
- An empty query returned from upstream (if the demo is extended to read from flags/stdin) triggers the `"search query must not be empty"` error before any Bedrock call.
- `AWS_REGION` from config is propagated to the Bedrock client (not just the default AWS SDK chain).
- The `libsql.Client` is `Close()`d on exit even when the CLI errors after its construction (`defer`).

## 7. Edge Cases

- **AWS credentials absent** — The demo fails on first `EmbedText`; config load succeeds (AWS config loading is lazy).
- **DB missing embeddings** — Stage 1 returns nothing, Stage 2 has nothing to expand; the CLI prints no result lines and exits 0.
- **Context cancellation** — Not wired today (the demo uses `context.Background()`); a signal-aware context would need to be added before exposing long-running HTTP handlers.
- **Future HTTP server** — The `SERVER_PORT` and `SERVER_HOSTNAME` fields already exist in `Config` but are unused. The contract the frontend expects is `POST /api/search` with JSON body `{query, knn?}` returning `{results: SearchResult[]}`. This is deferred.
