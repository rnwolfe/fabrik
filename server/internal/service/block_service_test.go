package service_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// --- Fake block repository ---

type fakeBlockRepo struct {
	blocks       map[int64]*models.Block
	aggs         map[int64]*models.TierAggregation
	portConns    map[int64]*models.TierPortConnection
	deviceModels map[int64]*models.DeviceModel
	racks        map[int64]*models.Rack
	devices      map[int64]*models.Device

	nextBlockID  int64
	nextAggID    int64
	nextConnID   int64
	nextDeviceID int64
	nextRackID   int64
}

func newFakeBlockRepo() *fakeBlockRepo {
	return &fakeBlockRepo{
		blocks:       make(map[int64]*models.Block),
		aggs:         make(map[int64]*models.TierAggregation),
		portConns:    make(map[int64]*models.TierPortConnection),
		deviceModels: make(map[int64]*models.DeviceModel),
		racks:        make(map[int64]*models.Rack),
		devices:      make(map[int64]*models.Device),
	}
}

func (r *fakeBlockRepo) addDeviceModel(portCount int) *models.DeviceModel {
	r.nextDeviceID++
	dm := &models.DeviceModel{ID: r.nextDeviceID, Vendor: "Generic", Model: "agg", PortCount: portCount, HeightU: 1}
	r.deviceModels[dm.ID] = dm
	return dm
}

func (r *fakeBlockRepo) addRack(blockID *int64) *models.Rack {
	r.nextRackID++
	rack := &models.Rack{ID: r.nextRackID, BlockID: blockID, Name: "rack", HeightU: 42}
	r.racks[rack.ID] = rack
	return rack
}

func (r *fakeBlockRepo) addLeaf(rackID int64, name string) {
	r.nextDeviceID++
	d := &models.Device{ID: r.nextDeviceID, RackID: rackID, Name: name, Role: models.DeviceRoleLeaf}
	r.devices[d.ID] = d
}

// --- BlockRepository implementation ---

func (r *fakeBlockRepo) CreateBlock(b *models.Block) (*models.Block, error) {
	r.nextBlockID++
	out := *b
	out.ID = r.nextBlockID
	r.blocks[out.ID] = &out
	return &out, nil
}

func (r *fakeBlockRepo) GetBlock(id int64) (*models.Block, error) {
	b, ok := r.blocks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *fakeBlockRepo) ListBlocks(superBlockID int64) ([]*models.Block, error) {
	var out []*models.Block
	for _, b := range r.blocks {
		if b.SuperBlockID == superBlockID {
			cp := *b
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) GetDefaultBlock(superBlockID int64) (*models.Block, error) {
	for _, b := range r.blocks {
		if b.SuperBlockID == superBlockID && b.Name == "default" {
			cp := *b
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeBlockRepo) SetAggregation(agg *models.TierAggregation) (*models.TierAggregation, error) {
	for id, a := range r.aggs {
		if a.ScopeType == agg.ScopeType && a.ScopeID == agg.ScopeID && a.Plane == agg.Plane {
			a.DeviceModelID = agg.DeviceModelID
			a.SpineCount = agg.SpineCount
			r.aggs[id] = a
			cp := *a
			return &cp, nil
		}
	}
	r.nextAggID++
	out := *agg
	out.ID = r.nextAggID
	r.aggs[out.ID] = &out
	return &out, nil
}

func (r *fakeBlockRepo) GetAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) (*models.TierAggregation, error) {
	for _, a := range r.aggs {
		if a.ScopeType == scopeType && a.ScopeID == scopeID && a.Plane == plane {
			cp := *a
			return &cp, nil
		}
	}
	return nil, models.ErrNotFound
}

func (r *fakeBlockRepo) ListAggregations(scopeType models.AggregationScope, scopeID int64) ([]*models.TierAggregation, error) {
	var out []*models.TierAggregation
	for _, a := range r.aggs {
		if a.ScopeType == scopeType && a.ScopeID == scopeID {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) DeleteAggregation(scopeType models.AggregationScope, scopeID int64, plane models.NetworkPlane) error {
	for id, a := range r.aggs {
		if a.ScopeType == scopeType && a.ScopeID == scopeID && a.Plane == plane {
			for cid, pc := range r.portConns {
				if pc.TierAggregationID == id {
					delete(r.portConns, cid)
				}
			}
			delete(r.aggs, id)
			return nil
		}
	}
	return models.ErrNotFound
}

func (r *fakeBlockRepo) AllocatePorts(aggID, childID int64, childNames []string, startPortIndex int) ([]*models.TierPortConnection, error) {
	var out []*models.TierPortConnection
	for i, name := range childNames {
		r.nextConnID++
		pc := &models.TierPortConnection{
			ID:                r.nextConnID,
			TierAggregationID: aggID,
			ChildID:           childID,
			AggPortIndex:      startPortIndex + i,
			ChildDeviceName:   name,
		}
		r.portConns[pc.ID] = pc
		out = append(out, pc)
	}
	return out, nil
}

func (r *fakeBlockRepo) DeallocatePorts(aggID, childID int64) error {
	for id, pc := range r.portConns {
		if pc.TierAggregationID == aggID && pc.ChildID == childID {
			delete(r.portConns, id)
		}
	}
	return nil
}

func (r *fakeBlockRepo) DeallocatePortsByChild(childID int64) error {
	for id, pc := range r.portConns {
		if pc.ChildID == childID {
			delete(r.portConns, id)
		}
	}
	return nil
}

func (r *fakeBlockRepo) CountAllocatedPorts(aggID int64) (int, error) {
	count := 0
	for _, pc := range r.portConns {
		if pc.TierAggregationID == aggID {
			count++
		}
	}
	return count, nil
}

func (r *fakeBlockRepo) ListPortConnections(aggID int64) ([]*models.TierPortConnection, error) {
	var out []*models.TierPortConnection
	for _, pc := range r.portConns {
		if pc.TierAggregationID == aggID {
			cp := *pc
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) ListPortConnectionsByChild(aggID, childID int64) ([]*models.TierPortConnection, error) {
	var out []*models.TierPortConnection
	for _, pc := range r.portConns {
		if pc.TierAggregationID == aggID && pc.ChildID == childID {
			cp := *pc
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm, ok := r.deviceModels[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *dm
	return &cp, nil
}

func (r *fakeBlockRepo) ListDevicesInRack(rackID int64) ([]*models.Device, error) {
	var out []*models.Device
	for _, d := range r.devices {
		if d.RackID == rackID {
			cp := *d
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) UpdateRackBlock(rackID int64, blockID *int64) error {
	rack, ok := r.racks[rackID]
	if !ok {
		return models.ErrNotFound
	}
	rack.BlockID = blockID
	return nil
}

func (r *fakeBlockRepo) GetRack(id int64) (*models.Rack, error) {
	rack, ok := r.racks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *rack
	return &cp, nil
}

func (r *fakeBlockRepo) CreateRack(rack *models.Rack) (*models.Rack, error) {
	r.nextRackID++
	out := *rack
	out.ID = r.nextRackID
	r.racks[out.ID] = &out
	return &out, nil
}

func (r *fakeBlockRepo) PlaceDevice(d *models.Device) (*models.Device, error) {
	r.nextDeviceID++
	out := *d
	out.ID = r.nextDeviceID
	r.devices[out.ID] = &out
	return &out, nil
}

// --- helper to create a TierAggregation at block scope (used in test setup) ---

func blockAgg(blockID, deviceModelID int64, plane models.NetworkPlane) *models.TierAggregation {
	return &models.TierAggregation{
		ScopeType:     models.ScopeBlock,
		ScopeID:       blockID,
		Plane:         plane,
		DeviceModelID: deviceModelID,
	}
}

// --- BlockService tests ---

func TestBlockService_CreateBlock(t *testing.T) {
	tests := []struct {
		name        string
		blockName   string
		wantErr     bool
		wantErrType error
	}{
		{"valid block", "row-A", false, nil},
		{"empty name", "", true, models.ErrConstraintViolation},
		{"whitespace name", "   ", true, models.ErrConstraintViolation},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeBlockRepo()
			svc := service.NewBlockService(repo)

			result, err := svc.CreateBlock(1, tc.blockName, "", nil, nil, 0)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
					t.Errorf("expected error %v, got %v", tc.wantErrType, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Block.ID == 0 {
				t.Error("expected non-zero block ID")
			}
			if result.Block.SuperBlockID != 1 {
				t.Errorf("expected super_block_id 1, got %d", result.Block.SuperBlockID)
			}
			if len(result.Racks) != 0 {
				t.Errorf("expected 0 racks without leaf model, got %d", len(result.Racks))
			}
		})
	}
}

func TestBlockService_CreateBlockWithLeafModel(t *testing.T) {
	repo := newFakeBlockRepo()
	svc := service.NewBlockService(repo)

	leafModel := repo.addDeviceModel(54)
	leafModel.HeightU = 1
	repo.deviceModels[leafModel.ID] = leafModel
	leafID := leafModel.ID

	t.Run("auto-creates 2 racks with 4 leaf devices", func(t *testing.T) {
		result, err := svc.CreateBlock(1, "block-A", "", &leafID, nil, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 2 {
			t.Fatalf("expected 2 racks, got %d", len(result.Racks))
		}

		// Each rack should have 2 leaf devices.
		totalLeaves := 0
		for _, rack := range result.Racks {
			devices, _ := repo.ListDevicesInRack(rack.ID)
			for _, d := range devices {
				if d.Role == models.DeviceRoleLeaf {
					totalLeaves++
				}
			}
		}
		if totalLeaves != 4 {
			t.Errorf("expected 4 leaf devices total, got %d", totalLeaves)
		}
	})

	t.Run("with spine model distributes spines across racks", func(t *testing.T) {
		spineModel := repo.addDeviceModel(36)
		spineModel.HeightU = 1
		repo.deviceModels[spineModel.ID] = spineModel
		spineID := spineModel.ID

		result, err := svc.CreateBlock(1, "block-B", "", &leafID, &spineID, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 4 spines distributed across 2 racks = 2 per rack.
		for _, rack := range result.Racks {
			devices, _ := repo.ListDevicesInRack(rack.ID)
			spines := 0
			for _, d := range devices {
				if d.Role == models.DeviceRoleSpine {
					spines++
				}
			}
			if spines != 2 {
				t.Errorf("rack %d: expected 2 spines, got %d", rack.ID, spines)
			}
		}
	})
}

func TestBlockService_AssignAggregation(t *testing.T) {
	repo := newFakeBlockRepo()
	svc := service.NewBlockService(repo)

	// Create a block.
	block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "row-A"})

	// Add a device model with 32 ports.
	dm32 := repo.addDeviceModel(32)

	t.Run("assign to block", func(t *testing.T) {
		summary, err := svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm32.ID, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.TotalPorts != 32 {
			t.Errorf("expected total_ports 32, got %d", summary.TotalPorts)
		}
		if summary.AllocatedPorts != 0 {
			t.Errorf("expected allocated_ports 0, got %d", summary.AllocatedPorts)
		}
		if summary.AvailablePorts != 32 {
			t.Errorf("expected available_ports 32, got %d", summary.AvailablePorts)
		}
		if summary.SpineCount != 2 {
			t.Errorf("expected spine_count 2, got %d", summary.SpineCount)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		_, err := svc.AssignAggregation(9999, models.NetworkPlaneFrontEnd, dm32.ID, 0)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("downsize rejected when over-allocated", func(t *testing.T) {
		agg, _ := repo.GetAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneFrontEnd)
		repo.AllocatePorts(agg.ID, 99, []string{"leaf-1", "leaf-2", "leaf-3"}, 0)

		dm2 := repo.addDeviceModel(2)
		_, err := svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm2.ID, 0)
		if !errors.Is(err, models.ErrAggModelDownsize) {
			t.Errorf("expected ErrAggModelDownsize, got %v", err)
		}
	})
}

func TestBlockService_AddRackToBlock(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(repo *fakeBlockRepo) (rackID int64, blockID *int64, superBlockID int64)
		wantConns   int
		wantWarning bool
		wantErr     bool
		wantErrType error
	}{
		{
			name: "no leaf devices with agg — auto-places leaves and connects",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "b1"})
				rack := repo.addRack(nil)
				dm := repo.addDeviceModel(32)
				repo.SetAggregation(blockAgg(block.ID, dm.ID, models.NetworkPlaneFrontEnd))
				return rack.ID, &block.ID, 0
			},
			wantConns:   2,
			wantWarning: false,
		},
		{
			name: "no agg assigned — success with warning",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 2, Name: "b2"})
				rack := repo.addRack(nil)
				repo.addLeaf(rack.ID, "leaf-1")
				return rack.ID, &block.ID, 0
			},
			wantConns:   0,
			wantWarning: true,
		},
		{
			name: "one leaf, one agg plane — 1 connection",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 3, Name: "b3"})
				rack := repo.addRack(nil)
				repo.addLeaf(rack.ID, "leaf-1")
				dm := repo.addDeviceModel(32)
				repo.SetAggregation(blockAgg(block.ID, dm.ID, models.NetworkPlaneFrontEnd))
				return rack.ID, &block.ID, 0
			},
			wantConns:   1,
			wantWarning: false,
		},
		{
			name: "two leaves, two planes — 4 connections",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 4, Name: "b4"})
				rack := repo.addRack(nil)
				repo.addLeaf(rack.ID, "leaf-1")
				repo.addLeaf(rack.ID, "leaf-2")
				dm := repo.addDeviceModel(32)
				repo.SetAggregation(blockAgg(block.ID, dm.ID, models.NetworkPlaneFrontEnd))
				dm2 := repo.addDeviceModel(32)
				repo.SetAggregation(blockAgg(block.ID, dm2.ID, models.NetworkPlaneManagement))
				return rack.ID, &block.ID, 0
			},
			wantConns:   4,
			wantWarning: false,
		},
		{
			name: "agg ports full — error",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 5, Name: "b5"})
				rack := repo.addRack(nil)
				repo.addLeaf(rack.ID, "leaf-1")
				dm := repo.addDeviceModel(1)
				agg, _ := repo.SetAggregation(blockAgg(block.ID, dm.ID, models.NetworkPlaneFrontEnd))
				repo.AllocatePorts(agg.ID, 888, []string{"existing-leaf"}, 0)
				return rack.ID, &block.ID, 0
			},
			wantErr:     true,
			wantErrType: models.ErrAggPortsFull,
		},
		{
			name: "nil blockID with superBlockID — auto-creates default block",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				rack := repo.addRack(nil)
				return rack.ID, nil, 10
			},
			wantConns:   0,
			wantWarning: false,
		},
		{
			name: "nil blockID and no superBlockID — error",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				rack := repo.addRack(nil)
				return rack.ID, nil, 0
			},
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeBlockRepo()
			svc := service.NewBlockService(repo)

			rackID, blockID, superBlockID := tc.setup(repo)

			result, err := svc.AddRackToBlock(rackID, blockID, superBlockID)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
					t.Errorf("expected error %v, got %v", tc.wantErrType, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Connections) != tc.wantConns {
				t.Errorf("expected %d connections, got %d", tc.wantConns, len(result.Connections))
			}
			if tc.wantWarning && result.Warning == "" {
				t.Error("expected a warning, got empty string")
			}
			if !tc.wantWarning && result.Warning != "" {
				t.Errorf("expected no warning, got %q", result.Warning)
			}
		})
	}
}

func TestBlockService_CapacityEnforcement(t *testing.T) {
	const portsPerAgg = 4

	tests := []struct {
		name    string
		numRacks int
		wantErr bool
	}{
		{"N-1 racks (capacity not full)", portsPerAgg - 1, false},
		{"N racks (exactly full)", portsPerAgg, false},
		{"N+1 racks (over capacity)", portsPerAgg + 1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeBlockRepo()
			svc := service.NewBlockService(repo)

			block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "cap-test"})
			dm := repo.addDeviceModel(portsPerAgg)
			repo.SetAggregation(blockAgg(block.ID, dm.ID, models.NetworkPlaneFrontEnd))

			var lastErr error
			for i := 0; i < tc.numRacks; i++ {
				rack := repo.addRack(nil)
				repo.addLeaf(rack.ID, "leaf-1")
				_, lastErr = svc.AddRackToBlock(rack.ID, &block.ID, 0)
				if lastErr != nil {
					break
				}
			}

			if tc.wantErr {
				if lastErr == nil {
					t.Fatal("expected error at capacity limit, got nil")
				}
				if !errors.Is(lastErr, models.ErrAggPortsFull) {
					t.Errorf("expected ErrAggPortsFull, got %v", lastErr)
				}
			} else {
				if lastErr != nil {
					t.Fatalf("unexpected error: %v", lastErr)
				}
			}
		})
	}
}

func TestBlockService_RemoveRackFromBlock(t *testing.T) {
	t.Run("removes connections and clears block", func(t *testing.T) {
		repo := newFakeBlockRepo()
		svc := service.NewBlockService(repo)

		block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "row-A"})
		rack := repo.addRack(&block.ID)
		repo.addLeaf(rack.ID, "leaf-1")
		dm := repo.addDeviceModel(32)
		repo.SetAggregation(blockAgg(block.ID, dm.ID, models.NetworkPlaneFrontEnd))

		_, err := svc.AddRackToBlock(rack.ID, &block.ID, 0)
		if err != nil {
			t.Fatalf("add rack: %v", err)
		}

		agg, _ := repo.GetAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneFrontEnd)
		count, _ := repo.CountAllocatedPorts(agg.ID)
		if count != 1 {
			t.Fatalf("expected 1 connection before remove, got %d", count)
		}

		if err := svc.RemoveRackFromBlock(rack.ID); err != nil {
			t.Fatalf("remove rack: %v", err)
		}

		count, _ = repo.CountAllocatedPorts(agg.ID)
		if count != 0 {
			t.Errorf("expected 0 connections after remove, got %d", count)
		}
	})

	t.Run("rack not in any block — error", func(t *testing.T) {
		repo := newFakeBlockRepo()
		svc := service.NewBlockService(repo)

		rack := repo.addRack(nil)
		err := svc.RemoveRackFromBlock(rack.ID)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestBlockService_DefaultBlock(t *testing.T) {
	t.Run("auto-creates default block on first placement", func(t *testing.T) {
		repo := newFakeBlockRepo()
		svc := service.NewBlockService(repo)

		rack := repo.addRack(nil)
		const superBlockID int64 = 42

		result, err := svc.AddRackToBlock(rack.ID, nil, superBlockID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updatedRack, _ := repo.GetRack(rack.ID)
		if updatedRack.BlockID == nil {
			t.Error("expected rack to have a block_id after placement")
		}

		def, err := repo.GetDefaultBlock(superBlockID)
		if err != nil {
			t.Fatalf("GetDefaultBlock: %v", err)
		}
		if def == nil {
			t.Fatal("expected default block to exist")
		}
		if def.Name != "default" {
			t.Errorf("expected block name 'default', got %q", def.Name)
		}

		_ = result
	})

	t.Run("second placement reuses existing default block", func(t *testing.T) {
		repo := newFakeBlockRepo()
		svc := service.NewBlockService(repo)

		rack1 := repo.addRack(nil)
		rack2 := repo.addRack(nil)
		const superBlockID int64 = 43

		_, _ = svc.AddRackToBlock(rack1.ID, nil, superBlockID)
		_, _ = svc.AddRackToBlock(rack2.ID, nil, superBlockID)

		blocks, _ := repo.ListBlocks(superBlockID)
		defaultCount := 0
		for _, b := range blocks {
			if b.Name == "default" {
				defaultCount++
			}
		}
		if defaultCount != 1 {
			t.Errorf("expected exactly 1 default block, got %d", defaultCount)
		}
	})
}

func TestBlockService_AggregationDownsizeRejected(t *testing.T) {
	repo := newFakeBlockRepo()
	svc := service.NewBlockService(repo)

	block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "row-A"})
	dm32 := repo.addDeviceModel(32)

	_, err := svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm32.ID, 0)
	if err != nil {
		t.Fatalf("initial assign: %v", err)
	}

	agg, _ := repo.GetAggregation(models.ScopeBlock, block.ID, models.NetworkPlaneFrontEnd)
	repo.AllocatePorts(agg.ID, 100, make([]string, 10), 0)

	dm8 := repo.addDeviceModel(8)
	_, err = svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm8.ID, 0)
	if !errors.Is(err, models.ErrAggModelDownsize) {
		t.Errorf("expected ErrAggModelDownsize, got %v", err)
	}

	dm12 := repo.addDeviceModel(12)
	_, err = svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm12.ID, 0)
	if err != nil {
		t.Errorf("expected success resizing to 12, got %v", err)
	}
}
