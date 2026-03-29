package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// TierAggregationStore provides CRUD operations for TierAggregation and TierPortConnection records.
type TierAggregationStore struct {
	db *sql.DB
}

// NewTierAggregationStore returns a new TierAggregationStore backed by db.
func NewTierAggregationStore(db *sql.DB) *TierAggregationStore {
	return &TierAggregationStore{db: db}
}

// --- TierAggregation CRUD ---

// SetAggregation upserts a TierAggregation for (scope_type, scope_id, plane).
func (s *TierAggregationStore) SetAggregation(agg *models.TierAggregation) (*models.TierAggregation, error) {
	const q = `
		INSERT INTO tier_aggregations (scope_type, scope_id, plane, device_model_id, spine_count)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(scope_type, scope_id, plane) DO UPDATE SET
			device_model_id = excluded.device_model_id,
			spine_count = excluded.spine_count,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		RETURNING id, scope_type, scope_id, plane, device_model_id, spine_count, created_at, updated_at`

	out := &models.TierAggregation{}
	err := s.db.QueryRow(q, agg.ScopeType, agg.ScopeID, agg.Plane, agg.DeviceModelID, agg.SpineCount).
		Scan(&out.ID, &out.ScopeType, &out.ScopeID, &out.Plane, &out.DeviceModelID, &out.SpineCount, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("set aggregation: %w", err)
	}
	return out, nil
}

// GetAggregation returns the TierAggregation for (scopeType, scopeID, plane), or models.ErrNotFound.
func (s *TierAggregationStore) GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error) {
	const q = `
		SELECT id, scope_type, scope_id, plane, device_model_id, spine_count, created_at, updated_at
		FROM tier_aggregations WHERE scope_type = ? AND scope_id = ? AND plane = ?`

	out := &models.TierAggregation{}
	err := s.db.QueryRow(q, scopeType, scopeID, plane).
		Scan(&out.ID, &out.ScopeType, &out.ScopeID, &out.Plane, &out.DeviceModelID, &out.SpineCount, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get aggregation: %w", err)
	}
	return out, nil
}

// ListAggregations returns all TierAggregation records for a given scope.
func (s *TierAggregationStore) ListAggregations(scopeType models.AggregationScope, scopeID int64) ([]*models.TierAggregation, error) {
	const q = `
		SELECT id, scope_type, scope_id, plane, device_model_id, spine_count, created_at, updated_at
		FROM tier_aggregations WHERE scope_type = ? AND scope_id = ? ORDER BY plane`

	rows, err := s.db.Query(q, scopeType, scopeID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations: %w", err)
	}
	defer rows.Close()

	var out []*models.TierAggregation
	for rows.Next() {
		a := &models.TierAggregation{}
		if err := rows.Scan(&a.ID, &a.ScopeType, &a.ScopeID, &a.Plane, &a.DeviceModelID, &a.SpineCount, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan aggregation: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate aggregations: %w", err)
	}
	return out, nil
}

// DeleteAggregation removes the TierAggregation for (scopeType, scopeID, plane).
// Associated port connections are removed via CASCADE.
func (s *TierAggregationStore) DeleteAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) error {
	result, err := s.db.Exec(
		`DELETE FROM tier_aggregations WHERE scope_type = ? AND scope_id = ? AND plane = ?`,
		scopeType, scopeID, plane)
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

// --- TierPortConnection operations ---

// AllocatePorts inserts port connections for a child entity against a tier aggregation.
// Each call allocates one port per child device name provided.
func (s *TierAggregationStore) AllocatePorts(aggID, childID int64, childNames []string, startPortIndex int) ([]*models.TierPortConnection, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
		INSERT INTO tier_port_connections (tier_aggregation_id, child_id, agg_port_index, child_device_name)
		VALUES (?, ?, ?, ?)
		RETURNING id, tier_aggregation_id, child_id, agg_port_index, child_device_name, created_at`

	var out []*models.TierPortConnection
	for i, name := range childNames {
		pc := &models.TierPortConnection{}
		err := tx.QueryRow(q, aggID, childID, startPortIndex+i, name).
			Scan(&pc.ID, &pc.TierAggregationID, &pc.ChildID, &pc.AggPortIndex, &pc.ChildDeviceName, &pc.CreatedAt)
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

// DeallocatePorts removes all port connections for a given (aggID, childID) pair.
func (s *TierAggregationStore) DeallocatePorts(aggID, childID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM tier_port_connections WHERE tier_aggregation_id = ? AND child_id = ?`,
		aggID, childID)
	if err != nil {
		return fmt.Errorf("deallocate ports for child %d agg %d: %w", childID, aggID, err)
	}
	return nil
}

// DeallocatePortsByChild removes all port connections for a child across all aggregations.
func (s *TierAggregationStore) DeallocatePortsByChild(childID int64) error {
	_, err := s.db.Exec(`DELETE FROM tier_port_connections WHERE child_id = ?`, childID)
	if err != nil {
		return fmt.Errorf("deallocate ports for child %d: %w", childID, err)
	}
	return nil
}

// CountAllocatedPorts returns the number of ports already allocated on a tier aggregation.
func (s *TierAggregationStore) CountAllocatedPorts(aggID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tier_port_connections WHERE tier_aggregation_id = ?`, aggID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count allocated ports for agg %d: %w", aggID, err)
	}
	return count, nil
}

// ListPortConnections returns all TierPortConnections for a tier aggregation, ordered by agg_port_index.
func (s *TierAggregationStore) ListPortConnections(aggID int64) ([]*models.TierPortConnection, error) {
	const q = `
		SELECT id, tier_aggregation_id, child_id, agg_port_index, child_device_name, created_at
		FROM tier_port_connections
		WHERE tier_aggregation_id = ?
		ORDER BY agg_port_index`

	rows, err := s.db.Query(q, aggID)
	if err != nil {
		return nil, fmt.Errorf("list port connections: %w", err)
	}
	defer rows.Close()

	var out []*models.TierPortConnection
	for rows.Next() {
		pc := &models.TierPortConnection{}
		if err := rows.Scan(&pc.ID, &pc.TierAggregationID, &pc.ChildID, &pc.AggPortIndex, &pc.ChildDeviceName, &pc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan port connection: %w", err)
		}
		out = append(out, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate port connections: %w", err)
	}
	return out, nil
}

// ListPortConnectionsByChild returns all TierPortConnections for a specific child within a tier aggregation.
func (s *TierAggregationStore) ListPortConnectionsByChild(aggID, childID int64) ([]*models.TierPortConnection, error) {
	const q = `
		SELECT id, tier_aggregation_id, child_id, agg_port_index, child_device_name, created_at
		FROM tier_port_connections
		WHERE tier_aggregation_id = ? AND child_id = ?
		ORDER BY agg_port_index`

	rows, err := s.db.Query(q, aggID, childID)
	if err != nil {
		return nil, fmt.Errorf("list port connections for child: %w", err)
	}
	defer rows.Close()

	var out []*models.TierPortConnection
	for rows.Next() {
		pc := &models.TierPortConnection{}
		if err := rows.Scan(&pc.ID, &pc.TierAggregationID, &pc.ChildID, &pc.AggPortIndex, &pc.ChildDeviceName, &pc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan port connection: %w", err)
		}
		out = append(out, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate port connections: %w", err)
	}
	return out, nil
}
