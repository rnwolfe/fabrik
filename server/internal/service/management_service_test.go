package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// --- Fake management agg repository ---

type fakeManagementAggRepo struct {
	aggs       map[string]*models.TierAggregation // key: "scopeType:scopeID:plane"
	portCounts map[int64]int                      // key: aggID → allocated port count
	models     map[int64]*models.DeviceModel      // key: deviceModelID
	nextID     int64
}

func newFakeManagementAggRepo() *fakeManagementAggRepo {
	return &fakeManagementAggRepo{
		aggs:       make(map[string]*models.TierAggregation),
		portCounts: make(map[int64]int),
		models:     make(map[int64]*models.DeviceModel),
	}
}

func (r *fakeManagementAggRepo) aggKey(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) string {
	return fmt.Sprintf("%s:%d:%s", scopeType, scopeID, plane)
}

func (r *fakeManagementAggRepo) addDeviceModel(id int64, portCount int) {
	r.models[id] = &models.DeviceModel{ID: id, PortCount: portCount, Vendor: "test", Model: "test"}
}

func (r *fakeManagementAggRepo) setAllocatedPorts(aggID int64, count int) {
	r.portCounts[aggID] = count
}

func (r *fakeManagementAggRepo) SetAggregation(agg *models.TierAggregation) (*models.TierAggregation, error) {
	key := r.aggKey(agg.ScopeType, agg.ScopeID, agg.Plane)
	if existing, ok := r.aggs[key]; ok {
		cp := *existing
		cp.DeviceModelID = agg.DeviceModelID
		cp.SpineCount = agg.SpineCount
		r.aggs[key] = &cp
		return &cp, nil
	}
	r.nextID++
	out := *agg
	out.ID = r.nextID
	r.aggs[key] = &out
	return &out, nil
}

func (r *fakeManagementAggRepo) GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error) {
	key := r.aggKey(scopeType, scopeID, plane)
	a, ok := r.aggs[key]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *fakeManagementAggRepo) ListAggregations(scopeType models.AggregationScope, scopeID int64) ([]*models.TierAggregation, error) {
	var out []*models.TierAggregation
	for _, a := range r.aggs {
		if a.ScopeType == scopeType && a.ScopeID == scopeID {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeManagementAggRepo) DeleteAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) error {
	key := r.aggKey(scopeType, scopeID, plane)
	if _, ok := r.aggs[key]; !ok {
		return models.ErrNotFound
	}
	delete(r.aggs, key)
	return nil
}

func (r *fakeManagementAggRepo) CountAllocatedPorts(aggID int64) (int, error) {
	return r.portCounts[aggID], nil
}

func (r *fakeManagementAggRepo) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm, ok := r.models[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return dm, nil
}

// --- Tests ---

func TestValidateDeviceRole(t *testing.T) {
	tests := []struct {
		role    string
		wantErr bool
	}{
		{"spine", false},
		{"leaf", false},
		{"super_spine", false},
		{"server", false},
		{"other", false},
		{"management_tor", false},
		{"management_agg", false},
		{"unknown_role", true},
		{"", true},
		{"SPINE", true},
	}
	for _, tc := range tests {
		t.Run(tc.role, func(t *testing.T) {
			err := service.ValidateDeviceRole(tc.role)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateDeviceRole(%q) error = %v, wantErr = %v", tc.role, err, tc.wantErr)
			}
			if tc.wantErr && err != nil {
				if !errors.Is(err, models.ErrConstraintViolation) {
					t.Errorf("expected ErrConstraintViolation, got %v", err)
				}
			}
		})
	}
}

func TestSetManagementAgg(t *testing.T) {
	repo := newFakeManagementAggRepo()
	repo.addDeviceModel(42, 48)
	svc := service.NewManagementService(repo)

	agg, err := svc.SetManagementAgg(1, 42)
	if err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}
	if agg.ScopeID != 1 {
		t.Errorf("ScopeID = %d, want 1", agg.ScopeID)
	}
	if agg.Plane != models.PlaneManagement {
		t.Errorf("Plane = %v, want %v", agg.Plane, models.PlaneManagement)
	}
	if agg.DeviceModelID != 42 {
		t.Errorf("DeviceModelID = %d, want 42", agg.DeviceModelID)
	}
}

func TestSetManagementAgg_ZeroDeviceModelID(t *testing.T) {
	repo := newFakeManagementAggRepo()
	svc := service.NewManagementService(repo)

	_, err := svc.SetManagementAgg(1, 0)
	if err == nil {
		t.Fatal("expected error for zero device_model_id")
	}
	if !errors.Is(err, models.ErrConstraintViolation) {
		t.Errorf("expected ErrConstraintViolation, got %v", err)
	}
}

func TestSetManagementAgg_Idempotent(t *testing.T) {
	repo := newFakeManagementAggRepo()
	repo.addDeviceModel(1, 24)
	repo.addDeviceModel(2, 48)
	svc := service.NewManagementService(repo)

	_, err := svc.SetManagementAgg(1, 1)
	if err != nil {
		t.Fatalf("first SetManagementAgg() error = %v", err)
	}

	updated, err := svc.SetManagementAgg(1, 2)
	if err != nil {
		t.Fatalf("second SetManagementAgg() error = %v", err)
	}
	if updated.DeviceModelID != 2 {
		t.Errorf("DeviceModelID = %d, want 2 after update", updated.DeviceModelID)
	}
}

func TestGetManagementAgg_NotFound(t *testing.T) {
	repo := newFakeManagementAggRepo()
	svc := service.NewManagementService(repo)

	_, err := svc.GetManagementAgg(99)
	if err == nil {
		t.Fatal("expected error for missing block agg")
	}
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRemoveManagementAgg(t *testing.T) {
	repo := newFakeManagementAggRepo()
	repo.addDeviceModel(1, 24)
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, 1); err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}

	if err := svc.RemoveManagementAgg(1); err != nil {
		t.Fatalf("RemoveManagementAgg() error = %v", err)
	}

	_, err := svc.GetManagementAgg(1)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound after removal, got %v", err)
	}
}

func TestAllocateManagementPort_NoAgg(t *testing.T) {
	repo := newFakeManagementAggRepo()
	svc := service.NewManagementService(repo)

	agg, warning, err := svc.AllocateManagementPort(1)
	if err != nil {
		t.Fatalf("AllocateManagementPort() error = %v", err)
	}
	if agg != nil {
		t.Errorf("expected nil agg when no agg assigned, got %+v", agg)
	}
	if warning == "" {
		t.Error("expected a warning string when no agg assigned")
	}
}

func TestAllocateManagementPort_WithinCapacity(t *testing.T) {
	repo := newFakeManagementAggRepo()
	repo.addDeviceModel(1, 4)
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, 1); err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}

	for i := 0; i < 4; i++ {
		agg, err := repo.GetAggregation(models.ScopeBlock, 1, models.PlaneManagement)
		if err != nil {
			t.Fatalf("GetAggregation() error = %v", err)
		}
		repo.setAllocatedPorts(agg.ID, i)

		result, warning, err := svc.AllocateManagementPort(1)
		if err != nil {
			t.Fatalf("AllocateManagementPort() iteration %d error = %v", i, err)
		}
		if warning != "" {
			t.Errorf("unexpected warning on iteration %d: %s", i, warning)
		}
		if result == nil {
			t.Errorf("expected non-nil agg on iteration %d", i)
		}
	}
}

func TestAllocateManagementPort_CapacityExceeded(t *testing.T) {
	repo := newFakeManagementAggRepo()
	repo.addDeviceModel(1, 2)
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, 1); err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}

	agg, err := repo.GetAggregation(models.ScopeBlock, 1, models.PlaneManagement)
	if err != nil {
		t.Fatalf("GetAggregation() error = %v", err)
	}
	repo.setAllocatedPorts(agg.ID, 2)

	_, _, err = svc.AllocateManagementPort(1)
	if err == nil {
		t.Fatal("expected error when exceeding management agg port capacity")
	}
	if !errors.Is(err, models.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestListBlockAggregations_Empty(t *testing.T) {
	repo := newFakeManagementAggRepo()
	svc := service.NewManagementService(repo)

	aggs, err := svc.ListBlockAggregations(99)
	if err != nil {
		t.Fatalf("ListBlockAggregations() error = %v", err)
	}
	if len(aggs) != 0 {
		t.Errorf("expected empty list, got %d items", len(aggs))
	}
}
