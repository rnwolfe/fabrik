package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DeviceModelStore provides CRUD operations for DeviceModel records.
type DeviceModelStore struct {
	db *sql.DB
}

// NewDeviceModelStore returns a new DeviceModelStore backed by db.
func NewDeviceModelStore(db *sql.DB) *DeviceModelStore {
	return &DeviceModelStore{db: db}
}

// scanDeviceModel scans a row into a DeviceModel.
func scanDeviceModel(row interface {
	Scan(dest ...any) error
}) (*models.DeviceModel, error) {
	dm := &models.DeviceModel{}
	var archivedAt sql.NullTime
	err := row.Scan(
		&dm.ID, &dm.Vendor, &dm.Model, &dm.DeviceModelType,
		&dm.PortCount, &dm.HeightU,
		&dm.PowerWattsIdle, &dm.PowerWattsTypical, &dm.PowerWattsMax,
		&dm.CPUSockets, &dm.CoresPerSocket, &dm.RAMGB, &dm.StorageTB, &dm.GPUCount,
		&dm.Description, &dm.IsSeed, &archivedAt,
		&dm.CreatedAt, &dm.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if archivedAt.Valid {
		t := archivedAt.Time
		dm.ArchivedAt = &t
	}
	return dm, nil
}

// Create inserts a new DeviceModel and returns the saved record.
func (s *DeviceModelStore) Create(dm *models.DeviceModel) (*models.DeviceModel, error) {
	// Default type to "network" if not specified.
	modelType := dm.DeviceModelType
	if modelType == "" {
		modelType = models.DeviceModelTypeNetwork
	}

	const q = `
		INSERT INTO device_models (
			vendor, model, device_model_type,
			port_count, height_u,
			power_watts_idle, power_watts_typical, power_watts_max,
			cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
			description, is_seed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, vendor, model, device_model_type,
		          port_count, height_u,
		          power_watts_idle, power_watts_typical, power_watts_max,
		          cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
		          description, is_seed, archived_at, created_at, updated_at`

	row := s.db.QueryRow(q,
		dm.Vendor, dm.Model, modelType,
		dm.PortCount, dm.HeightU,
		dm.PowerWattsIdle, dm.PowerWattsTypical, dm.PowerWattsMax,
		dm.CPUSockets, dm.CoresPerSocket, dm.RAMGB, dm.StorageTB, dm.GPUCount,
		dm.Description, dm.IsSeed,
	)
	out, err := scanDeviceModel(row)
	if err != nil {
		if isDuplicateErr(err) {
			return nil, fmt.Errorf("create device model: %w", models.ErrDuplicate)
		}
		return nil, fmt.Errorf("create device model: %w", err)
	}
	return out, nil
}

// List returns all non-archived DeviceModel records. If includeArchived is true,
// archived models are included as well.
func (s *DeviceModelStore) List(includeArchived bool) ([]*models.DeviceModel, error) {
	q := `
		SELECT id, vendor, model, device_model_type,
		       port_count, height_u,
		       power_watts_idle, power_watts_typical, power_watts_max,
		       cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
		       description, is_seed, archived_at, created_at, updated_at
		FROM device_models`
	if !includeArchived {
		q += " WHERE archived_at IS NULL"
	}
	q += " ORDER BY vendor, model"

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("list device models: %w", err)
	}
	defer rows.Close()

	var out []*models.DeviceModel
	for rows.Next() {
		dm, err := scanDeviceModel(rows)
		if err != nil {
			return nil, fmt.Errorf("scan device model: %w", err)
		}
		out = append(out, dm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device models: %w", err)
	}
	return out, nil
}

// Get returns the DeviceModel with the given id, or models.ErrNotFound.
func (s *DeviceModelStore) Get(id int64) (*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, device_model_type,
		       port_count, height_u,
		       power_watts_idle, power_watts_typical, power_watts_max,
		       cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
		       description, is_seed, archived_at, created_at, updated_at
		FROM device_models
		WHERE id = ?`

	dm, err := scanDeviceModel(s.db.QueryRow(q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", id, err)
	}
	return dm, nil
}

// Update modifies an existing DeviceModel. Returns models.ErrNotFound if it
// does not exist.
func (s *DeviceModelStore) Update(dm *models.DeviceModel) (*models.DeviceModel, error) {
	// Default type to "network" if not specified.
	modelType := dm.DeviceModelType
	if modelType == "" {
		modelType = models.DeviceModelTypeNetwork
	}

	const q = `
		UPDATE device_models
		SET vendor = ?, model = ?, device_model_type = ?,
		    port_count = ?, height_u = ?,
		    power_watts_idle = ?, power_watts_typical = ?, power_watts_max = ?,
		    cpu_sockets = ?, cores_per_socket = ?, ram_gb = ?, storage_tb = ?, gpu_count = ?,
		    description = ?,
		    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, vendor, model, device_model_type,
		          port_count, height_u,
		          power_watts_idle, power_watts_typical, power_watts_max,
		          cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
		          description, is_seed, archived_at, created_at, updated_at`

	row := s.db.QueryRow(q,
		dm.Vendor, dm.Model, modelType,
		dm.PortCount, dm.HeightU,
		dm.PowerWattsIdle, dm.PowerWattsTypical, dm.PowerWattsMax,
		dm.CPUSockets, dm.CoresPerSocket, dm.RAMGB, dm.StorageTB, dm.GPUCount,
		dm.Description, dm.ID,
	)
	out, err := scanDeviceModel(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		if isDuplicateErr(err) {
			return nil, fmt.Errorf("update device model: %w", models.ErrDuplicate)
		}
		return nil, fmt.Errorf("update device model %d: %w", dm.ID, err)
	}
	return out, nil
}

// Archive sets archived_at to the current time for the given device model.
// Returns models.ErrNotFound if the record does not exist.
func (s *DeviceModelStore) Archive(id int64) error {
	const q = `
		UPDATE device_models
		SET archived_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?`

	result, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("archive device model %d: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive device model rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// Duplicate creates a copy of the device model identified by sourceID, giving
// it the provided vendor and model name. The copy is never a seed row.
func (s *DeviceModelStore) Duplicate(sourceID int64, newVendor, newModel string) (*models.DeviceModel, error) {
	src, err := s.Get(sourceID)
	if err != nil {
		return nil, fmt.Errorf("duplicate device model: %w", err)
	}

	copy := &models.DeviceModel{
		Vendor:            newVendor,
		Model:             newModel,
		DeviceModelType:   src.DeviceModelType,
		PortCount:         src.PortCount,
		HeightU:           src.HeightU,
		PowerWattsIdle:    src.PowerWattsIdle,
		PowerWattsTypical: src.PowerWattsTypical,
		PowerWattsMax:     src.PowerWattsMax,
		CPUSockets:        src.CPUSockets,
		CoresPerSocket:    src.CoresPerSocket,
		RAMGB:             src.RAMGB,
		StorageTB:         src.StorageTB,
		GPUCount:          src.GPUCount,
		Description:       src.Description,
		IsSeed:            false,
		ArchivedAt:        nil,
	}

	// Ensure the new name is a copy marker if defaults were not supplied.
	if copy.Vendor == "" {
		copy.Vendor = src.Vendor
	}
	if copy.Model == "" {
		copy.Model = src.Model + " (copy)"
	}

	// Add a timestamp suffix when the default copy name would collide.
	copy.Model = uniqueName(copy.Model, func(name string) bool {
		const check = `SELECT 1 FROM device_models WHERE vendor = ? AND model = ? LIMIT 1`
		var v int
		return s.db.QueryRow(check, copy.Vendor, name).Scan(&v) == nil
	})

	out, err := s.Create(copy)
	if err != nil {
		return nil, fmt.Errorf("duplicate device model %d: %w", sourceID, err)
	}
	return out, nil
}

// isDuplicateErr reports whether err is a SQLite UNIQUE constraint violation.
func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// uniqueName returns name if it does not exist according to exists, otherwise
// appends an incrementing timestamp-based suffix until a free name is found.
func uniqueName(name string, exists func(string) bool) string {
	if !exists(name) {
		return name
	}
	suffix := time.Now().Format("20060102-150405")
	candidate := name + " " + suffix
	if !exists(candidate) {
		return candidate
	}
	// Last resort: append a counter.
	for i := 2; ; i++ {
		c := fmt.Sprintf("%s %d", candidate, i)
		if !exists(c) {
			return c
		}
	}
}
