package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// BlockAggregationStore provides CRUD operations for BlockAggregation records.
type BlockAggregationStore struct {
	db *sql.DB
}

// NewBlockAggregationStore returns a new BlockAggregationStore backed by db.
func NewBlockAggregationStore(db *sql.DB) *BlockAggregationStore {
	return &BlockAggregationStore{db: db}
}

// Upsert inserts or updates a block aggregation assignment for the given block and plane.
// If an aggregation record already exists for (block_id, plane), it is updated.
func (s *BlockAggregationStore) Upsert(agg *models.BlockAggregation) (*models.BlockAggregation, error) {
	const q = `
		INSERT INTO block_aggregations (block_id, plane, device_id, max_ports, used_ports, description)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(block_id, plane) DO UPDATE SET
			device_id   = excluded.device_id,
			max_ports   = excluded.max_ports,
			description = excluded.description,
			updated_at  = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		RETURNING id, block_id, plane, device_id, max_ports, used_ports, description, created_at, updated_at`

	out := &models.BlockAggregation{}
	err := s.db.QueryRow(q, agg.BlockID, string(agg.Plane), agg.DeviceID, agg.MaxPorts, agg.UsedPorts, agg.Description).
		Scan(&out.ID, &out.BlockID, &out.Plane, &out.DeviceID, &out.MaxPorts, &out.UsedPorts, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert block aggregation: %w", err)
	}
	return out, nil
}

// Get returns the block aggregation for the given block and plane.
func (s *BlockAggregationStore) Get(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error) {
	const q = `
		SELECT id, block_id, plane, device_id, max_ports, used_ports, description, created_at, updated_at
		FROM block_aggregations
		WHERE block_id = ? AND plane = ?`

	out := &models.BlockAggregation{}
	err := s.db.QueryRow(q, blockID, string(plane)).
		Scan(&out.ID, &out.BlockID, &out.Plane, &out.DeviceID, &out.MaxPorts, &out.UsedPorts, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: block aggregation for block %d plane %s", models.ErrNotFound, blockID, plane)
	}
	if err != nil {
		return nil, fmt.Errorf("get block aggregation: %w", err)
	}
	return out, nil
}

// ListForBlock returns all aggregation records for the given block.
func (s *BlockAggregationStore) ListForBlock(blockID int64) ([]*models.BlockAggregation, error) {
	const q = `
		SELECT id, block_id, plane, device_id, max_ports, used_ports, description, created_at, updated_at
		FROM block_aggregations
		WHERE block_id = ?
		ORDER BY plane`

	rows, err := s.db.Query(q, blockID)
	if err != nil {
		return nil, fmt.Errorf("list block aggregations: %w", err)
	}
	defer rows.Close()

	var out []*models.BlockAggregation
	for rows.Next() {
		a := &models.BlockAggregation{}
		if err := rows.Scan(&a.ID, &a.BlockID, &a.Plane, &a.DeviceID, &a.MaxPorts, &a.UsedPorts, &a.Description, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan block aggregation: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list block aggregations rows: %w", err)
	}
	return out, nil
}

// Delete removes the aggregation record for the given block and plane.
func (s *BlockAggregationStore) Delete(blockID int64, plane models.NetworkPlane) error {
	const q = `DELETE FROM block_aggregations WHERE block_id = ? AND plane = ?`
	res, err := s.db.Exec(q, blockID, string(plane))
	if err != nil {
		return fmt.Errorf("delete block aggregation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("block aggregation delete rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("%w: block aggregation for block %d plane %s", models.ErrNotFound, blockID, plane)
	}
	return nil
}

// IncrementUsedPorts atomically increments used_ports for the given block+plane.
// Returns ErrNotFound if no record exists.
// Returns an error string embedding the port capacity if the increment would exceed max_ports.
func (s *BlockAggregationStore) IncrementUsedPorts(blockID int64, plane models.NetworkPlane, delta int) (*models.BlockAggregation, error) {
	agg, err := s.Get(blockID, plane)
	if err != nil {
		return nil, err
	}

	newUsed := agg.UsedPorts + delta
	if agg.MaxPorts > 0 && newUsed > agg.MaxPorts {
		return nil, fmt.Errorf("management agg port capacity exceeded: %d used + %d = %d > %d max",
			agg.UsedPorts, delta, newUsed, agg.MaxPorts)
	}

	const q = `
		UPDATE block_aggregations SET used_ports = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE block_id = ? AND plane = ?`
	if _, err := s.db.Exec(q, newUsed, blockID, string(plane)); err != nil {
		return nil, fmt.Errorf("increment used ports: %w", err)
	}

	agg.UsedPorts = newUsed
	return agg, nil
}

// DecrementUsedPorts atomically decrements used_ports, flooring at 0.
func (s *BlockAggregationStore) DecrementUsedPorts(blockID int64, plane models.NetworkPlane, delta int) (*models.BlockAggregation, error) {
	agg, err := s.Get(blockID, plane)
	if err != nil {
		return nil, err
	}

	newUsed := agg.UsedPorts - delta
	if newUsed < 0 {
		newUsed = 0
	}

	const q = `
		UPDATE block_aggregations SET used_ports = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE block_id = ? AND plane = ?`
	if _, err := s.db.Exec(q, newUsed, blockID, string(plane)); err != nil {
		return nil, fmt.Errorf("decrement used ports: %w", err)
	}

	agg.UsedPorts = newUsed
	return agg, nil
}
