package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

const powerWarningThreshold = 0.80

// RackTypeRepository is the store interface for rack type operations.
type RackTypeRepository interface {
	Create(rt *models.RackTemplate) (*models.RackTemplate, error)
	List() ([]*models.RackTemplate, error)
	Get(id int64) (*models.RackTemplate, error)
	Update(rt *models.RackTemplate) (*models.RackTemplate, error)
	Delete(id int64) error
	ListRackIDsForType(typeID int64) ([]int64, error)
}

// RackRepository is the store interface for rack and device placement operations.
type RackRepository interface {
	Create(r *models.Rack) (*models.Rack, error)
	List(blockID *int64) ([]*models.Rack, error)
	Get(id int64) (*models.Rack, error)
	Update(r *models.Rack) (*models.Rack, error)
	Delete(id int64) error
	PlaceDevice(d *models.Device) (*models.Device, error)
	GetDevice(id int64) (*models.Device, error)
	MoveDevice(deviceID, rackID int64, position int) (*models.Device, error)
	RemoveDevice(deviceID int64, compact bool) error
	ListDevicesInRack(rackID int64) ([]*models.DeviceSummary, error)
	GetDeviceModel(id int64) (*models.DeviceModel, error)
}

// RackService implements business logic for rack types, racks, and device placement.
type RackService struct {
	typeRepo RackTypeRepository
	rackRepo RackRepository
}

// NewRackService returns a new RackService.
func NewRackService(typeRepo RackTypeRepository, rackRepo RackRepository) *RackService {
	return &RackService{typeRepo: typeRepo, rackRepo: rackRepo}
}

// --- Rack Type operations ---

// CreateRackType validates and creates a new RackTemplate.
func (s *RackService) CreateRackType(name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: rack type name is required", models.ErrConstraintViolation)
	}
	if heightU <= 0 {
		return nil, fmt.Errorf("%w: height_u must be positive", models.ErrConstraintViolation)
	}
	if powerCapacityW < 0 {
		return nil, fmt.Errorf("%w: power_capacity_w must be non-negative", models.ErrConstraintViolation)
	}
	rt, err := s.typeRepo.Create(&models.RackTemplate{
		Name:           name,
		HeightU:        heightU,
		PowerCapacityW: powerCapacityW,
		Description:    description,
	})
	if err != nil {
		return nil, fmt.Errorf("create rack type: %w", err)
	}
	slog.Info("rack type created", "rackTypeID", rt.ID, "name", rt.Name)
	return rt, nil
}

// ListRackTypes returns all rack types.
func (s *RackService) ListRackTypes() ([]*models.RackTemplate, error) {
	rts, err := s.typeRepo.List()
	if err != nil {
		return nil, fmt.Errorf("list rack types: %w", err)
	}
	return rts, nil
}

// GetRackType returns the rack type with the given id.
func (s *RackService) GetRackType(id int64) (*models.RackTemplate, error) {
	rt, err := s.typeRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get rack type %d: %w", id, err)
	}
	return rt, nil
}

// UpdateRackType updates a rack type's fields.
func (s *RackService) UpdateRackType(id int64, name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: rack type name is required", models.ErrConstraintViolation)
	}
	if heightU <= 0 {
		return nil, fmt.Errorf("%w: height_u must be positive", models.ErrConstraintViolation)
	}
	if powerCapacityW < 0 {
		return nil, fmt.Errorf("%w: power_capacity_w must be non-negative", models.ErrConstraintViolation)
	}
	rt, err := s.typeRepo.Update(&models.RackTemplate{
		ID:             id,
		Name:           name,
		HeightU:        heightU,
		PowerCapacityW: powerCapacityW,
		Description:    description,
	})
	if err != nil {
		return nil, fmt.Errorf("update rack type %d: %w", id, err)
	}
	return rt, nil
}

// DeleteRackType removes a rack type. Returns ErrConflict if racks reference it,
// with the list of affected rack IDs embedded in the error message.
func (s *RackService) DeleteRackType(id int64) error {
	ids, err := s.typeRepo.ListRackIDsForType(id)
	if err != nil {
		return fmt.Errorf("list rack ids for type: %w", err)
	}
	if len(ids) > 0 {
		return fmt.Errorf("%w: rack type is referenced by rack IDs %v", models.ErrConflict, ids)
	}
	if err := s.typeRepo.Delete(id); err != nil {
		return fmt.Errorf("delete rack type %d: %w", id, err)
	}
	slog.Info("rack type deleted", "rackTypeID", id)
	return nil
}

// --- Rack operations ---

// CreateRack creates a new rack, optionally seeding specs from a rack type.
func (s *RackService) CreateRack(name, description string, blockID, rackTypeID *int64, heightU, powerCapacityW int) (*models.Rack, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: rack name is required", models.ErrConstraintViolation)
	}

	// If a rack type is provided, copy its specs (rack can diverge afterward).
	if rackTypeID != nil {
		rt, err := s.typeRepo.Get(*rackTypeID)
		if err != nil {
			return nil, fmt.Errorf("get rack type for rack creation: %w", err)
		}
		if heightU == 0 {
			heightU = rt.HeightU
		}
		if powerCapacityW == 0 {
			powerCapacityW = rt.PowerCapacityW
		}
	}

	if heightU <= 0 {
		heightU = 42
	}

	r, err := s.rackRepo.Create(&models.Rack{
		BlockID:        blockID,
		RackTypeID:     rackTypeID,
		Name:           name,
		HeightU:        heightU,
		PowerCapacityW: powerCapacityW,
		Description:    description,
	})
	if err != nil {
		return nil, fmt.Errorf("create rack: %w", err)
	}
	slog.Info("rack created", "rackID", r.ID, "name", r.Name)
	return r, nil
}

// ListRacks returns racks, optionally filtered by block.
func (s *RackService) ListRacks(blockID *int64) ([]*models.Rack, error) {
	racks, err := s.rackRepo.List(blockID)
	if err != nil {
		return nil, fmt.Errorf("list racks: %w", err)
	}
	return racks, nil
}

// GetRackSummary returns a rack with computed usage metrics and device list.
func (s *RackService) GetRackSummary(id int64) (*models.RackSummary, error) {
	rack, err := s.rackRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get rack %d: %w", id, err)
	}

	devices, err := s.rackRepo.ListDevicesInRack(id)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack %d: %w", id, err)
	}

	usedU := 0
	usedWatts := 0
	for _, d := range devices {
		usedU += d.HeightU
		usedWatts += d.PowerWatts
	}

	summary := &models.RackSummary{
		Rack:       *rack,
		UsedU:      usedU,
		AvailableU: rack.HeightU - usedU,
		UsedWatts:  usedWatts,
		Devices:    devices,
	}

	// Power warning at >80% utilization.
	if rack.PowerCapacityW > 0 {
		ratio := float64(usedWatts) / float64(rack.PowerCapacityW)
		if ratio > powerWarningThreshold {
			summary.Warning = fmt.Sprintf("power utilization at %.0f%% (%.0fW / %dW)", ratio*100, float64(usedWatts), rack.PowerCapacityW)
		}
	}

	if devices == nil {
		summary.Devices = []*models.DeviceSummary{}
	}

	return summary, nil
}

// UpdateRack updates a rack's name, description and block assignment.
func (s *RackService) UpdateRack(id int64, name, description string, blockID *int64) (*models.Rack, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: rack name is required", models.ErrConstraintViolation)
	}
	r, err := s.rackRepo.Update(&models.Rack{
		ID:          id,
		Name:        name,
		Description: description,
		BlockID:     blockID,
	})
	if err != nil {
		return nil, fmt.Errorf("update rack %d: %w", id, err)
	}
	return r, nil
}

// DeleteRack removes a rack (devices cascade automatically via FK).
func (s *RackService) DeleteRack(id int64) error {
	if err := s.rackRepo.Delete(id); err != nil {
		return fmt.Errorf("delete rack %d: %w", id, err)
	}
	slog.Info("rack deleted", "rackID", id)
	return nil
}

// --- Device placement ---

// PlaceDevice places a device in a rack at a specific position.
// If position is 0, it auto-suggests the lowest available slot.
// Returns ErrRUOverflow (hard) if the device doesn't fit.
// Returns ErrPositionOverlap (hard) if the position is already occupied.
// Returns a warning in the result if power capacity would be exceeded (soft).
func (s *RackService) PlaceDevice(rackID, deviceModelID int64, name, description, role string, position int) (*models.PlaceDeviceResult, error) {
	rack, err := s.rackRepo.Get(rackID)
	if err != nil {
		return nil, fmt.Errorf("get rack %d: %w", rackID, err)
	}

	dm, err := s.rackRepo.GetDeviceModel(deviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", deviceModelID, err)
	}
	if dm.HeightU <= 0 {
		return nil, fmt.Errorf("%w: device model has invalid height", models.ErrConstraintViolation)
	}

	existing, err := s.rackRepo.ListDevicesInRack(rackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack: %w", err)
	}

	// Compute used RU and power.
	usedU := 0
	usedWatts := 0
	for _, d := range existing {
		usedU += d.HeightU
		usedWatts += d.PowerWatts
	}

	// For management switches: RU/power overflow is a soft warning, not a hard block.
	isManagementRole := models.DeviceRole(role) == models.DeviceRoleManagementToR ||
		models.DeviceRole(role) == models.DeviceRoleManagementAgg

	// Hard reject: device doesn't fit at all (soft for management).
	if dm.HeightU > rack.HeightU-usedU && !isManagementRole {
		return nil, fmt.Errorf("%w: device needs %dU but only %dU available", models.ErrRUOverflow, dm.HeightU, rack.HeightU-usedU)
	}

	// Validate the device role before placement.
	if role == "" {
		role = string(models.DeviceRoleOther)
	}
	if err := ValidateDeviceRole(role); err != nil {
		return nil, err
	}

	var ruWarning string

	// Auto-suggest position if not specified.
	if position == 0 {
		position = suggestPosition(rack.HeightU, existing, dm.HeightU)
		if position == 0 {
			if isManagementRole {
				// Management devices: no free slot is a soft warning; place at position 1.
				position = 1
				ruWarning = fmt.Sprintf("management switch placement: no contiguous %dU slot available in rack", dm.HeightU)
			} else {
				return nil, fmt.Errorf("%w: no contiguous %dU slot available", models.ErrRUOverflow, dm.HeightU)
			}
		}
	}

	// Validate position bounds.
	if position < 1 {
		return nil, fmt.Errorf("%w: position must be >= 1", models.ErrConstraintViolation)
	}
	if position+dm.HeightU-1 > rack.HeightU {
		if isManagementRole {
			ruWarning = fmt.Sprintf("management switch placement: device at position %d with height %dU exceeds rack height %dU",
				position, dm.HeightU, rack.HeightU)
			// Place at last valid position.
			position = rack.HeightU - dm.HeightU + 1
			if position < 1 {
				position = 1
			}
		} else {
			return nil, fmt.Errorf("%w: device at position %d with height %dU exceeds rack height %dU", models.ErrRUOverflow, position, dm.HeightU, rack.HeightU)
		}
	}

	// Validate no overlap (management devices also cannot overlap).
	if err := checkOverlap(existing, position, dm.HeightU, 0); err != nil {
		if isManagementRole {
			if ruWarning == "" {
				ruWarning = fmt.Sprintf("management switch placement: %s", err.Error())
			}
			// Find first open slot, ignoring capacity.
			pos := suggestPosition(rack.HeightU, existing, dm.HeightU)
			if pos == 0 {
				pos = 1
			}
			position = pos
		} else {
			return nil, err
		}
	}

	d, err := s.rackRepo.PlaceDevice(&models.Device{
		RackID:        rackID,
		DeviceModelID: deviceModelID,
		Name:          name,
		Role:          models.DeviceRole(role),
		Position:      position,
		Description:   description,
	})
	if err != nil {
		return nil, fmt.Errorf("place device: %w", err)
	}

	result := &models.PlaceDeviceResult{Device: d}

	// Soft warning: power capacity exceeded or approaching threshold.
	if rack.PowerCapacityW > 0 {
		newTotal := usedWatts + dm.PowerWatts
		if newTotal > rack.PowerCapacityW {
			pwrMsg := fmt.Sprintf("power capacity exceeded: %dW used + %dW new = %dW > %dW capacity",
				usedWatts, dm.PowerWatts, newTotal, rack.PowerCapacityW)
			if ruWarning != "" {
				result.Warning = ruWarning + "; " + pwrMsg
			} else {
				result.Warning = pwrMsg
			}
		} else if float64(newTotal)/float64(rack.PowerCapacityW) > powerWarningThreshold {
			pwrMsg := fmt.Sprintf("power utilization at %.0f%% (%dW / %dW)",
				float64(newTotal)/float64(rack.PowerCapacityW)*100, newTotal, rack.PowerCapacityW)
			if ruWarning != "" {
				result.Warning = ruWarning + "; " + pwrMsg
			} else {
				result.Warning = pwrMsg
			}
		} else if ruWarning != "" {
			result.Warning = ruWarning
		}
	} else if ruWarning != "" {
		result.Warning = ruWarning
	}

	slog.Info("device placed", "rackID", rackID, "deviceID", d.ID, "position", position, "role", role)
	return result, nil
}

// MoveDeviceInRack moves a device to a new position within the same rack.
func (s *RackService) MoveDeviceInRack(rackID, deviceID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	rack, err := s.rackRepo.Get(rackID)
	if err != nil {
		return nil, fmt.Errorf("get rack %d: %w", rackID, err)
	}

	device, err := s.rackRepo.GetDevice(deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device %d: %w", deviceID, err)
	}
	if device.RackID != rackID {
		return nil, fmt.Errorf("%w: device %d is not in rack %d", models.ErrNotFound, deviceID, rackID)
	}

	dm, err := s.rackRepo.GetDeviceModel(device.DeviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model: %w", err)
	}

	if newPosition < 1 {
		return nil, fmt.Errorf("%w: position must be >= 1", models.ErrConstraintViolation)
	}
	if newPosition+dm.HeightU-1 > rack.HeightU {
		return nil, fmt.Errorf("%w: device at position %d with height %dU exceeds rack height %dU", models.ErrRUOverflow, newPosition, dm.HeightU, rack.HeightU)
	}

	existing, err := s.rackRepo.ListDevicesInRack(rackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in rack: %w", err)
	}

	// Check overlap excluding the device being moved.
	if err := checkOverlap(existing, newPosition, dm.HeightU, deviceID); err != nil {
		return nil, err
	}

	d, err := s.rackRepo.MoveDevice(deviceID, rackID, newPosition)
	if err != nil {
		return nil, fmt.Errorf("move device: %w", err)
	}

	return &models.PlaceDeviceResult{Device: d}, nil
}

// MoveDeviceCrossRack moves a device from one rack to another rack at a given position.
func (s *RackService) MoveDeviceCrossRack(srcRackID, deviceID, dstRackID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	srcDevice, err := s.rackRepo.GetDevice(deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device %d: %w", deviceID, err)
	}
	if srcDevice.RackID != srcRackID {
		return nil, fmt.Errorf("%w: device %d is not in rack %d", models.ErrNotFound, deviceID, srcRackID)
	}

	dstRack, err := s.rackRepo.Get(dstRackID)
	if err != nil {
		return nil, fmt.Errorf("get destination rack %d: %w", dstRackID, err)
	}

	dm, err := s.rackRepo.GetDeviceModel(srcDevice.DeviceModelID)
	if err != nil {
		return nil, fmt.Errorf("get device model: %w", err)
	}

	dstDevices, err := s.rackRepo.ListDevicesInRack(dstRackID)
	if err != nil {
		return nil, fmt.Errorf("list devices in destination rack: %w", err)
	}

	usedU := 0
	usedWatts := 0
	for _, d := range dstDevices {
		usedU += d.HeightU
		usedWatts += d.PowerWatts
	}

	// Hard reject: doesn't fit in destination rack.
	if dm.HeightU > dstRack.HeightU-usedU {
		return nil, fmt.Errorf("%w: device needs %dU but destination rack only has %dU available", models.ErrRUOverflow, dm.HeightU, dstRack.HeightU-usedU)
	}

	if newPosition == 0 {
		newPosition = suggestPosition(dstRack.HeightU, dstDevices, dm.HeightU)
		if newPosition == 0 {
			return nil, fmt.Errorf("%w: no contiguous %dU slot available in destination rack", models.ErrRUOverflow, dm.HeightU)
		}
	}

	if newPosition < 1 {
		return nil, fmt.Errorf("%w: position must be >= 1", models.ErrConstraintViolation)
	}
	if newPosition+dm.HeightU-1 > dstRack.HeightU {
		return nil, fmt.Errorf("%w: device at position %d with height %dU exceeds destination rack height %dU", models.ErrRUOverflow, newPosition, dm.HeightU, dstRack.HeightU)
	}

	if err := checkOverlap(dstDevices, newPosition, dm.HeightU, 0); err != nil {
		return nil, err
	}

	d, err := s.rackRepo.MoveDevice(deviceID, dstRackID, newPosition)
	if err != nil {
		return nil, fmt.Errorf("move device cross-rack: %w", err)
	}

	result := &models.PlaceDeviceResult{Device: d}

	// Soft warning: power capacity exceeded in destination rack.
	if dstRack.PowerCapacityW > 0 && usedWatts+dm.PowerWatts > dstRack.PowerCapacityW {
		result.Warning = fmt.Sprintf("power capacity exceeded in destination rack: %dW used + %dW new = %dW > %dW capacity",
			usedWatts, dm.PowerWatts, usedWatts+dm.PowerWatts, dstRack.PowerCapacityW)
	}

	slog.Info("device moved cross-rack", "srcRackID", srcRackID, "dstRackID", dstRackID, "deviceID", deviceID)
	return result, nil
}

// RemoveDevice deletes a device from a rack.
func (s *RackService) RemoveDevice(rackID, deviceID int64, compact bool) error {
	device, err := s.rackRepo.GetDevice(deviceID)
	if err != nil {
		return fmt.Errorf("get device %d: %w", deviceID, err)
	}
	if device.RackID != rackID {
		return fmt.Errorf("%w: device %d is not in rack %d", models.ErrNotFound, deviceID, rackID)
	}
	if err := s.rackRepo.RemoveDevice(deviceID, compact); err != nil {
		return fmt.Errorf("remove device %d: %w", deviceID, err)
	}
	slog.Info("device removed", "rackID", rackID, "deviceID", deviceID, "compact", compact)
	return nil
}

// --- helpers ---

// checkOverlap returns ErrPositionOverlap if the given position+height range overlaps
// any existing device (excluding skipDeviceID, use 0 to skip nothing).
func checkOverlap(existing []*models.DeviceSummary, position, heightU int, skipDeviceID int64) error {
	for _, d := range existing {
		if d.Device.ID == skipDeviceID {
			continue
		}
		// Overlap if ranges intersect: [position, position+heightU) ∩ [d.Position, d.Position+d.HeightU)
		if position < d.Device.Position+d.HeightU && position+heightU > d.Device.Position {
			return fmt.Errorf("%w: position %d overlaps device %d at position %d (height %dU)",
				models.ErrPositionOverlap, position, d.Device.ID, d.Device.Position, d.HeightU)
		}
	}
	return nil
}

// suggestPosition returns the lowest RU position that can accommodate a device of the given height.
// Returns 0 if no slot is available.
func suggestPosition(rackHeightU int, existing []*models.DeviceSummary, neededU int) int {
	// Build a set of occupied positions.
	occupied := make(map[int]bool)
	for _, d := range existing {
		for u := d.Device.Position; u < d.Device.Position+d.HeightU; u++ {
			occupied[u] = true
		}
	}

	for start := 1; start <= rackHeightU-neededU+1; start++ {
		fits := true
		for u := start; u < start+neededU; u++ {
			if occupied[u] {
				fits = false
				break
			}
		}
		if fits {
			return start
		}
	}
	return 0
}
