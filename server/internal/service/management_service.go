package service

import (
	"fmt"
	"log/slog"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// BlockAggregationRepository is the store interface for block aggregation operations.
type BlockAggregationRepository interface {
	Upsert(agg *models.BlockAggregation) (*models.BlockAggregation, error)
	Get(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error)
	ListForBlock(blockID int64) ([]*models.BlockAggregation, error)
	Delete(blockID int64, plane models.NetworkPlane) error
	IncrementUsedPorts(blockID int64, plane models.NetworkPlane, delta int) (*models.BlockAggregation, error)
	DecrementUsedPorts(blockID int64, plane models.NetworkPlane, delta int) (*models.BlockAggregation, error)
}

// ManagementService implements management plane business logic:
//   - Block aggregation switch assignment for the management plane
//   - Management ToR port allocation against the management agg
type ManagementService struct {
	aggRepo BlockAggregationRepository
}

// NewManagementService returns a new ManagementService.
func NewManagementService(aggRepo BlockAggregationRepository) *ManagementService {
	return &ManagementService{aggRepo: aggRepo}
}

// SetManagementAgg assigns (or updates) the management aggregation switch for a block.
// deviceID may be nil to clear the assignment. maxPorts controls port capacity enforcement.
// Passing maxPorts=0 disables capacity enforcement (unlimited).
func (s *ManagementService) SetManagementAgg(blockID int64, deviceID *int64, maxPorts int, description string) (*models.BlockAggregation, error) {
	if maxPorts < 0 {
		return nil, fmt.Errorf("%w: max_ports must be non-negative", models.ErrConstraintViolation)
	}

	agg, err := s.aggRepo.Upsert(&models.BlockAggregation{
		BlockID:     blockID,
		Plane:       models.PlaneManagement,
		DeviceID:    deviceID,
		MaxPorts:    maxPorts,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("set management agg for block %d: %w", blockID, err)
	}
	slog.Info("management agg assigned", "blockID", blockID, "deviceID", deviceID, "maxPorts", maxPorts)
	return agg, nil
}

// GetManagementAgg returns the management aggregation record for a block.
// Returns ErrNotFound if no management agg has been assigned.
func (s *ManagementService) GetManagementAgg(blockID int64) (*models.BlockAggregation, error) {
	agg, err := s.aggRepo.Get(blockID, models.PlaneManagement)
	if err != nil {
		return nil, fmt.Errorf("get management agg for block %d: %w", blockID, err)
	}
	return agg, nil
}

// RemoveManagementAgg removes the management aggregation assignment from a block.
func (s *ManagementService) RemoveManagementAgg(blockID int64) error {
	if err := s.aggRepo.Delete(blockID, models.PlaneManagement); err != nil {
		return fmt.Errorf("remove management agg for block %d: %w", blockID, err)
	}
	slog.Info("management agg removed", "blockID", blockID)
	return nil
}

// ListBlockAggregations returns all aggregation assignments for a block.
func (s *ManagementService) ListBlockAggregations(blockID int64) ([]*models.BlockAggregation, error) {
	aggs, err := s.aggRepo.ListForBlock(blockID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations for block %d: %w", blockID, err)
	}
	if aggs == nil {
		aggs = []*models.BlockAggregation{}
	}
	return aggs, nil
}

// AllocateManagementPort increments the used_ports count for a block's management agg,
// enforcing the max_ports hard limit. Returns an error if capacity is exceeded.
// If no management agg is assigned for the block, it returns a warning result instead.
func (s *ManagementService) AllocateManagementPort(blockID int64) (*models.BlockAggregation, string, error) {
	agg, err := s.aggRepo.Get(blockID, models.PlaneManagement)
	if err != nil {
		// No management agg assigned — not a hard failure, just a warning.
		return nil, "no management aggregation assigned to this block; management ToR has no upstream connectivity", nil
	}

	updated, err := s.aggRepo.IncrementUsedPorts(blockID, models.PlaneManagement, 1)
	if err != nil {
		// Port capacity exceeded is a hard block for management agg.
		return nil, "", fmt.Errorf("%w: %s", models.ErrConflict, err.Error())
	}

	_ = agg
	return updated, "", nil
}

// ReleaseManagementPort decrements the used_ports count for a block's management agg.
func (s *ManagementService) ReleaseManagementPort(blockID int64) error {
	_, err := s.aggRepo.Get(blockID, models.PlaneManagement)
	if err != nil {
		// No agg record — nothing to release.
		return nil
	}
	if _, err := s.aggRepo.DecrementUsedPorts(blockID, models.PlaneManagement, 1); err != nil {
		return fmt.Errorf("release management port for block %d: %w", blockID, err)
	}
	return nil
}

// ValidateDeviceRole returns an error if the given role string is not a known DeviceRole.
func ValidateDeviceRole(role string) error {
	switch models.DeviceRole(role) {
	case models.DeviceRoleSpine,
		models.DeviceRoleLeaf,
		models.DeviceRoleSuperSpine,
		models.DeviceRoleServer,
		models.DeviceRoleOther,
		models.DeviceRoleManagementToR,
		models.DeviceRoleManagementAgg:
		return nil
	default:
		return fmt.Errorf("%w: unknown device role %q; valid roles are spine, leaf, super_spine, server, other, management_tor, management_agg",
			models.ErrConstraintViolation, role)
	}
}
