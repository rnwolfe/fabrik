package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// MetricsStore provides queries for the metrics service.
type MetricsStore struct {
	db *sql.DB
}

// NewMetricsStore returns a new MetricsStore backed by db.
func NewMetricsStore(db *sql.DB) *MetricsStore {
	return &MetricsStore{db: db}
}

// GetDesignName returns the design name for the given ID, or models.ErrNotFound.
func (s *MetricsStore) GetDesignName(designID int64) (string, error) {
	var name string
	err := s.db.QueryRow(`SELECT name FROM designs WHERE id = ?`, designID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", models.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup design %d: %w", designID, err)
	}
	return name, nil
}

// ListFabricsByDesign returns all FabricRecord rows belonging to the given design.
func (s *MetricsStore) ListFabricsByDesign(designID int64) ([]*FabricRecord, error) {
	const q = `
		SELECT id, design_id, name, tier, stages, radix, oversubscription,
		       leaf_model_id, spine_model_id, super_spine_model_id, description,
		       created_at, updated_at
		FROM fabrics
		WHERE design_id = ?
		ORDER BY id`

	rows, err := s.db.Query(q, designID)
	if err != nil {
		return nil, fmt.Errorf("list fabrics for design %d: %w", designID, err)
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

// GetDeviceModelByID returns the device model for the given ID, or models.ErrNotFound.
func (s *MetricsStore) GetDeviceModelByID(id int64) (*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, device_model_type, port_count, height_u,
		       power_watts_idle, power_watts_typical, power_watts_max,
		       cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
		       description, is_seed, archived_at, created_at, updated_at
		FROM device_models
		WHERE id = ?`

	dm := &models.DeviceModel{}
	err := s.db.QueryRow(q, id).Scan(
		&dm.ID, &dm.Vendor, &dm.Model, &dm.DeviceModelType, &dm.PortCount, &dm.HeightU,
		&dm.PowerWattsIdle, &dm.PowerWattsTypical, &dm.PowerWattsMax,
		&dm.CPUSockets, &dm.CoresPerSocket, &dm.RAMGB, &dm.StorageTB, &dm.GPUCount,
		&dm.Description, &dm.IsSeed, &dm.ArchivedAt, &dm.CreatedAt, &dm.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", id, err)
	}
	return dm, nil
}

// QueryDesignCapacity returns aggregated resource capacity for all devices in the design.
// It delegates to the shared aggregateCols defined in capacity_store.go.
func (s *MetricsStore) QueryDesignCapacity(designID int64) (*models.CapacitySummary, error) {
	var name string
	if err := s.db.QueryRow(`SELECT name FROM designs WHERE id = ?`, designID).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("lookup design %d: %w", designID, err)
	}

	q := `SELECT` + aggregateCols + `
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		JOIN racks r ON d.rack_id = r.id
		JOIN blocks b ON r.block_id = b.id
		JOIN super_blocks sb ON b.super_block_id = sb.id
		JOIN sites s ON sb.site_id = s.id
		WHERE s.design_id = ?`

	c, err := scanCapacity(s.db.QueryRow(q, designID), models.CapacityLevelDesign, designID, name)
	if err != nil {
		return nil, fmt.Errorf("query design capacity %d: %w", designID, err)
	}
	return c, nil
}

// QueryDesignPowerAndRacks returns the total device power draw (typical watts)
// and the total rack power capacity (watts) for all racks belonging to the design.
func (s *MetricsStore) QueryDesignPowerAndRacks(designID int64) (totalDrawW int, totalRackCapacityW int, err error) {
	// Sum device power draw via the design hierarchy.
	drawQ := `
		SELECT COALESCE(SUM(dm.power_watts_typical), 0)
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		JOIN racks r ON d.rack_id = r.id
		JOIN blocks b ON r.block_id = b.id
		JOIN super_blocks sb ON b.super_block_id = sb.id
		JOIN sites s ON sb.site_id = s.id
		WHERE s.design_id = ?`

	if err = s.db.QueryRow(drawQ, designID).Scan(&totalDrawW); err != nil {
		return 0, 0, fmt.Errorf("query device power draw for design %d: %w", designID, err)
	}

	// Sum rack power capacity via the design hierarchy.
	capacityQ := `
		SELECT COALESCE(SUM(r.power_capacity_w), 0)
		FROM racks r
		JOIN blocks b ON r.block_id = b.id
		JOIN super_blocks sb ON b.super_block_id = sb.id
		JOIN sites s ON sb.site_id = s.id
		WHERE s.design_id = ?`

	if err = s.db.QueryRow(capacityQ, designID).Scan(&totalRackCapacityW); err != nil {
		return 0, 0, fmt.Errorf("query rack capacity for design %d: %w", designID, err)
	}

	return totalDrawW, totalRackCapacityW, nil
}
