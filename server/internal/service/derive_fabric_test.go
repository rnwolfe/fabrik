package service_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// fakeDeriveFabricRepo is an in-memory implementation of DeriveFabricRepository.
type fakeDeriveFabricRepo struct {
	designs      map[int64]*models.Design
	sites        map[int64][]*models.Site        // keyed by designID
	superBlocks  map[int64][]*models.SuperBlock  // keyed by siteID
	blocks       map[int64][]*models.Block       // keyed by superBlockID
	aggs         map[string]*models.TierAggregation
	portCounts   map[int64]int
	deviceModels map[int64]*models.DeviceModel
}

func newFakeDeriveFabricRepo() *fakeDeriveFabricRepo {
	return &fakeDeriveFabricRepo{
		designs:      make(map[int64]*models.Design),
		sites:        make(map[int64][]*models.Site),
		superBlocks:  make(map[int64][]*models.SuperBlock),
		blocks:       make(map[int64][]*models.Block),
		aggs:         make(map[string]*models.TierAggregation),
		portCounts:   make(map[int64]int),
		deviceModels: make(map[int64]*models.DeviceModel),
	}
}

func (r *fakeDeriveFabricRepo) aggKey(st models.AggregationScope, id int64, p models.NetworkPlane) string {
	return string(st) + ":" + string(p) + ":" + string(rune(id))
}

func (r *fakeDeriveFabricRepo) GetDesign(id int64) (*models.Design, error) {
	d, ok := r.designs[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return d, nil
}

func (r *fakeDeriveFabricRepo) ListSitesByDesign(designID int64) ([]*models.Site, error) {
	return r.sites[designID], nil
}

func (r *fakeDeriveFabricRepo) ListSuperBlocksBySite(siteID int64) ([]*models.SuperBlock, error) {
	return r.superBlocks[siteID], nil
}

func (r *fakeDeriveFabricRepo) ListBlocksBySuperBlock(superBlockID int64) ([]*models.Block, error) {
	return r.blocks[superBlockID], nil
}

func (r *fakeDeriveFabricRepo) GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error) {
	key := string(scopeType) + ":" + string(plane) + ":" + itoa(scopeID)
	a, ok := r.aggs[key]
	if !ok {
		return nil, models.ErrNotFound
	}
	return a, nil
}

func (r *fakeDeriveFabricRepo) CountAllocatedPorts(aggID int64) (int, error) {
	return r.portCounts[aggID], nil
}

func (r *fakeDeriveFabricRepo) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm, ok := r.deviceModels[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return dm, nil
}

func itoa(i int64) string {
	return string(rune(i))
}

// setAgg registers an aggregation in the fake repo using consistent key format.
func (r *fakeDeriveFabricRepo) setAgg(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane, id int64, deviceModelID int64, spineCount int) {
	key := string(scopeType) + ":" + string(plane) + ":" + itoa(scopeID)
	r.aggs[key] = &models.TierAggregation{
		ID:            id,
		ScopeType:     scopeType,
		ScopeID:       scopeID,
		Plane:         plane,
		DeviceModelID: deviceModelID,
		SpineCount:    spineCount,
	}
}

// ── helpers to populate hierarchy ───────────────────────────────────────────

func (r *fakeDeriveFabricRepo) addDesign(id int64) {
	r.designs[id] = &models.Design{ID: id, Name: "test"}
}

func (r *fakeDeriveFabricRepo) addSite(designID, siteID int64) {
	r.sites[designID] = append(r.sites[designID], &models.Site{ID: siteID, DesignID: designID, Name: "site-1"})
}

func (r *fakeDeriveFabricRepo) addSuperBlock(siteID, sbID int64) {
	r.superBlocks[siteID] = append(r.superBlocks[siteID], &models.SuperBlock{ID: sbID, SiteID: siteID, Name: "sb-1"})
}

func (r *fakeDeriveFabricRepo) addBlock(sbID, blockID int64) {
	r.blocks[sbID] = append(r.blocks[sbID], &models.Block{ID: blockID, SuperBlockID: sbID, Name: "row-A"})
}

func (r *fakeDeriveFabricRepo) addDeviceModel(id int64, portCount int) {
	r.deviceModels[id] = &models.DeviceModel{ID: id, PortCount: portCount, Vendor: "test", Model: "switch"}
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestDeriveFabric_DesignNotFound(t *testing.T) {
	repo := newFakeDeriveFabricRepo()
	svc := service.NewDeriveFabricService(repo)

	_, err := svc.DeriveFabric(999, models.PlaneFrontEnd)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeriveFabric_EmptyDesign_Returns2Stage(t *testing.T) {
	repo := newFakeDeriveFabricRepo()
	repo.addDesign(1)
	svc := service.NewDeriveFabricService(repo)

	df, err := svc.DeriveFabric(1, models.PlaneFrontEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if df.Stages < 2 {
		t.Errorf("expected stages >= 2, got %d", df.Stages)
	}
	if len(df.Tiers) != 0 {
		t.Errorf("expected empty tiers, got %d", len(df.Tiers))
	}
}

func TestDeriveFabric_2Stage_BlockOnly(t *testing.T) {
	repo := newFakeDeriveFabricRepo()
	repo.addDesign(1)
	repo.addSite(1, 10)
	repo.addSuperBlock(10, 20)
	repo.addBlock(20, 30)
	repo.addDeviceModel(100, 48) // 48-port spine switch
	repo.setAgg(models.ScopeBlock, 30, models.PlaneFrontEnd, 1, 100, 4)

	svc := service.NewDeriveFabricService(repo)

	df, err := svc.DeriveFabric(1, models.PlaneFrontEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if df.Stages != 2 {
		t.Errorf("stages = %d, want 2", df.Stages)
	}
	if len(df.Tiers) != 1 {
		t.Fatalf("expected 1 tier, got %d", len(df.Tiers))
	}
	if df.Tiers[0].ScopeType != models.ScopeBlock {
		t.Errorf("tier scope = %s, want block", df.Tiers[0].ScopeType)
	}
	if df.Tiers[0].SpineCount != 4 {
		t.Errorf("spine count = %d, want 4", df.Tiers[0].SpineCount)
	}
	if df.Tiers[0].PortCount != 48 {
		t.Errorf("port count = %d, want 48", df.Tiers[0].PortCount)
	}
}

func TestDeriveFabric_3Stage_BlockAndSuperBlock(t *testing.T) {
	repo := newFakeDeriveFabricRepo()
	repo.addDesign(1)
	repo.addSite(1, 10)
	repo.addSuperBlock(10, 20)
	repo.addBlock(20, 30)
	repo.addDeviceModel(100, 48) // block-level spine (leaf radix=48)
	repo.addDeviceModel(200, 64) // super-block-level super-spine
	repo.setAgg(models.ScopeBlock, 30, models.PlaneFrontEnd, 1, 100, 4)
	repo.setAgg(models.ScopeSuperBlock, 20, models.PlaneFrontEnd, 2, 200, 2)

	svc := service.NewDeriveFabricService(repo)

	df, err := svc.DeriveFabric(1, models.PlaneFrontEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if df.Stages != 3 {
		t.Errorf("stages = %d, want 3", df.Stages)
	}
	if len(df.Tiers) != 2 {
		t.Errorf("expected 2 tiers, got %d", len(df.Tiers))
	}
}

func TestDeriveFabric_TopologyDerivedFromBlockTier(t *testing.T) {
	repo := newFakeDeriveFabricRepo()
	repo.addDesign(1)
	repo.addSite(1, 10)
	repo.addSuperBlock(10, 20)
	repo.addBlock(20, 30)
	repo.addDeviceModel(100, 48) // radix=48, 4 uplinks → 44 downlinks, ~11:1 oversub
	repo.setAgg(models.ScopeBlock, 30, models.PlaneFrontEnd, 1, 100, 4)

	svc := service.NewDeriveFabricService(repo)

	df, err := svc.DeriveFabric(1, models.PlaneFrontEnd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if df.Topology == nil {
		t.Fatal("expected non-nil topology")
	}
	if df.Topology.SpineCount != 4 {
		t.Errorf("topology spine count = %d, want 4", df.Topology.SpineCount)
	}
	if df.Topology.LeafUplinks != 4 {
		t.Errorf("leaf uplinks = %d, want 4", df.Topology.LeafUplinks)
	}
}

func TestDeriveFabric_ManagementPlane_Independent(t *testing.T) {
	repo := newFakeDeriveFabricRepo()
	repo.addDesign(1)
	repo.addSite(1, 10)
	repo.addSuperBlock(10, 20)
	repo.addBlock(20, 30)
	repo.addDeviceModel(100, 48)
	// only front_end agg at block level; management has none
	repo.setAgg(models.ScopeBlock, 30, models.PlaneFrontEnd, 1, 100, 4)

	svc := service.NewDeriveFabricService(repo)

	mgmtDF, err := svc.DeriveFabric(1, models.PlaneManagement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mgmtDF.Tiers) != 0 {
		t.Errorf("expected empty management tiers, got %d", len(mgmtDF.Tiers))
	}
}
