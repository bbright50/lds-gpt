package main

import (
	"context"
	"fmt"
	"lds-gpt/cmd/dataloader/config"
	"lds-gpt/internal/app"
	"lds-gpt/internal/bedrockembedding"
	"lds-gpt/internal/libsql"
	"log/slog"
	"os"

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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// ctx := context.Background()

	// LibSQL client
	libsqlClient, err := libsql.NewClient(libsql.Config{
		Path: cfg.MainDBURL,
	})
	if err != nil {
		logger.Error("failed to create libsql client", "error", err)
		os.Exit(1)
	}
	defer libsqlClient.Close()

	// AWS embedding client
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		logger.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}
	embeddingClient := bedrockembedding.NewClient(awsCfg)

	// create application
	app := app.NewApp(libsqlClient, embeddingClient)

	// Do contextual search
	results, err := app.DoContextualSearch(context.Background(), "What is faith?", libsql.WithKNN(10))
	if err != nil {
		logger.Error("failed to do contextual search", "error", err)
		os.Exit(1)
	}

	for _, result := range results {
		fmt.Printf("(%.3f) %s [%d]: %s\n", result.Distance, result.EntityType, result.ID, result.Text)
	}
}
