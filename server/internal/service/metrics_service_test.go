package service_test

import (
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// fakeMetricsRepo implements service.MetricsRepository for testing.
type fakeMetricsRepo struct {
	designName       string
	designErr        error
	fabrics          []*store.FabricRecord
	deviceModels     map[int64]*models.DeviceModel
	capacitySummary  *models.CapacitySummary
	capacityErr      error
	totalDrawW       int
	totalCapacityW   int
	powerErr         error
}

func (r *fakeMetricsRepo) GetDesignName(_ int64) (string, error) {
	return r.designName, r.designErr
}

func (r *fakeMetricsRepo) ListFabricsByDesign(_ int64) ([]*store.FabricRecord, error) {
	return r.fabrics, nil
}

func (r *fakeMetricsRepo) GetDeviceModelByID(id int64) (*models.DeviceModel, error) {
	dm, ok := r.deviceModels[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return dm, nil
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
		designName:   "test-design",
		deviceModels: make(map[int64]*models.DeviceModel),
	}
}

func makeFabricRecord(id, designID int64, stages, radix int, oversubscription float64) *store.FabricRecord {
	return &store.FabricRecord{
		Fabric: models.Fabric{
			ID:       id,
			DesignID: designID,
			Name:     "fabric-" + string(rune('0'+id)),
			Tier:     models.FabricTierFrontEnd,
		},
		Stages:           stages,
		Radix:            radix,
		Oversubscription: oversubscription,
	}
}

func TestMetricsService_GetDesignMetrics_empty(t *testing.T) {
	repo := newFakeMetricsRepo()
	svc := service.NewMetricsService(repo)

	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.Empty {
		t.Error("expected Empty=true for design with no fabrics or devices")
	}
	if len(m.Fabrics) != 0 {
		t.Errorf("expected 0 fabric entries, got %d", len(m.Fabrics))
	}
}

func TestMetricsService_GetDesignMetrics_notFound(t *testing.T) {
	repo := newFakeMetricsRepo()
	repo.designErr = models.ErrNotFound
	svc := service.NewMetricsService(repo)

	_, err := svc.GetDesignMetrics(42)
	if err == nil {
		t.Fatal("expected error for missing design")
	}
}

func TestMetricsService_GetDesignMetrics_2stage(t *testing.T) {
	repo := newFakeMetricsRepo()
	// 2-stage fabric: radix=32, oversubscription=3 → uplinks=8, downlinks=24
	repo.fabrics = []*store.FabricRecord{makeFabricRecord(1, 1, 2, 32, 3.0)}

	svc := service.NewMetricsService(repo)
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
	repo.fabrics = []*store.FabricRecord{makeFabricRecord(2, 1, 3, 32, 3.0)}

	svc := service.NewMetricsService(repo)
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
}

func TestMetricsService_GetDesignMetrics_chokePoint(t *testing.T) {
	repo := newFakeMetricsRepo()
	// Two fabrics: one 3:1, one 2:1 oversubscription
	repo.fabrics = []*store.FabricRecord{
		makeFabricRecord(1, 1, 2, 32, 3.0), // oversubscription 3:1
		makeFabricRecord(2, 1, 2, 32, 2.0), // oversubscription 2:1 (radix snaps to 33→36 for divisor 3)
	}

	svc := service.NewMetricsService(repo)
	m, err := svc.GetDesignMetrics(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.ChokePoint == nil {
		t.Fatal("expected a choke point, got nil")
	}
	if m.ChokePoint.FabricID != 1 {
		t.Errorf("expected choke point at fabric 1 (3:1), got fabric %d", m.ChokePoint.FabricID)
	}
}

func TestMetricsService_GetDesignMetrics_powerMetrics(t *testing.T) {
	repo := newFakeMetricsRepo()
	repo.totalDrawW = 5000
	repo.totalCapacityW = 10000

	svc := service.NewMetricsService(repo)
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
	repo.fabrics = []*store.FabricRecord{makeFabricRecord(1, 1, 2, 32, 3.0)}

	svc := service.NewMetricsService(repo)
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
	// leafCount=1, radix=32 → total = 1*32 = 32
	if leafEntry.TotalPorts != 32 {
		t.Errorf("expected leaf total ports=32, got %d", leafEntry.TotalPorts)
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

	svc := service.NewMetricsService(repo)
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
