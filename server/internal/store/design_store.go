package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DesignStore provides CRUD operations for Design records.
type DesignStore struct {
	db *sql.DB
}

// NewDesignStore returns a new DesignStore backed by db.
func NewDesignStore(db *sql.DB) *DesignStore {
	return &DesignStore{db: db}
}

// Create inserts a new Design and returns the saved record (with ID and timestamps).
func (s *DesignStore) Create(d *models.Design) (*models.Design, error) {
	const q = `
		INSERT INTO designs (name, description)
		VALUES (?, ?)
		RETURNING id, name, description, created_at, updated_at`

	row := s.db.QueryRow(q, d.Name, d.Description)
	out := &models.Design{}
	if err := row.Scan(&out.ID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create design: %w", err)
	}
	return out, nil
}

// List returns all Design records ordered by id.
func (s *DesignStore) List() ([]*models.Design, error) {
	const q = `
		SELECT id, name, description, created_at, updated_at
		FROM designs
		ORDER BY id`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list designs: %w", err)
	}
	defer rows.Close()

	var designs []*models.Design
	for rows.Next() {
		d := &models.Design{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan design: %w", err)
		}
		designs = append(designs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate designs: %w", err)
	}
	return designs, nil
}

// Get returns the Design with the given id, or models.ErrNotFound.
func (s *DesignStore) Get(id int64) (*models.Design, error) {
	const q = `
		SELECT id, name, description, created_at, updated_at
		FROM designs
		WHERE id = ?`

	d := &models.Design{}
	err := s.db.QueryRow(q, id).Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get design %d: %w", id, err)
	}
	return d, nil
}

// DesignScaffold holds the auto-created default site and super-block IDs for a design.
type DesignScaffold struct {
	SiteID       int64 `json:"site_id"`
	SuperBlockID int64 `json:"super_block_id"`
}

// GetOrCreateScaffold returns the default site and super-block for a design,
// creating them if they don't already exist.
func (s *DesignStore) GetOrCreateScaffold(designID int64) (*DesignScaffold, error) {
	// Try to find existing site for this design.
	var siteID int64
	err := s.db.QueryRow(`SELECT id FROM sites WHERE design_id = ? ORDER BY id LIMIT 1`, designID).Scan(&siteID)
	if err != nil {
		// No site exists — create one.
		err = s.db.QueryRow(
			`INSERT INTO sites (design_id, name, description) VALUES (?, 'Default Site', '') RETURNING id`,
			designID,
		).Scan(&siteID)
		if err != nil {
			return nil, fmt.Errorf("create default site: %w", err)
		}
	}

	// Try to find existing super-block for this site.
	var superBlockID int64
	err = s.db.QueryRow(`SELECT id FROM super_blocks WHERE site_id = ? ORDER BY id LIMIT 1`, siteID).Scan(&superBlockID)
	if err != nil {
		// No super-block exists — create one.
		err = s.db.QueryRow(
			`INSERT INTO super_blocks (site_id, name, description) VALUES (?, 'Default Pod', '') RETURNING id`,
			siteID,
		).Scan(&superBlockID)
		if err != nil {
			return nil, fmt.Errorf("create default super-block: %w", err)
		}
	}

	return &DesignScaffold{SiteID: siteID, SuperBlockID: superBlockID}, nil
}

// Delete removes the Design with the given id. Returns models.ErrNotFound if it
// does not exist.
func (s *DesignStore) Delete(id int64) error {
	const q = `DELETE FROM designs WHERE id = ?`

	result, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("delete design %d: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete design rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}
