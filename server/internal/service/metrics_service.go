package service

import (
	"errors"
	"fmt"
	"math"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// MetricsRepository defines the database queries required by MetricsService.
type MetricsRepository interface {
	// GetDesignName returns the name of the design or models.ErrNotFound.
	GetDesignName(designID int64) (string, error)
	// QueryDesignCapacity returns aggregated resource capacity for the design.
	QueryDesignCapacity(designID int64) (*models.CapacitySummary, error)
	// QueryDesignPowerAndRacks returns total draw and total rack capacity for the design.
	QueryDesignPowerAndRacks(designID int64) (totalDrawW int, totalRackCapacityW int, err error)
}

// metricsDeriveFabric is the subset of DeriveFabricService used by MetricsService.
type metricsDeriveFabric interface {
	DeriveFabric(designID int64, plane models.NetworkPlane) (*DerivedFabric, error)
}

// MetricsService computes on-demand metrics for a design.
type MetricsService struct {
	repo    MetricsRepository
	derived metricsDeriveFabric
}

// NewMetricsService returns a new MetricsService backed by repo and deriveSvc.
func NewMetricsService(repo MetricsRepository, deriveSvc metricsDeriveFabric) *MetricsService {
	return &MetricsService{repo: repo, derived: deriveSvc}
}

// GetDesignMetrics computes all metrics for the given design.
func (s *MetricsService) GetDesignMetrics(designID int64) (*models.DesignMetrics, error) {
	_, err := s.repo.GetDesignName(designID)
	if err != nil {
		return nil, fmt.Errorf("get design metrics %d: %w", designID, err)
	}

	df, err := s.derived.DeriveFabric(designID, models.PlaneFrontEnd)
	if err != nil {
		return nil, fmt.Errorf("derive fabric for design %d: %w", designID, err)
	}

	// Compute fabric-level metrics from derived topology.
	fabricEntries, portEntries, totalSwitches, totalHostPorts, bisectionBW := computeFabricMetrics(df)

	// Identify choke point.
	chokePoint := findChokePoint(fabricEntries)

	// Compute power metrics.
	power, err := s.computePowerMetrics(designID)
	if err != nil {
		return nil, err
	}

	// Compute resource capacity.
	capacity, deviceCount, err := s.computeCapacity(designID)
	if err != nil {
		return nil, err
	}

	m := &models.DesignMetrics{
		DesignID:               designID,
		TotalHosts:             totalHostPorts,
		TotalSwitches:          totalSwitches,
		BisectionBandwidthGbps: bisectionBW,
		Fabrics:                fabricEntries,
		ChokePoint:             chokePoint,
		Power:                  power,
		Capacity:               capacity,
		PortUtilization:        portEntries,
		Empty:                  deviceCount == 0 && len(df.Tiers) == 0,
	}

	return m, nil
}

// computeFabricMetrics builds metric entries from the derived fabric topology.
func computeFabricMetrics(df *DerivedFabric) (
	entries []models.FabricMetricEntry,
	portEntries []models.PortUtilizationEntry,
	totalSwitches int,
	totalHostPorts int,
	bisectionBW float64,
) {
	entries = []models.FabricMetricEntry{}
	portEntries = []models.PortUtilizationEntry{}

	if df == nil || df.Topology == nil {
		return
	}

	topo := df.Topology
	fabricID := df.DesignID
	fabricName := "Derived Fabric"

	leafSpineOversub := 0.0
	if topo.LeafUplinks > 0 {
		leafSpineOversub = float64(topo.LeafDownlinks) / float64(topo.LeafUplinks)
	}

	spineSuperSpineOversub := 0.0
	if topo.Stages >= 3 && topo.SpineCount > 0 {
		spineUplinks := topo.Radix - topo.LeafCount
		spineDownlinks := topo.LeafCount
		if spineUplinks > 0 {
			spineSuperSpineOversub = float64(spineDownlinks) / float64(spineUplinks)
		}
	}

	entry := models.FabricMetricEntry{
		FabricID:                        fabricID,
		FabricName:                      fabricName,
		Tier:                            "front_end",
		Stages:                          topo.Stages,
		LeafSpineOversubscription:       leafSpineOversub,
		SpineSuperSpineOversubscription: spineSuperSpineOversub,
		TotalSwitches:                   topo.TotalSwitches,
		TotalHostPorts:                  topo.TotalHostPorts,
	}
	entries = append(entries, entry)
	totalSwitches += topo.TotalSwitches
	totalHostPorts += topo.TotalHostPorts

	// Port utilization entries per tier.
	leafTotal := topo.LeafCount * topo.Radix
	leafAllocated := topo.LeafCount * (topo.LeafDownlinks + topo.LeafUplinks)
	portEntries = append(portEntries, models.PortUtilizationEntry{
		FabricID:       fabricID,
		FabricName:     fabricName,
		TierName:       "leaf",
		TotalPorts:     leafTotal,
		AllocatedPorts: leafAllocated,
		AvailablePorts: leafTotal - leafAllocated,
	})

	if topo.Stages >= 2 {
		spineTotal := topo.SpineCount * topo.Radix
		spinePortsPerSwitch := topo.LeafCount + topo.SuperSpineCount
		spineAllocated := topo.SpineCount * spinePortsPerSwitch
		portEntries = append(portEntries, models.PortUtilizationEntry{
			FabricID:       fabricID,
			FabricName:     fabricName,
			TierName:       "spine",
			TotalPorts:     spineTotal,
			AllocatedPorts: spineAllocated,
			AvailablePorts: spineTotal - spineAllocated,
		})
	}

	if topo.Stages >= 3 && topo.SuperSpineCount > 0 {
		ssTotal := topo.SuperSpineCount * topo.Radix
		ssAllocated := topo.SuperSpineCount * topo.SpineCount
		portEntries = append(portEntries, models.PortUtilizationEntry{
			FabricID:       fabricID,
			FabricName:     fabricName,
			TierName:       "super-spine",
			TotalPorts:     ssTotal,
			AllocatedPorts: ssAllocated,
			AvailablePorts: ssTotal - ssAllocated,
		})
	}

	return entries, portEntries, totalSwitches, totalHostPorts, bisectionBW
}

// findChokePoint returns the fabric entry with the highest oversubscription ratio.
func findChokePoint(entries []models.FabricMetricEntry) *models.ChokePoint {
	if len(entries) == 0 {
		return nil
	}

	var worst *models.ChokePoint
	worstRatio := 0.0

	for _, e := range entries {
		if e.LeafSpineOversubscription > worstRatio {
			worstRatio = e.LeafSpineOversubscription
			worst = &models.ChokePoint{
				FabricID:   e.FabricID,
				FabricName: e.FabricName,
				Tier:       "leaf→spine",
				Ratio:      e.LeafSpineOversubscription,
			}
		}
		if e.SpineSuperSpineOversubscription > worstRatio {
			worstRatio = e.SpineSuperSpineOversubscription
			worst = &models.ChokePoint{
				FabricID:   e.FabricID,
				FabricName: e.FabricName,
				Tier:       "spine→super-spine",
				Ratio:      e.SpineSuperSpineOversubscription,
			}
		}
	}

	return worst
}

// computePowerMetrics queries power draw and rack capacity for the design.
func (s *MetricsService) computePowerMetrics(designID int64) (models.PowerMetrics, error) {
	drawW, capacityW, err := s.repo.QueryDesignPowerAndRacks(designID)
	if err != nil {
		return models.PowerMetrics{}, fmt.Errorf("query power for design %d: %w", designID, err)
	}

	utilPct := 0.0
	if capacityW > 0 {
		utilPct = math.Min(float64(drawW)/float64(capacityW)*100.0, 100.0)
	}

	return models.PowerMetrics{
		TotalCapacityW: capacityW,
		TotalDrawW:     drawW,
		UtilizationPct: utilPct,
	}, nil
}

// computeCapacity returns resource capacity totals and total device count.
func (s *MetricsService) computeCapacity(designID int64) (models.ResourceCapacity, int, error) {
	cs, err := s.repo.QueryDesignCapacity(designID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.ResourceCapacity{}, 0, nil
		}
		return models.ResourceCapacity{}, 0, fmt.Errorf("query capacity for design %d: %w", designID, err)
	}

	return models.ResourceCapacity{
		TotalVCPU:      cs.TotalVCPU,
		TotalRAMGB:     cs.TotalRAMGB,
		TotalStorageTB: cs.TotalStorageTB,
		TotalGPUCount:  cs.TotalGPUCount,
	}, cs.DeviceCount, nil
}
