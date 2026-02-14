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
