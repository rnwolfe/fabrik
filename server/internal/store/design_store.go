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
