package store_test

import (
	"database/sql"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/migrations"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// openTestDB returns an in-memory SQLite database with migrations applied.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open("file::memory:?cache=shared", migrations.FS)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpen_MigrationsApplied(t *testing.T) {
	db := openTestDB(t)

	// Verify the designs table exists by querying it.
	_, err := db.Exec("SELECT id FROM designs LIMIT 1")
	if err != nil {
		t.Fatalf("designs table not created: %v", err)
	}
}

func TestOpen_WALMode(t *testing.T) {
	// In-memory SQLite always uses "memory" journal mode; WAL only applies to
	// file-backed databases. This test verifies the PRAGMA executes without
	// error on a file-backed DB.
	dir := t.TempDir()
	db, err := store.Open(dir+"/wal_test.db", migrations.FS)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("expected journal_mode=wal, got %q", mode)
	}
}

func TestOpen_ForeignKeys(t *testing.T) {
	db := openTestDB(t)

	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fk)
	}
}

// TestOpen_ForeignKeysPerConnection verifies that foreign_keys=1 is set on
// every connection in the pool, not just the first one. Two separate
// connections are acquired and both must report PRAGMA foreign_keys = 1.
func TestOpen_ForeignKeysPerConnection(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(dir+"/fk_multi.db", migrations.FS)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Allow multiple connections so we actually get two distinct connections.
	db.SetMaxOpenConns(5)

	check := func(label string) {
		conn, err := db.Conn(t.Context())
		if err != nil {
			t.Fatalf("%s: acquire conn: %v", label, err)
		}
		defer conn.Close()
		var fk int
		if err := conn.QueryRowContext(t.Context(), "PRAGMA foreign_keys").Scan(&fk); err != nil {
			t.Fatalf("%s: query foreign_keys: %v", label, err)
		}
		if fk != 1 {
			t.Errorf("%s: expected foreign_keys=1, got %d", label, fk)
		}
	}

	check("conn1")
	check("conn2")
}
