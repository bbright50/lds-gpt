package app

import (
	"context"
	"fmt"

	"lds-gpt/internal/embedding"
	"lds-gpt/internal/falkor"
)

// App wires a FalkorDB client and an embedding client into the CLI /
// (future) HTTP entry point. DoContextualSearch embeds the user's query
// once and hands the []float32 straight to falkor — no more byte-packing
// via internal/utils/vec because FalkorDB's vecf32() accepts floats directly.
type App struct {
	fc          *falkor.Client
	embedClient embedding.Client
}

func NewApp(fc *falkor.Client, embedClient embedding.Client) *App {
	return &App{fc: fc, embedClient: embedClient}
}

func (a *App) DoContextualSearch(
	ctx context.Context,
	query string,
	options ...falkor.ContextSearchOption,
) ([]falkor.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("app: search query must not be empty")
	}
	f64s, err := a.embedClient.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("app: embedding query: %w", err)
	}
	f32s := make([]float32, len(f64s))
	for i, x := range f64s {
		f32s[i] = float32(x)
	}
	results, err := a.fc.DoContextualSearch(ctx, f32s, options...)
	if err != nil {
		return nil, fmt.Errorf("app: contextual search: %w", err)
	}
	return results, nil
}
