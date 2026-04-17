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

// nodeLabels is the set of @node types from internal/falkor/schema.graphql
// that use `id: ID!` as their lookup key. Every batched write in the
// dataloader does MATCH (n:<Label> {id: $id}); without a property index on
// (n.id) every MATCH is a full label scan — a 41k-verse graph then blows
// past the driver read timeout after a couple hundred rows. The index
// turns each lookup into O(1) and keeps batch times tens-of-ms flat.
var nodeLabels = []string{
	"Volume", "Book", "Chapter", "Verse", "VerseGroup",
	"TopicalGuideEntry", "BibleDictEntry", "IndexEntry", "JSTPassage",
}

// Migrate idempotently creates every vector index declared in the GraphQL
// schema plus a property index on id for every @node label, and lifts the
// server-side RESULTSET_SIZE cap so bulk target queries (e.g. Phase 6's
// "find every placeholder VerseGroup") don't get silently truncated at the
// default 10000. Safe to re-run.
func (c *Client) Migrate(ctx context.Context) error {
	// RESULTSET_SIZE = -1 → unlimited. Without this a 13k-row MATCH against
	// VerseGroup gets truncated to 10k and Phase 6 skips 25% of the nodes.
	if err := c.db.ConfigSet("RESULTSET_SIZE", -1); err != nil {
		return fmt.Errorf("falkor: raise RESULTSET_SIZE cap: %w", err)
	}
	if err := generated.CreateIndexes(ctx, c.drv); err != nil {
		return err
	}
	for _, label := range nodeLabels {
		q := fmt.Sprintf("CREATE INDEX FOR (n:%s) ON (n.id)", label)
		if _, err := c.graph.Query(q, nil, nil); err != nil && !isAlreadyIndexedErr(err) {
			return fmt.Errorf("falkor: create id index on %s: %w", label, err)
		}
	}
	return nil
}

func isAlreadyIndexedErr(err error) bool {
	return err != nil && (containsCI(err.Error(), "already indexed") || containsCI(err.Error(), "already exists"))
}

func containsCI(haystack, needle string) bool {
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			hc, nc := haystack[i+j], needle[j]
			if hc >= 'A' && hc <= 'Z' {
				hc += 'a' - 'A'
			}
			if nc >= 'A' && nc <= 'Z' {
				nc += 'a' - 'A'
			}
			if hc != nc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
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
