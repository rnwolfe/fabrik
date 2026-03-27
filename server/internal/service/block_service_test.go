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
	aggs         map[int64]*models.BlockAggregation // keyed by (blockID*100 + plane-hash) — we use aggID key
	portConns    map[int64]*models.PortConnection
	deviceModels map[int64]*models.DeviceModel
	racks        map[int64]*models.Rack
	devices      map[int64]*models.Device

	nextBlockID   int64
	nextAggID     int64
	nextConnID    int64
	nextDeviceID  int64
	nextRackID    int64
}

func newFakeBlockRepo() *fakeBlockRepo {
	return &fakeBlockRepo{
		blocks:       make(map[int64]*models.Block),
		aggs:         make(map[int64]*models.BlockAggregation),
		portConns:    make(map[int64]*models.PortConnection),
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

func (r *fakeBlockRepo) SetAggregation(agg *models.BlockAggregation) (*models.BlockAggregation, error) {
	// Check if exists — update or insert.
	for id, a := range r.aggs {
		if a.BlockID == agg.BlockID && a.Plane == agg.Plane {
			a.DeviceModelID = agg.DeviceModelID
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

func (r *fakeBlockRepo) GetAggregation(blockID int64, plane models.NetworkPlane) (*models.BlockAggregation, error) {
	for _, a := range r.aggs {
		if a.BlockID == blockID && a.Plane == plane {
			cp := *a
			return &cp, nil
		}
	}
	return nil, models.ErrNotFound
}

func (r *fakeBlockRepo) ListAggregations(blockID int64) ([]*models.BlockAggregation, error) {
	var out []*models.BlockAggregation
	for _, a := range r.aggs {
		if a.BlockID == blockID {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) DeleteAggregation(blockID int64, plane models.NetworkPlane) error {
	for id, a := range r.aggs {
		if a.BlockID == blockID && a.Plane == plane {
			// Remove all port connections for this agg.
			for cid, pc := range r.portConns {
				if pc.BlockAggregationID == id {
					delete(r.portConns, cid)
				}
			}
			delete(r.aggs, id)
			return nil
		}
	}
	return models.ErrNotFound
}

func (r *fakeBlockRepo) AllocatePorts(aggID, rackID int64, leafNames []string, startPortIndex int) ([]*models.PortConnection, error) {
	var out []*models.PortConnection
	for i, name := range leafNames {
		r.nextConnID++
		pc := &models.PortConnection{
			ID:                 r.nextConnID,
			BlockAggregationID: aggID,
			RackID:             rackID,
			AggPortIndex:       startPortIndex + i,
			LeafDeviceName:     name,
		}
		r.portConns[pc.ID] = pc
		out = append(out, pc)
	}
	return out, nil
}

func (r *fakeBlockRepo) DeallocatePorts(aggID, rackID int64) error {
	for id, pc := range r.portConns {
		if pc.BlockAggregationID == aggID && pc.RackID == rackID {
			delete(r.portConns, id)
		}
	}
	return nil
}

func (r *fakeBlockRepo) DeallocatePortsByRack(rackID int64) error {
	for id, pc := range r.portConns {
		if pc.RackID == rackID {
			delete(r.portConns, id)
		}
	}
	return nil
}

func (r *fakeBlockRepo) CountAllocatedPorts(aggID int64) (int, error) {
	count := 0
	for _, pc := range r.portConns {
		if pc.BlockAggregationID == aggID {
			count++
		}
	}
	return count, nil
}

func (r *fakeBlockRepo) ListPortConnections(aggID int64) ([]*models.PortConnection, error) {
	var out []*models.PortConnection
	for _, pc := range r.portConns {
		if pc.BlockAggregationID == aggID {
			cp := *pc
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeBlockRepo) ListPortConnectionsByRack(aggID, rackID int64) ([]*models.PortConnection, error) {
	var out []*models.PortConnection
	for _, pc := range r.portConns {
		if pc.BlockAggregationID == aggID && pc.RackID == rackID {
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

			b, err := svc.CreateBlock(1, tc.blockName, "")
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
			if b.ID == 0 {
				t.Error("expected non-zero block ID")
			}
			if b.SuperBlockID != 1 {
				t.Errorf("expected super_block_id 1, got %d", b.SuperBlockID)
			}
		})
	}
}

func TestBlockService_AssignAggregation(t *testing.T) {
	repo := newFakeBlockRepo()
	svc := service.NewBlockService(repo)

	// Create a block.
	block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "row-A"})

	// Add a device model with 32 ports.
	dm32 := repo.addDeviceModel(32)

	t.Run("assign to block", func(t *testing.T) {
		summary, err := svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm32.ID)
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
	})

	t.Run("block not found", func(t *testing.T) {
		_, err := svc.AssignAggregation(9999, models.NetworkPlaneFrontEnd, dm32.ID)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("downsize rejected when over-allocated", func(t *testing.T) {
		// Pre-allocate some ports manually.
		agg, _ := repo.GetAggregation(block.ID, models.NetworkPlaneFrontEnd)
		repo.AllocatePorts(agg.ID, 99, []string{"leaf-1", "leaf-2", "leaf-3"}, 0)

		// Try to assign a smaller model with only 2 ports.
		dm2 := repo.addDeviceModel(2)
		_, err := svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm2.ID)
		if !errors.Is(err, models.ErrAggModelDownsize) {
			t.Errorf("expected ErrAggModelDownsize, got %v", err)
		}
	})
}

func TestBlockService_AddRackToBlock(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(repo *fakeBlockRepo) (rackID int64, blockID *int64, superBlockID int64)
		wantConns     int
		wantWarning   bool
		wantErr       bool
		wantErrType   error
	}{
		{
			name: "no leaf devices — success with no connections",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 1, Name: "b1"})
				rack := repo.addRack(nil) // no devices
				dm := repo.addDeviceModel(32)
				repo.SetAggregation(&models.BlockAggregation{BlockID: block.ID, Plane: models.NetworkPlaneFrontEnd, DeviceModelID: dm.ID})
				return rack.ID, &block.ID, 0
			},
			wantConns:   0,
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
				repo.SetAggregation(&models.BlockAggregation{BlockID: block.ID, Plane: models.NetworkPlaneFrontEnd, DeviceModelID: dm.ID})
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
				repo.SetAggregation(&models.BlockAggregation{BlockID: block.ID, Plane: models.NetworkPlaneFrontEnd, DeviceModelID: dm.ID})
				dm2 := repo.addDeviceModel(32)
				repo.SetAggregation(&models.BlockAggregation{BlockID: block.ID, Plane: models.NetworkPlaneManagement, DeviceModelID: dm2.ID})
				return rack.ID, &block.ID, 0
			},
			wantConns:   4, // 2 leaves × 2 planes
			wantWarning: false,
		},
		{
			name: "agg ports full — error",
			setup: func(repo *fakeBlockRepo) (int64, *int64, int64) {
				block, _ := repo.CreateBlock(&models.Block{SuperBlockID: 5, Name: "b5"})
				rack := repo.addRack(nil)
				repo.addLeaf(rack.ID, "leaf-1")
				dm := repo.addDeviceModel(1) // only 1 port
				agg, _ := repo.SetAggregation(&models.BlockAggregation{BlockID: block.ID, Plane: models.NetworkPlaneFrontEnd, DeviceModelID: dm.ID})
				// Fill the only port.
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
	// Table-driven capacity test: N-1, N, N+1 racks.
	const portsPerAgg = 4

	tests := []struct {
		name       string
		numRacks   int // racks with 1 leaf each
		wantErr    bool
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
			repo.SetAggregation(&models.BlockAggregation{
				BlockID: block.ID, Plane: models.NetworkPlaneFrontEnd, DeviceModelID: dm.ID,
			})

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
		repo.SetAggregation(&models.BlockAggregation{BlockID: block.ID, Plane: models.NetworkPlaneFrontEnd, DeviceModelID: dm.ID})

		// Add rack to block first.
		_, err := svc.AddRackToBlock(rack.ID, &block.ID, 0)
		if err != nil {
			t.Fatalf("add rack: %v", err)
		}

		// Check connections exist.
		agg, _ := repo.GetAggregation(block.ID, models.NetworkPlaneFrontEnd)
		count, _ := repo.CountAllocatedPorts(agg.ID)
		if count != 1 {
			t.Fatalf("expected 1 connection before remove, got %d", count)
		}

		// Remove rack from block.
		if err := svc.RemoveRackFromBlock(rack.ID); err != nil {
			t.Fatalf("remove rack: %v", err)
		}

		// Connections should be gone.
		count, _ = repo.CountAllocatedPorts(agg.ID)
		if count != 0 {
			t.Errorf("expected 0 connections after remove, got %d", count)
		}
	})

	t.Run("rack not in any block — error", func(t *testing.T) {
		repo := newFakeBlockRepo()
		svc := service.NewBlockService(repo)

		rack := repo.addRack(nil) // no block
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

		// Rack should now have a block assigned.
		updatedRack, _ := repo.GetRack(rack.ID)
		if updatedRack.BlockID == nil {
			t.Error("expected rack to have a block_id after placement")
		}

		// Default block should exist.
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

	_, err := svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm32.ID)
	if err != nil {
		t.Fatalf("initial assign: %v", err)
	}

	// Allocate 10 ports.
	agg, _ := repo.GetAggregation(block.ID, models.NetworkPlaneFrontEnd)
	repo.AllocatePorts(agg.ID, 100, make([]string, 10), 0)

	// Try to downsize to 8 ports — should fail.
	dm8 := repo.addDeviceModel(8)
	_, err = svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm8.ID)
	if !errors.Is(err, models.ErrAggModelDownsize) {
		t.Errorf("expected ErrAggModelDownsize, got %v", err)
	}

	// Resize to 12 ports — should succeed (10 allocated < 12).
	dm12 := repo.addDeviceModel(12)
	_, err = svc.AssignAggregation(block.ID, models.NetworkPlaneFrontEnd, dm12.ID)
	if err != nil {
		t.Errorf("expected success resizing to 12, got %v", err)
	}
}
