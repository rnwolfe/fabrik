package store_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func TestRackTypeStore_CRUD(t *testing.T) {
	db := openTestDB(t)
	s := store.NewRackTypeStore(db)

	t.Run("create", func(t *testing.T) {
		rt, err := s.Create(&models.RackTemplate{Name: "42U-std", HeightU: 42, PowerCapacityW: 10000})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if rt.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if rt.Name != "42U-std" {
			t.Errorf("expected name %q, got %q", "42U-std", rt.Name)
		}
		if rt.HeightU != 42 {
			t.Errorf("expected HeightU 42, got %d", rt.HeightU)
		}
		if rt.PowerCapacityW != 10000 {
			t.Errorf("expected PowerCapacityW 10000, got %d", rt.PowerCapacityW)
		}
		if rt.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("list", func(t *testing.T) {
		s.Create(&models.RackTemplate{Name: "list-type-1", HeightU: 24})
		s.Create(&models.RackTemplate{Name: "list-type-2", HeightU: 48})
		rts, err := s.List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(rts) < 2 {
			t.Errorf("expected at least 2 rack types, got %d", len(rts))
		}
	})

	t.Run("get", func(t *testing.T) {
		created, _ := s.Create(&models.RackTemplate{Name: "get-type", HeightU: 42})
		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := s.Get(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		created, _ := s.Create(&models.RackTemplate{Name: "update-type", HeightU: 42, PowerCapacityW: 5000})
		updated, err := s.Update(&models.RackTemplate{
			ID:             created.ID,
			Name:           "update-type-v2",
			HeightU:        48,
			PowerCapacityW: 8000,
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Name != "update-type-v2" {
			t.Errorf("expected name %q, got %q", "update-type-v2", updated.Name)
		}
		if updated.HeightU != 48 {
			t.Errorf("expected HeightU 48, got %d", updated.HeightU)
		}
	})

	t.Run("update not found", func(t *testing.T) {
		_, err := s.Update(&models.RackTemplate{ID: 999999, Name: "x", HeightU: 42})
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		created, _ := s.Create(&models.RackTemplate{Name: "delete-type", HeightU: 42})
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

func TestRackStore_CRUD(t *testing.T) {
	db := openTestDB(t)
	s := store.NewRackStore(db)

	t.Run("create standalone rack", func(t *testing.T) {
		r, err := s.Create(&models.Rack{Name: "rack-01", HeightU: 42, PowerCapacityW: 10000})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if r.Name != "rack-01" {
			t.Errorf("expected name %q, got %q", "rack-01", r.Name)
		}
		if r.BlockID != nil {
			t.Errorf("expected nil BlockID, got %v", r.BlockID)
		}
	})

	t.Run("list all", func(t *testing.T) {
		s.Create(&models.Rack{Name: "list-rack-1", HeightU: 42})
		s.Create(&models.Rack{Name: "list-rack-2", HeightU: 42})
		racks, err := s.List(nil)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(racks) < 2 {
			t.Errorf("expected at least 2 racks, got %d", len(racks))
		}
	})

	t.Run("get", func(t *testing.T) {
		created, _ := s.Create(&models.Rack{Name: "get-rack", HeightU: 42})
		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := s.Get(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		created, _ := s.Create(&models.Rack{Name: "upd-rack", HeightU: 42})
		updated, err := s.Update(&models.Rack{ID: created.ID, Name: "upd-rack-v2", Description: "updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Name != "upd-rack-v2" {
			t.Errorf("expected name %q, got %q", "upd-rack-v2", updated.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		created, _ := s.Create(&models.Rack{Name: "del-rack", HeightU: 42})
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

func TestRackStore_DevicePlacement(t *testing.T) {
	db := openTestDB(t)
	s := store.NewRackStore(db)

	// Seed a device model.
	var dmID int64
	err := db.QueryRow(`
		INSERT INTO device_models (vendor, model, port_count, height_u, power_watts_typical)
		VALUES ('Cisco', 'ASR9000', 32, 2, 500)
		RETURNING id`).Scan(&dmID)
	if err != nil {
		t.Fatalf("seed device model: %v", err)
	}

	rack, err := s.Create(&models.Rack{Name: "test-rack", HeightU: 42, PowerCapacityW: 5000})
	if err != nil {
		t.Fatalf("create rack: %v", err)
	}

	t.Run("place device", func(t *testing.T) {
		d, err := s.PlaceDevice(&models.Device{
			RackID:        rack.ID,
			DeviceModelID: dmID,
			Name:          "device-1",
			Role:          models.DeviceRoleLeaf,
			Position:      1,
		})
		if err != nil {
			t.Fatalf("PlaceDevice: %v", err)
		}
		if d.ID == 0 {
			t.Error("expected non-zero device ID")
		}
		if d.Position != 1 {
			t.Errorf("expected position 1, got %d", d.Position)
		}
	})

	t.Run("list devices in rack", func(t *testing.T) {
		devices, err := s.ListDevicesInRack(rack.ID)
		if err != nil {
			t.Fatalf("ListDevicesInRack: %v", err)
		}
		if len(devices) == 0 {
			t.Error("expected at least one device")
		}
		// Verify model info is populated.
		if devices[0].ModelVendor == "" {
			t.Error("expected non-empty ModelVendor")
		}
		if devices[0].HeightU == 0 {
			t.Error("expected non-zero HeightU in summary")
		}
	})

	t.Run("move device", func(t *testing.T) {
		d, _ := s.PlaceDevice(&models.Device{
			RackID:        rack.ID,
			DeviceModelID: dmID,
			Name:          "device-move",
			Role:          models.DeviceRoleLeaf,
			Position:      5,
		})
		moved, err := s.MoveDevice(d.ID, rack.ID, 10)
		if err != nil {
			t.Fatalf("MoveDevice: %v", err)
		}
		if moved.Position != 10 {
			t.Errorf("expected position 10, got %d", moved.Position)
		}
	})

	t.Run("remove device no compact", func(t *testing.T) {
		d, _ := s.PlaceDevice(&models.Device{
			RackID:        rack.ID,
			DeviceModelID: dmID,
			Name:          "device-remove",
			Role:          models.DeviceRoleLeaf,
			Position:      20,
		})
		if err := s.RemoveDevice(d.ID, false); err != nil {
			t.Fatalf("RemoveDevice: %v", err)
		}
		_, err := s.GetDevice(d.ID)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound after remove, got %v", err)
		}
	})

	t.Run("get device model", func(t *testing.T) {
		dm, err := s.GetDeviceModel(dmID)
		if err != nil {
			t.Fatalf("GetDeviceModel: %v", err)
		}
		if dm.Vendor != "Cisco" {
			t.Errorf("expected vendor Cisco, got %s", dm.Vendor)
		}
	})

	t.Run("get device model not found", func(t *testing.T) {
		_, err := s.GetDeviceModel(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
