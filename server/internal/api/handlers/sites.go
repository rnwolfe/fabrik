package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// SiteService is the business logic interface required by SiteHandler.
type SiteService interface {
	CreateSite(designID int64, name, description string) (*models.Site, error)
	GetSite(id int64) (*models.Site, error)
	ListSites(designID int64) ([]*models.Site, error)
	UpdateSite(id int64, name, description string) (*models.Site, error)
	DeleteSite(id int64) error
	GetDesignHierarchy(designID int64) (*store.DesignHierarchy, error)
}

// SiteHandler handles HTTP requests for Site resources.
type SiteHandler struct {
	svc SiteService
}

// NewSiteHandler returns a new SiteHandler using svc.
func NewSiteHandler(svc SiteService) *SiteHandler {
	return &SiteHandler{svc: svc}
}

type createSiteRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateSite handles POST /api/designs/{designId}/sites.
func (h *SiteHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	designID, ok := parseID(w, r, "designId")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	site, err := h.svc.CreateSite(designID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("create site", "err", err, "designID", designID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, site)
}

// ListSites handles GET /api/designs/{designId}/sites.
func (h *SiteHandler) ListSites(w http.ResponseWriter, r *http.Request) {
	designID, ok := parseID(w, r, "designId")
	if !ok {
		return
	}

	sites, err := h.svc.ListSites(designID)
	if err != nil {
		slog.Error("list sites", "err", err, "designID", designID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if sites == nil {
		sites = []*models.Site{}
	}
	writeJSON(w, http.StatusOK, sites)
}

// GetSite handles GET /api/sites/{id}.
func (h *SiteHandler) GetSite(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	site, err := h.svc.GetSite(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		slog.Error("get site", "err", err, "siteID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, site)
}

type updateSiteRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateSite handles PUT /api/sites/{id}.
func (h *SiteHandler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req updateSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	site, err := h.svc.UpdateSite(id, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("update site", "err", err, "siteID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, site)
}

// DeleteSite handles DELETE /api/sites/{id}.
func (h *SiteHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.DeleteSite(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("delete site", "err", err, "siteID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetDesignHierarchy handles GET /api/designs/{id}/hierarchy.
func (h *SiteHandler) GetDesignHierarchy(w http.ResponseWriter, r *http.Request) {
	designID, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	hierarchy, err := h.svc.GetDesignHierarchy(designID)
	if err != nil {
		slog.Error("get design hierarchy", "err", err, "designID", designID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, hierarchy)
}
