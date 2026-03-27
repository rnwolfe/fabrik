package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// --- Fake block aggregation repository ---

type fakeBlockAggRepo struct {
	records map[string]*models.BlockAggregation // key: "blockID:plane"
	nextID  int64
}

func newFakeBlockAggRepo() *fakeBlockAggRepo {
	return &fakeBlockAggRepo{records: make(map[string]*models.BlockAggregation)}
}

func aggKey(blockID int64, plane models.NetworkPlane) string {
	return fmt.Sprintf("%d:%s", blockID, plane)
}

func (r *fakeBlockAggRepo) Upsert(agg *models.BlockAggregation) (*models.BlockAggregation, error) {
	k := aggKey(agg.BlockID, agg.Plane)
	existing, ok := r.records[k]
	if ok {
		cp := *existing
		cp.DeviceID = agg.DeviceID
		cp.MaxPorts = agg.MaxPorts
		cp.Description = agg.Description
		r.records[k] = &cp
		return &cp, nil
	}
	r.nextID++
	out := *agg
	out.ID = r.nextID
	r.records[k] = &out
	return &out, nil
}

func (r *fakeBlockAggRepo) Get(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error) {
	k := aggKey(blockID, plane)
	a, ok := r.records[k]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r *fakeBlockAggRepo) ListForBlock(blockID int64) ([]*models.BlockAggregation, error) {
	var out []*models.BlockAggregation
	for _, a := range r.records {
		if a.BlockID == blockID {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockAggRepo) Delete(blockID int64, plane models.NetworkPlane) error {
	k := aggKey(blockID, plane)
	if _, ok := r.records[k]; !ok {
		return models.ErrNotFound
	}
	delete(r.records, k)
	return nil
}

func (r *fakeBlockAggRepo) IncrementUsedPorts(blockID int64, plane models.NetworkPlane, delta int) (*models.BlockAggregation, error) {
	k := aggKey(blockID, plane)
	a, ok := r.records[k]
	if !ok {
		return nil, models.ErrNotFound
	}
	newUsed := a.UsedPorts + delta
	if a.MaxPorts > 0 && newUsed > a.MaxPorts {
		return nil, errors.New("port capacity exceeded")
	}
	cp := *a
	cp.UsedPorts = newUsed
	r.records[k] = &cp
	return &cp, nil
}

func (r *fakeBlockAggRepo) DecrementUsedPorts(blockID int64, plane models.NetworkPlane, delta int) (*models.BlockAggregation, error) {
	k := aggKey(blockID, plane)
	a, ok := r.records[k]
	if !ok {
		return nil, models.ErrNotFound
	}
	newUsed := a.UsedPorts - delta
	if newUsed < 0 {
		newUsed = 0
	}
	cp := *a
	cp.UsedPorts = newUsed
	r.records[k] = &cp
	return &cp, nil
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
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	deviceID := int64(42)
	agg, err := svc.SetManagementAgg(1, &deviceID, 48, "test agg")
	if err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}
	if agg.BlockID != 1 {
		t.Errorf("BlockID = %d, want 1", agg.BlockID)
	}
	if agg.Plane != models.PlaneManagement {
		t.Errorf("Plane = %v, want %v", agg.Plane, models.PlaneManagement)
	}
	if *agg.DeviceID != 42 {
		t.Errorf("DeviceID = %v, want 42", agg.DeviceID)
	}
	if agg.MaxPorts != 48 {
		t.Errorf("MaxPorts = %d, want 48", agg.MaxPorts)
	}
}

func TestSetManagementAgg_NegativeMaxPorts(t *testing.T) {
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	_, err := svc.SetManagementAgg(1, nil, -1, "")
	if err == nil {
		t.Fatal("expected error for negative max_ports")
	}
	if !errors.Is(err, models.ErrConstraintViolation) {
		t.Errorf("expected ErrConstraintViolation, got %v", err)
	}
}

func TestSetManagementAgg_Idempotent(t *testing.T) {
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	_, err := svc.SetManagementAgg(1, nil, 24, "first")
	if err != nil {
		t.Fatalf("first SetManagementAgg() error = %v", err)
	}

	deviceID := int64(5)
	updated, err := svc.SetManagementAgg(1, &deviceID, 48, "second")
	if err != nil {
		t.Fatalf("second SetManagementAgg() error = %v", err)
	}
	if updated.MaxPorts != 48 {
		t.Errorf("MaxPorts = %d, want 48 after update", updated.MaxPorts)
	}
}

func TestGetManagementAgg_NotFound(t *testing.T) {
	repo := newFakeBlockAggRepo()
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
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, nil, 0, ""); err != nil {
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
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	// No agg assigned — should return a warning string, not an error.
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
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, nil, 4, ""); err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}

	for i := 0; i < 4; i++ {
		agg, warning, err := svc.AllocateManagementPort(1)
		if err != nil {
			t.Fatalf("AllocateManagementPort() iteration %d error = %v", i, err)
		}
		if warning != "" {
			t.Errorf("unexpected warning on iteration %d: %s", i, warning)
		}
		if agg.UsedPorts != i+1 {
			t.Errorf("UsedPorts = %d, want %d", agg.UsedPorts, i+1)
		}
	}
}

func TestAllocateManagementPort_CapacityExceeded(t *testing.T) {
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, nil, 2, ""); err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}

	// Fill to capacity.
	for i := 0; i < 2; i++ {
		if _, _, err := svc.AllocateManagementPort(1); err != nil {
			t.Fatalf("AllocateManagementPort() error = %v", err)
		}
	}

	// Exceed capacity — should return ErrConflict.
	_, _, err := svc.AllocateManagementPort(1)
	if err == nil {
		t.Fatal("expected error when exceeding management agg port capacity")
	}
	if !errors.Is(err, models.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestReleaseManagementPort(t *testing.T) {
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	if _, err := svc.SetManagementAgg(1, nil, 4, ""); err != nil {
		t.Fatalf("SetManagementAgg() error = %v", err)
	}

	if _, _, err := svc.AllocateManagementPort(1); err != nil {
		t.Fatalf("AllocateManagementPort() error = %v", err)
	}
	if _, _, err := svc.AllocateManagementPort(1); err != nil {
		t.Fatalf("AllocateManagementPort() error = %v", err)
	}

	if err := svc.ReleaseManagementPort(1); err != nil {
		t.Fatalf("ReleaseManagementPort() error = %v", err)
	}

	agg, err := svc.GetManagementAgg(1)
	if err != nil {
		t.Fatalf("GetManagementAgg() error = %v", err)
	}
	if agg.UsedPorts != 1 {
		t.Errorf("UsedPorts = %d, want 1 after release", agg.UsedPorts)
	}
}

func TestListBlockAggregations_Empty(t *testing.T) {
	repo := newFakeBlockAggRepo()
	svc := service.NewManagementService(repo)

	aggs, err := svc.ListBlockAggregations(99)
	if err != nil {
		t.Fatalf("ListBlockAggregations() error = %v", err)
	}
	if len(aggs) != 0 {
		t.Errorf("expected empty list, got %d items", len(aggs))
	}
}
