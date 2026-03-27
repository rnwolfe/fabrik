package service

import (
	"fmt"
	"log/slog"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// ManagementAggRepository is the store interface for management aggregation operations.
// BlockStore satisfies this interface.
type ManagementAggRepository interface {
	SetAggregation(agg *models.BlockAggregation) (*models.BlockAggregation, error)
	GetAggregation(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error)
	ListAggregations(blockID int64) ([]*models.BlockAggregation, error)
	DeleteAggregation(blockID int64, plane models.NetworkPlane) error
	CountAllocatedPorts(aggID int64) (int, error)
	GetDeviceModel(id int64) (*models.DeviceModel, error)
}

// ManagementService implements management plane business logic:
//   - Block aggregation switch assignment for the management plane
//   - Management ToR port allocation against the management agg
type ManagementService struct {
	repo ManagementAggRepository
}

// NewManagementService returns a new ManagementService.
func NewManagementService(repo ManagementAggRepository) *ManagementService {
	return &ManagementService{repo: repo}
}

// SetManagementAgg assigns (or updates) the management aggregation switch model for a block.
// deviceModelID is required; it determines port capacity via device_model.port_count.
func (s *ManagementService) SetManagementAgg(blockID int64, deviceModelID int64) (*models.BlockAggregation, error) {
	if deviceModelID <= 0 {
		return nil, fmt.Errorf("%w: device_model_id is required", models.ErrConstraintViolation)
	}

	agg, err := s.repo.SetAggregation(&models.BlockAggregation{
		BlockID:       blockID,
		Plane:         models.PlaneManagement,
		DeviceModelID: deviceModelID,
	})
	if err != nil {
		return nil, fmt.Errorf("set management agg for block %d: %w", blockID, err)
	}
	slog.Info("management agg assigned", "blockID", blockID, "deviceModelID", deviceModelID)
	return agg, nil
}

// GetManagementAgg returns the management aggregation record for a block.
// Returns ErrNotFound if no management agg has been assigned.
func (s *ManagementService) GetManagementAgg(blockID int64) (*models.BlockAggregation, error) {
	agg, err := s.repo.GetAggregation(blockID, models.PlaneManagement)
	if err != nil {
		return nil, fmt.Errorf("get management agg for block %d: %w", blockID, err)
	}
	return agg, nil
}

// RemoveManagementAgg removes the management aggregation assignment from a block.
func (s *ManagementService) RemoveManagementAgg(blockID int64) error {
	if err := s.repo.DeleteAggregation(blockID, models.PlaneManagement); err != nil {
		return fmt.Errorf("remove management agg for block %d: %w", blockID, err)
	}
	slog.Info("management agg removed", "blockID", blockID)
	return nil
}

// ListBlockAggregations returns all aggregation assignments for a block.
func (s *ManagementService) ListBlockAggregations(blockID int64) ([]*models.BlockAggregation, error) {
	aggs, err := s.repo.ListAggregations(blockID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations for block %d: %w", blockID, err)
	}
	if aggs == nil {
		aggs = []*models.BlockAggregation{}
	}
	return aggs, nil
}

// AllocateManagementPort checks whether the block's management agg has capacity for one more
// ToR uplink. Returns a warning (not an error) if no management agg is assigned.
// Returns ErrConflict if the agg's port_count is fully allocated.
func (s *ManagementService) AllocateManagementPort(blockID int64) (*models.BlockAggregation, string, error) {
	agg, err := s.repo.GetAggregation(blockID, models.PlaneManagement)
	if err != nil {
		// No management agg assigned — not a hard failure, just a warning.
		return nil, "no management aggregation assigned to this block; management ToR has no upstream connectivity", nil
	}

	dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
	if err != nil {
		return nil, "", fmt.Errorf("get device model %d for management agg: %w", agg.DeviceModelID, err)
	}

	allocated, err := s.repo.CountAllocatedPorts(agg.ID)
	if err != nil {
		return nil, "", fmt.Errorf("count allocated ports for management agg %d: %w", agg.ID, err)
	}

	if dm.PortCount > 0 && allocated >= dm.PortCount {
		return nil, "", fmt.Errorf("%w: management agg port capacity exceeded: %d/%d ports allocated",
			models.ErrConflict, allocated, dm.PortCount)
	}

	return agg, "", nil
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
