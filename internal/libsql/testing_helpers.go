package libsql

import (
	"context"
	"testing"
)

// TestClient creates an in-memory SQLite client with auto-migration applied.
// The client is automatically closed when the test completes.
func TestClient(t *testing.T) *Client {
	t.Helper()

	client, err := NewClient(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}

	// In-memory SQLite creates a separate database per connection.
	// Limit to 1 so all queries (including concurrent ones) share the same DB.
	client.db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("closing test client: %v", err)
		}
	})

	if err := client.Migrate(context.Background()); err != nil {
		t.Fatalf("running test migration: %v", err)
	}

	return client
}
