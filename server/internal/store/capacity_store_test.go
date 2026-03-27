package store_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func TestCapacityStore_QueryRackCapacity(t *testing.T) {
	db := openTestDB(t)
	cs := store.NewCapacityStore(db)

	// Set up hierarchy.
	var designID, siteID, superBlockID, blockID, rackID int64
	db.QueryRow(`INSERT INTO designs (name) VALUES ('d') RETURNING id`).Scan(&designID)
	db.QueryRow(`INSERT INTO sites (design_id, name) VALUES (?, 's') RETURNING id`, designID).Scan(&siteID)
	db.QueryRow(`INSERT INTO super_blocks (site_id, name) VALUES (?, 'sb') RETURNING id`, siteID).Scan(&superBlockID)
	db.QueryRow(`INSERT INTO blocks (super_block_id, name) VALUES (?, 'b') RETURNING id`, superBlockID).Scan(&blockID)
	db.QueryRow(`INSERT INTO racks (block_id, name, height_u) VALUES (?, 'r', 42) RETURNING id`, blockID).Scan(&rackID)

	var dmID int64
	db.QueryRow(`
		INSERT INTO device_models (vendor, model, port_count, height_u, power_watts_typical, power_watts_max, cpu_sockets, cores_per_socket, ram_gb, storage_tb, gpu_count)
		VALUES ('V', 'M', 0, 1, 300, 400, 2, 16, 128, 2.0, 1) RETURNING id`).Scan(&dmID)
	db.QueryRow(`INSERT INTO devices (rack_id, device_model_id, name, role, position) VALUES (?, ?, 'srv', 'server', 1) RETURNING id`, rackID, dmID).Scan(new(int64))

	c, err := cs.QueryRackCapacity(rackID)
	if err != nil {
		t.Fatalf("QueryRackCapacity: %v", err)
	}
	if c.Level != models.CapacityLevelRack {
		t.Errorf("level: want rack got %q", c.Level)
	}
	if c.PowerWattsTypical != 300 {
		t.Errorf("power_watts_typical: want 300 got %d", c.PowerWattsTypical)
	}
	if c.PowerWattsMax != 400 {
		t.Errorf("power_watts_max: want 400 got %d", c.PowerWattsMax)
	}
	// vCPU = cpu_sockets * cores_per_socket = 2 * 16 = 32
	if c.TotalVCPU != 32 {
		t.Errorf("total_vcpu: want 32 got %d", c.TotalVCPU)
	}
	if c.TotalRAMGB != 128 {
		t.Errorf("total_ram_gb: want 128 got %d", c.TotalRAMGB)
	}
	if c.TotalStorageTB != 2.0 {
		t.Errorf("total_storage_tb: want 2.0 got %f", c.TotalStorageTB)
	}
	if c.TotalGPUCount != 1 {
		t.Errorf("total_gpu_count: want 1 got %d", c.TotalGPUCount)
	}
	if c.DeviceCount != 1 {
		t.Errorf("device_count: want 1 got %d", c.DeviceCount)
	}
}

func TestCapacityStore_QueryRackCapacity_Empty(t *testing.T) {
	db := openTestDB(t)
	cs := store.NewCapacityStore(db)

	var rackID int64
	var designID, siteID, superBlockID, blockID int64
	db.QueryRow(`INSERT INTO designs (name) VALUES ('d2') RETURNING id`).Scan(&designID)
	db.QueryRow(`INSERT INTO sites (design_id, name) VALUES (?, 's2') RETURNING id`, designID).Scan(&siteID)
	db.QueryRow(`INSERT INTO super_blocks (site_id, name) VALUES (?, 'sb2') RETURNING id`, siteID).Scan(&superBlockID)
	db.QueryRow(`INSERT INTO blocks (super_block_id, name) VALUES (?, 'b2') RETURNING id`, superBlockID).Scan(&blockID)
	db.QueryRow(`INSERT INTO racks (block_id, name, height_u) VALUES (?, 'empty-rack', 42) RETURNING id`, blockID).Scan(&rackID)

	c, err := cs.QueryRackCapacity(rackID)
	if err != nil {
		t.Fatalf("QueryRackCapacity empty rack: %v", err)
	}
	if c.DeviceCount != 0 {
		t.Errorf("device_count: want 0 got %d", c.DeviceCount)
	}
	if c.PowerWattsTypical != 0 {
		t.Errorf("power_watts_typical: want 0 got %d", c.PowerWattsTypical)
	}
}

func TestCapacityStore_QueryRackCapacity_NotFound(t *testing.T) {
	db := openTestDB(t)
	cs := store.NewCapacityStore(db)

	_, err := cs.QueryRackCapacity(999999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCapacityStore_QueryDesignCapacity(t *testing.T) {
	db := openTestDB(t)
	cs := store.NewCapacityStore(db)

	var designID, siteID, superBlockID, blockID, rackID int64
	db.QueryRow(`INSERT INTO designs (name) VALUES ('full-design') RETURNING id`).Scan(&designID)
	db.QueryRow(`INSERT INTO sites (design_id, name) VALUES (?, 's') RETURNING id`, designID).Scan(&siteID)
	db.QueryRow(`INSERT INTO super_blocks (site_id, name) VALUES (?, 'sb') RETURNING id`, siteID).Scan(&superBlockID)
	db.QueryRow(`INSERT INTO blocks (super_block_id, name) VALUES (?, 'b') RETURNING id`, superBlockID).Scan(&blockID)
	db.QueryRow(`INSERT INTO racks (block_id, name, height_u) VALUES (?, 'r', 42) RETURNING id`, blockID).Scan(&rackID)

	var dmID int64
	db.QueryRow(`
		INSERT INTO device_models (vendor, model, port_count, height_u, power_watts_typical, cpu_sockets, cores_per_socket, ram_gb)
		VALUES ('V', 'Server', 0, 1, 500, 4, 32, 512) RETURNING id`).Scan(&dmID)
	db.QueryRow(`INSERT INTO devices (rack_id, device_model_id, name, role, position) VALUES (?, ?, 'srv', 'server', 1) RETURNING id`, rackID, dmID).Scan(new(int64))

	c, err := cs.QueryDesignCapacity(designID)
	if err != nil {
		t.Fatalf("QueryDesignCapacity: %v", err)
	}
	if c.Level != models.CapacityLevelDesign {
		t.Errorf("level: want design got %q", c.Level)
	}
	if c.PowerWattsTypical != 500 {
		t.Errorf("power_watts_typical: want 500 got %d", c.PowerWattsTypical)
	}
	// vCPU = 4 * 32 = 128
	if c.TotalVCPU != 128 {
		t.Errorf("total_vcpu: want 128 got %d", c.TotalVCPU)
	}
	if c.TotalRAMGB != 512 {
		t.Errorf("total_ram_gb: want 512 got %d", c.TotalRAMGB)
	}
}

func TestCapacityStore_QueryDesignCapacity_NotFound(t *testing.T) {
	db := openTestDB(t)
	cs := store.NewCapacityStore(db)

	_, err := cs.QueryDesignCapacity(999999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCapacityStore_MigrationPreservesData(t *testing.T) {
	// Verify that the migration correctly set device_model_type for seed data.
	db := openTestDB(t)
	s := store.NewDeviceModelStore(db)

	list, err := s.List(false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected seed device models, got none")
	}
	for _, dm := range list {
		if dm.DeviceModelType == "" {
			t.Errorf("device model %d (%s %s) has empty device_model_type", dm.ID, dm.Vendor, dm.Model)
		}
	}
}
