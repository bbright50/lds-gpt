package libsql

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty path returns error",
			cfg:     Config{Path: ""},
			wantErr: true,
			errMsg:  "path must not be empty",
		},
		{
			name:    "memory database succeeds",
			cfg:     Config{Path: ":memory:"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer client.Close()

			if client.Ent() == nil {
				t.Error("Ent() returned nil")
			}
			if client.Sqlx() == nil {
				t.Error("Sqlx() returned nil")
			}
		})
	}
}

func TestMigrate(t *testing.T) {
	client := TestClient(t)

	// Verify the volumes table exists by querying it via sqlx.
	var count int
	err := client.Sqlx().Get(&count, "SELECT COUNT(*) FROM volumes")
	if err != nil {
		t.Fatalf("querying volumes table: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestSharedConnection(t *testing.T) {
	client := TestClient(t)
	ctx := context.Background()

	// Ent write -> sqlx read.
	vol, err := client.Ent().Volume.Create().
		SetName("Old Testament").
		SetAbbreviation("ot").
		Save(ctx)
	if err != nil {
		t.Fatalf("creating volume via ent: %v", err)
	}

	var name string
	err = client.Sqlx().Get(&name, "SELECT name FROM volumes WHERE id = ?", vol.ID)
	if err != nil {
		t.Fatalf("querying via sqlx: %v", err)
	}
	if name != "Old Testament" {
		t.Errorf("sqlx: expected 'Old Testament', got %q", name)
	}

	// sqlx write -> Ent read (bidirectional verification).
	_, err = client.Sqlx().Exec(
		"INSERT INTO volumes (name, abbreviation) VALUES (?, ?)",
		"New Testament", "nt",
	)
	if err != nil {
		t.Fatalf("inserting via sqlx: %v", err)
	}

	entVol, err := client.Ent().Volume.Query().
		Where().
		All(ctx)
	if err != nil {
		t.Fatalf("reading via ent: %v", err)
	}
	if len(entVol) != 2 {
		t.Errorf("ent: expected 2 volumes, got %d", len(entVol))
	}
}

func TestCloseNilClient(t *testing.T) {
	var client *Client
	if err := client.Close(); err != nil {
		t.Errorf("Close() on nil client: %v", err)
	}
}

func TestEnsureDatabaseDir(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "creates nested directories",
			dbPath:  filepath.Join(t.TempDir(), "a", "b", "test.db"),
			wantErr: false,
		},
		{
			name:    "handles existing directory",
			dbPath:  filepath.Join(t.TempDir(), "test.db"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDatabaseDir(tt.dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDatabaseDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				dir := filepath.Dir(tt.dbPath)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("directory %q was not created", dir)
				}
			}
		})
	}
}
