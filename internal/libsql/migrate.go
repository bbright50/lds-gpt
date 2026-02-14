package libsql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Migrate runs Ent's auto-migration to create or update the database schema.
func (c *Client) Migrate(ctx context.Context) error {
	if err := c.ent.Schema.Create(ctx); err != nil {
		return fmt.Errorf("libsql: running migration: %w", err)
	}
	return nil
}

// EnsureDatabaseDir creates the parent directory for dbPath if it does not exist.
func EnsureDatabaseDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("libsql: creating database directory %q: %w", dir, err)
	}
	return nil
}
