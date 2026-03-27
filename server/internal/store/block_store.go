package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// BlockStore provides CRUD operations for Block, BlockAggregation, and PortConnection records.
type BlockStore struct {
	db *sql.DB
}

// NewBlockStore returns a new BlockStore backed by db.
func NewBlockStore(db *sql.DB) *BlockStore {
	return &BlockStore{db: db}
}

// --- Block CRUD ---

// CreateBlock inserts a new Block and returns the saved record.
func (s *BlockStore) CreateBlock(b *models.Block) (*models.Block, error) {
	const q = `
		INSERT INTO blocks (super_block_id, name, description)
		VALUES (?, ?, ?)
		RETURNING id, super_block_id, name, description, created_at, updated_at`

	out := &models.Block{}
	err := s.db.QueryRow(q, b.SuperBlockID, b.Name, b.Description).
		Scan(&out.ID, &out.SuperBlockID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create block: %w", err)
	}
	return out, nil
}

// GetBlock returns the Block with the given id, or models.ErrNotFound.
func (s *BlockStore) GetBlock(id int64) (*models.Block, error) {
	const q = `
		SELECT id, super_block_id, name, description, created_at, updated_at
		FROM blocks WHERE id = ?`

	out := &models.Block{}
	err := s.db.QueryRow(q, id).
		Scan(&out.ID, &out.SuperBlockID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get block %d: %w", id, err)
	}
	return out, nil
}

// ListBlocks returns all blocks for a given super_block_id, ordered by id.
func (s *BlockStore) ListBlocks(superBlockID int64) ([]*models.Block, error) {
	const q = `
		SELECT id, super_block_id, name, description, created_at, updated_at
		FROM blocks WHERE super_block_id = ? ORDER BY id`

	rows, err := s.db.Query(q, superBlockID)
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}
	defer rows.Close()

	var out []*models.Block
	for rows.Next() {
		b := &models.Block{}
		if err := rows.Scan(&b.ID, &b.SuperBlockID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate blocks: %w", err)
	}
	return out, nil
}

// GetDefaultBlock returns the default block (named "default") for a super-block, or nil if none.
func (s *BlockStore) GetDefaultBlock(superBlockID int64) (*models.Block, error) {
	const q = `
		SELECT id, super_block_id, name, description, created_at, updated_at
		FROM blocks WHERE super_block_id = ? AND name = 'default' LIMIT 1`

	out := &models.Block{}
	err := s.db.QueryRow(q, superBlockID).
		Scan(&out.ID, &out.SuperBlockID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get default block: %w", err)
	}
	return out, nil
}

// --- BlockAggregation CRUD ---

// SetAggregation upserts a BlockAggregation for (blockID, plane).
// If one already exists, it is replaced with the new device_model_id.
func (s *BlockStore) SetAggregation(agg *models.BlockAggregation) (*models.BlockAggregation, error) {
	const q = `
		INSERT INTO block_aggregations (block_id, plane, device_model_id)
		VALUES (?, ?, ?)
		ON CONFLICT(block_id, plane) DO UPDATE SET
			device_model_id = excluded.device_model_id,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		RETURNING id, block_id, plane, device_model_id, created_at, updated_at`

	out := &models.BlockAggregation{}
	err := s.db.QueryRow(q, agg.BlockID, agg.Plane, agg.DeviceModelID).
		Scan(&out.ID, &out.BlockID, &out.Plane, &out.DeviceModelID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("set aggregation: %w", err)
	}
	return out, nil
}

// GetAggregation returns the BlockAggregation for (blockID, plane), or models.ErrNotFound.
func (s *BlockStore) GetAggregation(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error) {
	const q = `
		SELECT id, block_id, plane, device_model_id, created_at, updated_at
		FROM block_aggregations WHERE block_id = ? AND plane = ?`

	out := &models.BlockAggregation{}
	err := s.db.QueryRow(q, blockID, plane).
		Scan(&out.ID, &out.BlockID, &out.Plane, &out.DeviceModelID, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get aggregation: %w", err)
	}
	return out, nil
}

// ListAggregations returns all BlockAggregation records for a block.
func (s *BlockStore) ListAggregations(blockID int64) ([]*models.BlockAggregation, error) {
	const q = `
		SELECT id, block_id, plane, device_model_id, created_at, updated_at
		FROM block_aggregations WHERE block_id = ? ORDER BY plane`

	rows, err := s.db.Query(q, blockID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations: %w", err)
	}
	defer rows.Close()

	var out []*models.BlockAggregation
	for rows.Next() {
		a := &models.BlockAggregation{}
		if err := rows.Scan(&a.ID, &a.BlockID, &a.Plane, &a.DeviceModelID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan aggregation: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate aggregations: %w", err)
	}
	return out, nil
}

// DeleteAggregation removes the BlockAggregation for (blockID, plane).
// Associated port connections are removed via CASCADE.
func (s *BlockStore) DeleteAggregation(blockID int64, plane models.NetworkPlane) error {
	result, err := s.db.Exec(
		`DELETE FROM block_aggregations WHERE block_id = ? AND plane = ?`, blockID, plane)
	if err != nil {
		return fmt.Errorf("delete aggregation: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete aggregation rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// --- PortConnection operations ---

// AllocatePorts inserts port connections for a rack against a block aggregation.
// Each call allocates one port per leaf device name provided.
// Returns the newly created connections.
func (s *BlockStore) AllocatePorts(aggID, rackID int64, leafNames []string, startPortIndex int) ([]*models.PortConnection, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
		INSERT INTO port_connections (block_aggregation_id, rack_id, agg_port_index, leaf_device_name)
		VALUES (?, ?, ?, ?)
		RETURNING id, block_aggregation_id, rack_id, agg_port_index, leaf_device_name, created_at`

	var out []*models.PortConnection
	for i, name := range leafNames {
		pc := &models.PortConnection{}
		err := tx.QueryRow(q, aggID, rackID, startPortIndex+i, name).
			Scan(&pc.ID, &pc.BlockAggregationID, &pc.RackID, &pc.AggPortIndex, &pc.LeafDeviceName, &pc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("allocate port %d: %w", startPortIndex+i, err)
		}
		out = append(out, pc)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit port allocation: %w", err)
	}
	return out, nil
}

// DeallocatePorts removes all port connections for a given (aggID, rackID) pair.
func (s *BlockStore) DeallocatePorts(aggID, rackID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM port_connections WHERE block_aggregation_id = ? AND rack_id = ?`,
		aggID, rackID)
	if err != nil {
		return fmt.Errorf("deallocate ports for rack %d agg %d: %w", rackID, aggID, err)
	}
	return nil
}

// DeallocatePortsByRack removes all port connections for a rack across all aggregations.
func (s *BlockStore) DeallocatePortsByRack(rackID int64) error {
	_, err := s.db.Exec(`DELETE FROM port_connections WHERE rack_id = ?`, rackID)
	if err != nil {
		return fmt.Errorf("deallocate ports for rack %d: %w", rackID, err)
	}
	return nil
}

// CountAllocatedPorts returns the number of ports already allocated on a block aggregation.
func (s *BlockStore) CountAllocatedPorts(aggID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM port_connections WHERE block_aggregation_id = ?`, aggID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count allocated ports for agg %d: %w", aggID, err)
	}
	return count, nil
}

// ListPortConnections returns all PortConnections for a block aggregation, ordered by agg_port_index.
func (s *BlockStore) ListPortConnections(aggID int64) ([]*models.PortConnection, error) {
	const q = `
		SELECT id, block_aggregation_id, rack_id, agg_port_index, leaf_device_name, created_at
		FROM port_connections
		WHERE block_aggregation_id = ?
		ORDER BY agg_port_index`

	rows, err := s.db.Query(q, aggID)
	if err != nil {
		return nil, fmt.Errorf("list port connections: %w", err)
	}
	defer rows.Close()

	var out []*models.PortConnection
	for rows.Next() {
		pc := &models.PortConnection{}
		if err := rows.Scan(&pc.ID, &pc.BlockAggregationID, &pc.RackID, &pc.AggPortIndex, &pc.LeafDeviceName, &pc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan port connection: %w", err)
		}
		out = append(out, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate port connections: %w", err)
	}
	return out, nil
}

// ListPortConnectionsByRack returns all PortConnections for a specific rack within a block aggregation.
func (s *BlockStore) ListPortConnectionsByRack(aggID, rackID int64) ([]*models.PortConnection, error) {
	const q = `
		SELECT id, block_aggregation_id, rack_id, agg_port_index, leaf_device_name, created_at
		FROM port_connections
		WHERE block_aggregation_id = ? AND rack_id = ?
		ORDER BY agg_port_index`

	rows, err := s.db.Query(q, aggID, rackID)
	if err != nil {
		return nil, fmt.Errorf("list port connections for rack: %w", err)
	}
	defer rows.Close()

	var out []*models.PortConnection
	for rows.Next() {
		pc := &models.PortConnection{}
		if err := rows.Scan(&pc.ID, &pc.BlockAggregationID, &pc.RackID, &pc.AggPortIndex, &pc.LeafDeviceName, &pc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan port connection: %w", err)
		}
		out = append(out, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate port connections: %w", err)
	}
	return out, nil
}

// GetDeviceModel returns the DeviceModel with the given id (used for port_count lookups).
func (s *BlockStore) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, port_count, height_u, power_watts, description, created_at, updated_at
		FROM device_models WHERE id = ?`

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

// ListDevicesInRack returns all device records in a rack (for leaf counting).
func (s *BlockStore) ListDevicesInRack(rackID int64) ([]*models.Device, error) {
	const q = `
		SELECT id, rack_id, device_model_id, name, role, position, description, created_at, updated_at
		FROM devices WHERE rack_id = ? ORDER BY position`

	rows, err := s.db.Query(q, rackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack %d: %w", rackID, err)
	}
	defer rows.Close()

	var out []*models.Device
	for rows.Next() {
		d := &models.Device{}
		if err := rows.Scan(&d.ID, &d.RackID, &d.DeviceModelID, &d.Name, &d.Role, &d.Position, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate devices: %w", err)
	}
	return out, nil
}

// UpdateRackBlock sets the block_id for a rack.
func (s *BlockStore) UpdateRackBlock(rackID int64, blockID *int64) error {
	const q = `
		UPDATE racks SET block_id = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?`
	result, err := s.db.Exec(q, blockID, rackID)
	if err != nil {
		return fmt.Errorf("update rack block for rack %d: %w", rackID, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update rack block rows affected: %w", err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// GetRack returns the Rack with the given id, or models.ErrNotFound.
func (s *BlockStore) GetRack(id int64) (*models.Rack, error) {
	const q = `
		SELECT id, block_id, rack_type_id, name, height_u, power_capacity_w, description, created_at, updated_at
		FROM racks WHERE id = ?`

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
