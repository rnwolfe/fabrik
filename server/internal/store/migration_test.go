package store_test

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"

	"github.com/rnwolfe/fabrik/server/internal/migrations"
)

// newMigrator opens a fresh in-memory DB and returns a migrator and the underlying *sql.DB.
// The DB is automatically closed via t.Cleanup when the test finishes.
func newMigrator(t *testing.T) (*migrate.Migrate, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", "file::memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("iofs source: %v", err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		t.Fatalf("sqlite driver: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		t.Fatalf("create migrator: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return m, db
}

func TestMigration_UpDown(t *testing.T) {
	m, db := newMigrator(t)

	// Apply all UP migrations.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	// Verify tables exist.
	tables := []string{
		"designs", "sites", "super_blocks", "blocks",
		"racks", "devices", "ports", "device_models", "fabrics",
	}
	for _, tbl := range tables {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM "+tbl).Scan(&n); err != nil {
			t.Errorf("table %q not found after up migration: %v", tbl, err)
		}
	}

	// Roll back all migrations.
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate down: %v", err)
	}

	// Verify designs table is gone.
	_, err := db.Exec("SELECT id FROM designs LIMIT 1")
	if err == nil {
		t.Error("expected error querying designs after down migration")
	}
}

func TestMigration_Rollback(t *testing.T) {
	m, _ := newMigrator(t)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	ver, _, err := m.Version()
	if err != nil {
		t.Fatalf("version before down: %v", err)
	}
	if ver == 0 {
		t.Fatal("expected non-zero version after up")
	}

	// Roll back all migrations using Down().
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate down: %v", err)
	}

	// After full rollback, Version returns ErrNilVersion.
	_, _, err = m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		t.Fatalf("unexpected error after down: %v", err)
	}
}
