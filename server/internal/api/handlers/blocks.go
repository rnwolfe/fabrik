package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// BlockService is the business logic interface required by BlockHandler.
type BlockService interface {
	// Block operations
	CreateBlock(superBlockID int64, name, description string, leafModelID, spineModelID *int64, spineCount int) (*models.CreateBlockResult, error)
	GetBlock(id int64) (*models.Block, error)
	ListBlocks(superBlockID int64) ([]*models.Block, error)

	// Aggregation operations (block-level)
	AssignAggregation(blockID int64, plane models.NetworkPlane, deviceModelID int64, spineCount int, hostLinkSpeedGbps int) (*models.TierAggregationSummary, error)
	GetAggregationSummary(blockID int64, plane models.NetworkPlane) (*models.TierAggregationSummary, error)
	ListAggregationSummaries(blockID int64) ([]*models.TierAggregationSummary, error)
	DeleteAggregation(blockID int64, plane models.NetworkPlane) error

	// Aggregation operations (super-block-level)
	AssignSuperBlockAggregation(superBlockID int64, plane models.NetworkPlane, deviceModelID int64, spineCount int, hostLinkSpeedGbps int) (*models.TierAggregationSummary, error)
	GetSuperBlockAggregationSummary(superBlockID int64, plane models.NetworkPlane) (*models.TierAggregationSummary, error)

	// Rack placement
	AddRackToBlock(rackID int64, blockID *int64, superBlockID int64) (*models.AddRackToBlockResult, error)
	RemoveRackFromBlock(rackID int64) error
	ListPortConnections(blockID int64, plane models.NetworkPlane) ([]*models.TierPortConnection, error)
	PlaceSpineDevices(blockID, spineModelID int64, count int) error
	DeleteBlock(id int64) error
}

// BlockHandler handles HTTP requests for block and block aggregation resources.
type BlockHandler struct {
	svc BlockService
}

// NewBlockHandler returns a new BlockHandler using svc.
func NewBlockHandler(svc BlockService) *BlockHandler {
	return &BlockHandler{svc: svc}
}

// --- Block handlers ---

type createBlockRequest struct {
	SuperBlockID int64  `json:"super_block_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	LeafModelID  *int64 `json:"leaf_model_id"`
	SpineModelID *int64 `json:"spine_model_id"`
	SpineCount   int    `json:"spine_count"`
}

// CreateBlock handles POST /api/blocks.
func (h *BlockHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.CreateBlock(req.SuperBlockID, req.Name, req.Description, req.LeafModelID, req.SpineModelID, req.SpineCount)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("create block", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// GetBlock handles GET /api/blocks/{id}.
func (h *BlockHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	b, err := h.svc.GetBlock(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "block not found")
			return
		}
		slog.Error("get block", "err", err, "blockID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// ListBlocks handles GET /api/blocks?super_block_id=N.
func (h *BlockHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	var superBlockID int64
	if v := r.URL.Query().Get("super_block_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid super_block_id")
			return
		}
		superBlockID = id
	}
	if superBlockID <= 0 {
		writeError(w, http.StatusBadRequest, "super_block_id is required")
		return
	}

	blocks, err := h.svc.ListBlocks(superBlockID)
	if err != nil {
		slog.Error("list blocks", "err", err, "superBlockID", superBlockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if blocks == nil {
		blocks = []*models.Block{}
	}
	writeJSON(w, http.StatusOK, blocks)
}

// --- Aggregation handlers ---

type assignAggregationRequest struct {
	Plane             string `json:"plane"`
	DeviceModelID     int64  `json:"device_model_id"`
	SpineCount        int    `json:"spine_count"`
	HostLinkSpeedGbps int    `json:"host_link_speed_gbps"`
}

// AssignAggregation handles PUT /api/blocks/{id}/aggregations/{plane}.
func (h *BlockHandler) AssignAggregation(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	planeStr := r.PathValue("plane")
	plane, ok := parsePlane(w, planeStr)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req assignAggregationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	summary, err := h.svc.AssignAggregation(blockID, plane, req.DeviceModelID, req.SpineCount, req.HostLinkSpeedGbps)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, models.ErrAggModelDownsize) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("assign aggregation", "err", err, "blockID", blockID, "plane", plane)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// GetAggregation handles GET /api/blocks/{id}/aggregations/{plane}.
func (h *BlockHandler) GetAggregation(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	planeStr := r.PathValue("plane")
	plane, ok := parsePlane(w, planeStr)
	if !ok {
		return
	}

	summary, err := h.svc.GetAggregationSummary(blockID, plane)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "aggregation not found")
			return
		}
		slog.Error("get aggregation", "err", err, "blockID", blockID, "plane", plane)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// ListAggregations handles GET /api/blocks/{id}/aggregations.
func (h *BlockHandler) ListAggregations(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	summaries, err := h.svc.ListAggregationSummaries(blockID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "block not found")
			return
		}
		slog.Error("list aggregations", "err", err, "blockID", blockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if summaries == nil {
		summaries = []*models.TierAggregationSummary{}
	}
	writeJSON(w, http.StatusOK, summaries)
}

// DeleteAggregation handles DELETE /api/blocks/{id}/aggregations/{plane}.
func (h *BlockHandler) DeleteAggregation(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	planeStr := r.PathValue("plane")
	plane, ok := parsePlane(w, planeStr)
	if !ok {
		return
	}

	if err := h.svc.DeleteAggregation(blockID, plane); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "aggregation not found")
			return
		}
		slog.Error("delete aggregation", "err", err, "blockID", blockID, "plane", plane)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Rack placement handlers ---

type addRackToBlockRequest struct {
	RackID       int64  `json:"rack_id"`
	BlockID      *int64 `json:"block_id"`
	SuperBlockID int64  `json:"super_block_id"`
}

// AddRackToBlock handles POST /api/blocks/add-rack.
func (h *BlockHandler) AddRackToBlock(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req addRackToBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RackID <= 0 {
		writeError(w, http.StatusBadRequest, "rack_id is required")
		return
	}

	result, err := h.svc.AddRackToBlock(req.RackID, req.BlockID, req.SuperBlockID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, models.ErrAggPortsFull) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("add rack to block", "err", err, "rackID", req.RackID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// RemoveRackFromBlock handles DELETE /api/blocks/racks/{rack_id}.
func (h *BlockHandler) RemoveRackFromBlock(w http.ResponseWriter, r *http.Request) {
	rackID, ok := parseID(w, r, "rack_id")
	if !ok {
		return
	}

	if err := h.svc.RemoveRackFromBlock(rackID); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("remove rack from block", "err", err, "rackID", rackID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListPortConnections handles GET /api/blocks/{id}/aggregations/{plane}/connections.
func (h *BlockHandler) ListPortConnections(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	planeStr := r.PathValue("plane")
	plane, ok := parsePlane(w, planeStr)
	if !ok {
		return
	}

	conns, err := h.svc.ListPortConnections(blockID, plane)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "aggregation not found")
			return
		}
		slog.Error("list port connections", "err", err, "blockID", blockID, "plane", plane)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if conns == nil {
		conns = []*models.TierPortConnection{}
	}
	writeJSON(w, http.StatusOK, conns)
}

// AssignSuperBlockAggregation handles PUT /api/super-blocks/{id}/aggregations/{plane}.
func (h *BlockHandler) AssignSuperBlockAggregation(w http.ResponseWriter, r *http.Request) {
	superBlockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	planeStr := r.PathValue("plane")
	plane, ok := parsePlane(w, planeStr)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req assignAggregationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	summary, err := h.svc.AssignSuperBlockAggregation(superBlockID, plane, req.DeviceModelID, req.SpineCount, req.HostLinkSpeedGbps)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("assign super-block aggregation", "err", err, "superBlockID", superBlockID, "plane", plane)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// GetSuperBlockAggregation handles GET /api/super-blocks/{id}/aggregations/{plane}.
func (h *BlockHandler) GetSuperBlockAggregation(w http.ResponseWriter, r *http.Request) {
	superBlockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	planeStr := r.PathValue("plane")
	plane, ok := parsePlane(w, planeStr)
	if !ok {
		return
	}

	summary, err := h.svc.GetSuperBlockAggregationSummary(superBlockID, plane)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "aggregation not found")
			return
		}
		slog.Error("get super-block aggregation", "err", err, "superBlockID", superBlockID, "plane", plane)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// placeSpineDevicesRequest is the request body for PlaceSpineDevices.
type placeSpineDevicesRequest struct {
	DeviceModelID int64 `json:"device_model_id"`
	Count         int   `json:"count"`
}

// PlaceSpineDevices handles POST /api/blocks/{id}/spine-devices.
func (h *BlockHandler) PlaceSpineDevices(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req placeSpineDevicesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DeviceModelID <= 0 {
		writeError(w, http.StatusBadRequest, "device_model_id is required")
		return
	}

	if err := h.svc.PlaceSpineDevices(blockID, req.DeviceModelID, req.Count); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("place spine devices", "err", err, "blockID", blockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteBlock handles DELETE /api/blocks/{id}.
func (h *BlockHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteBlock(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "block not found")
			return
		}
		slog.Error("delete block", "err", err, "blockID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parsePlane parses a network plane string from a path variable.
func parsePlane(w http.ResponseWriter, s string) (models.NetworkPlane, bool) {
	switch s {
	case string(models.NetworkPlaneFrontEnd), string(models.NetworkPlaneManagement):
		return models.NetworkPlane(s), true
	default:
		writeError(w, http.StatusBadRequest, "plane must be 'front_end' or 'management'")
		return "", false
	}
}
