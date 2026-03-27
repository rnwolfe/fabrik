package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// FabricStore provides CRUD operations for Fabric records.
type FabricStore struct {
	db *sql.DB
}

// NewFabricStore returns a new FabricStore backed by db.
func NewFabricStore(db *sql.DB) *FabricStore {
	return &FabricStore{db: db}
}

// FabricParams holds the parameters for creating or updating a Fabric.
type FabricParams struct {
	DesignID        int64
	Name            string
	Tier            models.FabricTier
	Stages          int
	Radix           int
	Oversubscription float64
	Description     string
	// Device model assignments per role (optional).
	LeafModelID       int64
	SpineModelID      int64
	SuperSpineModelID int64
}

// FabricRecord extends models.Fabric with topology parameters.
type FabricRecord struct {
	models.Fabric
	Stages            int     `json:"stages"`
	Radix             int     `json:"radix"`
	Oversubscription  float64 `json:"oversubscription"`
	LeafModelID       *int64  `json:"leaf_model_id,omitempty"`
	SpineModelID      *int64  `json:"spine_model_id,omitempty"`
	SuperSpineModelID *int64  `json:"super_spine_model_id,omitempty"`
}

// Create inserts a new Fabric and returns the saved record.
// It uses the provided tx if non-nil, otherwise uses the store's db.
func (s *FabricStore) Create(p FabricParams) (*FabricRecord, error) {
	const q = `
		INSERT INTO fabrics (design_id, name, tier, stages, radix, oversubscription,
		                     leaf_model_id, spine_model_id, super_spine_model_id, description)
		VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, 0), NULLIF(?, 0), NULLIF(?, 0), ?)
		RETURNING id, design_id, name, tier, stages, radix, oversubscription,
		          leaf_model_id, spine_model_id, super_spine_model_id, description,
		          created_at, updated_at`

	row := s.db.QueryRow(q,
		p.DesignID, p.Name, string(p.Tier), p.Stages, p.Radix, p.Oversubscription,
		p.LeafModelID, p.SpineModelID, p.SuperSpineModelID, p.Description)
	return scanFabric(row)
}

// List returns all Fabric records ordered by id.
func (s *FabricStore) List() ([]*FabricRecord, error) {
	const q = `
		SELECT id, design_id, name, tier, stages, radix, oversubscription,
		       leaf_model_id, spine_model_id, super_spine_model_id, description,
		       created_at, updated_at
		FROM fabrics
		ORDER BY id`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list fabrics: %w", err)
	}
	defer rows.Close()

	var out []*FabricRecord
	for rows.Next() {
		f, err := scanFabricRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan fabric: %w", err)
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fabrics: %w", err)
	}
	return out, nil
}

// Get returns the Fabric with the given id, or models.ErrNotFound.
func (s *FabricStore) Get(id int64) (*FabricRecord, error) {
	const q = `
		SELECT id, design_id, name, tier, stages, radix, oversubscription,
		       leaf_model_id, spine_model_id, super_spine_model_id, description,
		       created_at, updated_at
		FROM fabrics
		WHERE id = ?`

	row := s.db.QueryRow(q, id)
	f, err := scanFabric(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	return f, err
}

// Update modifies an existing Fabric and returns the updated record.
func (s *FabricStore) Update(id int64, p FabricParams) (*FabricRecord, error) {
	const q = `
		UPDATE fabrics
		SET name = ?, tier = ?, stages = ?, radix = ?, oversubscription = ?,
		    leaf_model_id = NULLIF(?, 0), spine_model_id = NULLIF(?, 0),
		    super_spine_model_id = NULLIF(?, 0), description = ?,
		    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, design_id, name, tier, stages, radix, oversubscription,
		          leaf_model_id, spine_model_id, super_spine_model_id, description,
		          created_at, updated_at`

	row := s.db.QueryRow(q,
		p.Name, string(p.Tier), p.Stages, p.Radix, p.Oversubscription,
		p.LeafModelID, p.SpineModelID, p.SuperSpineModelID, p.Description, id)
	f, err := scanFabric(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	return f, err
}

// Delete removes the Fabric with the given id.
func (s *FabricStore) Delete(id int64) error {
	result, err := s.db.Exec("DELETE FROM fabrics WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete fabric %d: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete fabric rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// GetDeviceModelByID fetches a device model for validation.
func (s *FabricStore) GetDeviceModelByID(id int64) (*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, port_count, height_u,
		       power_watts_typical, description, created_at, updated_at
		FROM device_models
		WHERE id = ?`

	dm := &models.DeviceModel{}
	err := s.db.QueryRow(q, id).Scan(
		&dm.ID, &dm.Vendor, &dm.Model, &dm.PortCount, &dm.HeightU,
		&dm.PowerWattsTypical, &dm.Description, &dm.CreatedAt, &dm.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", id, err)
	}
	return dm, nil
}

// ListDeviceModels returns all device models.
func (s *FabricStore) ListDeviceModels() ([]*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, port_count, height_u,
		       power_watts_typical, description, created_at, updated_at
		FROM device_models
		ORDER BY vendor, model`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list device models: %w", err)
	}
	defer rows.Close()

	var out []*models.DeviceModel
	for rows.Next() {
		dm := &models.DeviceModel{}
		if err := rows.Scan(&dm.ID, &dm.Vendor, &dm.Model, &dm.PortCount, &dm.HeightU,
			&dm.PowerWattsTypical, &dm.Description, &dm.CreatedAt, &dm.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device model: %w", err)
		}
		out = append(out, dm)
	}
	return out, rows.Err()
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanFabric(row *sql.Row) (*FabricRecord, error) {
	f := &FabricRecord{}
	err := row.Scan(
		&f.ID, &f.DesignID, &f.Name, &f.Tier,
		&f.Stages, &f.Radix, &f.Oversubscription,
		&f.LeafModelID, &f.SpineModelID, &f.SuperSpineModelID,
		&f.Description, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func scanFabricRow(rows *sql.Rows) (*FabricRecord, error) {
	f := &FabricRecord{}
	err := rows.Scan(
		&f.ID, &f.DesignID, &f.Name, &f.Tier,
		&f.Stages, &f.Radix, &f.Oversubscription,
		&f.LeafModelID, &f.SpineModelID, &f.SuperSpineModelID,
		&f.Description, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}
