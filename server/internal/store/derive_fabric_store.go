package store

import (
	"database/sql"
	"errors"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DeriveFabricStore provides read-only hierarchy queries for the DeriveFabric service.
type DeriveFabricStore struct {
	db *sql.DB
}

// NewDeriveFabricStore returns a new DeriveFabricStore backed by db.
func NewDeriveFabricStore(db *sql.DB) *DeriveFabricStore {
	return &DeriveFabricStore{db: db}
}

// GetDesign returns a design by ID.
func (s *DeriveFabricStore) GetDesign(id int64) (*models.Design, error) {
	d := &models.Design{}
	err := s.db.QueryRow(
		`SELECT id, name, description, created_at, updated_at FROM designs WHERE id = ?`, id,
	).Scan(&d.ID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	return d, err
}

// ListSitesByDesign returns all sites for a design, ordered by id.
func (s *DeriveFabricStore) ListSitesByDesign(designID int64) ([]*models.Site, error) {
	rows, err := s.db.Query(
		`SELECT id, design_id, name, description, created_at, updated_at
		 FROM sites WHERE design_id = ? ORDER BY id`, designID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Site
	for rows.Next() {
		s := &models.Site{}
		if err := rows.Scan(&s.ID, &s.DesignID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ListSuperBlocksBySite returns all super_blocks for a site, ordered by id.
func (s *DeriveFabricStore) ListSuperBlocksBySite(siteID int64) ([]*models.SuperBlock, error) {
	rows, err := s.db.Query(
		`SELECT id, site_id, name, description, created_at, updated_at
		 FROM super_blocks WHERE site_id = ? ORDER BY id`, siteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.SuperBlock
	for rows.Next() {
		sb := &models.SuperBlock{}
		if err := rows.Scan(&sb.ID, &sb.SiteID, &sb.Name, &sb.Description, &sb.CreatedAt, &sb.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, sb)
	}
	return out, rows.Err()
}

// ListBlocksBySuperBlock returns all blocks for a super_block, ordered by id.
func (s *DeriveFabricStore) ListBlocksBySuperBlock(superBlockID int64) ([]*models.Block, error) {
	rows, err := s.db.Query(
		`SELECT id, super_block_id, name, description, created_at, updated_at
		 FROM blocks WHERE super_block_id = ? ORDER BY id`, superBlockID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Block
	for rows.Next() {
		b := &models.Block{}
		if err := rows.Scan(&b.ID, &b.SuperBlockID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetAggregation returns the TierAggregation for a given scope + plane, or ErrNotFound.
func (s *DeriveFabricStore) GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error) {
	a := &models.TierAggregation{}
	err := s.db.QueryRow(
		`SELECT id, scope_type, scope_id, plane, device_model_id, spine_count, created_at, updated_at
		 FROM tier_aggregations WHERE scope_type = ? AND scope_id = ? AND plane = ?`,
		scopeType, scopeID, plane,
	).Scan(&a.ID, &a.ScopeType, &a.ScopeID, &a.Plane, &a.DeviceModelID, &a.SpineCount, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	return a, err
}

// CountAllocatedPorts returns how many tier_port_connections exist for an aggregation.
func (s *DeriveFabricStore) CountAllocatedPorts(aggID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tier_port_connections WHERE tier_aggregation_id = ?`, aggID).Scan(&n)
	return n, err
}

// GetDeviceModel returns a device model by ID.
func (s *DeriveFabricStore) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm := &models.DeviceModel{}
	err := s.db.QueryRow(
		`SELECT id, vendor, model, device_model_type, port_count, height_u,
		        power_watts_idle, power_watts_typical, power_watts_max,
		        cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count,
		        description, is_seed, archived_at, created_at, updated_at
		 FROM device_models WHERE id = ?`, id,
	).Scan(
		&dm.ID, &dm.Vendor, &dm.Model, &dm.DeviceModelType, &dm.PortCount, &dm.HeightU,
		&dm.PowerWattsIdle, &dm.PowerWattsTypical, &dm.PowerWattsMax,
		&dm.CPUSockets, &dm.CoresPerSocket, &dm.RAMGB, &dm.StorageTB, &dm.GPUCount,
		&dm.Description, &dm.IsSeed, &dm.ArchivedAt, &dm.CreatedAt, &dm.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	return dm, err
}
