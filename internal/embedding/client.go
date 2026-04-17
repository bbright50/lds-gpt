package embedding

import "context"

// Client is consumed by the dataloader (Phase 6) and the runtime app
// (query-time embedding). 1024 dimensions are assumed across the codebase
// because the FalkorDB @vector indexes in internal/falkor/schema.graphql
// are wired that way. Swap the underlying model only after confirming
// output dimensionality matches — otherwise the schema needs a regen and
// a full re-embed.
//
// EmbedText is the latency-optimised single-item path (used for the
// query-time hot loop). EmbedBatch is the throughput-optimised path
// (used by the Phase 6 loader to amortise HTTP round-trips across many
// chunks); implementations should forward a single string through
// EmbedBatch rather than duplicating the request code.
//
//go:generate mockgen -source=client.go -destination=mocks/mock_embedding_client.go -package=mocks
type Client interface {
	EmbedText(ctx context.Context, text string) ([]float64, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}
