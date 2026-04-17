package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"lds-gpt/cmd/dataloader/config"
	"lds-gpt/internal/app"
	"lds-gpt/internal/embedding"
	"lds-gpt/internal/falkor"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		fmt.Printf("failed to validate config: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx := context.Background()

	client, err := falkor.NewClient(falkor.Config{
		URL:       cfg.FalkorDBURL,
		GraphName: cfg.FalkorDBGraph,
	})
	if err != nil {
		logger.Error("falkor client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	if cfg.OllamaURL == "" || cfg.OllamaModel == "" {
		logger.Error("OLLAMA_URL and OLLAMA_MODEL are required")
		os.Exit(1)
	}
	embedClient := embedding.NewOllamaClient(cfg.OllamaURL, cfg.OllamaModel)

	a := app.NewApp(client, embedClient)

	results, err := a.DoContextualSearch(ctx, "What is faith?", falkor.WithKNN(10))
	if err != nil {
		logger.Error("contextual search", "error", err)
		os.Exit(1)
	}

	for _, r := range results {
		fmt.Printf("(%.3f) %s [%s]: %s\n", r.Distance, r.EntityType, r.ID, r.Text)
	}
}
