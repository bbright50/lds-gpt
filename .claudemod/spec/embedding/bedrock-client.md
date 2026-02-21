# Specification: Bedrock Embedding Client

## 1. Goal

Generate 1024-dimensional vector embeddings for text content using AWS Bedrock Titan Embed Text v2, with rate limiting for bulk operations.

## 2. User Stories

- **As a data loader**, I want to embed all entity text during the ETL pipeline.
- **As a search system**, I want to embed user queries at search time for vector comparison.

## 3. Technical Requirements

- **Entry Point**: `internal/bedrockembedding/client.go`
- **Interface**: `Client` with `GenerateEmbedding(ctx, text) ([]float64, error)`
- **AWS Model**: `amazon.titan-embed-text-v2:0`
- **Output Dimension**: 1024
- **Rate Limiting**: `internal/utils/rate_limiter/embeddable.go` using pond worker pool (20 concurrent)
- **Vector Encoding**: `internal/utils/vec/encoding.go`
  - `Float64sToFloat32Bytes([]float64) []byte` — for DB storage
  - `BytesToFloat32s([]byte) []float32` — for search queries
- **Mock**: `internal/bedrockembedding/mocks/` via go.uber.org/mock

## 4. Acceptance Criteria

- Single text input produces a 1024-dimensional float64 vector.
- Bulk embedding respects rate limits (max 20 concurrent requests).
- Encoding round-trips correctly (float64 -> float32 bytes -> float32).
- Mock client enables testing without AWS credentials.

## 5. Edge Cases

- Empty text input (should return error or zero vector).
- AWS rate limiting / throttling (retried via SDK).
- Very long text input (Titan has token limits; may need truncation).
