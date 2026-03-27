package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DeviceModelRepository is the store interface required by DeviceModelService.
type DeviceModelRepository interface {
	Create(dm *models.DeviceModel) (*models.DeviceModel, error)
	List(includeArchived bool) ([]*models.DeviceModel, error)
	Get(id int64) (*models.DeviceModel, error)
	Update(dm *models.DeviceModel) (*models.DeviceModel, error)
	Archive(id int64) error
	Duplicate(sourceID int64, newVendor, newModel string) (*models.DeviceModel, error)
}

// DeviceModelService implements business logic for DeviceModel resources.
type DeviceModelService struct {
	repo DeviceModelRepository
}

// NewDeviceModelService returns a new DeviceModelService backed by repo.
func NewDeviceModelService(repo DeviceModelRepository) *DeviceModelService {
	return &DeviceModelService{repo: repo}
}

// CreateDeviceModel validates and creates a new DeviceModel.
func (s *DeviceModelService) CreateDeviceModel(dm *models.DeviceModel) (*models.DeviceModel, error) {
	if err := validateDeviceModel(dm); err != nil {
		return nil, err
	}

	out, err := s.repo.Create(dm)
	if err != nil {
		if isErrDuplicate(err) {
			return nil, fmt.Errorf("%w: vendor+model combination already exists", models.ErrDuplicate)
		}
		return nil, fmt.Errorf("create device model: %w", err)
	}
	slog.Info("device model created", "deviceModelID", out.ID, "vendor", out.Vendor, "model", out.Model)
	return out, nil
}

// ListDeviceModels returns device models. When includeArchived is true, archived
// models are included.
func (s *DeviceModelService) ListDeviceModels(includeArchived bool) ([]*models.DeviceModel, error) {
	out, err := s.repo.List(includeArchived)
	if err != nil {
		return nil, fmt.Errorf("list device models: %w", err)
	}
	return out, nil
}

// GetDeviceModel returns the device model with the given id.
func (s *DeviceModelService) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm, err := s.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get device model %d: %w", id, err)
	}
	return dm, nil
}

// UpdateDeviceModel validates and updates an existing DeviceModel. Returns
// models.ErrSeedReadOnly for seed models.
func (s *DeviceModelService) UpdateDeviceModel(dm *models.DeviceModel) (*models.DeviceModel, error) {
	existing, err := s.repo.Get(dm.ID)
	if err != nil {
		return nil, fmt.Errorf("update device model %d: %w", dm.ID, err)
	}
	if existing.IsSeed {
		return nil, models.ErrSeedReadOnly
	}

	if err := validateDeviceModel(dm); err != nil {
		return nil, err
	}

	out, err := s.repo.Update(dm)
	if err != nil {
		if isErrDuplicate(err) {
			return nil, fmt.Errorf("%w: vendor+model combination already exists", models.ErrDuplicate)
		}
		return nil, fmt.Errorf("update device model %d: %w", dm.ID, err)
	}
	slog.Info("device model updated", "deviceModelID", out.ID)
	return out, nil
}

// ArchiveDeviceModel soft-deletes a device model by setting its archived_at.
// Returns models.ErrSeedReadOnly for seed models.
func (s *DeviceModelService) ArchiveDeviceModel(id int64) error {
	existing, err := s.repo.Get(id)
	if err != nil {
		return fmt.Errorf("archive device model %d: %w", id, err)
	}
	if existing.IsSeed {
		return models.ErrSeedReadOnly
	}

	if err := s.repo.Archive(id); err != nil {
		return fmt.Errorf("archive device model %d: %w", id, err)
	}
	slog.Info("device model archived", "deviceModelID", id)
	return nil
}

// DuplicateDeviceModel clones the device model identified by id. The copy is
// never a seed and receives a unique vendor+model name.
func (s *DeviceModelService) DuplicateDeviceModel(id int64) (*models.DeviceModel, error) {
	src, err := s.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("duplicate device model %d: %w", id, err)
	}

	out, err := s.repo.Duplicate(id, src.Vendor, src.Model+" (copy)")
	if err != nil {
		return nil, fmt.Errorf("duplicate device model %d: %w", id, err)
	}
	slog.Info("device model duplicated", "sourceID", id, "newID", out.ID)
	return out, nil
}

// validateDeviceModel enforces field-level constraints for create and update.
func validateDeviceModel(dm *models.DeviceModel) error {
	dm.Vendor = strings.TrimSpace(dm.Vendor)
	dm.Model = strings.TrimSpace(dm.Model)

	if dm.Vendor == "" {
		return fmt.Errorf("%w: vendor is required", models.ErrConstraintViolation)
	}
	if dm.Model == "" {
		return fmt.Errorf("%w: model name is required", models.ErrConstraintViolation)
	}
	if dm.PortCount < 0 {
		return fmt.Errorf("%w: port_count must not be negative", models.ErrConstraintViolation)
	}
	if dm.HeightU < 1 || dm.HeightU > 50 {
		return fmt.Errorf("%w: height_u must be between 1 and 50", models.ErrConstraintViolation)
	}
	if dm.PowerWattsIdle < 0 {
		return fmt.Errorf("%w: power_watts_idle must not be negative", models.ErrConstraintViolation)
	}
	if dm.PowerWattsTypical < 0 {
		return fmt.Errorf("%w: power_watts_typical must not be negative", models.ErrConstraintViolation)
	}
	if dm.PowerWattsMax < 0 {
		return fmt.Errorf("%w: power_watts_max must not be negative", models.ErrConstraintViolation)
	}
	if dm.CPUSockets < 0 {
		return fmt.Errorf("%w: cpu_sockets must not be negative", models.ErrConstraintViolation)
	}
	if dm.CoresPerSocket < 0 {
		return fmt.Errorf("%w: cores_per_socket must not be negative", models.ErrConstraintViolation)
	}
	if dm.RAMGB < 0 {
		return fmt.Errorf("%w: ram_gb must not be negative", models.ErrConstraintViolation)
	}
	if dm.StorageTB < 0 {
		return fmt.Errorf("%w: storage_tb must not be negative", models.ErrConstraintViolation)
	}
	if dm.GPUCount < 0 {
		return fmt.Errorf("%w: gpu_count must not be negative", models.ErrConstraintViolation)
	}
	// Default device_model_type to "network" if not set.
	if dm.DeviceModelType == "" {
		dm.DeviceModelType = models.DeviceModelTypeNetwork
	}
	validTypes := map[models.DeviceModelType]bool{
		models.DeviceModelTypeNetwork: true,
		models.DeviceModelTypeServer:  true,
		models.DeviceModelTypeStorage: true,
		models.DeviceModelTypeOther:   true,
	}
	if !validTypes[dm.DeviceModelType] {
		return fmt.Errorf("%w: device_model_type must be one of: network, server, storage, other", models.ErrConstraintViolation)
	}
	return nil
}

// isErrDuplicate reports whether err wraps models.ErrDuplicate.
func isErrDuplicate(err error) bool {
	return errors.Is(err, models.ErrDuplicate)
}
