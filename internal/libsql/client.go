package libsql

import (
	"database/sql"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/tursodatabase/go-libsql"

	"lds-gpt/internal/libsql/generated"
)

func init() {
	sqlx.BindDriver("libsql", sqlx.QUESTION)
}

// buildDSN converts a path into a libsql-compatible DSN.
// ":memory:" is returned as-is; file paths get a "file:" prefix.
func buildDSN(path string) string {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return path
	}
	return "file:" + path
}

// Config holds the configuration for the database client.
type Config struct {
	Path string
}

// Client wraps both an Ent ORM client and a sqlx client,
// sharing a single *sql.DB connection pool.
type Client struct {
	db   *sql.DB
	ent  *generated.Client
	sqlx *sqlx.DB
}

// NewClient opens a SQLite database at cfg.Path and returns a Client
// that provides access via both Ent and sqlx over a shared connection.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("libsql: database path must not be empty")
	}

	db, err := sql.Open("libsql", buildDSN(cfg.Path))
	if err != nil {
		return nil, fmt.Errorf("libsql: opening database: %w", err)
	}

	cleanup := true
	defer func() {
		if cleanup {
			db.Close()
		}
	}()

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("libsql: pinging database: %w", err)
	}

	// Enable WAL mode for concurrent read access (file-based DBs only)
	// and set a busy timeout to avoid "database is locked" errors.
	// The libsql driver returns rows for PRAGMAs, so we use QueryRow.
	if cfg.Path != ":memory:" {
		var mode string
		if err := db.QueryRow("PRAGMA journal_mode=WAL").Scan(&mode); err != nil {
			return nil, fmt.Errorf("libsql: enabling WAL mode: %w", err)
		}
	}
	var timeout int
	if err := db.QueryRow("PRAGMA busy_timeout=5000").Scan(&timeout); err != nil {
		return nil, fmt.Errorf("libsql: setting busy_timeout: %w", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	entClient := generated.NewClient(generated.Driver(drv))

	sqlxDB := sqlx.NewDb(db, "libsql")

	cleanup = false
	return &Client{
		db:   db,
		ent:  entClient,
		sqlx: sqlxDB,
	}, nil
}

// Ent returns the Ent ORM client.
func (c *Client) Ent() *generated.Client {
	return c.ent
}

// Sqlx returns the sqlx client for raw SQL queries.
func (c *Client) Sqlx() *sqlx.DB {
	return c.sqlx
}

// Close closes the Ent driver and the underlying database connection.
// Both the Ent and sqlx clients become unusable after Close.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	if c.ent != nil {
		if err := c.ent.Close(); err != nil {
			return fmt.Errorf("libsql: closing ent client: %w", err)
		}
	}

	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return fmt.Errorf("libsql: closing database: %w", err)
		}
	}

	return nil
}
