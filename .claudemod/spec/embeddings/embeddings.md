# Specification: Embeddings

## 1. Purpose

Generates 1024-dimensional vector embeddings via AWS Bedrock's Titan Text Embeddings V2 and encodes them for LibSQL storage. Used by the Dataloader (phase 6, bulk) and the App CLI (per-query, one-shot). This domain is the project's only interface to a paid external service, so it owns concurrency control and on-the-wire concerns in one place.

## 2. Key Components

- `internal/bedrockembedding/client.go` — `Client` interface with a single method `EmbedText(ctx, text) ([]float64, error)`; concrete `client` struct composed with a pond-backed `rate_limiter.Embeddable[[]float64]` (max 20 concurrent) and a Bedrock Runtime client.
- `internal/bedrockembedding/mocks/` — `gomock`-generated mock for test wiring (see `//go:generate mockgen` directive in `client.go`).
- `internal/utils/vec/encoding.go` — `Float64sToFloat32Bytes`, `Float32sToBytes`, `BytesToFloat32s`. Little-endian packing.
- `internal/utils/rate_limiter/embeddable.go` — Generic `Embeddable[T]` wrapping a `pond.ResultPool[T]` with `Submit` / `SubmitErr` / `StopAndWait`. Any client that needs bounded concurrency embeds this.

## 3. Data Models

- **`embedRequest`** — `{"inputText": "..."}` sent as the Bedrock InvokeModel body.
- **`embedResponse`** — `{"embedding": [float64, ...]}` — Titan returns JSON floats, so the internal type is `[]float64`.
- **Storage format** — LibSQL's `F32_BLOB(1024)` expects packed little-endian `float32` bytes. `Float64sToFloat32Bytes` does the downcast + pack in one pass. Round-tripping via `BytesToFloat32s` returns a new `[]float32` suitable for any further math.
- **Embedding dimensionality** — Hardcoded at 1024 in the column type (`F32_BLOB(1024)`) and implicitly in the model choice (`EMBEDDING_MODEL_ID = "amazon.titan-embed-text-v2:0"`). The two must stay in sync.

## 4. Interfaces

- **`NewClient(awsConfig, options...) Client`** — Default 20-concurrent rate limit; `WithMaxConcurrentRequests(n)` (defined but not currently wired into construction).
- **`Client.EmbedText(ctx, text) ([]float64, error)`** — Thread-safe; calls queue through the pond pool. Returns the raw float64 vector.
- **Pre-storage conversion** — Callers must apply `vec.Float64sToFloat32Bytes` before writing to any `F32_BLOB` column. The embedding type `[]float64` is deliberately different from the storage type `[]byte` so callers cannot accidentally skip the conversion.

## 5. Dependencies

- **Depends on:** `github.com/aws/aws-sdk-go-v2/service/bedrockruntime`, `github.com/alitto/pond/v2`, `go.uber.org/mock` (tests only).
- **Depended on by:** Dataloader (phase 6, 8-way outer concurrency multiplied by this package's own 20-way limit), App CLI (single-call per search query).

## 6. Acceptance Criteria

- `EmbedText` rejects no input shapes client-side — whitespace-only and empty strings are passed through so the service's own behavior is observed (callers are expected to have already composed meaningful text).
- Return value length is always 1024 on success; a response missing `"embedding"` returns `"embedding not found in response"` rather than an empty slice.
- Concurrency across many `EmbedText` calls from multiple goroutines is bounded by the configured pool size (default 20) — the caller does not need to add external rate limiting.
- `Float64sToFloat32Bytes(xs)` produces exactly `4 * len(xs)` bytes in little-endian order.
- `BytesToFloat32s(Float32sToBytes(xs))` round-trips bit-exactly.
- The mock in `mocks/` stays in sync with the `Client` interface (`go generate ./internal/bedrockembedding` regenerates it).

## 7. Edge Cases

- **AWS credentials not configured** — Surfaces on first `InvokeModel` call as a loader-phase error; the client constructor does not eagerly validate.
- **Titan API throttling** — Returned verbatim as the err; Dataloader's phase-6 worker downgrades it to a `stats.Warn` and leaves the row's embedding null so a later `--embed-only` run can retry just those rows.
- **Input text > Titan's token limit** — The service truncates or errors per its own rules; callers that anticipate long inputs (Dataloader) pre-truncate to 25 000 chars.
- **Float32 precision loss from float64 downcast** — Acceptable; cosine distance at 1024 dims is not affected meaningfully, and storage halves in size.
- **Non-little-endian hosts** — Not supported. `encoding/binary.LittleEndian` is used unconditionally because LibSQL's on-disk format is little-endian; reading a DB produced on a big-endian host would corrupt vectors. (Go's official supported platforms are all LE, so this is a non-issue in practice.)
