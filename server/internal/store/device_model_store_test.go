package store_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func TestDeviceModelStore_Create(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	dm, err := s.Create(&models.DeviceModel{
		Vendor:     "Acme",
		Model:      "Switch X",
		PortCount:  48,
		HeightU:    1,
		PowerWattsTypical: 300,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if dm.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if dm.Vendor != "Acme" {
		t.Errorf("vendor: want %q got %q", "Acme", dm.Vendor)
	}
	if dm.IsSeed {
		t.Error("expected IsSeed=false for user-created model")
	}
	if dm.ArchivedAt != nil {
		t.Error("expected ArchivedAt=nil for new model")
	}
	if dm.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestDeviceModelStore_CreateDuplicate(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	s.Create(&models.DeviceModel{Vendor: "Acme", Model: "Dup", HeightU: 1})
	_, err := s.Create(&models.DeviceModel{Vendor: "Acme", Model: "Dup", HeightU: 1})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !errors.Is(err, models.ErrDuplicate) {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestDeviceModelStore_List(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	// Non-archived
	s.Create(&models.DeviceModel{Vendor: "VendorA", Model: "Active", HeightU: 1})

	// Archived
	dm, _ := s.Create(&models.DeviceModel{Vendor: "VendorA", Model: "Archived", HeightU: 1})
	s.Archive(dm.ID)

	tests := []struct {
		includeArchived bool
		wantMin         int
	}{
		{includeArchived: false, wantMin: 1},
		{includeArchived: true, wantMin: 2},
	}

	for _, tc := range tests {
		list, err := s.List(tc.includeArchived)
		if err != nil {
			t.Fatalf("List(includeArchived=%v): %v", tc.includeArchived, err)
		}
		// Seed rows from migration are also present
		if len(list) < tc.wantMin {
			t.Errorf("List(includeArchived=%v): want at least %d, got %d", tc.includeArchived, tc.wantMin, len(list))
		}
	}
}

func TestDeviceModelStore_Get(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	created, _ := s.Create(&models.DeviceModel{Vendor: "VendorG", Model: "ModelG", HeightU: 1})

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: want %d got %d", created.ID, got.ID)
	}

	_, err = s.Get(999999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelStore_Update(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	created, _ := s.Create(&models.DeviceModel{Vendor: "OldVendor", Model: "OldModel", HeightU: 1, PortCount: 24})

	updated, err := s.Update(&models.DeviceModel{
		ID:         created.ID,
		Vendor:     "NewVendor",
		Model:      "NewModel",
		HeightU:    2,
		PortCount:  48,
		PowerWattsTypical: 500,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Vendor != "NewVendor" {
		t.Errorf("vendor: want %q got %q", "NewVendor", updated.Vendor)
	}
	if updated.PortCount != 48 {
		t.Errorf("port_count: want 48 got %d", updated.PortCount)
	}

	// Not found
	_, err = s.Update(&models.DeviceModel{ID: 999999, Vendor: "X", Model: "Y", HeightU: 1})
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelStore_Archive(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	created, _ := s.Create(&models.DeviceModel{Vendor: "ArchVendor", Model: "ArchModel", HeightU: 1})

	if err := s.Archive(created.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Should still be retrievable by Get
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get after archive: %v", err)
	}
	if got.ArchivedAt == nil {
		t.Error("expected ArchivedAt to be set after archive")
	}

	// Should not appear in default list
	list, _ := s.List(false)
	for _, dm := range list {
		if dm.ID == created.ID {
			t.Error("archived model should not appear in default list")
		}
	}

	// Should appear when includeArchived=true
	listAll, _ := s.List(true)
	found := false
	for _, dm := range listAll {
		if dm.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Error("archived model should appear when includeArchived=true")
	}

	// Archive non-existent
	if err := s.Archive(999999); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelStore_Duplicate(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	src, _ := s.Create(&models.DeviceModel{
		Vendor:      "DupVendor",
		Model:       "DupModel",
		PortCount:   32,
		HeightU:     2,
		PowerWattsTypical:  400,
		Description: "original",
	})

	cp, err := s.Duplicate(src.ID, "DupVendor", "DupModel (copy)")
	if err != nil {
		t.Fatalf("Duplicate: %v", err)
	}
	if cp.ID == src.ID {
		t.Error("copy should have a different ID")
	}
	if cp.IsSeed {
		t.Error("copy should not be a seed")
	}
	if cp.PortCount != src.PortCount {
		t.Errorf("port_count: want %d got %d", src.PortCount, cp.PortCount)
	}
	if cp.Vendor != src.Vendor {
		t.Errorf("vendor: want %q got %q", src.Vendor, cp.Vendor)
	}

	// Duplicate non-existent
	_, err = s.Duplicate(999999, "X", "Y")
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceModelStore_SeedData(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	list, err := s.List(false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	// Expect at least the 6 seed rows from migration 0002.
	if len(list) < 6 {
		t.Errorf("expected at least 6 seed rows, got %d", len(list))
	}

	seedCount := 0
	for _, dm := range list {
		if dm.IsSeed {
			seedCount++
		}
	}
	if seedCount < 6 {
		t.Errorf("expected at least 6 seed models, got %d", seedCount)
	}
}
