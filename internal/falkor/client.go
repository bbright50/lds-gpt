// Package falkor wraps a FalkorDB graph behind a Client that owns two handles:
//
//   - a typed go-ormql client built from internal/falkor/generated, used for
//     CRUD, graph traversal, and @vector similarity queries.
//
//   - a raw *falkordb.Graph handle, used for DDL the generator does not cover
//     (Graph.Delete, administrative introspection) and for bulk writes whose
//     GraphQL shape is blocked by the @vector non-null requirement (Phase 1
//     of the dataloader inserts nodes before their embeddings exist).
package falkor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/FalkorDB/falkordb-go/v2"
	ormql "github.com/tab58/go-ormql/pkg/client"
	ormqldriver "github.com/tab58/go-ormql/pkg/driver"
	ormqlfalkor "github.com/tab58/go-ormql/pkg/driver/falkordb"

	"lds-gpt/internal/falkor/generated"
)

type Config struct {
	URL       string
	GraphName string
	// Logger, when non-nil, receives slog.Debug("cypher.execute", ...) events
	// for every Cypher statement the go-ormql driver runs. Useful for
	// debugging the GraphQL → Cypher translation.
	Logger *slog.Logger
}

type Client struct {
	db    *falkordb.FalkorDB
	graph *falkordb.Graph

	drv     ormqldriver.Driver
	gclient *ormql.Client
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("falkor: URL is required")
	}
	if cfg.GraphName == "" {
		return nil, errors.New("falkor: GraphName is required")
	}

	db, err := falkordb.FromURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("falkor: connect %q: %w", cfg.URL, err)
	}

	host, port, err := parseHostPort(cfg.URL)
	if err != nil {
		_ = db.Conn.Close()
		return nil, fmt.Errorf("falkor: parse URL: %w", err)
	}

	drv, err := ormqlfalkor.NewFalkorDBDriver(ormqldriver.Config{
		Host:          host,
		Port:          port,
		Scheme:        "redis",
		Database:      cfg.GraphName,
		VectorIndexes: generated.VectorIndexes,
		Logger:        cfg.Logger,
	})
	if err != nil {
		_ = db.Conn.Close()
		return nil, fmt.Errorf("falkor: go-ormql driver: %w", err)
	}

	return &Client{
		db:      db,
		graph:   db.SelectGraph(cfg.GraphName),
		drv:     drv,
		gclient: generated.NewClient(drv, ormql.WithBatchSize(500)),
	}, nil
}

// Ping verifies the underlying Redis connection is live.
func (c *Client) Ping(ctx context.Context) error {
	return c.db.Conn.Ping(ctx).Err()
}

// Raw returns the raw FalkorDB graph handle. Used for DDL, graph-level admin,
// and bulk writes whose GraphQL shape is blocked by @vector non-null inputs.
func (c *Client) Raw() *falkordb.Graph {
	return c.graph
}

// GraphQL returns the typed go-ormql client. All typed reads and Stage-1 kNN
// search queries flow through this handle.
func (c *Client) GraphQL() *ormql.Client {
	return c.gclient
}

// Migrate idempotently creates every vector index declared in the GraphQL
// schema. Safe to re-run — the generator emits `already indexed`-tolerant DDL.
func (c *Client) Migrate(ctx context.Context) error {
	return generated.CreateIndexes(ctx, c.drv)
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var firstErr error
	if c.drv != nil {
		if err := c.drv.Close(ctx); err != nil {
			firstErr = err
		}
	}
	if c.db != nil {
		if err := c.db.Conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// parseHostPort extracts host and TCP port from a redis:// or rediss:// URL.
// The go-ormql FalkorDB driver takes Host/Port separately rather than a URL.
func parseHostPort(raw string) (string, int, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", 0, err
	}
	host := u.Hostname()
	if host == "" {
		return "", 0, fmt.Errorf("missing host in %q", raw)
	}
	portStr := u.Port()
	if portStr == "" {
		portStr = "6379"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port %q: %w", portStr, err)
	}
	return host, port, nil
}
