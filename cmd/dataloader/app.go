package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	"lds-gpt/cmd/dataloader/config"
	"lds-gpt/internal/bedrockembedding"
	"lds-gpt/internal/dataloader"
	"lds-gpt/internal/falkor"
)

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		fmt.Printf("failed to validate config: %v", err)
		os.Exit(1)
	}
}

func main() {
	embed := flag.Bool("embed", false, "Run phases 1-5 + phase 6 (embedding generation)")
	embedOnly := flag.Bool("embed-only", false, "Run phase 6 only (embedding generation against an existing graph)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx := context.Background()

	client, err := falkor.NewClient(falkor.Config{
		URL:       cfg.FalkorDBURL,
		GraphName: cfg.FalkorDBGraph,
	})
	if err != nil {
		logger.Error("failed to create falkor client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// For a full load (not embed-only) destroy any prior graph with this name
	// so the loader starts from a clean slate, then recreate the vector
	// indexes. Phase 6 (embed-only) skips this — it operates against an
	// existing populated graph.
	if !*embedOnly {
		logger.Info("preparing graph", "url", cfg.FalkorDBURL, "graph", cfg.FalkorDBGraph)
		if err := client.Raw().Delete(); err != nil {
			// GRAPH.DELETE returns an error if the graph doesn't exist; that's fine.
			logger.Warn("graph delete (first run is expected to be a no-op)", "error", err)
		}
	}

	logger.Info("running migrations (vector indexes)")
	if err := client.Migrate(ctx); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	var opts []dataloader.LoaderOption
	if *embed || *embedOnly {
		embedClient, err := newEmbedClient(ctx, cfg.AWSRegion)
		if err != nil {
			logger.Error("failed to create embedding client", "error", err)
			os.Exit(1)
		}
		opts = append(opts, dataloader.WithEmbedClient(embedClient))
	}

	loader := dataloader.New(client, cfg.DataDir, logger, opts...)

	if *embedOnly {
		if err := loader.EmbedOnly(ctx); err != nil {
			logger.Error("embed-only failed", "error", err)
			os.Exit(1)
		}
	} else {
		if err := loader.Run(ctx); err != nil {
			logger.Error("dataloader failed", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("dataloader completed successfully")
}

func newEmbedClient(ctx context.Context, region string) (bedrockembedding.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return bedrockembedding.NewClient(awsCfg), nil
}
