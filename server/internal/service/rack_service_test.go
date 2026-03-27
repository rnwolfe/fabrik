package service_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// --- Fake rack type repository ---

type fakeRackTypeRepo struct {
	types  map[int64]*models.RackTemplate
	nextID int64
}

func newFakeRackTypeRepo() *fakeRackTypeRepo {
	return &fakeRackTypeRepo{types: make(map[int64]*models.RackTemplate)}
}

func (r *fakeRackTypeRepo) Create(rt *models.RackTemplate) (*models.RackTemplate, error) {
	r.nextID++
	out := *rt
	out.ID = r.nextID
	r.types[out.ID] = &out
	return &out, nil
}

func (r *fakeRackTypeRepo) List() ([]*models.RackTemplate, error) {
	out := make([]*models.RackTemplate, 0, len(r.types))
	for _, rt := range r.types {
		cp := *rt
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeRackTypeRepo) Get(id int64) (*models.RackTemplate, error) {
	rt, ok := r.types[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *rt
	return &cp, nil
}

func (r *fakeRackTypeRepo) Update(rt *models.RackTemplate) (*models.RackTemplate, error) {
	if _, ok := r.types[rt.ID]; !ok {
		return nil, models.ErrNotFound
	}
	cp := *rt
	r.types[rt.ID] = &cp
	return &cp, nil
}

func (r *fakeRackTypeRepo) Delete(id int64) error {
	if _, ok := r.types[id]; !ok {
		return models.ErrNotFound
	}
	delete(r.types, id)
	return nil
}

func (r *fakeRackTypeRepo) ListRackIDsForType(typeID int64) ([]int64, error) {
	return nil, nil
}

// --- Fake rack repository ---

type fakeRackRepo struct {
	racks        map[int64]*models.Rack
	devices      map[int64]*models.Device
	deviceModels map[int64]*models.DeviceModel
	nextRackID   int64
	nextDeviceID int64
}

func newFakeRackRepo() *fakeRackRepo {
	return &fakeRackRepo{
		racks:        make(map[int64]*models.Rack),
		devices:      make(map[int64]*models.Device),
		deviceModels: make(map[int64]*models.DeviceModel),
	}
}

func (r *fakeRackRepo) addDeviceModel(dm *models.DeviceModel) int64 {
	r.nextDeviceID++
	dm.ID = r.nextDeviceID
	r.deviceModels[dm.ID] = dm
	r.nextDeviceID++ // bump so IDs don't collide with devices
	return dm.ID
}

func (r *fakeRackRepo) Create(rack *models.Rack) (*models.Rack, error) {
	r.nextRackID++
	out := *rack
	out.ID = r.nextRackID
	r.racks[out.ID] = &out
	return &out, nil
}

func (r *fakeRackRepo) List(blockID *int64) ([]*models.Rack, error) {
	out := make([]*models.Rack, 0)
	for _, rack := range r.racks {
		if blockID != nil {
			if rack.BlockID == nil || *rack.BlockID != *blockID {
				continue
			}
		}
		cp := *rack
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeRackRepo) Get(id int64) (*models.Rack, error) {
	rack, ok := r.racks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *rack
	return &cp, nil
}

func (r *fakeRackRepo) Update(rack *models.Rack) (*models.Rack, error) {
	existing, ok := r.racks[rack.ID]
	if !ok {
		return nil, models.ErrNotFound
	}
	existing.Name = rack.Name
	existing.Description = rack.Description
	existing.BlockID = rack.BlockID
	cp := *existing
	return &cp, nil
}

func (r *fakeRackRepo) Delete(id int64) error {
	if _, ok := r.racks[id]; !ok {
		return models.ErrNotFound
	}
	delete(r.racks, id)
	// Cascade delete devices.
	for id2, d := range r.devices {
		if d.RackID == id {
			delete(r.devices, id2)
		}
	}
	return nil
}

func (r *fakeRackRepo) PlaceDevice(d *models.Device) (*models.Device, error) {
	r.nextDeviceID++
	out := *d
	out.ID = r.nextDeviceID
	r.devices[out.ID] = &out
	return &out, nil
}

func (r *fakeRackRepo) GetDevice(id int64) (*models.Device, error) {
	d, ok := r.devices[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (r *fakeRackRepo) MoveDevice(deviceID, rackID int64, position int) (*models.Device, error) {
	d, ok := r.devices[deviceID]
	if !ok {
		return nil, models.ErrNotFound
	}
	d.RackID = rackID
	d.Position = position
	cp := *d
	return &cp, nil
}

func (r *fakeRackRepo) RemoveDevice(deviceID int64, compact bool) error {
	d, ok := r.devices[deviceID]
	if !ok {
		return models.ErrNotFound
	}
	dm := r.deviceModels[d.DeviceModelID]
	heightU := 1
	if dm != nil {
		heightU = dm.HeightU
	}
	removedPos := d.Position
	rackID := d.RackID
	delete(r.devices, deviceID)

	if compact {
		for _, other := range r.devices {
			if other.RackID == rackID && other.Position > removedPos {
				other.Position -= heightU
			}
		}
	}
	return nil
}

func (r *fakeRackRepo) ListDevicesInRack(rackID int64) ([]*models.DeviceSummary, error) {
	var out []*models.DeviceSummary
	for _, d := range r.devices {
		if d.RackID != rackID {
			continue
		}
		dm := r.deviceModels[d.DeviceModelID]
		heightU := 1
		powerWatts := 0
		vendor := ""
		modelName := ""
		if dm != nil {
			heightU = dm.HeightU
			powerWatts = dm.PowerWatts
			vendor = dm.Vendor
			modelName = dm.Model
		}
		cp := *d
		out = append(out, &models.DeviceSummary{
			Device:      cp,
			ModelVendor: vendor,
			ModelName:   modelName,
			HeightU:     heightU,
			PowerWatts:  powerWatts,
		})
	}
	return out, nil
}

func (r *fakeRackRepo) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm, ok := r.deviceModels[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *dm
	return &cp, nil
}

// --- Tests ---

func newRackSvc() (*service.RackService, *fakeRackTypeRepo, *fakeRackRepo) {
	tr := newFakeRackTypeRepo()
	rr := newFakeRackRepo()
	return service.NewRackService(tr, rr), tr, rr
}

func TestRackService_RackTypeCRUD(t *testing.T) {
	svc, _, _ := newRackSvc()

	t.Run("create valid", func(t *testing.T) {
		rt, err := svc.CreateRackType("42U-standard", "desc", 42, 10000)
		if err != nil {
			t.Fatalf("CreateRackType: %v", err)
		}
		if rt.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if rt.Name != "42U-standard" {
			t.Errorf("expected name %q, got %q", "42U-standard", rt.Name)
		}
	})

	t.Run("create empty name", func(t *testing.T) {
		_, err := svc.CreateRackType("", "", 42, 0)
		if !errors.Is(err, models.ErrConstraintViolation) {
			t.Errorf("expected ErrConstraintViolation, got %v", err)
		}
	})

	t.Run("create zero height", func(t *testing.T) {
		_, err := svc.CreateRackType("name", "", 0, 0)
		if !errors.Is(err, models.ErrConstraintViolation) {
			t.Errorf("expected ErrConstraintViolation, got %v", err)
		}
	})

	t.Run("create negative power", func(t *testing.T) {
		_, err := svc.CreateRackType("name", "", 42, -1)
		if !errors.Is(err, models.ErrConstraintViolation) {
			t.Errorf("expected ErrConstraintViolation, got %v", err)
		}
	})

	t.Run("delete with no racks", func(t *testing.T) {
		rt, _ := svc.CreateRackType("delete-me", "", 42, 0)
		if err := svc.DeleteRackType(rt.ID); err != nil {
			t.Fatalf("DeleteRackType: %v", err)
		}
	})
}

func TestRackService_RackCRUD(t *testing.T) {
	svc, _, _ := newRackSvc()

	t.Run("create standalone rack", func(t *testing.T) {
		r, err := svc.CreateRack("rack-01", "desc", nil, nil, 42, 5000)
		if err != nil {
			t.Fatalf("CreateRack: %v", err)
		}
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if r.HeightU != 42 {
			t.Errorf("expected HeightU 42, got %d", r.HeightU)
		}
	})

	t.Run("create with rack type inherits specs", func(t *testing.T) {
		svc2, tr, _ := newRackSvc()
		rt, _ := tr.Create(&models.RackTemplate{Name: "type1", HeightU: 24, PowerCapacityW: 3000})
		r, err := svc2.CreateRack("typed-rack", "", nil, &rt.ID, 0, 0)
		if err != nil {
			t.Fatalf("CreateRack with type: %v", err)
		}
		if r.HeightU != 24 {
			t.Errorf("expected HeightU 24 from rack type, got %d", r.HeightU)
		}
		if r.PowerCapacityW != 3000 {
			t.Errorf("expected PowerCapacityW 3000 from rack type, got %d", r.PowerCapacityW)
		}
	})

	t.Run("create empty name", func(t *testing.T) {
		_, err := svc.CreateRack("", "", nil, nil, 42, 0)
		if !errors.Is(err, models.ErrConstraintViolation) {
			t.Errorf("expected ErrConstraintViolation, got %v", err)
		}
	})

	t.Run("list racks", func(t *testing.T) {
		svc2, _, _ := newRackSvc()
		svc2.CreateRack("r1", "", nil, nil, 42, 0)
		svc2.CreateRack("r2", "", nil, nil, 42, 0)
		racks, err := svc2.ListRacks(nil)
		if err != nil {
			t.Fatalf("ListRacks: %v", err)
		}
		if len(racks) != 2 {
			t.Errorf("expected 2 racks, got %d", len(racks))
		}
	})
}

func TestRackService_PlaceDevice_RUOverflow(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "Cisco", Model: "ASR", HeightU: 2, PowerWatts: 500}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("small-rack", "", nil, nil, 2, 10000) // exactly 2U

	// First placement fills the rack.
	_, err := svc.PlaceDevice(rack.ID, dmID, "dev-1", "", "leaf", 1)
	if err != nil {
		t.Fatalf("first PlaceDevice: %v", err)
	}

	// Second placement should fail: RU overflow (hard reject).
	_, err = svc.PlaceDevice(rack.ID, dmID, "dev-2", "", "leaf", 0)
	if !errors.Is(err, models.ErrRUOverflow) {
		t.Errorf("expected ErrRUOverflow, got %v", err)
	}
}

func TestRackService_PlaceDevice_PositionOverlap(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "Arista", Model: "7050", HeightU: 1, PowerWatts: 200}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("overlap-rack", "", nil, nil, 10, 5000)

	// Place at position 3.
	_, err := svc.PlaceDevice(rack.ID, dmID, "dev-a", "", "leaf", 3)
	if err != nil {
		t.Fatalf("first place: %v", err)
	}

	// Attempt to place at position 3 again — overlap.
	_, err = svc.PlaceDevice(rack.ID, dmID, "dev-b", "", "leaf", 3)
	if !errors.Is(err, models.ErrPositionOverlap) {
		t.Errorf("expected ErrPositionOverlap, got %v", err)
	}
}

func TestRackService_PlaceDevice_PowerSoftWarning(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	// Device uses 900W; rack capacity is 1000W → 90% > 80% threshold.
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "Juniper", Model: "MX", HeightU: 1, PowerWatts: 900}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("power-rack", "", nil, nil, 42, 1000)

	result, err := svc.PlaceDevice(rack.ID, dmID, "dev-power", "", "spine", 1)
	if err != nil {
		t.Fatalf("PlaceDevice: %v", err)
	}
	// Should succeed with a warning.
	if result.Warning == "" {
		t.Error("expected power warning, got empty")
	}
}

func TestRackService_PlaceDevice_MultiRU(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "Cisco", Model: "Nexus9500", HeightU: 7, PowerWatts: 3000}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("multi-ru-rack", "", nil, nil, 42, 20000)

	result, err := svc.PlaceDevice(rack.ID, dmID, "chassis", "", "spine", 1)
	if err != nil {
		t.Fatalf("PlaceDevice multi-RU: %v", err)
	}
	if result.Device.Position != 1 {
		t.Errorf("expected position 1, got %d", result.Device.Position)
	}

	// Placing at position 2 should overlap with the 7U device at 1.
	_, err = svc.PlaceDevice(rack.ID, dmID, "overlap-dev", "", "leaf", 2)
	if !errors.Is(err, models.ErrPositionOverlap) {
		t.Errorf("expected ErrPositionOverlap for multi-RU overlap, got %v", err)
	}

	// Placing at position 8 should succeed.
	_, err = svc.PlaceDevice(rack.ID, dmID, "next-dev", "", "leaf", 8)
	if err != nil {
		t.Fatalf("PlaceDevice at 8 after 7U device: %v", err)
	}
}

func TestRackService_PlaceDevice_PositionBounds(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "X", Model: "Y", HeightU: 1, PowerWatts: 0}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("bounds-rack", "", nil, nil, 5, 0)

	// Position 0 is invalid.
	_, err := svc.PlaceDevice(rack.ID, dmID, "d", "", "other", -1)
	if !errors.Is(err, models.ErrConstraintViolation) {
		t.Errorf("expected ErrConstraintViolation for position -1, got %v", err)
	}

	// Position beyond rack height.
	_, err = svc.PlaceDevice(rack.ID, dmID, "d", "", "other", 6)
	if !errors.Is(err, models.ErrRUOverflow) {
		t.Errorf("expected ErrRUOverflow for position beyond rack height, got %v", err)
	}
}

func TestRackService_PlaceDevice_ExactEndOfRack(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "X", Model: "Y", HeightU: 1, PowerWatts: 0}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("exact-end-rack", "", nil, nil, 42, 0)

	// A 1U device placed at position 42 in a 42U rack should succeed (fills last slot exactly).
	result, err := svc.PlaceDevice(rack.ID, dmID, "last-device", "", "other", 42)
	if err != nil {
		t.Fatalf("PlaceDevice at exact end of rack: %v", err)
	}
	if result.Device.Position != 42 {
		t.Errorf("expected position 42, got %d", result.Device.Position)
	}
}

func TestRackService_MoveDeviceInRack(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "A", Model: "B", HeightU: 1, PowerWatts: 100}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("move-rack", "", nil, nil, 10, 5000)
	result, _ := svc.PlaceDevice(rack.ID, dmID, "dev", "", "leaf", 1)

	moved, err := svc.MoveDeviceInRack(rack.ID, result.Device.ID, 5)
	if err != nil {
		t.Fatalf("MoveDeviceInRack: %v", err)
	}
	if moved.Device.Position != 5 {
		t.Errorf("expected position 5, got %d", moved.Device.Position)
	}
}

func TestRackService_MoveDeviceCrossRack(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "A", Model: "B", HeightU: 1, PowerWatts: 100}
	rr.nextDeviceID = dmID

	src, _ := svc.CreateRack("src-rack", "", nil, nil, 10, 5000)
	dst, _ := svc.CreateRack("dst-rack", "", nil, nil, 10, 5000)

	placed, _ := svc.PlaceDevice(src.ID, dmID, "dev", "", "leaf", 1)

	result, err := svc.MoveDeviceCrossRack(src.ID, placed.Device.ID, dst.ID, 3)
	if err != nil {
		t.Fatalf("MoveDeviceCrossRack: %v", err)
	}
	if result.Device.RackID != dst.ID {
		t.Errorf("expected rack %d, got %d", dst.ID, result.Device.RackID)
	}
	if result.Device.Position != 3 {
		t.Errorf("expected position 3, got %d", result.Device.Position)
	}
}

func TestRackService_MoveDeviceCrossRack_RUOverflow(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "A", Model: "B", HeightU: 3, PowerWatts: 100}
	rr.nextDeviceID = dmID

	src, _ := svc.CreateRack("src", "", nil, nil, 10, 5000)
	// dst only has 2U available (2U total, filled with nothing but size-3 device can't fit)
	dst, _ := svc.CreateRack("dst", "", nil, nil, 2, 5000)

	placed, _ := svc.PlaceDevice(src.ID, dmID, "dev", "", "leaf", 1)

	_, err := svc.MoveDeviceCrossRack(src.ID, placed.Device.ID, dst.ID, 0)
	if !errors.Is(err, models.ErrRUOverflow) {
		t.Errorf("expected ErrRUOverflow, got %v", err)
	}
}

func TestRackService_RemoveDevice(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "A", Model: "B", HeightU: 1, PowerWatts: 100}
	rr.nextDeviceID = dmID

	rack, _ := svc.CreateRack("rem-rack", "", nil, nil, 10, 5000)
	placed, _ := svc.PlaceDevice(rack.ID, dmID, "dev", "", "leaf", 1)

	err := svc.RemoveDevice(rack.ID, placed.Device.ID, false)
	if err != nil {
		t.Fatalf("RemoveDevice: %v", err)
	}

	// Removing from wrong rack should fail.
	placed2, _ := svc.PlaceDevice(rack.ID, dmID, "dev2", "", "leaf", 2)
	err = svc.RemoveDevice(999, placed2.Device.ID, false)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound for wrong rack, got %v", err)
	}
}

func TestRackService_GetRackSummary_PowerWarning(t *testing.T) {
	svc, _, rr := newRackSvc()

	dmID := rr.nextDeviceID + 1
	rr.deviceModels[dmID] = &models.DeviceModel{ID: dmID, Vendor: "A", Model: "B", HeightU: 1, PowerWatts: 850}
	rr.nextDeviceID = dmID

	// Rack with 1000W capacity; device uses 850W = 85% → warning.
	rack, _ := svc.CreateRack("summary-rack", "", nil, nil, 42, 1000)
	svc.PlaceDevice(rack.ID, dmID, "dev", "", "leaf", 1)

	summary, err := svc.GetRackSummary(rack.ID)
	if err != nil {
		t.Fatalf("GetRackSummary: %v", err)
	}
	if summary.UsedU != 1 {
		t.Errorf("expected UsedU 1, got %d", summary.UsedU)
	}
	if summary.AvailableU != 41 {
		t.Errorf("expected AvailableU 41, got %d", summary.AvailableU)
	}
	if summary.UsedWatts != 850 {
		t.Errorf("expected UsedWatts 850, got %d", summary.UsedWatts)
	}
	if summary.Warning == "" {
		t.Error("expected power warning in summary, got empty")
	}
}
