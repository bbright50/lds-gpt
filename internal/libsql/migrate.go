package libsql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// vectorIndices maps index name to the CREATE INDEX DDL for DiskANN vector indices.
var vectorIndices = []string{
	"CREATE INDEX IF NOT EXISTS idx_verse_groups_embedding ON verse_groups(libsql_vector_idx(embedding))",
	"CREATE INDEX IF NOT EXISTS idx_chapters_summary_embedding ON chapters(libsql_vector_idx(summary_embedding))",
	"CREATE INDEX IF NOT EXISTS idx_tg_entries_embedding ON topical_guide_entries(libsql_vector_idx(embedding))",
	"CREATE INDEX IF NOT EXISTS idx_bd_entries_embedding ON bible_dict_entries(libsql_vector_idx(embedding))",
	"CREATE INDEX IF NOT EXISTS idx_index_entries_embedding ON index_entries(libsql_vector_idx(embedding))",
	"CREATE INDEX IF NOT EXISTS idx_jst_passages_embedding ON jst_passages(libsql_vector_idx(embedding))",
}

// Migrate runs Ent's auto-migration to create or update the database schema,
// then creates DiskANN vector indices on all embedding columns.
func (c *Client) Migrate(ctx context.Context) error {
	if err := c.ent.Schema.Create(ctx); err != nil {
		return fmt.Errorf("libsql: running migration: %w", err)
	}

	for _, ddl := range vectorIndices {
		if _, err := c.db.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("libsql: creating vector index: %w", err)
		}
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
