// Package store provides the SQLite repository layer for fabrik.
package store

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

// Open opens (or creates) the SQLite database at path, enables WAL mode and
// foreign key enforcement, then runs all pending migrations from migrationsFS.
// It returns the open *sql.DB ready for use.
//
// Foreign key enforcement is set via a DSN pragma so that every connection in
// the pool has it enabled from the moment it is opened, rather than relying on
// a single PRAGMA statement that is scoped to only the connection it runs on.
func Open(path string, migrationsFS fs.FS) (*sql.DB, error) {
	// Normalize plain file paths to SQLite URIs so pragma query params work.
	// Without the file: scheme, modernc.org/sqlite treats query parameters as
	// part of the filename, creating a literal file named "fabrik.db?_pragma=…".
	dsnPath := path
	if !strings.HasPrefix(dsnPath, "file:") && !strings.HasPrefix(dsnPath, ":") {
		dsnPath = "file:" + dsnPath
	}
	sep := "?"
	if strings.Contains(dsnPath, "?") {
		sep = "&"
	}
	dsn := dsnPath + sep + "_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	if err := runMigrations(db, migrationsFS); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// runMigrations applies all pending UP migrations from migrationsFS.
func runMigrations(db *sql.DB, migrationsFS fs.FS) error {
	src, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("create sqlite migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("get migration version: %w", err)
	}
	slog.Info("migrations applied", "version", version, "dirty", dirty)
	return nil
}
