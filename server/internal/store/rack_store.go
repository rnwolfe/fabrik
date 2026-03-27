package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// RackStore provides CRUD operations for Rack records and device placement.
type RackStore struct {
	db *sql.DB
}

// NewRackStore returns a new RackStore backed by db.
func NewRackStore(db *sql.DB) *RackStore {
	return &RackStore{db: db}
}

// Create inserts a new Rack and returns the saved record.
func (s *RackStore) Create(r *models.Rack) (*models.Rack, error) {
	const q = `
		INSERT INTO racks (block_id, rack_type_id, name, height_u, power_capacity_w, description)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, block_id, rack_type_id, name, height_u, power_capacity_w, description, created_at, updated_at`

	out := &models.Rack{}
	err := s.db.QueryRow(q, r.BlockID, r.RackTypeID, r.Name, r.HeightU, r.PowerCapacityW, r.Description).
		Scan(&out.ID, &out.BlockID, &out.RackTypeID, &out.Name, &out.HeightU, &out.PowerCapacityW, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create rack: %w", err)
	}
	return out, nil
}

// List returns all Rack records, optionally filtered by block_id.
func (s *RackStore) List(blockID *int64) ([]*models.Rack, error) {
	q := `
		SELECT id, block_id, rack_type_id, name, height_u, power_capacity_w, description, created_at, updated_at
		FROM racks`
	args := []any{}

	if blockID != nil {
		q += " WHERE block_id = ?"
		args = append(args, *blockID)
	}
	q += " ORDER BY id"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list racks: %w", err)
	}
	defer rows.Close()

	var out []*models.Rack
	for rows.Next() {
		r := &models.Rack{}
		if err := rows.Scan(&r.ID, &r.BlockID, &r.RackTypeID, &r.Name, &r.HeightU, &r.PowerCapacityW, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rack: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate racks: %w", err)
	}
	return out, nil
}

// Get returns the Rack with the given id, or models.ErrNotFound.
func (s *RackStore) Get(id int64) (*models.Rack, error) {
	const q = `
		SELECT id, block_id, rack_type_id, name, height_u, power_capacity_w, description, created_at, updated_at
		FROM racks
		WHERE id = ?`

	r := &models.Rack{}
	err := s.db.QueryRow(q, id).
		Scan(&r.ID, &r.BlockID, &r.RackTypeID, &r.Name, &r.HeightU, &r.PowerCapacityW, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get rack %d: %w", id, err)
	}
	return r, nil
}

// Update modifies the Rack with the given id and returns the updated record.
func (s *RackStore) Update(r *models.Rack) (*models.Rack, error) {
	const q = `
		UPDATE racks
		SET block_id = ?, name = ?, description = ?,
		    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, block_id, rack_type_id, name, height_u, power_capacity_w, description, created_at, updated_at`

	out := &models.Rack{}
	err := s.db.QueryRow(q, r.BlockID, r.Name, r.Description, r.ID).
		Scan(&out.ID, &out.BlockID, &out.RackTypeID, &out.Name, &out.HeightU, &out.PowerCapacityW, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update rack %d: %w", r.ID, err)
	}
	return out, nil
}

// Delete removes the Rack with the given id and cascades to devices.
func (s *RackStore) Delete(id int64) error {
	result, err := s.db.Exec(`DELETE FROM racks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete rack %d: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete rack rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// PlaceDevice inserts a device into a rack at the given position.
func (s *RackStore) PlaceDevice(d *models.Device) (*models.Device, error) {
	const q = `
		INSERT INTO devices (rack_id, device_model_id, name, role, position, description)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, rack_id, device_model_id, name, role, position, description, created_at, updated_at`

	out := &models.Device{}
	err := s.db.QueryRow(q, d.RackID, d.DeviceModelID, d.Name, d.Role, d.Position, d.Description).
		Scan(&out.ID, &out.RackID, &out.DeviceModelID, &out.Name, &out.Role, &out.Position, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("place device: %w", err)
	}
	return out, nil
}

// GetDevice returns the device with the given id, or models.ErrNotFound.
func (s *RackStore) GetDevice(id int64) (*models.Device, error) {
	const q = `
		SELECT id, rack_id, device_model_id, name, role, position, description, created_at, updated_at
		FROM devices
		WHERE id = ?`

	d := &models.Device{}
	err := s.db.QueryRow(q, id).
		Scan(&d.ID, &d.RackID, &d.DeviceModelID, &d.Name, &d.Role, &d.Position, &d.Description, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device %d: %w", id, err)
	}
	return d, nil
}

// MoveDevice updates a device's rack and position.
func (s *RackStore) MoveDevice(deviceID, rackID int64, position int) (*models.Device, error) {
	const q = `
		UPDATE devices
		SET rack_id = ?, position = ?,
		    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, rack_id, device_model_id, name, role, position, description, created_at, updated_at`

	d := &models.Device{}
	err := s.db.QueryRow(q, rackID, position, deviceID).
		Scan(&d.ID, &d.RackID, &d.DeviceModelID, &d.Name, &d.Role, &d.Position, &d.Description, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("move device %d: %w", deviceID, err)
	}
	return d, nil
}

// RemoveDevice deletes a device. If compact is true, devices above the gap are shifted down.
func (s *RackStore) RemoveDevice(deviceID int64, compact bool) error {
	// Load the device first to know its position and height.
	device, err := s.GetDevice(deviceID)
	if err != nil {
		return err
	}

	// Get model height.
	var heightU int
	if err := s.db.QueryRow(`SELECT height_u FROM device_models WHERE id = ?`, device.DeviceModelID).Scan(&heightU); err != nil {
		// Default to 1 if model not found.
		heightU = 1
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(`DELETE FROM devices WHERE id = ?`, deviceID); err != nil {
		return fmt.Errorf("delete device %d: %w", deviceID, err)
	}

	if compact {
		// Shift all devices above the deleted device's position down by heightU.
		_, err = tx.Exec(`
			UPDATE devices
			SET position = position - ?,
			    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
			WHERE rack_id = ? AND position > ?`,
			heightU, device.RackID, device.Position)
		if err != nil {
			return fmt.Errorf("compact devices: %w", err)
		}
	}

	return tx.Commit()
}

// ListDevicesInRack returns all devices in a rack with model info.
func (s *RackStore) ListDevicesInRack(rackID int64) ([]*models.DeviceSummary, error) {
	const q = `
		SELECT d.id, d.rack_id, d.device_model_id, d.name, d.role, d.position, d.description, d.created_at, d.updated_at,
		       dm.vendor, dm.model, dm.height_u, dm.power_watts
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		WHERE d.rack_id = ?
		ORDER BY d.position`

	rows, err := s.db.Query(q, rackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack %d: %w", rackID, err)
	}
	defer rows.Close()

	var out []*models.DeviceSummary
	for rows.Next() {
		ds := &models.DeviceSummary{}
		if err := rows.Scan(
			&ds.Device.ID, &ds.Device.RackID, &ds.Device.DeviceModelID,
			&ds.Device.Name, &ds.Device.Role, &ds.Device.Position, &ds.Device.Description,
			&ds.Device.CreatedAt, &ds.Device.UpdatedAt,
			&ds.ModelVendor, &ds.ModelName, &ds.HeightU, &ds.PowerWatts,
		); err != nil {
			return nil, fmt.Errorf("scan device summary: %w", err)
		}
		out = append(out, ds)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate devices: %w", err)
	}
	return out, nil
}

// GetDeviceModel returns the DeviceModel with the given id.
func (s *RackStore) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, port_count, height_u, power_watts, description, created_at, updated_at
		FROM device_models
		WHERE id = ?`

	dm := &models.DeviceModel{}
	err := s.db.QueryRow(q, id).
		Scan(&dm.ID, &dm.Vendor, &dm.Model, &dm.PortCount, &dm.HeightU, &dm.PowerWatts, &dm.Description, &dm.CreatedAt, &dm.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", id, err)
	}
	return dm, nil
}
