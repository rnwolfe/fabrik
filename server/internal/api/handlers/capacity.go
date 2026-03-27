package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// CapacityService is the business logic interface required by CapacityHandler.
type CapacityService interface {
	GetRackCapacity(rackID int64) (*models.CapacitySummary, error)
	GetBlockCapacity(blockID int64) (*models.CapacitySummary, error)
	GetSuperBlockCapacity(superBlockID int64) (*models.CapacitySummary, error)
	GetSiteCapacity(siteID int64) (*models.CapacitySummary, error)
	GetDesignCapacity(designID int64) (*models.CapacitySummary, error)
}

// CapacityHandler handles HTTP requests for capacity aggregation.
type CapacityHandler struct {
	svc CapacityService
}

// NewCapacityHandler returns a new CapacityHandler using svc.
func NewCapacityHandler(svc CapacityService) *CapacityHandler {
	return &CapacityHandler{svc: svc}
}

// GetDesignCapacity handles GET /api/designs/:id/capacity.
// It accepts an optional ?level=rack|block|superblock|site|design query param
// and an ?entity_id= param when the level is not "design".
// Defaults to design-level aggregation.
func (h *CapacityHandler) GetDesignCapacity(w http.ResponseWriter, r *http.Request) {
	designID, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	level := r.URL.Query().Get("level")
	if level == "" {
		level = string(models.CapacityLevelDesign)
	}

	switch models.CapacityLevel(level) {
	case models.CapacityLevelDesign:
		h.handleDesignLevel(w, r, designID)
	case models.CapacityLevelSite:
		entityID, ok := parseQueryID(w, r)
		if !ok {
			return
		}
		h.handleSiteLevel(w, r, entityID)
	case models.CapacityLevelSuperBlock:
		entityID, ok := parseQueryID(w, r)
		if !ok {
			return
		}
		h.handleSuperBlockLevel(w, r, entityID)
	case models.CapacityLevelBlock:
		entityID, ok := parseQueryID(w, r)
		if !ok {
			return
		}
		h.handleBlockLevel(w, r, entityID)
	case models.CapacityLevelRack:
		entityID, ok := parseQueryID(w, r)
		if !ok {
			return
		}
		h.handleRackLevel(w, r, entityID)
	default:
		writeError(w, http.StatusBadRequest, "invalid level: must be one of rack, block, superblock, site, design")
	}
}

func (h *CapacityHandler) handleDesignLevel(w http.ResponseWriter, _ *http.Request, designID int64) {
	c, err := h.svc.GetDesignCapacity(designID)
	if err != nil {
		handleCapacityErr(w, err, "design", designID)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CapacityHandler) handleSiteLevel(w http.ResponseWriter, _ *http.Request, siteID int64) {
	c, err := h.svc.GetSiteCapacity(siteID)
	if err != nil {
		handleCapacityErr(w, err, "site", siteID)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CapacityHandler) handleSuperBlockLevel(w http.ResponseWriter, _ *http.Request, id int64) {
	c, err := h.svc.GetSuperBlockCapacity(id)
	if err != nil {
		handleCapacityErr(w, err, "super-block", id)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CapacityHandler) handleBlockLevel(w http.ResponseWriter, _ *http.Request, id int64) {
	c, err := h.svc.GetBlockCapacity(id)
	if err != nil {
		handleCapacityErr(w, err, "block", id)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CapacityHandler) handleRackLevel(w http.ResponseWriter, _ *http.Request, id int64) {
	c, err := h.svc.GetRackCapacity(id)
	if err != nil {
		handleCapacityErr(w, err, "rack", id)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// parseQueryID extracts the "entity_id" query parameter as int64.
func parseQueryID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	val := r.URL.Query().Get("entity_id")
	if val == "" {
		writeError(w, http.StatusBadRequest, "entity_id query parameter is required for this level")
		return 0, false
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "entity_id must be a valid integer")
		return 0, false
	}
	return id, true
}

func handleCapacityErr(w http.ResponseWriter, err error, kind string, id int64) {
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, kind+" not found")
		return
	}
	slog.Error("compute capacity", "kind", kind, "id", id, "err", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}
