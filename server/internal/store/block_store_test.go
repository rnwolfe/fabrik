package store_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// seedBlockHier creates design → site → super_block, returns superBlockID.
func seedBlockHier(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var designID, siteID, superBlockID int64
	if err := db.QueryRow(`INSERT INTO designs (name) VALUES ('test') RETURNING id`).Scan(&designID); err != nil {
		t.Fatalf("insert design: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO sites (design_id, name) VALUES (?, 'site-1') RETURNING id`, designID).Scan(&siteID); err != nil {
		t.Fatalf("insert site: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO super_blocks (site_id, name) VALUES (?, 'sb-1') RETURNING id`, siteID).Scan(&superBlockID); err != nil {
		t.Fatalf("insert super_block: %v", err)
	}
	return superBlockID
}

var dmCounter int // package-level counter for unique device model names

// seedDeviceModel inserts a device_model with a unique name and returns its ID.
func seedDeviceModel(t *testing.T, db *sql.DB, portCount int) int64 {
	t.Helper()
	dmCounter++
	var id int64
	if err := db.QueryRow(
		`INSERT INTO device_models (vendor, model, port_count, height_u) VALUES ('test', ?, ?, 1) RETURNING id`,
		fmt.Sprintf("agg-%d", dmCounter), portCount,
	).Scan(&id); err != nil {
		t.Fatalf("insert device model: %v", err)
	}
	return id
}

// seedRack inserts a rack (no block) and returns its ID.
func seedRack(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(
		`INSERT INTO racks (name, height_u, description) VALUES (?, 42, '') RETURNING id`, name,
	).Scan(&id); err != nil {
		t.Fatalf("insert rack: %v", err)
	}
	return id
}

func TestBlockStore_CreateAndGetBlock(t *testing.T) {
	db := openTestDB(t)
	superBlockID := seedBlockHier(t, db)
	s := store.NewBlockStore(db)

	t.Run("create", func(t *testing.T) {
		b, err := s.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-A", Description: "first row"})
		if err != nil {
			t.Fatalf("CreateBlock: %v", err)
		}
		if b.ID == 0 {
			t.Error("expected non-zero block ID")
		}
		if b.Name != "row-A" {
			t.Errorf("expected name 'row-A', got %q", b.Name)
		}
		if b.SuperBlockID != superBlockID {
			t.Errorf("expected super_block_id %d, got %d", superBlockID, b.SuperBlockID)
		}
		if b.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("get existing", func(t *testing.T) {
		created, _ := s.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-B"})
		got, err := s.GetBlock(created.ID)
		if err != nil {
			t.Fatalf("GetBlock: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := s.GetBlock(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestBlockStore_ListBlocks(t *testing.T) {
	db := openTestDB(t)
	superBlockID := seedBlockHier(t, db)
	s := store.NewBlockStore(db)

	s.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-A"})
	s.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-B"})
	s.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-C"})

	blocks, err := s.ListBlocks(superBlockID)
	if err != nil {
		t.Fatalf("ListBlocks: %v", err)
	}
	if len(blocks) != 3 {
		t.Errorf("expected 3 blocks, got %d", len(blocks))
	}
}

func TestBlockStore_GetDefaultBlock(t *testing.T) {
	db := openTestDB(t)
	superBlockID := seedBlockHier(t, db)
	s := store.NewBlockStore(db)

	t.Run("no default block returns nil", func(t *testing.T) {
		b, err := s.GetDefaultBlock(superBlockID)
		if err != nil {
			t.Fatalf("GetDefaultBlock: %v", err)
		}
		if b != nil {
			t.Errorf("expected nil, got %+v", b)
		}
	})

	t.Run("finds default block by name", func(t *testing.T) {
		s.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "default"})
		b, err := s.GetDefaultBlock(superBlockID)
		if err != nil {
			t.Fatalf("GetDefaultBlock: %v", err)
		}
		if b == nil {
			t.Fatal("expected default block, got nil")
		}
		if b.Name != "default" {
			t.Errorf("expected name 'default', got %q", b.Name)
		}
	})
}

func TestTierAggregationStore_SetAndGetAggregation(t *testing.T) {
	db := openTestDB(t)
	superBlockID := seedBlockHier(t, db)
	deviceModelID := seedDeviceModel(t, db, 32)
	bs := store.NewBlockStore(db)
	s := store.NewTierAggregationStore(db)

	block, _ := bs.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-A"})

	t.Run("set aggregation", func(t *testing.T) {
		agg, err := s.SetAggregation(&models.TierAggregation{
			ScopeType:     models.ScopeBlock,
			ScopeID:       block.ID,
			Plane:         models.NetworkPlaneFrontEnd,
			DeviceModelID: deviceModelID,
			SpineCount:    2,
		})
		if err != nil {
			t.Fatalf("SetAggregation: %v", err)
		}
		if agg.ID == 0 {
			t.Error("expected non-zero agg ID")
		}
		if agg.Plane != models.NetworkPlaneFrontEnd {
			t.Errorf("expected plane frontend, got %q", agg.Plane)
		}
		if agg.SpineCount != 2 {
			t.Errorf("expected spine_count 2, got %d", agg.SpineCount)
		}
	})

	t.Run("get aggregation", func(t *testing.T) {
		got, err := s.GetAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneFrontEnd)
		if err != nil {
			t.Fatalf("GetAggregation: %v", err)
		}
		if got.DeviceModelID != deviceModelID {
			t.Errorf("expected device_model_id %d, got %d", deviceModelID, got.DeviceModelID)
		}
	})

	t.Run("get not found for management plane", func(t *testing.T) {
		_, err := s.GetAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneManagement)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("upsert replaces device model", func(t *testing.T) {
		dm2ID := seedDeviceModel(t, db, 64)
		agg, err := s.SetAggregation(&models.TierAggregation{
			ScopeType:     models.ScopeBlock,
			ScopeID:       block.ID,
			Plane:         models.NetworkPlaneFrontEnd,
			DeviceModelID: dm2ID,
			SpineCount:    4,
		})
		if err != nil {
			t.Fatalf("SetAggregation upsert: %v", err)
		}
		if agg.DeviceModelID != dm2ID {
			t.Errorf("expected updated device_model_id %d, got %d", dm2ID, agg.DeviceModelID)
		}
		if agg.SpineCount != 4 {
			t.Errorf("expected spine_count 4 after upsert, got %d", agg.SpineCount)
		}

		aggs, _ := s.ListAggregations(models.ScopeBlock, block.ID)
		if len(aggs) != 1 {
			t.Errorf("expected 1 agg row after upsert, got %d", len(aggs))
		}
	})
}

func TestTierAggregationStore_AllocateAndDeallocatePorts(t *testing.T) {
	db := openTestDB(t)
	superBlockID := seedBlockHier(t, db)
	deviceModelID := seedDeviceModel(t, db, 32)
	rackID := seedRack(t, db, "r1")

	bs := store.NewBlockStore(db)
	s := store.NewTierAggregationStore(db)
	block, _ := bs.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-A"})
	agg, _ := s.SetAggregation(&models.TierAggregation{
		ScopeType:     models.ScopeBlock,
		ScopeID:       block.ID,
		Plane:         models.NetworkPlaneFrontEnd,
		DeviceModelID: deviceModelID,
	})

	t.Run("allocate ports", func(t *testing.T) {
		conns, err := s.AllocatePorts(agg.ID, rackID, []string{"leaf-1", "leaf-2"}, 0)
		if err != nil {
			t.Fatalf("AllocatePorts: %v", err)
		}
		if len(conns) != 2 {
			t.Errorf("expected 2 connections, got %d", len(conns))
		}
		if conns[0].AggPortIndex != 0 {
			t.Errorf("expected port index 0, got %d", conns[0].AggPortIndex)
		}
		if conns[1].AggPortIndex != 1 {
			t.Errorf("expected port index 1, got %d", conns[1].AggPortIndex)
		}
	})

	t.Run("count allocated ports", func(t *testing.T) {
		count, err := s.CountAllocatedPorts(agg.ID)
		if err != nil {
			t.Fatalf("CountAllocatedPorts: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2 allocated, got %d", count)
		}
	})

	t.Run("list port connections", func(t *testing.T) {
		conns, err := s.ListPortConnections(agg.ID)
		if err != nil {
			t.Fatalf("ListPortConnections: %v", err)
		}
		if len(conns) != 2 {
			t.Errorf("expected 2 connections, got %d", len(conns))
		}
	})

	t.Run("deallocate ports by child", func(t *testing.T) {
		if err := s.DeallocatePortsByChild(rackID); err != nil {
			t.Fatalf("DeallocatePortsByChild: %v", err)
		}
		count, _ := s.CountAllocatedPorts(agg.ID)
		if count != 0 {
			t.Errorf("expected 0 allocated after dealloc, got %d", count)
		}
	})
}

func TestTierAggregationStore_DeleteAggregationCascade(t *testing.T) {
	db := openTestDB(t)
	superBlockID := seedBlockHier(t, db)
	deviceModelID := seedDeviceModel(t, db, 32)
	rackID := seedRack(t, db, "r1")

	bs := store.NewBlockStore(db)
	s := store.NewTierAggregationStore(db)
	block, _ := bs.CreateBlock(&models.Block{SuperBlockID: superBlockID, Name: "row-A"})
	agg, _ := s.SetAggregation(&models.TierAggregation{
		ScopeType:     models.ScopeBlock,
		ScopeID:       block.ID,
		Plane:         models.NetworkPlaneFrontEnd,
		DeviceModelID: deviceModelID,
	})

	s.AllocatePorts(agg.ID, rackID, []string{"leaf-1"}, 0)

	if err := s.DeleteAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneFrontEnd); err != nil {
		t.Fatalf("DeleteAggregation: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM tier_port_connections WHERE tier_aggregation_id = ?`, agg.ID).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 tier_port_connections after cascade delete, got %d", count)
	}

	_, err := s.GetAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneFrontEnd)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTierAggregationStore_DeleteAggregation_NotFound(t *testing.T) {
	db := openTestDB(t)
	s := store.NewTierAggregationStore(db)

	err := s.DeleteAggregation(models.ScopeBlock, 9999, models.NetworkPlaneFrontEnd)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
