package service_test

import (
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// fakeMetricsRepo implements service.MetricsRepository for testing.
type fakeMetricsRepo struct {
	designName     string
	designErr      error
	capacitySummary *models.CapacitySummary
	capacityErr    error
	totalDrawW     int
	totalCapacityW int
	powerErr       error
}

func (r *fakeMetricsRepo) GetDesignName(_ int64) (string, error) {
	return r.designName, r.designErr
}

func (r *fakeMetricsRepo) QueryDesignCapacity(_ int64) (*models.CapacitySummary, error) {
	if r.capacityErr != nil {
		return nil, r.capacityErr
	}
	if r.capacitySummary == nil {
		return nil, models.ErrNotFound
	}
	cp := *r.capacitySummary
	return &cp, nil
}

func (r *fakeMetricsRepo) QueryDesignPowerAndRacks(_ int64) (int, int, error) {
	return r.totalDrawW, r.totalCapacityW, r.powerErr
}

func newFakeMetricsRepo() *fakeMetricsRepo {
	return &fakeMetricsRepo{
		designName: "test-design",
	}
}

// fakeDeriveFabric implements service.metricsDeriveFabric for testing.
type fakeDeriveFabric struct {
	df  *service.DerivedFabric
	err error
}

func (f *fakeDeriveFabric) DeriveFabric(_ int64, _ models.NetworkPlane) (*service.DerivedFabric, error) {
	return f.df, f.err
}

// derivedFabricWithTopology returns a DerivedFabric with a topology for the given stages/radix/oversub.
func derivedFabricWithTopology(designID int64, stages, radix int, oversubscription float64) *service.DerivedFabric {
	topo, err := service.CalculateTopology(stages, radix, oversubscription)
	if err != nil {
		panic(err)
	}
	tiers := []service.DerivedTier{
		{ScopeType: models.ScopeBlock, ScopeID: 1, ScopeName: "row-A", SpineCount: topo.SpineCount, PortCount: radix},
	}
	return &service.DerivedFabric{
		DesignID: designID,
		Plane:    models.PlaneFrontEnd,
		Stages:   topo.Stages,
		Topology: topo,
		Tiers:    tiers,
	}
}

func TestMetricsService_GetDesignMetrics_empty(t *testing.T) {
	repo := newFakeMetricsRepo()
	derive := &fakeDeriveFabric{df: &service.DerivedFabric{DesignID: 1, Plane: models.PlaneFrontEnd, Stages: 2}}
	svc := service.NewMetricsService(repo, derive)

	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.Empty {
		t.Error("expected Empty=true for design with no tiers or devices")
	}
	if len(m.Fabrics) != 0 {
		t.Errorf("expected 0 fabric entries, got %d", len(m.Fabrics))
	}
}

func TestMetricsService_GetDesignMetrics_notFound(t *testing.T) {
	repo := newFakeMetricsRepo()
	repo.designErr = models.ErrNotFound
	derive := &fakeDeriveFabric{df: &service.DerivedFabric{DesignID: 42, Plane: models.PlaneFrontEnd, Stages: 2}}
	svc := service.NewMetricsService(repo, derive)

	_, err := svc.GetDesignMetrics(42)
	if err == nil {
		t.Fatal("expected error for missing design")
	}
}

func TestMetricsService_GetDesignMetrics_2stage(t *testing.T) {
	repo := newFakeMetricsRepo()
	// 2-stage fabric: radix=32, oversubscription=3 → uplinks=8, downlinks=24
	derive := &fakeDeriveFabric{df: derivedFabricWithTopology(1, 2, 32, 3.0)}
	svc := service.NewMetricsService(repo, derive)

	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Fabrics) != 1 {
		t.Fatalf("expected 1 fabric entry, got %d", len(m.Fabrics))
	}

	f := m.Fabrics[0]
	if f.Stages != 2 {
		t.Errorf("expected stages=2, got %d", f.Stages)
	}
	// oversubscription = downlinks/uplinks = 24/8 = 3
	if f.LeafSpineOversubscription != 3.0 {
		t.Errorf("expected leaf→spine oversubscription=3.0, got %f", f.LeafSpineOversubscription)
	}
	// 2-stage: no spine→super-spine oversubscription
	if f.SpineSuperSpineOversubscription != 0.0 {
		t.Errorf("expected spine→super-spine oversubscription=0, got %f", f.SpineSuperSpineOversubscription)
	}
}

func TestMetricsService_GetDesignMetrics_3stage(t *testing.T) {
	repo := newFakeMetricsRepo()
	// 3-stage fabric: radix=32, oversubscription=3
	derive := &fakeDeriveFabric{df: derivedFabricWithTopology(1, 3, 32, 3.0)}
	svc := service.NewMetricsService(repo, derive)

	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Fabrics) != 1 {
		t.Fatalf("expected 1 fabric entry, got %d", len(m.Fabrics))
	}

	f := m.Fabrics[0]
	if f.Stages != 3 {
		t.Errorf("expected stages=3, got %d", f.Stages)
	}
	if f.LeafSpineOversubscription != 3.0 {
		t.Errorf("expected leaf→spine oversubscription=3.0, got %f", f.LeafSpineOversubscription)
	}
	// 3-stage should have a non-zero spine→super-spine ratio.
	if f.SpineSuperSpineOversubscription <= 0.0 {
		t.Errorf("expected spine→super-spine oversubscription > 0, got %f", f.SpineSuperSpineOversubscription)
	}
}

func TestMetricsService_GetDesignMetrics_powerZeroCapacity(t *testing.T) {
	repo := newFakeMetricsRepo()
	repo.totalDrawW = 100
	repo.totalCapacityW = 0
	derive := &fakeDeriveFabric{df: &service.DerivedFabric{DesignID: 1, Plane: models.PlaneFrontEnd, Stages: 2}}

	svc := service.NewMetricsService(repo, derive)
	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Power.UtilizationPct != 0.0 {
		t.Errorf("expected utilization=0 when capacity=0, got %f", m.Power.UtilizationPct)
	}
}

func TestMetricsService_GetDesignMetrics_chokePoint(t *testing.T) {
	repo := newFakeMetricsRepo()
	// 2-stage, radix=32, oversub=3 → leaf→spine oversub=3.0
	derive := &fakeDeriveFabric{df: derivedFabricWithTopology(1, 2, 32, 3.0)}

	svc := service.NewMetricsService(repo, derive)
	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.ChokePoint == nil {
		t.Fatal("expected a choke point, got nil")
	}
	if m.ChokePoint.Ratio != 3.0 {
		t.Errorf("expected choke point ratio=3.0, got %f", m.ChokePoint.Ratio)
	}
}

func TestMetricsService_GetDesignMetrics_powerMetrics(t *testing.T) {
	repo := newFakeMetricsRepo()
	repo.totalDrawW = 5000
	repo.totalCapacityW = 10000
	derive := &fakeDeriveFabric{df: &service.DerivedFabric{DesignID: 1, Plane: models.PlaneFrontEnd, Stages: 2}}

	svc := service.NewMetricsService(repo, derive)
	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Power.TotalDrawW != 5000 {
		t.Errorf("expected draw=5000, got %d", m.Power.TotalDrawW)
	}
	if m.Power.TotalCapacityW != 10000 {
		t.Errorf("expected capacity=10000, got %d", m.Power.TotalCapacityW)
	}
	if m.Power.UtilizationPct != 50.0 {
		t.Errorf("expected utilization=50.0, got %f", m.Power.UtilizationPct)
	}
}

func TestMetricsService_GetDesignMetrics_portUtilization(t *testing.T) {
	repo := newFakeMetricsRepo()
	// 2-stage, radix=32, oversubscription=3 → uplinks=8, downlinks=24
	derive := &fakeDeriveFabric{df: derivedFabricWithTopology(1, 2, 32, 3.0)}

	svc := service.NewMetricsService(repo, derive)
	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2-stage produces leaf + spine port utilization entries.
	if len(m.PortUtilization) < 2 {
		t.Fatalf("expected at least 2 port utilization entries, got %d", len(m.PortUtilization))
	}

	var leafEntry *models.PortUtilizationEntry
	for i := range m.PortUtilization {
		if m.PortUtilization[i].TierName == "leaf" {
			leafEntry = &m.PortUtilization[i]
			break
		}
	}
	if leafEntry == nil {
		t.Fatal("expected a leaf tier port utilization entry")
	}
	// full fabric: leafCount=32, radix=32 → total = 32*32 = 1024
	if leafEntry.TotalPorts != 1024 {
		t.Errorf("expected leaf total ports=1024, got %d", leafEntry.TotalPorts)
	}
}

func TestMetricsService_GetDesignMetrics_capacityResources(t *testing.T) {
	repo := newFakeMetricsRepo()
	repo.capacitySummary = &models.CapacitySummary{
		Level:          models.CapacityLevelDesign,
		ID:             1,
		TotalVCPU:      128,
		TotalRAMGB:     512,
		TotalStorageTB: 10.5,
		TotalGPUCount:  8,
		DeviceCount:    4,
	}
	derive := &fakeDeriveFabric{df: &service.DerivedFabric{DesignID: 1, Plane: models.PlaneFrontEnd, Stages: 2}}

	svc := service.NewMetricsService(repo, derive)
	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Capacity.TotalVCPU != 128 {
		t.Errorf("expected TotalVCPU=128, got %d", m.Capacity.TotalVCPU)
	}
	if m.Capacity.TotalRAMGB != 512 {
		t.Errorf("expected TotalRAMGB=512, got %d", m.Capacity.TotalRAMGB)
	}
	if m.Capacity.TotalGPUCount != 8 {
		t.Errorf("expected TotalGPUCount=8, got %d", m.Capacity.TotalGPUCount)
	}
}
