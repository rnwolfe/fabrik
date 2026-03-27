package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// ManagementService is the business logic interface required by ManagementHandler.
type ManagementService interface {
	SetManagementAgg(blockID int64, deviceModelID int64) (*models.BlockAggregation, error)
	GetManagementAgg(blockID int64) (*models.BlockAggregation, error)
	RemoveManagementAgg(blockID int64) error
	ListBlockAggregations(blockID int64) ([]*models.BlockAggregation, error)
}

// ManagementHandler handles HTTP requests for the management plane resources.
type ManagementHandler struct {
	svc ManagementService
}

// NewManagementHandler returns a new ManagementHandler using svc.
func NewManagementHandler(svc ManagementService) *ManagementHandler {
	return &ManagementHandler{svc: svc}
}

type setManagementAggRequest struct {
	DeviceModelID int64 `json:"device_model_id"`
}

// SetManagementAgg handles PUT /api/blocks/{block_id}/management-agg.
func (h *ManagementHandler) SetManagementAgg(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "block_id")
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req setManagementAggRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	agg, err := h.svc.SetManagementAgg(blockID, req.DeviceModelID)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		slog.Error("set management agg", "err", err, "blockID", blockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, agg)
}

// GetManagementAgg handles GET /api/blocks/{block_id}/management-agg.
func (h *ManagementHandler) GetManagementAgg(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "block_id")
	if !ok {
		return
	}

	agg, err := h.svc.GetManagementAgg(blockID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no management aggregation assigned to this block")
			return
		}
		slog.Error("get management agg", "err", err, "blockID", blockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, agg)
}

// RemoveManagementAgg handles DELETE /api/blocks/{block_id}/management-agg.
func (h *ManagementHandler) RemoveManagementAgg(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "block_id")
	if !ok {
		return
	}

	if err := h.svc.RemoveManagementAgg(blockID); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no management aggregation assigned to this block")
			return
		}
		slog.Error("remove management agg", "err", err, "blockID", blockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListBlockAggregations handles GET /api/blocks/{block_id}/aggregations.
func (h *ManagementHandler) ListBlockAggregations(w http.ResponseWriter, r *http.Request) {
	blockID, ok := parseID(w, r, "block_id")
	if !ok {
		return
	}

	aggs, err := h.svc.ListBlockAggregations(blockID)
	if err != nil {
		slog.Error("list block aggregations", "err", err, "blockID", blockID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, aggs)
}
