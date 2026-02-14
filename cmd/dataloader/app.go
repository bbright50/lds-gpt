package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"lds-gpt/cmd/dataloader/config"
	"lds-gpt/internal/bedrockembedding"
	"lds-gpt/internal/dataloader"
	"lds-gpt/internal/libsql"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v", err)
		os.Exit(1)
	}
	err = config.Validate(cfg)
	if err != nil {
		fmt.Printf("failed to validate config: %v", err)
		os.Exit(1)
	}
}

func main() {
	embed := flag.Bool("embed", false, "Run phases 1-5 + phase 6 (embedding generation)")
	embedOnly := flag.Bool("embed-only", false, "Run phase 6 only (embedding generation against existing DB)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	// Only delete/recreate DB for full load (not embed-only)
	if !*embedOnly {
		logger.Info("preparing database", "path", cfg.MainDBURL)
		if err := os.Remove(cfg.MainDBURL); err != nil && !os.IsNotExist(err) {
			logger.Error("failed to remove existing database", "error", err)
			os.Exit(1)
		}
	}

	if err := libsql.EnsureDatabaseDir(cfg.MainDBURL); err != nil {
		logger.Error("failed to create database directory", "error", err)
		os.Exit(1)
	}

	libsqlClient, err := libsql.NewClient(libsql.Config{
		Path: cfg.MainDBURL,
	})
	if err != nil {
		logger.Error("failed to create libsql client", "error", err)
		os.Exit(1)
	}
	defer libsqlClient.Close()

	// Run migrations (safe for embed-only since migrations are idempotent)
	if !*embedOnly {
		logger.Info("running migrations")
		if err := libsqlClient.Migrate(ctx); err != nil {
			logger.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	}

	// Enable WAL mode for better write performance
	if _, err := libsqlClient.Sqlx().ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		logger.Error("failed to set WAL mode", "error", err)
		os.Exit(1)
	}
	if _, err := libsqlClient.Sqlx().ExecContext(ctx, "PRAGMA synchronous=NORMAL"); err != nil {
		logger.Error("failed to set synchronous mode", "error", err)
		os.Exit(1)
	}

	// Build loader options
	var opts []dataloader.LoaderOption
	if *embed || *embedOnly {
		embedClient, err := newEmbedClient(ctx, cfg.AWSRegion)
		if err != nil {
			logger.Error("failed to create embedding client", "error", err)
			os.Exit(1)
		}
		opts = append(opts, dataloader.WithEmbedClient(embedClient))
	}

	loader := dataloader.New(libsqlClient.Ent(), cfg.DataDir, logger, opts...)

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
