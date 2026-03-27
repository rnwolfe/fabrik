package store_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func TestFabricStore_CRUD(t *testing.T) {
	db := openTestDB(t)

	// We need a design to satisfy the FK constraint on fabrics.design_id.
	ds := store.NewDesignStore(db)
	design, err := ds.Create(&models.Design{Name: "fabric-test-design"})
	if err != nil {
		t.Fatalf("create test design: %v", err)
	}

	s := store.NewFabricStore(db)
	p := store.FabricParams{
		DesignID:         design.ID,
		Name:             "test-fabric",
		Tier:             models.FabricTierFrontEnd,
		Stages:           2,
		Radix:            64,
		Oversubscription: 1.0,
		Description:      "a test fabric",
	}

	t.Run("create", func(t *testing.T) {
		f, err := s.Create(p)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if f.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if f.Name != "test-fabric" {
			t.Errorf("expected name %q, got %q", "test-fabric", f.Name)
		}
		if f.Stages != 2 {
			t.Errorf("expected stages=2, got %d", f.Stages)
		}
		if f.Radix != 64 {
			t.Errorf("expected radix=64, got %d", f.Radix)
		}
		if f.Oversubscription != 1.0 {
			t.Errorf("expected oversubscription=1.0, got %.2f", f.Oversubscription)
		}
		if f.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("list", func(t *testing.T) {
		fabrics, err := s.List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(fabrics) == 0 {
			t.Error("expected at least one fabric")
		}
	})

	t.Run("get", func(t *testing.T) {
		created, _ := s.Create(store.FabricParams{
			DesignID: design.ID,
			Name:     "get-test",
			Tier:     models.FabricTierBackEnd,
			Stages:   3, Radix: 48, Oversubscription: 2.0,
		})
		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
		if got.Stages != 3 {
			t.Errorf("expected stages=3, got %d", got.Stages)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := s.Get(999999)
		if err == nil {
			t.Fatal("expected error for missing fabric")
		}
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		created, _ := s.Create(store.FabricParams{
			DesignID: design.ID,
			Name:     "update-test",
			Tier:     models.FabricTierFrontEnd,
			Stages:   2, Radix: 64, Oversubscription: 1.0,
		})
		updated, err := s.Update(created.ID, store.FabricParams{
			Name:             "updated-fabric",
			Tier:             models.FabricTierBackEnd,
			Stages:           3,
			Radix:            48,
			Oversubscription: 2.0,
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Name != "updated-fabric" {
			t.Errorf("expected name %q, got %q", "updated-fabric", updated.Name)
		}
		if updated.Stages != 3 {
			t.Errorf("expected stages=3, got %d", updated.Stages)
		}
	})

	t.Run("update not found", func(t *testing.T) {
		_, err := s.Update(999999, store.FabricParams{
			Name: "x", Tier: models.FabricTierFrontEnd,
			Stages: 2, Radix: 64, Oversubscription: 1.0,
		})
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		created, _ := s.Create(store.FabricParams{
			DesignID: design.ID,
			Name:     "delete-test",
			Tier:     models.FabricTierFrontEnd,
			Stages:   2, Radix: 64, Oversubscription: 1.0,
		})
		if err := s.Delete(created.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := s.Get(created.ID)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		err := s.Delete(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestFabricStore_WithModelIDs(t *testing.T) {
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	design, _ := ds.Create(&models.Design{Name: "model-test-design"})

	// Insert a device model directly.
	_, err := db.Exec(
		`INSERT INTO device_models (vendor, model, port_count, height_u, power_watts_typical)
		 VALUES ('Arista', '7050CX3', 64, 1, 400)`,
	)
	if err != nil {
		t.Fatalf("insert device model: %v", err)
	}
	var dmID int64
	db.QueryRow("SELECT id FROM device_models WHERE model = '7050CX3'").Scan(&dmID)

	s := store.NewFabricStore(db)
	f, err := s.Create(store.FabricParams{
		DesignID: design.ID,
		Name:     "model-fabric",
		Tier:     models.FabricTierFrontEnd,
		Stages:   2, Radix: 64, Oversubscription: 1.0,
		LeafModelID: dmID,
	})
	if err != nil {
		t.Fatalf("Create with model: %v", err)
	}
	if f.LeafModelID == nil || *f.LeafModelID != dmID {
		t.Errorf("expected LeafModelID=%d, got %v", dmID, f.LeafModelID)
	}

	// Verify we can look up the device model.
	dm, err := s.GetDeviceModelByID(dmID)
	if err != nil {
		t.Fatalf("GetDeviceModelByID: %v", err)
	}
	if dm.Model != "7050CX3" {
		t.Errorf("expected model=7050CX3, got %s", dm.Model)
	}

	// Not found case.
	_, err = s.GetDeviceModelByID(999999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
