package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// CapacityStore provides power and resource capacity aggregation queries.
type CapacityStore struct {
	db *sql.DB
}

// NewCapacityStore returns a new CapacityStore backed by db.
func NewCapacityStore(db *sql.DB) *CapacityStore {
	return &CapacityStore{db: db}
}

// aggregateRow is the shared column set returned by all capacity queries.
// SUM returns NULL when there are no rows; COALESCE ensures 0 is returned instead.
const aggregateCols = `
	COALESCE(SUM(dm.power_watts_idle),    0) AS power_watts_idle,
	COALESCE(SUM(dm.power_watts_typical), 0) AS power_watts_typical,
	COALESCE(SUM(dm.power_watts_max),     0) AS power_watts_max,
	COALESCE(SUM(dm.cpu_sockets),         0) AS cpu_sockets,
	COALESCE(SUM(dm.cores_per_socket),    0) AS cores_per_socket,
	COALESCE(SUM(dm.ram_gb),              0) AS ram_gb,
	COALESCE(SUM(dm.storage_tb),          0) AS storage_tb,
	COALESCE(SUM(dm.gpu_count),           0) AS gpu_count,
	COUNT(d.id)                              AS device_count`

// scanCapacity scans the aggregate columns from a *sql.Row into a CapacitySummary.
// vCPU is computed as cpu_sockets * cores_per_socket after the scan.
func scanCapacity(row *sql.Row, level models.CapacityLevel, id int64, name string) (*models.CapacitySummary, error) {
	c := &models.CapacitySummary{
		Level: level,
		ID:    id,
		Name:  name,
	}
	var cpuSockets, coresPerSocket int
	err := row.Scan(
		&c.PowerWattsIdle, &c.PowerWattsTypical, &c.PowerWattsMax,
		&cpuSockets, &coresPerSocket,
		&c.TotalRAMGB, &c.TotalStorageTB, &c.TotalGPUCount,
		&c.DeviceCount,
	)
	if err != nil {
		return nil, err
	}
	c.TotalVCPU = cpuSockets * coresPerSocket
	return c, nil
}

// QueryRackCapacity returns aggregated capacity for all devices in the given rack.
func (s *CapacityStore) QueryRackCapacity(rackID int64) (*models.CapacitySummary, error) {
	// Verify the rack exists.
	var rackName string
	if err := s.db.QueryRow(`SELECT name FROM racks WHERE id = ?`, rackID).Scan(&rackName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("lookup rack %d: %w", rackID, err)
	}

	q := `SELECT` + aggregateCols + `
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		WHERE d.rack_id = ?`

	c, err := scanCapacity(s.db.QueryRow(q, rackID), models.CapacityLevelRack, rackID, rackName)
	if err != nil {
		return nil, fmt.Errorf("query rack capacity %d: %w", rackID, err)
	}
	return c, nil
}

// QueryBlockCapacity returns aggregated capacity for all devices in the given block.
func (s *CapacityStore) QueryBlockCapacity(blockID int64) (*models.CapacitySummary, error) {
	var blockName string
	if err := s.db.QueryRow(`SELECT name FROM blocks WHERE id = ?`, blockID).Scan(&blockName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("lookup block %d: %w", blockID, err)
	}

	q := `SELECT` + aggregateCols + `
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		JOIN racks r ON d.rack_id = r.id
		WHERE r.block_id = ?`

	c, err := scanCapacity(s.db.QueryRow(q, blockID), models.CapacityLevelBlock, blockID, blockName)
	if err != nil {
		return nil, fmt.Errorf("query block capacity %d: %w", blockID, err)
	}
	return c, nil
}

// QuerySuperBlockCapacity returns aggregated capacity for all devices in the given super-block.
func (s *CapacityStore) QuerySuperBlockCapacity(superBlockID int64) (*models.CapacitySummary, error) {
	var name string
	if err := s.db.QueryRow(`SELECT name FROM super_blocks WHERE id = ?`, superBlockID).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("lookup super-block %d: %w", superBlockID, err)
	}

	q := `SELECT` + aggregateCols + `
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		JOIN racks r ON d.rack_id = r.id
		JOIN blocks b ON r.block_id = b.id
		WHERE b.super_block_id = ?`

	c, err := scanCapacity(s.db.QueryRow(q, superBlockID), models.CapacityLevelSuperBlock, superBlockID, name)
	if err != nil {
		return nil, fmt.Errorf("query super-block capacity %d: %w", superBlockID, err)
	}
	return c, nil
}

// QuerySiteCapacity returns aggregated capacity for all devices in the given site.
func (s *CapacityStore) QuerySiteCapacity(siteID int64) (*models.CapacitySummary, error) {
	var name string
	if err := s.db.QueryRow(`SELECT name FROM sites WHERE id = ?`, siteID).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("lookup site %d: %w", siteID, err)
	}

	q := `SELECT` + aggregateCols + `
		FROM devices d
		JOIN device_models dm ON d.device_model_id = dm.id
		JOIN racks r ON d.rack_id = r.id
		JOIN blocks b ON r.block_id = b.id
		JOIN super_blocks sb ON b.super_block_id = sb.id
		WHERE sb.site_id = ?`

	c, err := scanCapacity(s.db.QueryRow(q, siteID), models.CapacityLevelSite, siteID, name)
	if err != nil {
		return nil, fmt.Errorf("query site capacity %d: %w", siteID, err)
	}
	return c, nil
}

// QueryDesignCapacity returns aggregated capacity for all devices in the given design.
func (s *CapacityStore) QueryDesignCapacity(designID int64) (*models.CapacitySummary, error) {
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
