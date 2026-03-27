package service_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// inMemFabricRepo is an in-memory implementation of FabricRepository for tests.
type inMemFabricRepo struct {
	fabrics      map[int64]*store.FabricRecord
	deviceModels map[int64]*models.DeviceModel
	nextID       int64
}

func newInMemFabricRepo() *inMemFabricRepo {
	return &inMemFabricRepo{
		fabrics:      make(map[int64]*store.FabricRecord),
		deviceModels: make(map[int64]*models.DeviceModel),
	}
}

func (r *inMemFabricRepo) Create(p store.FabricParams) (*store.FabricRecord, error) {
	r.nextID++
	rec := &store.FabricRecord{
		Fabric: models.Fabric{
			ID:          r.nextID,
			DesignID:    p.DesignID,
			Name:        p.Name,
			Tier:        p.Tier,
			Description: p.Description,
		},
		Stages:           p.Stages,
		Radix:            p.Radix,
		Oversubscription: p.Oversubscription,
	}
	if p.LeafModelID != 0 {
		id := p.LeafModelID
		rec.LeafModelID = &id
	}
	if p.SpineModelID != 0 {
		id := p.SpineModelID
		rec.SpineModelID = &id
	}
	if p.SuperSpineModelID != 0 {
		id := p.SuperSpineModelID
		rec.SuperSpineModelID = &id
	}
	r.fabrics[rec.ID] = rec
	return rec, nil
}

func (r *inMemFabricRepo) List() ([]*store.FabricRecord, error) {
	out := make([]*store.FabricRecord, 0, len(r.fabrics))
	for _, f := range r.fabrics {
		cp := *f
		out = append(out, &cp)
	}
	return out, nil
}

func (r *inMemFabricRepo) Get(id int64) (*store.FabricRecord, error) {
	f, ok := r.fabrics[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *f
	return &cp, nil
}

func (r *inMemFabricRepo) Update(id int64, p store.FabricParams) (*store.FabricRecord, error) {
	if _, ok := r.fabrics[id]; !ok {
		return nil, models.ErrNotFound
	}
	rec := &store.FabricRecord{
		Fabric: models.Fabric{
			ID:          id,
			DesignID:    r.fabrics[id].DesignID,
			Name:        p.Name,
			Tier:        p.Tier,
			Description: p.Description,
		},
		Stages:           p.Stages,
		Radix:            p.Radix,
		Oversubscription: p.Oversubscription,
	}
	if p.LeafModelID != 0 {
		lid := p.LeafModelID
		rec.LeafModelID = &lid
	}
	r.fabrics[id] = rec
	return rec, nil
}

func (r *inMemFabricRepo) Delete(id int64) error {
	if _, ok := r.fabrics[id]; !ok {
		return models.ErrNotFound
	}
	delete(r.fabrics, id)
	return nil
}

func (r *inMemFabricRepo) GetDeviceModelByID(id int64) (*models.DeviceModel, error) {
	dm, ok := r.deviceModels[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *dm
	return &cp, nil
}

func (r *inMemFabricRepo) ListDeviceModels() ([]*models.DeviceModel, error) {
	out := make([]*models.DeviceModel, 0, len(r.deviceModels))
	for _, dm := range r.deviceModels {
		cp := *dm
		out = append(out, &cp)
	}
	return out, nil
}

// --- Tests ---

func TestFabricService_CreateFabric_Valid(t *testing.T) {
	repo := newInMemFabricRepo()
	svc := service.NewFabricService(repo)

	resp, err := svc.CreateFabric(service.CreateFabricRequest{
		DesignID:         1,
		Name:             "test-fabric",
		Tier:             models.FabricTierFrontEnd,
		Stages:           2,
		Radix:            64,
		Oversubscription: 1.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if resp.Name != "test-fabric" {
		t.Errorf("expected name %q, got %q", "test-fabric", resp.Name)
	}
	if resp.Topology == nil {
		t.Error("expected topology to be populated")
	}
	if resp.Metrics == nil {
		t.Error("expected metrics to be populated")
	}
}

func TestFabricService_CreateFabric_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     service.CreateFabricRequest
		wantErr error
	}{
		{
			name: "empty fabric name",
			req: service.CreateFabricRequest{
				Name: "", Stages: 2, Radix: 64, Oversubscription: 1.0, Tier: models.FabricTierFrontEnd,
			},
			wantErr: models.ErrConstraintViolation,
		},
		{
			name: "invalid stages",
			req: service.CreateFabricRequest{
				Name: "test", Stages: 4, Radix: 64, Oversubscription: 1.0, Tier: models.FabricTierFrontEnd,
			},
			wantErr: models.ErrConstraintViolation,
		},
		{
			name: "zero radix",
			req: service.CreateFabricRequest{
				Name: "test", Stages: 2, Radix: 0, Oversubscription: 1.0, Tier: models.FabricTierFrontEnd,
			},
			wantErr: models.ErrConstraintViolation,
		},
		{
			name: "oversubscription below 1",
			req: service.CreateFabricRequest{
				Name: "test", Stages: 2, Radix: 64, Oversubscription: 0.5, Tier: models.FabricTierFrontEnd,
			},
			wantErr: models.ErrConstraintViolation,
		},
		{
			name: "invalid tier",
			req: service.CreateFabricRequest{
				Name: "test", Stages: 2, Radix: 64, Oversubscription: 1.0, Tier: "invalid",
			},
			wantErr: models.ErrConstraintViolation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewFabricService(newInMemFabricRepo())
			_, err := svc.CreateFabric(tc.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestFabricService_ListFabrics(t *testing.T) {
	repo := newInMemFabricRepo()
	svc := service.NewFabricService(repo)

	// Empty list.
	fabrics, err := svc.ListFabrics()
	if err != nil {
		t.Fatalf("ListFabrics: %v", err)
	}
	if len(fabrics) != 0 {
		t.Errorf("expected 0 fabrics, got %d", len(fabrics))
	}

	// Add two fabrics.
	validReq := service.CreateFabricRequest{
		DesignID: 1, Name: "f1", Tier: models.FabricTierFrontEnd,
		Stages: 2, Radix: 64, Oversubscription: 1.0,
	}
	svc.CreateFabric(validReq)
	validReq.Name = "f2"
	svc.CreateFabric(validReq)

	fabrics, err = svc.ListFabrics()
	if err != nil {
		t.Fatalf("ListFabrics after create: %v", err)
	}
	if len(fabrics) != 2 {
		t.Errorf("expected 2 fabrics, got %d", len(fabrics))
	}
}

func TestFabricService_GetFabric(t *testing.T) {
	repo := newInMemFabricRepo()
	svc := service.NewFabricService(repo)

	created, err := svc.CreateFabric(service.CreateFabricRequest{
		DesignID: 1, Name: "get-test", Tier: models.FabricTierBackEnd,
		Stages: 3, Radix: 48, Oversubscription: 2.0,
	})
	if err != nil {
		t.Fatalf("CreateFabric: %v", err)
	}

	got, err := svc.GetFabric(created.ID)
	if err != nil {
		t.Fatalf("GetFabric: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Topology.Stages != 3 {
		t.Errorf("expected stages=3, got %d", got.Topology.Stages)
	}

	// Not found.
	_, err = svc.GetFabric(99999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFabricService_UpdateFabric(t *testing.T) {
	repo := newInMemFabricRepo()
	svc := service.NewFabricService(repo)

	created, err := svc.CreateFabric(service.CreateFabricRequest{
		DesignID: 1, Name: "update-test", Tier: models.FabricTierFrontEnd,
		Stages: 2, Radix: 64, Oversubscription: 1.0,
	})
	if err != nil {
		t.Fatalf("CreateFabric: %v", err)
	}

	updated, err := svc.UpdateFabric(created.ID, service.UpdateFabricRequest{
		Name: "updated-fabric", Tier: models.FabricTierBackEnd,
		Stages: 3, Radix: 48, Oversubscription: 2.0,
	})
	if err != nil {
		t.Fatalf("UpdateFabric: %v", err)
	}
	if updated.Name != "updated-fabric" {
		t.Errorf("expected name %q, got %q", "updated-fabric", updated.Name)
	}
	if updated.Topology.Stages != 3 {
		t.Errorf("expected stages=3 after update, got %d", updated.Topology.Stages)
	}

	// Not found.
	_, err = svc.UpdateFabric(99999, service.UpdateFabricRequest{
		Name: "x", Tier: models.FabricTierFrontEnd, Stages: 2, Radix: 64, Oversubscription: 1.0,
	})
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFabricService_DeleteFabric(t *testing.T) {
	repo := newInMemFabricRepo()
	svc := service.NewFabricService(repo)

	created, err := svc.CreateFabric(service.CreateFabricRequest{
		DesignID: 1, Name: "delete-test", Tier: models.FabricTierFrontEnd,
		Stages: 2, Radix: 64, Oversubscription: 1.0,
	})
	if err != nil {
		t.Fatalf("CreateFabric: %v", err)
	}

	if err := svc.DeleteFabric(created.ID); err != nil {
		t.Fatalf("DeleteFabric: %v", err)
	}

	_, err = svc.GetFabric(created.ID)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}

	// Delete non-existent.
	if err := svc.DeleteFabric(99999); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFabricService_PreviewTopology_Service(t *testing.T) {
	svc := service.NewFabricService(newInMemFabricRepo())

	plan, err := svc.PreviewTopology(2, 64, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Stages != 2 {
		t.Errorf("expected stages=2, got %d", plan.Stages)
	}
	if plan.TotalHostPorts == 0 {
		t.Error("expected non-zero TotalHostPorts")
	}

	// Invalid params should return error.
	_, err = svc.PreviewTopology(4, 64, 1.0)
	if err == nil {
		t.Fatal("expected error for invalid stages")
	}
	if !errors.Is(err, models.ErrConstraintViolation) {
		t.Errorf("expected ErrConstraintViolation, got %v", err)
	}
}

func TestFabricService_DeviceModelWarning(t *testing.T) {
	repo := newInMemFabricRepo()
	// Add a device model with port_count=32; leaf radix=64 — mismatch.
	repo.deviceModels[1] = &models.DeviceModel{
		ID: 1, Vendor: "Arista", Model: "7050CX3-32S", PortCount: 32,
	}

	svc := service.NewFabricService(repo)

	resp, err := svc.CreateFabric(service.CreateFabricRequest{
		DesignID: 1, Name: "dm-test", Tier: models.FabricTierFrontEnd,
		Stages: 2, Radix: 64, Oversubscription: 1.0,
		LeafModelID: 1,
	})
	if err != nil {
		t.Fatalf("CreateFabric: %v", err)
	}
	// The leaf model should be resolved.
	if resp.LeafModel == nil {
		t.Error("expected LeafModel to be populated")
	}
	if resp.LeafModel.ID != 1 {
		t.Errorf("expected LeafModel.ID=1, got %d", resp.LeafModel.ID)
	}
}
