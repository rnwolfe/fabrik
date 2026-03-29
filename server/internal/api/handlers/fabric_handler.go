package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// FabricService is the business logic interface required by FabricHandler.
type FabricService interface {
	CreateFabric(req service.CreateFabricRequest) (*service.FabricResponse, error)
	ListFabrics() ([]*service.FabricResponse, error)
	GetFabric(id int64) (*service.FabricResponse, error)
	UpdateFabric(id int64, req service.UpdateFabricRequest) (*service.FabricResponse, error)
	DeleteFabric(id int64) error
	PreviewTopology(req service.PreviewTopologyRequest) (*service.TopologyPlan, error)
	ListDeviceModels() ([]*models.DeviceModel, error)
}

// FabricHandler handles HTTP requests for Fabric resources.
type FabricHandler struct {
	svc FabricService
}

// NewFabricHandler returns a new FabricHandler using svc.
func NewFabricHandler(svc FabricService) *FabricHandler {
	return &FabricHandler{svc: svc}
}

// Create handles POST /api/fabrics.
func (h *FabricHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req service.CreateFabricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.CreateFabric(req)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("create fabric", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// List handles GET /api/fabrics.
func (h *FabricHandler) List(w http.ResponseWriter, r *http.Request) {
	fabrics, err := h.svc.ListFabrics()
	if err != nil {
		slog.Error("list fabrics", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if fabrics == nil {
		fabrics = []*service.FabricResponse{}
	}
	writeJSON(w, http.StatusOK, fabrics)
}

// Get handles GET /api/fabrics/{id}.
func (h *FabricHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	resp, err := h.svc.GetFabric(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "fabric not found")
			return
		}
		slog.Error("get fabric", "err", err, "fabricID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Update handles PUT /api/fabrics/{id}.
func (h *FabricHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req service.UpdateFabricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.UpdateFabric(id, req)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "fabric not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("update fabric", "err", err, "fabricID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /api/fabrics/{id}.
func (h *FabricHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.DeleteFabric(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "fabric not found")
			return
		}
		slog.Error("delete fabric", "err", err, "fabricID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Preview handles POST /api/fabrics/preview — topology preview without persistence.
func (h *FabricHandler) Preview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req service.PreviewTopologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.svc.PreviewTopology(req)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("preview topology", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, plan)
}
