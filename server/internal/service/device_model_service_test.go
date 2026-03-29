package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// fakeDeviceModelRepo is an in-memory DeviceModelRepository for service tests.
type fakeDeviceModelRepo struct {
	models map[int64]*models.DeviceModel
	nextID int64
}

func newFakeDMRepo() *fakeDeviceModelRepo {
	return &fakeDeviceModelRepo{models: make(map[int64]*models.DeviceModel)}
}

func (r *fakeDeviceModelRepo) Create(dm *models.DeviceModel) (*models.DeviceModel, error) {
	for _, existing := range r.models {
		if existing.Vendor == dm.Vendor && existing.Model == dm.Model {
			return nil, models.ErrDuplicate
		}
	}
	r.nextID++
	out := *dm
	out.ID = r.nextID
	now := time.Now()
	out.CreatedAt = now
	out.UpdatedAt = now
	r.models[out.ID] = &out
	return &out, nil
}

func (r *fakeDeviceModelRepo) List(includeArchived bool) ([]*models.DeviceModel, error) {
	out := make([]*models.DeviceModel, 0)
	for _, dm := range r.models {
		if !includeArchived && dm.ArchivedAt != nil {
			continue
		}
		cp := *dm
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeDeviceModelRepo) Get(id int64) (*models.DeviceModel, error) {
	dm, ok := r.models[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *dm
	return &cp, nil
}

func (r *fakeDeviceModelRepo) Update(dm *models.DeviceModel) (*models.DeviceModel, error) {
	if _, ok := r.models[dm.ID]; !ok {
		return nil, models.ErrNotFound
	}
	for id, existing := range r.models {
		if id != dm.ID && existing.Vendor == dm.Vendor && existing.Model == dm.Model {
			return nil, models.ErrDuplicate
		}
	}
	out := *dm
	r.models[dm.ID] = &out
	return &out, nil
}

func (r *fakeDeviceModelRepo) Archive(id int64) error {
	dm, ok := r.models[id]
	if !ok {
		return models.ErrNotFound
	}
	now := time.Now()
	dm.ArchivedAt = &now
	return nil
}

func (r *fakeDeviceModelRepo) Duplicate(sourceID int64, newVendor, newModel string) (*models.DeviceModel, error) {
	src, ok := r.models[sourceID]
	if !ok {
		return nil, models.ErrNotFound
	}
	r.nextID++
	out := *src
	out.ID = r.nextID
	out.Vendor = newVendor
	out.Model = newModel
	out.IsSeed = false
	out.ArchivedAt = nil
	r.models[out.ID] = &out
	return &out, nil
}

func (r *fakeDeviceModelRepo) SetPortGroups(deviceModelID int64, groups []models.PortGroup) ([]models.PortGroup, error) {
	dm, ok := r.models[deviceModelID]
	if !ok {
		return nil, models.ErrNotFound
	}
	out := make([]models.PortGroup, len(groups))
	for i, g := range groups {
		out[i] = models.PortGroup{
			ID:            int64(i + 1),
			DeviceModelID: deviceModelID,
			Count:         g.Count,
			SpeedGbps:     g.SpeedGbps,
			Label:         g.Label,
		}
	}
	dm.PortGroups = out
	return out, nil
}

// --- tests ---

func TestDeviceModelService_Create(t *testing.T) {
	tests := []struct {
		name        string
		vendor      string
		model       string
		portCount   int
		heightU     int
		powerWatts  int
		wantErr     bool
		wantErrType error
	}{
		{
			name:    "valid switch",
			vendor:  "Cisco",
			model:   "Nexus 9k",
			heightU: 1,
		},
		{
			name:        "empty vendor",
			vendor:      "",
			model:       "Model X",
			heightU:     1,
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "empty model",
			vendor:      "Vendor Y",
			model:       "",
			heightU:     1,
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "negative port count",
			vendor:      "V",
			model:       "M",
			portCount:   -1,
			heightU:     1,
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "height_u zero",
			vendor:      "V",
			model:       "M",
			heightU:     0,
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "height_u above 50",
			vendor:      "V",
			model:       "M",
			heightU:     51,
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "negative power",
			vendor:      "V",
			model:       "M",
			heightU:     1,
			powerWatts:  -1,
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewDeviceModelService(newFakeDMRepo())
			dm := &models.DeviceModel{
				Vendor:     tc.vendor,
				Model:      tc.model,
				PortCount:  tc.portCount,
				HeightU:    tc.heightU,
				PowerWattsTypical: tc.powerWatts,
			}
			out, err := svc.CreateDeviceModel(dm)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
					t.Errorf("expected %v, got %v", tc.wantErrType, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.ID == 0 {
				t.Error("expected non-zero ID")
			}
		})
	}
}

func TestDeviceModelService_CreateDuplicate(t *testing.T) {
	svc := service.NewDeviceModelService(newFakeDMRepo())
	dm := &models.DeviceModel{Vendor: "V", Model: "M", HeightU: 1}
	svc.CreateDeviceModel(dm)
	_, err := svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "M", HeightU: 1})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !errors.Is(err, models.ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestDeviceModelService_List(t *testing.T) {
	svc := service.NewDeviceModelService(newFakeDMRepo())

	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "Active", HeightU: 1})

	created, _ := svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "ToArchive", HeightU: 1})
	svc.ArchiveDeviceModel(created.ID)

	list, err := svc.ListDeviceModels(false)
	if err != nil {
		t.Fatalf("ListDeviceModels: %v", err)
	}
	for _, dm := range list {
		if dm.ArchivedAt != nil {
			t.Error("archived model should not appear in default list")
		}
	}

	listAll, _ := svc.ListDeviceModels(true)
	if len(listAll) < 2 {
		t.Errorf("expected at least 2 with include_archived, got %d", len(listAll))
	}
}

func TestDeviceModelService_Get(t *testing.T) {
	svc := service.NewDeviceModelService(newFakeDMRepo())
	created, _ := svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "G", HeightU: 1})

	got, err := svc.GetDeviceModel(created.ID)
	if err != nil {
		t.Fatalf("GetDeviceModel: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID mismatch")
	}

	_, err = svc.GetDeviceModel(999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelService_Update(t *testing.T) {
	repo := newFakeDMRepo()
	svc := service.NewDeviceModelService(repo)

	// Seed model - should be protected
	repo.nextID++
	seedID := repo.nextID
	repo.models[seedID] = &models.DeviceModel{ID: seedID, Vendor: "Seed", Model: "SW", HeightU: 1, IsSeed: true}

	_, err := svc.UpdateDeviceModel(&models.DeviceModel{ID: seedID, Vendor: "X", Model: "Y", HeightU: 1})
	if !errors.Is(err, models.ErrSeedReadOnly) {
		t.Errorf("expected ErrSeedReadOnly for seed model update, got %v", err)
	}

	// Normal model update
	created, _ := svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "Orig", HeightU: 1, PortCount: 24})
	updated, err := svc.UpdateDeviceModel(&models.DeviceModel{
		ID:         created.ID,
		Vendor:     "V",
		Model:      "Updated",
		HeightU:    2,
		PortCount:  48,
		PowerWattsTypical: 600,
	})
	if err != nil {
		t.Fatalf("UpdateDeviceModel: %v", err)
	}
	if updated.Model != "Updated" {
		t.Errorf("model: want %q got %q", "Updated", updated.Model)
	}

	// Not found
	_, err = svc.UpdateDeviceModel(&models.DeviceModel{ID: 999999, Vendor: "V", Model: "X", HeightU: 1})
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelService_Archive(t *testing.T) {
	repo := newFakeDMRepo()
	svc := service.NewDeviceModelService(repo)

	// Seed model - should be protected
	repo.nextID++
	seedID := repo.nextID
	repo.models[seedID] = &models.DeviceModel{ID: seedID, Vendor: "Seed", Model: "SW", HeightU: 1, IsSeed: true}

	err := svc.ArchiveDeviceModel(seedID)
	if !errors.Is(err, models.ErrSeedReadOnly) {
		t.Errorf("expected ErrSeedReadOnly for seed archive, got %v", err)
	}

	// Normal archive
	created, _ := svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "Arch", HeightU: 1})
	if err := svc.ArchiveDeviceModel(created.ID); err != nil {
		t.Fatalf("ArchiveDeviceModel: %v", err)
	}

	list, _ := svc.ListDeviceModels(false)
	for _, dm := range list {
		if dm.ID == created.ID {
			t.Error("archived model appeared in default list")
		}
	}

	// Not found
	if err := svc.ArchiveDeviceModel(999999); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelService_Duplicate(t *testing.T) {
	svc := service.NewDeviceModelService(newFakeDMRepo())

	src, _ := svc.CreateDeviceModel(&models.DeviceModel{
		Vendor:     "OrigVendor",
		Model:      "OrigModel",
		PortCount:  32,
		HeightU:    2,
		PowerWattsTypical: 400,
	})

	cp, err := svc.DuplicateDeviceModel(src.ID)
	if err != nil {
		t.Fatalf("DuplicateDeviceModel: %v", err)
	}
	if cp.ID == src.ID {
		t.Error("copy should have different ID")
	}
	if cp.IsSeed {
		t.Error("copy should not be a seed")
	}
	if cp.PortCount != src.PortCount {
		t.Errorf("port_count: want %d got %d", src.PortCount, cp.PortCount)
	}

	// Not found
	_, err = svc.DuplicateDeviceModel(999999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
