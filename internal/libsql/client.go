package libsql

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"lds-gpt/internal/libsql/generated"
)

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

	db, err := sql.Open("sqlite3", cfg.Path)
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

	drv := entsql.OpenDB(dialect.SQLite, db)
	entClient := generated.NewClient(generated.Driver(drv))

	sqlxDB := sqlx.NewDb(db, "sqlite3")

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
