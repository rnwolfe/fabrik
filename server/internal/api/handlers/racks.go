package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// RackService is the business logic interface required by RackHandler.
type RackService interface {
	// Rack type operations
	CreateRackType(name, description string, heightU, powerCapacityW, powerOversubPctWarn, powerOversubPctMax int) (*models.RackTemplate, error)
	ListRackTypes() ([]*models.RackTemplate, error)
	GetRackType(id int64) (*models.RackTemplate, error)
	UpdateRackType(id int64, name, description string, heightU, powerCapacityW, powerOversubPctWarn, powerOversubPctMax int) (*models.RackTemplate, error)
	DeleteRackType(id int64) error
	// Rack operations
	CreateRack(name, description string, blockID, rackTypeID *int64, heightU, powerCapacityW int) (*models.Rack, error)
	ListRacks(blockID *int64) ([]*models.Rack, error)
	GetRackSummary(id int64) (*models.RackSummary, error)
	UpdateRack(id int64, name, description string, blockID *int64) (*models.Rack, error)
	DeleteRack(id int64) error
	// Device placement
	PlaceDevice(rackID, deviceModelID int64, name, description, role string, position int) (*models.PlaceDeviceResult, error)
	MoveDeviceInRack(rackID, deviceID int64, newPosition int) (*models.PlaceDeviceResult, error)
	MoveDeviceCrossRack(srcRackID, deviceID, dstRackID int64, newPosition int) (*models.PlaceDeviceResult, error)
	RemoveDevice(rackID, deviceID int64, compact bool) error
}

// RackHandler handles HTTP requests for rack type and rack resources.
type RackHandler struct {
	svc RackService
}

// NewRackHandler returns a new RackHandler using svc.
func NewRackHandler(svc RackService) *RackHandler {
	return &RackHandler{svc: svc}
}

// --- Rack Type handlers ---

type createRackTypeRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	HeightU            int    `json:"height_u"`
	PowerCapacityW     int    `json:"power_capacity_w"`
	PowerOversubPctWarn int   `json:"power_oversub_pct_warn"`
	PowerOversubPctMax  int   `json:"power_oversub_pct_max"`
}

// CreateRackType handles POST /api/rack-types.
func (h *RackHandler) CreateRackType(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createRackTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rt, err := h.svc.CreateRackType(req.Name, req.Description, req.HeightU, req.PowerCapacityW, req.PowerOversubPctWarn, req.PowerOversubPctMax)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("create rack type", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, rt)
}

// ListRackTypes handles GET /api/rack-types.
func (h *RackHandler) ListRackTypes(w http.ResponseWriter, r *http.Request) {
	rts, err := h.svc.ListRackTypes()
	if err != nil {
		slog.Error("list rack types", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if rts == nil {
		rts = []*models.RackTemplate{}
	}
	writeJSON(w, http.StatusOK, rts)
}

// GetRackType handles GET /api/rack-types/:id.
func (h *RackHandler) GetRackType(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	rt, err := h.svc.GetRackType(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rack type not found")
			return
		}
		slog.Error("get rack type", "err", err, "rackTypeID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, rt)
}

// UpdateRackType handles PUT /api/rack-types/:id.
func (h *RackHandler) UpdateRackType(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createRackTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rt, err := h.svc.UpdateRackType(id, req.Name, req.Description, req.HeightU, req.PowerCapacityW, req.PowerOversubPctWarn, req.PowerOversubPctMax)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rack type not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("update rack type", "err", err, "rackTypeID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, rt)
}

// DeleteRackType handles DELETE /api/rack-types/:id.
func (h *RackHandler) DeleteRackType(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteRackType(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rack type not found")
			return
		}
		if errors.Is(err, models.ErrConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		slog.Error("delete rack type", "err", err, "rackTypeID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Rack handlers ---

type createRackRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	BlockID        *int64  `json:"block_id"`
	RackTypeID     *int64  `json:"rack_type_id"`
	HeightU        int     `json:"height_u"`
	PowerCapacityW int     `json:"power_capacity_w"`
}

// CreateRack handles POST /api/racks.
func (h *RackHandler) CreateRack(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createRackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rack, err := h.svc.CreateRack(req.Name, req.Description, req.BlockID, req.RackTypeID, req.HeightU, req.PowerCapacityW)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("create rack", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, rack)
}

// ListRacks handles GET /api/racks with optional ?block_id= filter.
func (h *RackHandler) ListRacks(w http.ResponseWriter, r *http.Request) {
	var blockID *int64
	if v := r.URL.Query().Get("block_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid block_id")
			return
		}
		blockID = &id
	}

	racks, err := h.svc.ListRacks(blockID)
	if err != nil {
		slog.Error("list racks", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if racks == nil {
		racks = []*models.Rack{}
	}
	writeJSON(w, http.StatusOK, racks)
}

// GetRack handles GET /api/racks/:id (returns summary).
func (h *RackHandler) GetRack(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	summary, err := h.svc.GetRackSummary(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rack not found")
			return
		}
		slog.Error("get rack", "err", err, "rackID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

type updateRackRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BlockID     *int64 `json:"block_id"`
}

// UpdateRack handles PUT /api/racks/:id.
func (h *RackHandler) UpdateRack(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateRackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rack, err := h.svc.UpdateRack(id, req.Name, req.Description, req.BlockID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rack not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("update rack", "err", err, "rackID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, rack)
}

// DeleteRack handles DELETE /api/racks/:id.
func (h *RackHandler) DeleteRack(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteRack(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "rack not found")
			return
		}
		slog.Error("delete rack", "err", err, "rackID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Device placement handlers ---

type placeDeviceRequest struct {
	DeviceModelID int64  `json:"device_model_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Role          string `json:"role"`
	Position      int    `json:"position"`
}

// PlaceDevice handles POST /api/racks/:id/devices.
func (h *RackHandler) PlaceDevice(w http.ResponseWriter, r *http.Request) {
	rackID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req placeDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.PlaceDevice(rackID, req.DeviceModelID, req.Name, req.Description, req.Role, req.Position)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, models.ErrRUOverflow) || errors.Is(err, models.ErrPositionOverlap) || errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, models.ErrConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		slog.Error("place device", "err", err, "rackID", rackID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

type moveDeviceRequest struct {
	Position int `json:"position"`
}

// MoveDeviceInRack handles PUT /api/racks/:rack_id/devices/:device_id.
func (h *RackHandler) MoveDeviceInRack(w http.ResponseWriter, r *http.Request) {
	rackID, ok := parseID(w, r, "rack_id")
	if !ok {
		return
	}
	deviceID, ok := parseID(w, r, "device_id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req moveDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.MoveDeviceInRack(rackID, deviceID, req.Position)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, models.ErrRUOverflow) || errors.Is(err, models.ErrPositionOverlap) || errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("move device in rack", "err", err, "rackID", rackID, "deviceID", deviceID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type moveCrossRackRequest struct {
	DestRackID int64 `json:"dest_rack_id"`
	Position   int   `json:"position"`
}

// MoveDeviceCrossRack handles PUT /api/racks/:rack_id/devices/:device_id/move.
func (h *RackHandler) MoveDeviceCrossRack(w http.ResponseWriter, r *http.Request) {
	rackID, ok := parseID(w, r, "rack_id")
	if !ok {
		return
	}
	deviceID, ok := parseID(w, r, "device_id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req moveCrossRackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.MoveDeviceCrossRack(rackID, deviceID, req.DestRackID, req.Position)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, models.ErrRUOverflow) || errors.Is(err, models.ErrPositionOverlap) || errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("move device cross-rack", "err", err, "rackID", rackID, "deviceID", deviceID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// RemoveDevice handles DELETE /api/racks/:rack_id/devices/:device_id.
func (h *RackHandler) RemoveDevice(w http.ResponseWriter, r *http.Request) {
	rackID, ok := parseID(w, r, "rack_id")
	if !ok {
		return
	}
	deviceID, ok := parseID(w, r, "device_id")
	if !ok {
		return
	}
	compact := r.URL.Query().Get("compact") == "true"
	if err := h.svc.RemoveDevice(rackID, deviceID, compact); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("remove device", "err", err, "rackID", rackID, "deviceID", deviceID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
