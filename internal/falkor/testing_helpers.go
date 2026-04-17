package falkor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartFalkorContainer boots a fresh falkordb/falkordb container, waits for
// it to accept connections, returns a Client bound to an isolated graph, and
// registers t.Cleanup to tear everything down. Integration-only — requires a
// running Docker daemon.
func StartFalkorContainer(t *testing.T) *Client {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "falkordb/falkordb:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("falkor testcontainer: start: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("falkor testcontainer: terminate: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("falkor testcontainer: host: %v", err)
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("falkor testcontainer: port: %v", err)
	}
	url := fmt.Sprintf("redis://%s:%s", host, port.Port())

	cfg := Config{URL: url, GraphName: "test"}
	if os.Getenv("FALKOR_DEBUG") == "1" {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("falkor testcontainer: NewClient: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("falkor testcontainer: close: %v", err)
		}
	})
	return client
}
