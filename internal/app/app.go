package app

import (
	"context"
	"fmt"

	"lds-gpt/internal/bedrockembedding"
	"lds-gpt/internal/libsql"
	"lds-gpt/internal/utils/vec"
)

type App struct {
	libsqlClient    *libsql.Client
	embeddingClient bedrockembedding.Client
}

func NewApp(libsqlClient *libsql.Client, embeddingClient bedrockembedding.Client) *App {
	return &App{
		libsqlClient:    libsqlClient,
		embeddingClient: embeddingClient,
	}
}

func (a *App) DoContextualSearch(ctx context.Context, query string, options ...libsql.ContextSearchOption) ([]libsql.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("app: search query must not be empty")
	}

	floats, err := a.embeddingClient.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("app: embedding query: %w", err)
	}

	embeddingBytes := vec.Float64sToFloat32Bytes(floats)

	results, err := a.libsqlClient.DoContextualSearch(ctx, embeddingBytes, options...)
	if err != nil {
		return nil, fmt.Errorf("app: contextual search: %w", err)
	}

	return results, nil
}
