package service

import (
	"errors"
	"fmt"
	"sort"

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
	ScopeType         models.AggregationScope `json:"scope_type"`
	ScopeID           int64                   `json:"scope_id"`
	ScopeName         string                  `json:"scope_name"`
	DeviceModel       *models.DeviceModel     `json:"device_model,omitempty"`
	SpineCount        int                     `json:"spine_count"`
	PortCount         int                     `json:"port_count"`
	AllocatedPorts    int                     `json:"allocated_ports"`
	HostLinkSpeedGbps int                     `json:"host_link_speed_gbps,omitempty"`
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

	// Determine Clos stage count from the hierarchy.
	//
	// The block-level agg stores the leaf (ToR) model — its spine_count
	// defines how many uplink ports connect leaves to spines.  The super-block
	// agg stores the spine model shared by all blocks — it completes the
	// 2-stage leaf-spine fabric but does NOT add a 3rd stage.
	//
	// Only a site-level agg (super-spine) adds a 3rd stage:
	//   block only            → 2-stage  (leaf + spine)
	//   block + super_block   → 2-stage  (leaf + spine, spine model from super_block)
	//   block + … + site      → 3-stage  (leaf + spine + super-spine)
	stages := 2
	for _, t := range tiers {
		if t.ScopeType == models.ScopeSite {
			stages = 3
			break
		}
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
		ScopeType:         scopeType,
		ScopeID:           scopeID,
		ScopeName:         name,
		SpineCount:        agg.SpineCount,
		HostLinkSpeedGbps: agg.HostLinkSpeedGbps,
		AllocatedPorts:    allocated,
	}

	dm, err := s.repo.GetDeviceModel(agg.DeviceModelID)
	if errors.Is(err, models.ErrNotFound) {
		return nil, fmt.Errorf("device model %d not found for aggregation %d", agg.DeviceModelID, agg.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("get device model %d for aggregation %d: %w", agg.DeviceModelID, agg.ID, err)
	}

	tier.DeviceModel = dm
	tier.PortCount = dm.PortCount

	return tier, nil
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

		// Bandwidth overlay: derive port speeds from leaf device model port groups.
		s.applyBandwidthOverlay(topo, &t)

		return topo
	}
	return nil
}

// applyBandwidthOverlay computes bandwidth-based oversubscription and bisection BW
// from port group speeds and the optional host link speed override.
func (s *DeriveFabricService) applyBandwidthOverlay(topo *TopologyPlan, tier *DerivedTier) {
	dm := tier.DeviceModel
	if dm == nil || len(dm.PortGroups) < 2 {
		return
	}

	// Sort port groups by speed ascending (same heuristic as frontend deriveFromPortGroups).
	groups := make([]models.PortGroup, len(dm.PortGroups))
	copy(groups, dm.PortGroups)
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].SpeedGbps != groups[j].SpeedGbps {
			return groups[i].SpeedGbps < groups[j].SpeedGbps
		}
		return groups[i].Count > groups[j].Count
	})

	// Highest speed group = uplinks, rest = downlinks.
	uplinkGroup := groups[len(groups)-1]
	uplinkSpeed := uplinkGroup.SpeedGbps

	// Downlink speed from port groups (use lowest speed group as the representative speed).
	downlinkSpeed := groups[0].SpeedGbps

	// Override displayed downlink speed with host link speed if set.
	effectiveDownlinkSpeed := downlinkSpeed
	if tier.HostLinkSpeedGbps > 0 {
		effectiveDownlinkSpeed = tier.HostLinkSpeedGbps
	}

	// Compute utilized uplinks from the uplink port group, capped by the derived topology.
	utilizedUplinks := topo.LeafUplinks
	if utilizedUplinks > uplinkGroup.Count {
		utilizedUplinks = uplinkGroup.Count
	}

	// Compute downlinks and downlink bandwidth directly from all non-uplink groups.
	downlinks := 0
	downlinkBW := 0.0
	for _, group := range groups[:len(groups)-1] {
		downlinks += group.Count

		groupSpeed := group.SpeedGbps
		if tier.HostLinkSpeedGbps > 0 {
			groupSpeed = tier.HostLinkSpeedGbps
		}
		downlinkBW += float64(group.Count) * float64(groupSpeed)
	}

	if utilizedUplinks > 0 && uplinkSpeed > 0 && downlinkBW > 0 {
		uplinkBW := float64(utilizedUplinks) * float64(uplinkSpeed)
		topo.BandwidthOversubscription = downlinkBW / uplinkBW
	}

	topo.LeafDownlinks = downlinks
	topo.HostLinkSpeedGbps = tier.HostLinkSpeedGbps
	topo.UplinkSpeedGbps = uplinkSpeed
	topo.DownlinkSpeedGbps = effectiveDownlinkSpeed
	topo.BisectionBandwidthGbps = float64(topo.SpineCount) * float64(utilizedUplinks) * float64(uplinkSpeed)
}
