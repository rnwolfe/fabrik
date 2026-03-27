package service_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// fakeCapacityRepo implements service.CapacityRepository for testing.
type fakeCapacityRepo struct {
	racks       map[int64]*models.CapacitySummary
	blocks      map[int64]*models.CapacitySummary
	superBlocks map[int64]*models.CapacitySummary
	sites       map[int64]*models.CapacitySummary
	designs     map[int64]*models.CapacitySummary
}

func newFakeCapacityRepo() *fakeCapacityRepo {
	return &fakeCapacityRepo{
		racks:       make(map[int64]*models.CapacitySummary),
		blocks:      make(map[int64]*models.CapacitySummary),
		superBlocks: make(map[int64]*models.CapacitySummary),
		sites:       make(map[int64]*models.CapacitySummary),
		designs:     make(map[int64]*models.CapacitySummary),
	}
}

func (r *fakeCapacityRepo) QueryRackCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := r.racks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCapacityRepo) QueryBlockCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := r.blocks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCapacityRepo) QuerySuperBlockCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := r.superBlocks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCapacityRepo) QuerySiteCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := r.sites[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCapacityRepo) QueryDesignCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := r.designs[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

// --- tests ---

func newCapacitySvc() (*service.CapacityService, *fakeCapacityRepo) {
	repo := newFakeCapacityRepo()
	return service.NewCapacityService(repo), repo
}

func TestCapacityService_GetRackCapacity(t *testing.T) {
	svc, repo := newCapacitySvc()

	repo.racks[1] = &models.CapacitySummary{
		Level:             models.CapacityLevelRack,
		ID:                1,
		Name:              "rack-a",
		PowerWattsIdle:    200,
		PowerWattsTypical: 500,
		PowerWattsMax:     800,
		TotalVCPU:         0,
		TotalRAMGB:        0,
		TotalStorageTB:    0,
		TotalGPUCount:     0,
		DeviceCount:       2,
	}

	c, err := svc.GetRackCapacity(1)
	if err != nil {
		t.Fatalf("GetRackCapacity: %v", err)
	}
	if c.Level != models.CapacityLevelRack {
		t.Errorf("level: want %q got %q", models.CapacityLevelRack, c.Level)
	}
	if c.PowerWattsTypical != 500 {
		t.Errorf("power_watts_typical: want 500 got %d", c.PowerWattsTypical)
	}
	if c.DeviceCount != 2 {
		t.Errorf("device_count: want 2 got %d", c.DeviceCount)
	}
}

func TestCapacityService_GetRackCapacity_NotFound(t *testing.T) {
	svc, _ := newCapacitySvc()

	_, err := svc.GetRackCapacity(999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCapacityService_GetBlockCapacity(t *testing.T) {
	svc, repo := newCapacitySvc()

	repo.blocks[2] = &models.CapacitySummary{
		Level:             models.CapacityLevelBlock,
		ID:                2,
		Name:              "block-b",
		PowerWattsTypical: 2000,
		TotalVCPU:         128,
		TotalRAMGB:        1024,
		TotalStorageTB:    10.5,
		TotalGPUCount:     4,
		DeviceCount:       10,
	}

	c, err := svc.GetBlockCapacity(2)
	if err != nil {
		t.Fatalf("GetBlockCapacity: %v", err)
	}
	if c.TotalVCPU != 128 {
		t.Errorf("total_vcpu: want 128 got %d", c.TotalVCPU)
	}
	if c.TotalRAMGB != 1024 {
		t.Errorf("total_ram_gb: want 1024 got %d", c.TotalRAMGB)
	}
	if c.TotalStorageTB != 10.5 {
		t.Errorf("total_storage_tb: want 10.5 got %f", c.TotalStorageTB)
	}
	if c.TotalGPUCount != 4 {
		t.Errorf("total_gpu_count: want 4 got %d", c.TotalGPUCount)
	}
}

func TestCapacityService_GetSuperBlockCapacity(t *testing.T) {
	svc, repo := newCapacitySvc()

	repo.superBlocks[3] = &models.CapacitySummary{
		Level:             models.CapacityLevelSuperBlock,
		ID:                3,
		Name:              "sb-c",
		PowerWattsTypical: 8000,
		DeviceCount:       40,
	}

	c, err := svc.GetSuperBlockCapacity(3)
	if err != nil {
		t.Fatalf("GetSuperBlockCapacity: %v", err)
	}
	if c.PowerWattsTypical != 8000 {
		t.Errorf("power_watts_typical: want 8000 got %d", c.PowerWattsTypical)
	}
}

func TestCapacityService_GetSiteCapacity(t *testing.T) {
	svc, repo := newCapacitySvc()

	repo.sites[4] = &models.CapacitySummary{
		Level:             models.CapacityLevelSite,
		ID:                4,
		Name:              "site-d",
		PowerWattsTypical: 50000,
		DeviceCount:       200,
	}

	c, err := svc.GetSiteCapacity(4)
	if err != nil {
		t.Fatalf("GetSiteCapacity: %v", err)
	}
	if c.DeviceCount != 200 {
		t.Errorf("device_count: want 200 got %d", c.DeviceCount)
	}
}

func TestCapacityService_GetDesignCapacity(t *testing.T) {
	svc, repo := newCapacitySvc()

	repo.designs[5] = &models.CapacitySummary{
		Level:             models.CapacityLevelDesign,
		ID:                5,
		Name:              "design-e",
		PowerWattsIdle:    10000,
		PowerWattsTypical: 80000,
		PowerWattsMax:     120000,
		TotalVCPU:         2048,
		TotalRAMGB:        16384,
		TotalStorageTB:    512,
		TotalGPUCount:     32,
		DeviceCount:       1000,
	}

	c, err := svc.GetDesignCapacity(5)
	if err != nil {
		t.Fatalf("GetDesignCapacity: %v", err)
	}
	if c.Level != models.CapacityLevelDesign {
		t.Errorf("level: want %q got %q", models.CapacityLevelDesign, c.Level)
	}
	if c.PowerWattsMax != 120000 {
		t.Errorf("power_watts_max: want 120000 got %d", c.PowerWattsMax)
	}
}

func TestCapacityService_GetDesignCapacity_Empty(t *testing.T) {
	svc, repo := newCapacitySvc()

	// An empty design returns zero-value capacity (DeviceCount=0).
	repo.designs[6] = &models.CapacitySummary{
		Level:       models.CapacityLevelDesign,
		ID:          6,
		Name:        "empty-design",
		DeviceCount: 0,
	}

	c, err := svc.GetDesignCapacity(6)
	if err != nil {
		t.Fatalf("GetDesignCapacity empty: %v", err)
	}
	if c.DeviceCount != 0 {
		t.Errorf("device_count: want 0 got %d", c.DeviceCount)
	}
	if c.PowerWattsTypical != 0 {
		t.Errorf("power_watts_typical: want 0 got %d", c.PowerWattsTypical)
	}
}
