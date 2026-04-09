package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// BlockRepository is the store interface required by BlockService.
type BlockRepository interface {
	// Block CRUD
	CreateBlock(b *models.Block) (*models.Block, error)
	GetBlock(id int64) (*models.Block, error)
	ListBlocks(superBlockID int64) ([]*models.Block, error)
	GetDefaultBlock(superBlockID int64) (*models.Block, error)

	// TierAggregation operations
	SetAggregation(agg *models.TierAggregation) (*models.TierAggregation, error)
	GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error)
	ListAggregations(scopeType models.AggregationScope, scopeID int64) ([]*models.TierAggregation, error)
	DeleteAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) error

	// TierPortConnection operations
	AllocatePorts(aggID, childID int64, childNames []string, startPortIndex int) ([]*models.TierPortConnection, error)
	DeallocatePorts(aggID, childID int64) error
	DeallocatePortsByChild(childID int64) error
	CountAllocatedPorts(aggID int64) (int, error)
	ListPortConnections(aggID int64) ([]*models.TierPortConnection, error)
	ListPortConnectionsByChild(aggID, childID int64) ([]*models.TierPortConnection, error)

	// Rack and device creation (for auto-provisioning)
	CreateRack(r *models.Rack) (*models.Rack, error)
	PlaceDevice(d *models.Device) (*models.Device, error)

	// Support queries
	GetDeviceModel(id int64) (*models.DeviceModel, error)
	ListDevicesInRack(rackID int64) ([]*models.Device, error)
	ListRacksInBlock(blockID int64) ([]*models.Rack, error)
	RemoveDevicesByRackAndRole(rackID int64, role models.DeviceRole) error
	UpdateRackBlock(rackID int64, blockID *int64) error
	GetRack(id int64) (*models.Rack, error)
	DeleteBlock(id int64) error
	DeleteRack(id int64) error
}

// BlockService implements business logic for blocks and block-level aggregation.
type BlockService struct {
	repo BlockRepository
}

// NewBlockService returns a new BlockService backed by repo.
func NewBlockService(repo BlockRepository) *BlockService {
	return &BlockService{repo: repo}
}

// --- Block operations ---

// CreateBlock validates and creates a new Block under a super-block.
// When leafModelID is non-nil, 2 racks are auto-created with redundant ToR leaf pairs.
// When spineModelID is also provided, spine devices are distributed across the racks.
func (s *BlockService) CreateBlock(superBlockID int64, name, description string, leafModelID, spineModelID *int64, spineCount int) (*models.CreateBlockResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: block name is required", models.ErrConstraintViolation)
	}
	b, err := s.repo.CreateBlock(&models.Block{
		SuperBlockID: superBlockID,
		Name:         name,
		Description:  description,
	})
	if err != nil {
		return nil, fmt.Errorf("create block: %w", err)
	}
	slog.Info("block created", "blockID", b.ID, "superBlockID", superBlockID, "name", b.Name)

	result := &models.CreateBlockResult{Block: b}

	if leafModelID == nil {
		return result, nil
	}

	// Look up the leaf model for height_u.
	leafModel, err := s.repo.GetDeviceModel(*leafModelID)
	if err != nil {
		return nil, fmt.Errorf("get leaf model %d: %w", *leafModelID, err)
	}

	// Look up spine model if provided.
	var spineModel *models.DeviceModel
	if spineModelID != nil {
		spineModel, err = s.repo.GetDeviceModel(*spineModelID)
		if err != nil {
			return nil, fmt.Errorf("get spine model %d: %w", *spineModelID, err)
		}
	}

	// Auto-create 4 racks: 2 base racks (Net-1, Net-2) + 2 compute racks (Rack-1, Rack-2).
	const defaultRackHeightU = 42
	const defaultPowerCapacityW = 10000
	const leavesPerRack = 2

	racks := make([]*models.Rack, 0, 4)

	// Create 2 base racks (network infra).
	baseRacks := make([]*models.Rack, 0, 2)
	for i := 1; i <= 2; i++ {
		rack, err := s.repo.CreateRack(&models.Rack{
			BlockID:        &b.ID,
			Name:           fmt.Sprintf("Net-%d", i),
			Role:           models.RackRoleBase,
			HeightU:        defaultRackHeightU,
			PowerCapacityW: defaultPowerCapacityW,
		})
		if err != nil {
			return nil, fmt.Errorf("create base rack %d: %w", i, err)
		}
		baseRacks = append(baseRacks, rack)
		racks = append(racks, rack)

		// Place 2 leaf devices at the top of the rack (top-down).
		// Start at the highest valid bottom-position for this leaf height so the
		// device's top slot is flush with the rack ceiling.
		pos := defaultRackHeightU - leafModel.HeightU + 1
		for j := 0; j < leavesPerRack; j++ {
			leafName := fmt.Sprintf("leaf-%d%c", i, 'a'+j)
			_, err := s.repo.PlaceDevice(&models.Device{
				RackID:        rack.ID,
				DeviceModelID: leafModel.ID,
				Name:          leafName,
				Role:          models.DeviceRoleLeaf,
				Position:      pos,
			})
			if err != nil {
				return nil, fmt.Errorf("place leaf %s: %w", leafName, err)
			}
			pos -= leafModel.HeightU
		}

		// Place spine devices in base racks only, alternating HA.
		if spineModel != nil && spineCount > 0 {
			// Adjust pos so the first spine's top slot clears the lowest leaf.
			// pos currently points to the slot just below the last leaf's bottom.
			// A spine of height H at position P occupies P..P+H-1, so to keep
			// P+H-1 < lowestLeafPosition we need P = pos-(H-1).
			pos -= spineModel.HeightU - 1
			for si := 0; si < spineCount; si++ {
				if si%2 != i-1 {
					continue
				}
				spineName := fmt.Sprintf("spine-%d", si+1)
				_, err := s.repo.PlaceDevice(&models.Device{
					RackID:        rack.ID,
					DeviceModelID: spineModel.ID,
					Name:          spineName,
					Role:          models.DeviceRoleSpine,
					Position:      pos,
				})
				if err != nil {
					return nil, fmt.Errorf("place spine %s: %w", spineName, err)
				}
				pos -= spineModel.HeightU
			}
		}
	}

	// Create 2 compute racks.
	for i := 1; i <= 2; i++ {
		rack, err := s.repo.CreateRack(&models.Rack{
			BlockID:        &b.ID,
			Name:           fmt.Sprintf("Rack-%d", i),
			Role:           models.RackRoleCompute,
			HeightU:        defaultRackHeightU,
			PowerCapacityW: defaultPowerCapacityW,
		})
		if err != nil {
			return nil, fmt.Errorf("create compute rack %d: %w", i, err)
		}
		racks = append(racks, rack)

		// Place 2 leaf devices at the top of each compute rack.
		pos := defaultRackHeightU - leafModel.HeightU + 1
		for j := 0; j < leavesPerRack; j++ {
			leafName := fmt.Sprintf("leaf-%d%c", i+2, 'a'+j) // leaf-3a, leaf-3b, etc.
			_, err := s.repo.PlaceDevice(&models.Device{
				RackID:        rack.ID,
				DeviceModelID: leafModel.ID,
				Name:          leafName,
				Role:          models.DeviceRoleLeaf,
				Position:      pos,
			})
			if err != nil {
				return nil, fmt.Errorf("place leaf %s: %w", leafName, err)
			}
			pos -= leafModel.HeightU
		}
	}

	result.Racks = racks

	// Assign the leaf model as front_end aggregation with spine count.
	_, err = s.repo.SetAggregation(&models.TierAggregation{
		ScopeType:     models.ScopeBlock,
		ScopeID:       b.ID,
		Plane:         models.PlaneFrontEnd,
		DeviceModelID: leafModel.ID,
		SpineCount:    spineCount,
	})
	if err != nil {
		return nil, fmt.Errorf("set leaf aggregation: %w", err)
	}

	// Allocate spine ports for each leaf device.
	if spineModel != nil {
		agg, err := s.repo.GetAggregation(models.ScopeBlock, b.ID, models.PlaneFrontEnd)
		if err != nil {
			return nil, fmt.Errorf("get aggregation: %w", err)
		}
		portIndex := 0
		for _, rack := range racks {
			devices, err := s.repo.ListDevicesInRack(rack.ID)
			if err != nil {
				return nil, fmt.Errorf("list devices in rack %d: %w", rack.ID, err)
			}
			names := leafDeviceNames(devices)
			if len(names) > 0 {
				_, err = s.repo.AllocatePorts(agg.ID, rack.ID, names, portIndex)
				if err != nil {
					return nil, fmt.Errorf("allocate ports for rack %d: %w", rack.ID, err)
				}
				portIndex += len(names)
			}
		}
	}

	slog.Info("block created with auto-racks",
		"blockID", b.ID, "racks", len(racks), "leafModel", leafModel.Model,
		"spineModel", spineModel)
	return result, nil
}

// AssignLeafModel assigns leafModelID as the front_end aggregation for blockID.
// This is the PATCH /api/blocks/{id} leaf_model_id path.
// It validates that the device model exists and delegates to AssignAggregation.
func (s *BlockService) AssignLeafModel(blockID int64, leafModelID int64) (*models.TierAggregationSummary, error) {
	if _, err := s.repo.GetBlock(blockID); err != nil {
		return nil, fmt.Errorf("get block %d: %w", blockID, err)
	}
	if _, err := s.repo.GetDeviceModel(leafModelID); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, fmt.Errorf("%w: leaf_model_id %d not found", models.ErrConstraintViolation, leafModelID)
		}
		return nil, fmt.Errorf("get device model %d: %w", leafModelID, err)
	}
	// Preserve existing spine_count if an agg already exists.
	spineCount := 0
	existing, err := s.repo.GetAggregation(models.ScopeBlock, blockID, models.PlaneFrontEnd)
	if err == nil {
		spineCount = existing.SpineCount
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, fmt.Errorf("get aggregation for block %d: %w", blockID, err)
	}
	return s.AssignAggregation(blockID, models.NetworkPlaneFrontEnd, leafModelID, spineCount, 0)
}

// GetBlock returns the block with the given id.
func (s *BlockService) GetBlock(id int64) (*models.Block, error) {
	b, err := s.repo.GetBlock(id)
	if err != nil {
		return nil, fmt.Errorf("get block %d: %w", id, err)
	}
	return b, nil
}

// ListBlocks returns all blocks for a super-block.
func (s *BlockService) ListBlocks(superBlockID int64) ([]*models.Block, error) {
	blocks, err := s.repo.ListBlocks(superBlockID)
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}
	return blocks, nil
}

// --- Aggregation operations ---

// AssignAggregation assigns an aggregation device model to a block for a given plane.
// If the block already has an agg for this plane, it is replaced.
// Replacing with a smaller model is rejected when existing connections would exceed new capacity.
func (s *BlockService) AssignAggregation(blockID int64, plane models.NetworkPlane, deviceModelID int64, spineCount int, hostLinkSpeedGbps int) (*models.TierAggregationSummary, error) {
	if _, err := s.repo.GetBlock(blockID); err != nil {
		return nil, fmt.Errorf("get block %d: %w", blockID, err)
	}

	dm, err := s.repo.GetDeviceModel(deviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", deviceModelID, err)
	}

	// If an agg already exists, check that downsizing is safe.
	existing, err := s.repo.GetAggregation(models.ScopeBlock, blockID, plane)
	if err == nil {
		// Aggregation exists — check current allocations vs new capacity.
		allocated, err := s.repo.CountAllocatedPorts(existing.ID)
		if err != nil {
			return nil, fmt.Errorf("count allocated ports: %w", err)
		}
		if allocated > dm.PortCount {
			return nil, fmt.Errorf("%w: %d ports allocated but new model only has %d ports",
				models.ErrAggModelDownsize, allocated, dm.PortCount)
		}
	}

	agg, err := s.repo.SetAggregation(&models.TierAggregation{
		ScopeType:         models.ScopeBlock,
		ScopeID:           blockID,
		Plane:             plane,
		DeviceModelID:     deviceModelID,
		SpineCount:        spineCount,
		HostLinkSpeedGbps: hostLinkSpeedGbps,
	})
	if err != nil {
		return nil, fmt.Errorf("set aggregation: %w", err)
	}

	slog.Info("aggregation assigned", "blockID", blockID, "plane", plane, "deviceModelID", deviceModelID, "spineCount", spineCount, "hostLinkSpeedGbps", hostLinkSpeedGbps)
	return s.buildAggSummary(agg, dm)
}

// GetAggregationSummary returns the aggregation summary for a (blockID, plane) pair.
func (s *BlockService) GetAggregationSummary(blockID int64, plane models.NetworkPlane) (*models.TierAggregationSummary, error) {
	agg, err := s.repo.GetAggregation(models.ScopeBlock, blockID, plane)
	if err != nil {
		return nil, fmt.Errorf("get aggregation for block %d plane %s: %w", blockID, plane, err)
	}
	dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", agg.DeviceModelID, err)
	}
	return s.buildAggSummary(agg, dm)
}

// ListAggregationSummaries returns summaries for all agg assignments on a block.
func (s *BlockService) ListAggregationSummaries(blockID int64) ([]*models.TierAggregationSummary, error) {
	aggs, err := s.repo.ListAggregations(models.ScopeBlock, blockID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations for block %d: %w", blockID, err)
	}

	out := make([]*models.TierAggregationSummary, 0, len(aggs))
	for _, agg := range aggs {
		dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
		if err != nil {
			return nil, fmt.Errorf("get device model %d: %w", agg.DeviceModelID, err)
		}
		summary, err := s.buildAggSummary(agg, dm)
		if err != nil {
			return nil, err
		}
		out = append(out, summary)
	}
	return out, nil
}

// AssignSuperBlockAggregation assigns an aggregation device model to a super-block for a given plane.
// If the super-block already has an agg for this plane, it is replaced.
func (s *BlockService) AssignSuperBlockAggregation(superBlockID int64, plane models.NetworkPlane, deviceModelID int64, spineCount int, hostLinkSpeedGbps int) (*models.TierAggregationSummary, error) {
	dm, err := s.repo.GetDeviceModel(deviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", deviceModelID, err)
	}

	agg, err := s.repo.SetAggregation(&models.TierAggregation{
		ScopeType:         models.ScopeSuperBlock,
		ScopeID:           superBlockID,
		Plane:             plane,
		DeviceModelID:     deviceModelID,
		SpineCount:        spineCount,
		HostLinkSpeedGbps: hostLinkSpeedGbps,
	})
	if err != nil {
		return nil, fmt.Errorf("set super-block aggregation: %w", err)
	}

	slog.Info("super-block aggregation assigned", "superBlockID", superBlockID, "plane", plane, "deviceModelID", deviceModelID, "spineCount", spineCount, "hostLinkSpeedGbps", hostLinkSpeedGbps)
	return s.buildAggSummary(agg, dm)
}

// GetSuperBlockAggregationSummary returns the aggregation summary for a (superBlockID, plane) pair.
func (s *BlockService) GetSuperBlockAggregationSummary(superBlockID int64, plane models.NetworkPlane) (*models.TierAggregationSummary, error) {
	agg, err := s.repo.GetAggregation(models.ScopeSuperBlock, superBlockID, plane)
	if err != nil {
		return nil, fmt.Errorf("get aggregation for super-block %d plane %s: %w", superBlockID, plane, err)
	}
	dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", agg.DeviceModelID, err)
	}
	return s.buildAggSummary(agg, dm)
}

// DeleteAggregation removes the aggregation for (blockID, plane) and all associated port connections.
func (s *BlockService) DeleteAggregation(blockID int64, plane models.NetworkPlane) error {
	if err := s.repo.DeleteAggregation(models.ScopeBlock, blockID, plane); err != nil {
		return fmt.Errorf("delete aggregation for block %d plane %s: %w", blockID, plane, err)
	}
	slog.Info("aggregation deleted", "blockID", blockID, "plane", plane)
	return nil
}

// --- Rack-to-block placement with auto-connection ---

// AddRackToBlock assigns a rack to a block and auto-allocates agg ports for each leaf device.
// If the block has no agg assigned, the rack is placed with a warning (no connectivity).
// If superBlockID is non-zero and blockID is nil, a default block is auto-created.
func (s *BlockService) AddRackToBlock(rackID int64, blockID *int64, superBlockID int64) (*models.AddRackToBlockResult, error) {
	// Resolve or create the block.
	block, err := s.resolveBlock(blockID, superBlockID)
	if err != nil {
		return nil, err
	}

	rack, err := s.repo.GetRack(rackID)
	if err != nil {
		return nil, fmt.Errorf("get rack %d: %w", rackID, err)
	}

	// Assign the rack to the block.
	if err := s.repo.UpdateRackBlock(rackID, &block.ID); err != nil {
		return nil, fmt.Errorf("assign rack %d to block %d: %w", rackID, block.ID, err)
	}
	rack.BlockID = &block.ID

	// Auto-place redundant ToR leaf pair if the block has a front_end aggregation
	// and the rack has no leaf devices yet.
	devices, err := s.repo.ListDevicesInRack(rackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack %d: %w", rackID, err)
	}

	if len(leafDeviceNames(devices)) == 0 {
		aggs, lookupErr := s.repo.ListAggregations(models.ScopeBlock, block.ID)
		if lookupErr == nil {
			for _, agg := range aggs {
				if agg.Plane == models.PlaneFrontEnd {
					leafModel, modelErr := s.repo.GetDeviceModel(agg.DeviceModelID)
					if modelErr == nil {
						// Count existing racks to derive rack number for naming.
						rackNum := 1
						pos := rack.HeightU - leafModel.HeightU + 1 // highest valid bottom-position
						for j := 0; j < 2; j++ {
							leafName := fmt.Sprintf("leaf-%d%c", rackNum, 'a'+j)
							s.repo.PlaceDevice(&models.Device{
								RackID:        rackID,
								DeviceModelID: leafModel.ID,
								Name:          leafName,
								Role:          models.DeviceRoleLeaf,
								Position:      pos,
							})
							pos -= leafModel.HeightU
						}
						// Refresh devices list after placement.
						devices, _ = s.repo.ListDevicesInRack(rackID)
					}
					break
				}
			}
		}
	}

	leafNames := leafDeviceNames(devices)

	// No leaf devices — succeed without port allocation.
	if len(leafNames) == 0 {
		slog.Info("rack added to block (no leaf devices)", "rackID", rackID, "blockID", block.ID)
		return &models.AddRackToBlockResult{
			Rack:        rack,
			Connections: []*models.TierPortConnection{},
		}, nil
	}

	// Get all agg assignments for this block.
	aggs, err := s.repo.ListAggregations(models.ScopeBlock, block.ID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations for block %d: %w", block.ID, err)
	}

	if len(aggs) == 0 {
		slog.Info("rack added to block (no agg assigned)", "rackID", rackID, "blockID", block.ID)
		return &models.AddRackToBlockResult{
			Rack:        rack,
			Connections: []*models.TierPortConnection{},
			Warning:     "no aggregation switch assigned to this block; rack placed without connectivity",
		}, nil
	}

	var allConns []*models.TierPortConnection

	for _, agg := range aggs {
		dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
		if err != nil {
			return nil, fmt.Errorf("get device model %d: %w", agg.DeviceModelID, err)
		}

		allocated, err := s.repo.CountAllocatedPorts(agg.ID)
		if err != nil {
			return nil, fmt.Errorf("count allocated ports for agg %d: %w", agg.ID, err)
		}

		available := dm.PortCount - allocated
		if available < len(leafNames) {
			needed := len(leafNames) - available
			return nil, fmt.Errorf("%w: %d/%d ports allocated on %s agg; need %d more for %d leaves",
				models.ErrAggPortsFull, allocated, dm.PortCount, agg.Plane, needed, len(leafNames))
		}

		conns, err := s.repo.AllocatePorts(agg.ID, rackID, leafNames, allocated)
		if err != nil {
			return nil, fmt.Errorf("allocate ports for rack %d on agg %d: %w", rackID, agg.ID, err)
		}
		allConns = append(allConns, conns...)
	}

	slog.Info("rack added to block", "rackID", rackID, "blockID", block.ID, "connections", len(allConns))
	return &models.AddRackToBlockResult{
		Rack:        rack,
		Connections: allConns,
	}, nil
}

// RemoveRackFromBlock removes a rack from its block and deallocates all agg port connections.
func (s *BlockService) RemoveRackFromBlock(rackID int64) error {
	rack, err := s.repo.GetRack(rackID)
	if err != nil {
		return fmt.Errorf("get rack %d: %w", rackID, err)
	}

	if rack.BlockID == nil {
		return fmt.Errorf("%w: rack %d is not assigned to any block", models.ErrNotFound, rackID)
	}

	// Deallocate all port connections for this rack.
	if err := s.repo.DeallocatePortsByChild(rackID); err != nil {
		return fmt.Errorf("deallocate ports for rack %d: %w", rackID, err)
	}

	// Clear block assignment.
	if err := s.repo.UpdateRackBlock(rackID, nil); err != nil {
		return fmt.Errorf("clear block assignment for rack %d: %w", rackID, err)
	}

	slog.Info("rack removed from block", "rackID", rackID, "blockID", *rack.BlockID)
	return nil
}

// ListPortConnections returns all port connections for a block aggregation.
func (s *BlockService) ListPortConnections(blockID int64, plane models.NetworkPlane) ([]*models.TierPortConnection, error) {
	agg, err := s.repo.GetAggregation(models.ScopeBlock, blockID, plane)
	if err != nil {
		return nil, fmt.Errorf("get aggregation for block %d plane %s: %w", blockID, plane, err)
	}
	conns, err := s.repo.ListPortConnections(agg.ID)
	if err != nil {
		return nil, fmt.Errorf("list port connections for agg %d: %w", agg.ID, err)
	}
	return conns, nil
}

// PlaceSpineDevices places count spine devices (using spineModelID) across the base racks in the block.
// Existing spine-role devices in target racks are removed first. If the block has no base racks,
// all racks are used as a fallback for backward compatibility.
func (s *BlockService) PlaceSpineDevices(blockID, spineModelID int64, count int) error {
	if count < 0 {
		count = 0
	}

	spineModel, err := s.repo.GetDeviceModel(spineModelID)
	if err != nil {
		return fmt.Errorf("get spine model %d: %w", spineModelID, err)
	}

	allRacks, err := s.repo.ListRacksInBlock(blockID)
	if err != nil {
		return fmt.Errorf("list racks in block %d: %w", blockID, err)
	}

	// Filter to base racks; fall back to all racks if none exist.
	var targetRacks []*models.Rack
	for _, r := range allRacks {
		if r.Role == models.RackRoleBase {
			targetRacks = append(targetRacks, r)
		}
	}
	if len(targetRacks) == 0 {
		targetRacks = allRacks
	}

	if len(targetRacks) == 0 {
		return nil // nothing to do
	}

	// Remove all existing spine devices from ALL racks in the block to avoid
	// stale spines remaining in compute racks (e.g. from a previous fallback pass).
	for _, r := range allRacks {
		if err := s.repo.RemoveDevicesByRackAndRole(r.ID, models.DeviceRoleSpine); err != nil {
			return fmt.Errorf("remove spine devices from rack %d: %w", r.ID, err)
		}
	}

	// Distribute count spine devices across target racks alternating HA.
	for si := 0; si < count; si++ {
		rack := targetRacks[si%len(targetRacks)]

		// Find position: place below the lowest-positioned device in the rack.
		// Initialize to HeightU+1 so an exact match at HeightU is still caught.
		devices, err := s.repo.ListDevicesInRack(rack.ID)
		if err != nil {
			return fmt.Errorf("list devices in rack %d: %w", rack.ID, err)
		}
		pos := rack.HeightU + 1
		for _, d := range devices {
			if d.Position <= pos {
				pos = d.Position - 1
			}
		}
		// Adjust for multi-U spines: pos is the slot just below the lowest
		// existing device, but a spine of height H at position P occupies
		// P..P+H-1. Subtract H-1 so the top slot clears all existing devices.
		pos -= spineModel.HeightU - 1
		if pos < 1 {
			pos = 1
		}

		spineName := fmt.Sprintf("spine-%d", si+1)
		_, err = s.repo.PlaceDevice(&models.Device{
			RackID:        rack.ID,
			DeviceModelID: spineModel.ID,
			Name:          spineName,
			Role:          models.DeviceRoleSpine,
			Position:      pos,
		})
		if err != nil {
			return fmt.Errorf("place spine %s in rack %d: %w", spineName, rack.ID, err)
		}
	}

	// Sync spine count into the block's front_end aggregation so that DeriveFabric
	// can calculate topology metrics. If no aggregation exists yet, skip silently —
	// the block was created without a leaf model assignment and topology can't be
	// derived regardless.
	if agg, err := s.repo.GetAggregation(models.ScopeBlock, blockID, models.PlaneFrontEnd); err == nil {
		agg.SpineCount = count
		if _, err := s.repo.SetAggregation(agg); err != nil {
			slog.Warn("failed to sync spine count into aggregation", "blockID", blockID, "err", err)
		}
	}

	slog.Info("spine devices placed", "blockID", blockID, "spineModelID", spineModelID, "count", count)
	return nil
}

// DeleteBlock removes a block and all its racks, devices, and aggregations.
// If this is the last block in its super-block, the super-block's aggregations
// are also removed so stale spine configuration doesn't bleed into new blocks.
func (s *BlockService) DeleteBlock(id int64) error {
	blk, err := s.repo.GetBlock(id)
	if err != nil {
		return fmt.Errorf("get block %d: %w", id, err)
	}
	racks, err := s.repo.ListRacksInBlock(id)
	if err != nil {
		return fmt.Errorf("list racks in block %d: %w", id, err)
	}
	// Delete aggregations for this block (cascades port connections via FK)
	aggs, err := s.repo.ListAggregations(models.ScopeBlock, id)
	if err != nil {
		return fmt.Errorf("list aggregations for block %d: %w", id, err)
	}
	for _, agg := range aggs {
		if err := s.repo.DeleteAggregation(models.ScopeBlock, id, agg.Plane); err != nil {
			return fmt.Errorf("delete aggregation for block %d plane %s: %w", id, agg.Plane, err)
		}
	}
	// Delete each rack (cascades devices via FK)
	for _, r := range racks {
		if err := s.repo.DeleteRack(r.ID); err != nil {
			return fmt.Errorf("delete rack %d: %w", r.ID, err)
		}
	}
	if err := s.repo.DeleteBlock(id); err != nil {
		return fmt.Errorf("delete block %d: %w", id, err)
	}
	// If this was the last block in the super-block, clean up the super-block's
	// aggregations so stale spine config doesn't appear when new blocks are added.
	remaining, err := s.repo.ListBlocks(blk.SuperBlockID)
	if err == nil && len(remaining) == 0 {
		sbAggs, err := s.repo.ListAggregations(models.ScopeSuperBlock, blk.SuperBlockID)
		if err == nil {
			for _, agg := range sbAggs {
				_ = s.repo.DeleteAggregation(models.ScopeSuperBlock, blk.SuperBlockID, agg.Plane)
			}
		}
	}
	slog.Info("block deleted", "blockID", id)
	return nil
}

// --- helpers ---

// resolveBlock returns the block to place the rack in.
func (s *BlockService) resolveBlock(blockID *int64, superBlockID int64) (*models.Block, error) {
	if blockID != nil {
		b, err := s.repo.GetBlock(*blockID)
		if err != nil {
			return nil, fmt.Errorf("get block %d: %w", *blockID, err)
		}
		return b, nil
	}

	if superBlockID <= 0 {
		return nil, fmt.Errorf("%w: block_id or super_block_id is required", models.ErrConstraintViolation)
	}

	// Find or create the default block.
	def, err := s.repo.GetDefaultBlock(superBlockID)
	if err != nil {
		return nil, fmt.Errorf("get default block: %w", err)
	}
	if def != nil {
		return def, nil
	}

	// Auto-create default block.
	def, err = s.repo.CreateBlock(&models.Block{
		SuperBlockID: superBlockID,
		Name:         "default",
		Description:  "Auto-created default block",
	})
	if err != nil {
		return nil, fmt.Errorf("create default block: %w", err)
	}
	slog.Info("default block auto-created", "superBlockID", superBlockID, "blockID", def.ID)
	return def, nil
}

// leafDeviceNames returns the names of all devices with role "leaf" in the rack.
func leafDeviceNames(devices []*models.Device) []string {
	var names []string
	for _, d := range devices {
		if d.Role == models.DeviceRoleLeaf {
			names = append(names, d.Name)
		}
	}
	return names
}

// buildAggSummary constructs a TierAggregationSummary from an agg record and its device model.
func (s *BlockService) buildAggSummary(agg *models.TierAggregation, dm *models.DeviceModel) (*models.TierAggregationSummary, error) {
	allocated, err := s.repo.CountAllocatedPorts(agg.ID)
	if err != nil {
		return nil, fmt.Errorf("count allocated ports for agg %d: %w", agg.ID, err)
	}

	available := dm.PortCount - allocated
	summary := &models.TierAggregationSummary{
		TierAggregation: *agg,
		TotalPorts:      dm.PortCount,
		AllocatedPorts:  allocated,
		AvailablePorts:  available,
		Utilization:     fmt.Sprintf("%d/%d ports allocated on %s agg", allocated, dm.PortCount, agg.Plane),
	}
	if dm.PortCount > 0 && allocated >= dm.PortCount {
		summary.Warning = fmt.Sprintf("%d/%d ports allocated on %s agg; no capacity for additional racks",
			allocated, dm.PortCount, agg.Plane)
	}
	return summary, nil
}
