package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// BlockStore provides CRUD operations for Block records and device/rack helpers used by block service.
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

// GetDeviceModel returns the DeviceModel with the given id (used for port_count lookups).
func (s *BlockStore) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	const q = `
		SELECT id, vendor, model, port_count, height_u,
		       power_watts_typical, description, created_at, updated_at
		FROM device_models WHERE id = ?`

	dm := &models.DeviceModel{}
	err := s.db.QueryRow(q, id).
		Scan(&dm.ID, &dm.Vendor, &dm.Model, &dm.PortCount, &dm.HeightU,
			&dm.PowerWattsTypical, &dm.Description, &dm.CreatedAt, &dm.UpdatedAt)
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

// CreateRack inserts a new Rack and returns the saved record.
func (s *BlockStore) CreateRack(r *models.Rack) (*models.Rack, error) {
	const q = `
		INSERT INTO racks (block_id, rack_type_id, name, height_u, power_capacity_w,
		                   power_oversub_pct_warn, power_oversub_pct_max, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, block_id, rack_type_id, name, height_u, power_capacity_w,
		          power_oversub_pct_warn, power_oversub_pct_max, description, created_at, updated_at`

	out := &models.Rack{}
	err := s.db.QueryRow(q,
		r.BlockID, r.RackTypeID, r.Name, r.HeightU, r.PowerCapacityW,
		r.PowerOversubPctWarn, r.PowerOversubPctMax, r.Description,
	).Scan(&out.ID, &out.BlockID, &out.RackTypeID, &out.Name, &out.HeightU, &out.PowerCapacityW,
		&out.PowerOversubPctWarn, &out.PowerOversubPctMax, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create rack: %w", err)
	}
	return out, nil
}

// PlaceDevice inserts a device into a rack at the given position.
func (s *BlockStore) PlaceDevice(d *models.Device) (*models.Device, error) {
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
