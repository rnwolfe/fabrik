package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// SuperBlockService is the business logic interface required by SuperBlockHandler.
type SuperBlockService interface {
	CreateSuperBlock(siteID int64, name, description string) (*models.SuperBlock, error)
	GetSuperBlock(id int64) (*models.SuperBlock, error)
	ListSuperBlocks(siteID int64) ([]*models.SuperBlock, error)
	UpdateSuperBlock(id int64, name, description string) (*models.SuperBlock, error)
	DeleteSuperBlock(id int64) error
}

// SuperBlockHandler handles HTTP requests for SuperBlock resources.
type SuperBlockHandler struct {
	svc SuperBlockService
}

// NewSuperBlockHandler returns a new SuperBlockHandler using svc.
func NewSuperBlockHandler(svc SuperBlockService) *SuperBlockHandler {
	return &SuperBlockHandler{svc: svc}
}

type createSuperBlockRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateSuperBlock handles POST /api/sites/{siteId}/superblocks.
func (h *SuperBlockHandler) CreateSuperBlock(w http.ResponseWriter, r *http.Request) {
	siteID, ok := parseID(w, r, "siteId")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createSuperBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sb, err := h.svc.CreateSuperBlock(siteID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("create super-block", "err", err, "siteID", siteID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, sb)
}

// ListSuperBlocks handles GET /api/sites/{siteId}/superblocks.
func (h *SuperBlockHandler) ListSuperBlocks(w http.ResponseWriter, r *http.Request) {
	siteID, ok := parseID(w, r, "siteId")
	if !ok {
		return
	}

	sbs, err := h.svc.ListSuperBlocks(siteID)
	if err != nil {
		slog.Error("list super-blocks", "err", err, "siteID", siteID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if sbs == nil {
		sbs = []*models.SuperBlock{}
	}
	writeJSON(w, http.StatusOK, sbs)
}

// GetSuperBlock handles GET /api/superblocks/{id}.
func (h *SuperBlockHandler) GetSuperBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	sb, err := h.svc.GetSuperBlock(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "super-block not found")
			return
		}
		slog.Error("get super-block", "err", err, "superBlockID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, sb)
}

type updateSuperBlockRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateSuperBlock handles PUT /api/superblocks/{id}.
func (h *SuperBlockHandler) UpdateSuperBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateSuperBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sb, err := h.svc.UpdateSuperBlock(id, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "super-block not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("update super-block", "err", err, "superBlockID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, sb)
}

// DeleteSuperBlock handles DELETE /api/superblocks/{id}.
func (h *SuperBlockHandler) DeleteSuperBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.DeleteSuperBlock(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "super-block not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("delete super-block", "err", err, "superBlockID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
