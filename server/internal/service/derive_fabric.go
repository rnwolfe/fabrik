package service

import (
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DeriveFabricRepository is the read-only store interface needed to walk the
// design hierarchy and collect tier aggregations.
type DeriveFabricRepository interface {
	GetDesign(id int64) (*models.Design, error)
	ListSitesByDesign(designID int64) ([]*models.Site, error)
	ListSuperBlocksBySite(siteID int64) ([]*models.SuperBlock, error)
	ListBlocksBySuperBlock(superBlockID int64) ([]*models.Block, error)
	GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error)
	CountAllocatedPorts(aggID int64) (int, error)
	GetDeviceModel(id int64) (*models.DeviceModel, error)
}

// DerivedTier describes one aggregation level within the derived Clos topology.
type DerivedTier struct {
	ScopeType      models.AggregationScope `json:"scope_type"`
	ScopeID        int64                   `json:"scope_id"`
	ScopeName      string                  `json:"scope_name"`
	DeviceModel    *models.DeviceModel     `json:"device_model,omitempty"`
	SpineCount     int                     `json:"spine_count"`
	PortCount      int                     `json:"port_count"`
	AllocatedPorts int                     `json:"allocated_ports"`
}

// DerivedFabric is the computed Clos topology for a design.
// Stages emerge from how many hierarchy levels carry a front_end TierAggregation.
// This is the authoritative topology source; the declared Fabric entity is deprecated.
type DerivedFabric struct {
	DesignID int64          `json:"design_id"`
	Plane    models.NetworkPlane `json:"plane"`
	Stages   int            `json:"stages"`
	Topology *TopologyPlan  `json:"topology,omitempty"`
	Tiers    []DerivedTier  `json:"tiers"`
}

// DeriveFabricService derives a Clos fabric from the live hierarchy.
type DeriveFabricService struct {
	repo DeriveFabricRepository
}

// NewDeriveFabricService returns a new DeriveFabricService.
func NewDeriveFabricService(repo DeriveFabricRepository) *DeriveFabricService {
	return &DeriveFabricService{repo: repo}
}

// DeriveFabric walks Design → Site → SuperBlock → Block, collects front_end
// TierAggregations at each level, and returns a DerivedFabric.
//
// Stage count = 1 (leaf layer implicit) + count(levels with a front_end agg),
// floored at 2 (a design with only block-level spines is already 2-stage).
func (s *DeriveFabricService) DeriveFabric(designID int64, plane models.NetworkPlane) (*DerivedFabric, error) {
	if _, err := s.repo.GetDesign(designID); err != nil {
		return nil, fmt.Errorf("derive fabric for design %d: %w", designID, err)
	}

	sites, err := s.repo.ListSitesByDesign(designID)
	if err != nil {
		return nil, fmt.Errorf("list sites for design %d: %w", designID, err)
	}

	var tiers []DerivedTier

	// --- site level ---
	for _, site := range sites {
		tier, err := s.collectTier(models.ScopeSite, site.ID, site.Name, plane)
		if err != nil {
			return nil, err
		}
		if tier != nil {
			tiers = append(tiers, *tier)
		}

		// --- super_block level ---
		superBlocks, err := s.repo.ListSuperBlocksBySite(site.ID)
		if err != nil {
			return nil, fmt.Errorf("list super_blocks for site %d: %w", site.ID, err)
		}
		for _, sb := range superBlocks {
			tier, err := s.collectTier(models.ScopeSuperBlock, sb.ID, sb.Name, plane)
			if err != nil {
				return nil, err
			}
			if tier != nil {
				tiers = append(tiers, *tier)
			}

			// --- block level ---
			blocks, err := s.repo.ListBlocksBySuperBlock(sb.ID)
			if err != nil {
				return nil, fmt.Errorf("list blocks for super_block %d: %w", sb.ID, err)
			}
			for _, blk := range blocks {
				tier, err := s.collectTier(models.ScopeBlock, blk.ID, blk.Name, plane)
				if err != nil {
					return nil, err
				}
				if tier != nil {
					tiers = append(tiers, *tier)
				}
			}
		}
	}

	// Count distinct scope levels that have at least one aggregation.
	levelsWithAgg := distinctLevels(tiers)
	stages := levelsWithAgg + 1 // +1 for the leaf layer
	if stages < 2 {
		stages = 2
	}

	df := &DerivedFabric{
		DesignID: designID,
		Plane:    plane,
		Stages:   stages,
		Tiers:    tiers,
	}

	// Derive topology from the block-level tier (leaf/spine parameters live there).
	df.Topology = s.deriveTopology(tiers, stages)

	return df, nil
}

// collectTier looks up a TierAggregation for the given scope and plane.
// Returns nil (no error) if no aggregation is assigned at that level.
func (s *DeriveFabricService) collectTier(scopeType models.AggregationScope, scopeID int64, name string, plane models.NetworkPlane) (*DerivedTier, error) {
	agg, err := s.repo.GetAggregation(scopeType, scopeID, plane)
	if errors.Is(err, models.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get aggregation (%s %d %s): %w", scopeType, scopeID, plane, err)
	}

	allocated, err := s.repo.CountAllocatedPorts(agg.ID)
	if err != nil {
		return nil, fmt.Errorf("count allocated ports for agg %d: %w", agg.ID, err)
	}

	tier := &DerivedTier{
		ScopeType:      scopeType,
		ScopeID:        scopeID,
		ScopeName:      name,
		SpineCount:     agg.SpineCount,
		AllocatedPorts: allocated,
	}

	dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
	if err == nil {
		tier.DeviceModel = dm
		tier.PortCount = dm.PortCount
	}

	return tier, nil
}

// distinctLevels returns the number of unique AggregationScope values present in tiers.
func distinctLevels(tiers []DerivedTier) int {
	seen := make(map[models.AggregationScope]struct{})
	for _, t := range tiers {
		seen[t.ScopeType] = struct{}{}
	}
	return len(seen)
}

// deriveTopology calculates a TopologyPlan from the block-level tier, if present.
// If no block-level tier exists, returns nil.
func (s *DeriveFabricService) deriveTopology(tiers []DerivedTier, stages int) *TopologyPlan {
	// Find the first block-level tier — it has leaf radix and spine count.
	for _, t := range tiers {
		if t.ScopeType != models.ScopeBlock {
			continue
		}
		if t.PortCount == 0 || t.SpineCount == 0 {
			continue
		}

		radix := t.PortCount
		uplinks := t.SpineCount
		if uplinks >= radix {
			continue // degenerate: can't have more uplinks than ports
		}

		downlinks := radix - uplinks
		oversub := float64(downlinks) / float64(uplinks)

		hints := &TopologyHints{}
		// If there's a super-block tier, use its device model port count as SpineRadix.
		for _, st := range tiers {
			if st.ScopeType == models.ScopeSuperBlock && st.PortCount > 0 {
				hints.SpineRadix = st.PortCount
				break
			}
		}

		topo, err := CalculateTopology(stages, radix, oversub, hints)
		if err != nil {
			return nil
		}
		return topo
	}
	return nil
}
