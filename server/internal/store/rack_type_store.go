package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// RackTypeStore provides CRUD operations for RackTemplate records.
type RackTypeStore struct {
	db *sql.DB
}

// NewRackTypeStore returns a new RackTypeStore backed by db.
func NewRackTypeStore(db *sql.DB) *RackTypeStore {
	return &RackTypeStore{db: db}
}

// Create inserts a new RackTemplate and returns the saved record.
func (s *RackTypeStore) Create(rt *models.RackTemplate) (*models.RackTemplate, error) {
	const q = `
		INSERT INTO rack_types (name, height_u, power_capacity_w, description)
		VALUES (?, ?, ?, ?)
		RETURNING id, name, height_u, power_capacity_w, description, created_at, updated_at`

	out := &models.RackTemplate{}
	err := s.db.QueryRow(q, rt.Name, rt.HeightU, rt.PowerCapacityW, rt.Description).
		Scan(&out.ID, &out.Name, &out.HeightU, &out.PowerCapacityW, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create rack type: %w", err)
	}
	return out, nil
}

// List returns all RackTemplate records ordered by id.
func (s *RackTypeStore) List() ([]*models.RackTemplate, error) {
	const q = `
		SELECT id, name, height_u, power_capacity_w, description, created_at, updated_at
		FROM rack_types
		ORDER BY id`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list rack types: %w", err)
	}
	defer rows.Close()

	var out []*models.RackTemplate
	for rows.Next() {
		rt := &models.RackTemplate{}
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.HeightU, &rt.PowerCapacityW, &rt.Description, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rack type: %w", err)
		}
		out = append(out, rt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rack types: %w", err)
	}
	return out, nil
}

// Get returns the RackTemplate with the given id, or models.ErrNotFound.
func (s *RackTypeStore) Get(id int64) (*models.RackTemplate, error) {
	const q = `
		SELECT id, name, height_u, power_capacity_w, description, created_at, updated_at
		FROM rack_types
		WHERE id = ?`

	rt := &models.RackTemplate{}
	err := s.db.QueryRow(q, id).
		Scan(&rt.ID, &rt.Name, &rt.HeightU, &rt.PowerCapacityW, &rt.Description, &rt.CreatedAt, &rt.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get rack type %d: %w", id, err)
	}
	return rt, nil
}

// Update modifies the RackTemplate with the given id and returns the updated record.
func (s *RackTypeStore) Update(rt *models.RackTemplate) (*models.RackTemplate, error) {
	const q = `
		UPDATE rack_types
		SET name = ?, height_u = ?, power_capacity_w = ?, description = ?,
		    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, name, height_u, power_capacity_w, description, created_at, updated_at`

	out := &models.RackTemplate{}
	err := s.db.QueryRow(q, rt.Name, rt.HeightU, rt.PowerCapacityW, rt.Description, rt.ID).
		Scan(&out.ID, &out.Name, &out.HeightU, &out.PowerCapacityW, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update rack type %d: %w", rt.ID, err)
	}
	return out, nil
}

// Delete removes the RackTemplate with the given id.
// Returns models.ErrNotFound if it does not exist.
// Returns models.ErrConflict if racks reference this rack type.
func (s *RackTypeStore) Delete(id int64) error {
	// Check for referencing racks.
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM racks WHERE rack_type_id = ?`, id).Scan(&count); err != nil {
		return fmt.Errorf("check rack type references: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%w: %d rack(s) reference this rack type", models.ErrConflict, count)
	}

	result, err := s.db.Exec(`DELETE FROM rack_types WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete rack type %d: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete rack type rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// ListRackIDsForType returns the IDs of all racks referencing the given rack type.
func (s *RackTypeStore) ListRackIDsForType(typeID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT id FROM racks WHERE rack_type_id = ?`, typeID)
	if err != nil {
		return nil, fmt.Errorf("list rack ids for type %d: %w", typeID, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan rack id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
