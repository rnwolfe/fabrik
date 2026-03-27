package service

import (
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

	// BlockAggregation operations
	SetAggregation(agg *models.BlockAggregation) (*models.BlockAggregation, error)
	GetAggregation(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error)
	ListAggregations(blockID int64) ([]*models.BlockAggregation, error)
	DeleteAggregation(blockID int64, plane models.NetworkPlane) error

	// PortConnection operations
	AllocatePorts(aggID, rackID int64, leafNames []string, startPortIndex int) ([]*models.PortConnection, error)
	DeallocatePorts(aggID, rackID int64) error
	DeallocatePortsByRack(rackID int64) error
	CountAllocatedPorts(aggID int64) (int, error)
	ListPortConnections(aggID int64) ([]*models.PortConnection, error)
	ListPortConnectionsByRack(aggID, rackID int64) ([]*models.PortConnection, error)

	// Support queries
	GetDeviceModel(id int64) (*models.DeviceModel, error)
	ListDevicesInRack(rackID int64) ([]*models.Device, error)
	UpdateRackBlock(rackID int64, blockID *int64) error
	GetRack(id int64) (*models.Rack, error)
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
func (s *BlockService) CreateBlock(superBlockID int64, name, description string) (*models.Block, error) {
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
	return b, nil
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
func (s *BlockService) AssignAggregation(blockID int64, plane models.NetworkPlane, deviceModelID int64) (*models.BlockAggregationSummary, error) {
	if _, err := s.repo.GetBlock(blockID); err != nil {
		return nil, fmt.Errorf("get block %d: %w", blockID, err)
	}

	dm, err := s.repo.GetDeviceModel(deviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", deviceModelID, err)
	}

	// If an agg already exists, check that downsizing is safe.
	existing, err := s.repo.GetAggregation(blockID, plane)
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

	agg, err := s.repo.SetAggregation(&models.BlockAggregation{
		BlockID:       blockID,
		Plane:         plane,
		DeviceModelID: deviceModelID,
	})
	if err != nil {
		return nil, fmt.Errorf("set aggregation: %w", err)
	}

	slog.Info("aggregation assigned", "blockID", blockID, "plane", plane, "deviceModelID", deviceModelID)
	return s.buildAggSummary(agg, dm)
}

// GetAggregationSummary returns the aggregation summary for a (blockID, plane) pair.
func (s *BlockService) GetAggregationSummary(blockID int64, plane models.NetworkPlane) (*models.BlockAggregationSummary, error) {
	agg, err := s.repo.GetAggregation(blockID, plane)
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
func (s *BlockService) ListAggregationSummaries(blockID int64) ([]*models.BlockAggregationSummary, error) {
	aggs, err := s.repo.ListAggregations(blockID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations for block %d: %w", blockID, err)
	}

	out := make([]*models.BlockAggregationSummary, 0, len(aggs))
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

// DeleteAggregation removes the aggregation for (blockID, plane) and all associated port connections.
func (s *BlockService) DeleteAggregation(blockID int64, plane models.NetworkPlane) error {
	if err := s.repo.DeleteAggregation(blockID, plane); err != nil {
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

	// Get leaf devices in the rack.
	devices, err := s.repo.ListDevicesInRack(rackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack %d: %w", rackID, err)
	}

	leafNames := leafDeviceNames(devices)

	// No leaf devices — succeed without port allocation.
	if len(leafNames) == 0 {
		slog.Info("rack added to block (no leaf devices)", "rackID", rackID, "blockID", block.ID)
		return &models.AddRackToBlockResult{
			Rack:        rack,
			Connections: []*models.PortConnection{},
		}, nil
	}

	// Get all agg assignments for this block.
	aggs, err := s.repo.ListAggregations(block.ID)
	if err != nil {
		return nil, fmt.Errorf("list aggregations for block %d: %w", block.ID, err)
	}

	if len(aggs) == 0 {
		slog.Info("rack added to block (no agg assigned)", "rackID", rackID, "blockID", block.ID)
		return &models.AddRackToBlockResult{
			Rack:        rack,
			Connections: []*models.PortConnection{},
			Warning:     "no aggregation switch assigned to this block; rack placed without connectivity",
		}, nil
	}

	var allConns []*models.PortConnection
	var warnings []string

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
			return nil, fmt.Errorf("%w: %d/%d ports allocated on %s agg; need %d more for %d leaves",
				models.ErrAggPortsFull, allocated, dm.PortCount, agg.Plane, len(leafNames), len(leafNames))
		}

		conns, err := s.repo.AllocatePorts(agg.ID, rackID, leafNames, allocated)
		if err != nil {
			return nil, fmt.Errorf("allocate ports for rack %d on agg %d: %w", rackID, agg.ID, err)
		}
		allConns = append(allConns, conns...)
	}

	result := &models.AddRackToBlockResult{
		Rack:        rack,
		Connections: allConns,
	}
	if len(warnings) > 0 {
		result.Warning = strings.Join(warnings, "; ")
	}

	slog.Info("rack added to block", "rackID", rackID, "blockID", block.ID, "connections", len(allConns))
	return result, nil
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
	if err := s.repo.DeallocatePortsByRack(rackID); err != nil {
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
func (s *BlockService) ListPortConnections(blockID int64, plane models.NetworkPlane) ([]*models.PortConnection, error) {
	agg, err := s.repo.GetAggregation(blockID, plane)
	if err != nil {
		return nil, fmt.Errorf("get aggregation for block %d plane %s: %w", blockID, plane, err)
	}
	conns, err := s.repo.ListPortConnections(agg.ID)
	if err != nil {
		return nil, fmt.Errorf("list port connections for agg %d: %w", agg.ID, err)
	}
	return conns, nil
}

// --- helpers ---

// resolveBlock returns the block to place the rack in.
// If blockID is provided, it is used directly.
// If blockID is nil and superBlockID > 0, the default block is found or auto-created.
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

// buildAggSummary constructs a BlockAggregationSummary from an agg record and its device model.
func (s *BlockService) buildAggSummary(agg *models.BlockAggregation, dm *models.DeviceModel) (*models.BlockAggregationSummary, error) {
	allocated, err := s.repo.CountAllocatedPorts(agg.ID)
	if err != nil {
		return nil, fmt.Errorf("count allocated ports for agg %d: %w", agg.ID, err)
	}

	available := dm.PortCount - allocated
	summary := &models.BlockAggregationSummary{
		BlockAggregation: *agg,
		TotalPorts:       dm.PortCount,
		AllocatedPorts:   allocated,
		AvailablePorts:   available,
		Utilization:      fmt.Sprintf("%d/%d ports allocated on %s agg", allocated, dm.PortCount, agg.Plane),
	}
	if dm.PortCount > 0 && allocated >= dm.PortCount {
		summary.Warning = fmt.Sprintf("%d/%d ports allocated on %s agg; no capacity for additional racks",
			allocated, dm.PortCount, agg.Plane)
	}
	return summary, nil
}
