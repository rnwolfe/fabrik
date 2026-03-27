package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DeviceModelService is the business logic interface required by DeviceModelHandler.
type DeviceModelService interface {
	CreateDeviceModel(dm *models.DeviceModel) (*models.DeviceModel, error)
	ListDeviceModels(includeArchived bool) ([]*models.DeviceModel, error)
	GetDeviceModel(id int64) (*models.DeviceModel, error)
	UpdateDeviceModel(dm *models.DeviceModel) (*models.DeviceModel, error)
	ArchiveDeviceModel(id int64) error
	DuplicateDeviceModel(id int64) (*models.DeviceModel, error)
}

// DeviceModelHandler handles HTTP requests for DeviceModel resources.
type DeviceModelHandler struct {
	svc DeviceModelService
}

// NewDeviceModelHandler returns a new DeviceModelHandler using svc.
func NewDeviceModelHandler(svc DeviceModelService) *DeviceModelHandler {
	return &DeviceModelHandler{svc: svc}
}

// createDeviceModelRequest is the JSON body for POST /api/catalog/devices.
type createDeviceModelRequest struct {
	Vendor            string                  `json:"vendor"`
	Model             string                  `json:"model"`
	DeviceModelType   models.DeviceModelType  `json:"device_model_type"`
	PortCount         int                     `json:"port_count"`
	HeightU           int                     `json:"height_u"`
	PowerWattsIdle    int                     `json:"power_watts_idle"`
	PowerWattsTypical int                     `json:"power_watts_typical"`
	PowerWattsMax     int                     `json:"power_watts_max"`
	CPUSockets        int                     `json:"cpu_sockets"`
	CoresPerSocket    int                     `json:"cores_per_socket"`
	RAMGB             int                     `json:"ram_gb"`
	StorageTB         float64                 `json:"storage_tb"`
	GPUCount          int                     `json:"gpu_count"`
	Description       string                  `json:"description"`
}

// Create handles POST /api/catalog/devices.
func (h *DeviceModelHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createDeviceModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		slog.Error("decode device model request", "err", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dm := &models.DeviceModel{
		Vendor:            req.Vendor,
		Model:             req.Model,
		DeviceModelType:   req.DeviceModelType,
		PortCount:         req.PortCount,
		HeightU:           req.HeightU,
		PowerWattsIdle:    req.PowerWattsIdle,
		PowerWattsTypical: req.PowerWattsTypical,
		PowerWattsMax:     req.PowerWattsMax,
		CPUSockets:        req.CPUSockets,
		CoresPerSocket:    req.CoresPerSocket,
		RAMGB:             req.RAMGB,
		StorageTB:         req.StorageTB,
		GPUCount:          req.GPUCount,
		Description:       req.Description,
	}

	out, err := h.svc.CreateDeviceModel(dm)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, models.ErrDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		slog.Error("create device model", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, out)
}

// List handles GET /api/catalog/devices.
func (h *DeviceModelHandler) List(w http.ResponseWriter, r *http.Request) {
	includeArchived := r.URL.Query().Get("include_archived") == "true"

	dms, err := h.svc.ListDeviceModels(includeArchived)
	if err != nil {
		slog.Error("list device models", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if dms == nil {
		dms = []*models.DeviceModel{}
	}
	writeJSON(w, http.StatusOK, dms)
}

// Get handles GET /api/catalog/devices/:id.
func (h *DeviceModelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	dm, err := h.svc.GetDeviceModel(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device model not found")
			return
		}
		slog.Error("get device model", "err", err, "deviceModelID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, dm)
}

// updateDeviceModelRequest is the JSON body for PUT /api/catalog/devices/:id.
type updateDeviceModelRequest struct {
	Vendor            string                  `json:"vendor"`
	Model             string                  `json:"model"`
	DeviceModelType   models.DeviceModelType  `json:"device_model_type"`
	PortCount         int                     `json:"port_count"`
	HeightU           int                     `json:"height_u"`
	PowerWattsIdle    int                     `json:"power_watts_idle"`
	PowerWattsTypical int                     `json:"power_watts_typical"`
	PowerWattsMax     int                     `json:"power_watts_max"`
	CPUSockets        int                     `json:"cpu_sockets"`
	CoresPerSocket    int                     `json:"cores_per_socket"`
	RAMGB             int                     `json:"ram_gb"`
	StorageTB         float64                 `json:"storage_tb"`
	GPUCount          int                     `json:"gpu_count"`
	Description       string                  `json:"description"`
}

// Update handles PUT /api/catalog/devices/:id.
func (h *DeviceModelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateDeviceModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		slog.Error("decode device model request", "err", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dm := &models.DeviceModel{
		ID:                id,
		Vendor:            req.Vendor,
		Model:             req.Model,
		DeviceModelType:   req.DeviceModelType,
		PortCount:         req.PortCount,
		HeightU:           req.HeightU,
		PowerWattsIdle:    req.PowerWattsIdle,
		PowerWattsTypical: req.PowerWattsTypical,
		PowerWattsMax:     req.PowerWattsMax,
		CPUSockets:        req.CPUSockets,
		CoresPerSocket:    req.CoresPerSocket,
		RAMGB:             req.RAMGB,
		StorageTB:         req.StorageTB,
		GPUCount:          req.GPUCount,
		Description:       req.Description,
	}

	out, err := h.svc.UpdateDeviceModel(dm)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device model not found")
			return
		}
		if errors.Is(err, models.ErrSeedReadOnly) {
			writeError(w, http.StatusForbidden, "seed device models are read-only")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, models.ErrDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		slog.Error("update device model", "err", err, "deviceModelID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// Delete handles DELETE /api/catalog/devices/:id (archives instead of hard-delete).
func (h *DeviceModelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.ArchiveDeviceModel(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device model not found")
			return
		}
		if errors.Is(err, models.ErrSeedReadOnly) {
			writeError(w, http.StatusForbidden, "seed device models are read-only")
			return
		}
		slog.Error("archive device model", "err", err, "deviceModelID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Duplicate handles POST /api/catalog/devices/:id/duplicate.
func (h *DeviceModelHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	out, err := h.svc.DuplicateDeviceModel(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device model not found")
			return
		}
		slog.Error("duplicate device model", "err", err, "deviceModelID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, out)
}
