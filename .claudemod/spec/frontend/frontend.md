# Specification: Frontend

## 1. Purpose

A React single-page app that lets the user type a natural-language query and browse the ranked contextual-search results. Today it runs against an in-memory mock because the Go HTTP server has not been built yet — the shape of the real backend contract is already fixed by the types in `src/types/search.ts`.

## 2. Key Components

- `src/main.tsx` — Bootstraps React; chooses between `createSearchClient(VITE_API_BASE_URL)` and `createMockSearchAPI()` based on env.
- `src/App.tsx` — Top-level state (`results`, `loading`, `error`, `hasSearched`), owns an `AbortController` so a new search cancels an in-flight one.
- `src/api/client.ts` — `createSearchClient(baseUrl)` — `POST {baseUrl}/api/search` with `{query, knn?}` body; throws on non-OK HTTP.
- `src/api/mockSearch.ts` — Fixture-backed mock with a 150 ms simulated delay.
- `src/components/` — `SearchBar`, `SearchStatus`, `ResultList`, `ResultCard`, `EntityIcon`, `DistanceBadge`, `MetadataPanel` (+ matching `.test.tsx`).
- `src/types/search.ts` — The canonical TS contract with the Go backend: `EntityType`, `ResultMeta`, `SearchResult`, `SearchRequest`, `SearchResponse`, `SearchAPI`.
- `src/test/fixtures.ts` — Mock `SearchResult[]` used by `mockSearch` and component tests.

## 3. Data Models

- **`EntityType`** — TS union mirroring the Go `EntityType` constants. `ENTITY_TYPES` is the iterable copy.
- **`SearchResult`** — `{entityType, id, name, text, distance, metadata}` — must match Go's JSON shape exactly when the real server lands.
- **`ResultMeta`** — All fields optional; mirrors Go's flat `ResultMeta` where unused fields are zero.
- **`SearchRequest`** — `{query, knn?, signal?}`. `signal` is stripped before serialization.
- **`SearchResponse`** — `{results: SearchResult[]}`.

## 4. Interfaces

- **`SearchAPI.search(request): Promise<SearchResponse>`** — The only boundary `App` knows about; mock and real impls are interchangeable.
- **HTTP contract (deferred server)** — `POST {VITE_API_BASE_URL}/api/search`, `Content-Type: application/json`, body `{query: string, knn?: number}`, response `{results: SearchResult[]}`. Non-2xx throws `Search request failed: {status} {statusText}`.
- **Cancellation** — Each new `handleSearch` call aborts the previous `AbortController`. Aborted responses suppress both result updates and error updates (the new in-flight search owns the UI).

## 5. Dependencies

- **Depends on:** React 19, Tailwind v4 (via `@tailwindcss/vite`); dev: Vite 7, Vitest 4, Testing Library. No runtime state library — `useState` only.
- **Depended on by:** None — this is a leaf UI.
- **Implicit contract dependency:** Go's `internal/libsql/search_result.go` `EntityType` + `SearchResult` — a backend change to either requires a matching TS update.

## 6. Acceptance Criteria

- Submitting a query calls `api.search(...)` exactly once and renders the returned results in order.
- A second submission while the first is in flight aborts the first; the first's response never updates state (neither results nor error).
- On non-2xx response, the UI shows the error message from the thrown `Error` and clears the result list.
- The `mockSearch` fallback produces results within 150 ms so manual dev-loop testing works without a backend.
- `npm test` passes against the component + API tests checked into `src/`.
- `VITE_API_BASE_URL` is read at build/start time — setting it points the UI at a running Go server without code changes.

## 7. Edge Cases

- **`knn` not sent** — Server is expected to default (currently Go defaults to 20 via `WithKNN`). Mock honors this with `DEFAULT_KNN = 20`.
- **`knn <= 0`** — Mock returns `{results: []}`; the real server returns a 4xx (not yet implemented) that the UI will surface as an error string.
- **Empty query** — Currently no client-side guard; the component relies on `SearchBar` to enforce a non-empty submit. The real server is expected to return 4xx for empty queries.
- **Result with missing `metadata.url`** — `ResultCard` hides the "Source" link. No placeholder rendered.
- **Long result text** — Truncated at 300 chars with an ellipsis (`ResultCard`); full text is not accessible from the UI today.
- **Backend `EntityType` drift** — An unknown string is not guarded at runtime. The TS union would need updating and `EntityIcon` extended; otherwise the icon falls through to its default.
