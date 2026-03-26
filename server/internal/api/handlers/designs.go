// Package handlers contains HTTP request handlers grouped by domain.
package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// DesignService is the business logic interface required by DesignHandler.
type DesignService interface {
	CreateDesign(name, description string) (*models.Design, error)
	ListDesigns() ([]*models.Design, error)
	GetDesign(id int64) (*models.Design, error)
	DeleteDesign(id int64) error
}

// DesignHandler handles HTTP requests for Design resources.
type DesignHandler struct {
	svc DesignService
}

// NewDesignHandler returns a new DesignHandler using svc.
func NewDesignHandler(svc DesignService) *DesignHandler {
	return &DesignHandler{svc: svc}
}

// createRequest is the JSON body for POST /api/designs.
type createDesignRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Create handles POST /api/designs.
func (h *DesignHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1 MB to guard against large-payload DoS attacks.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createDesignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	d, err := h.svc.CreateDesign(req.Name, req.Description)
	if err != nil {
		if errors.Is(err, models.ErrConstraintViolation) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("create design", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, d)
}

// List handles GET /api/designs.
func (h *DesignHandler) List(w http.ResponseWriter, r *http.Request) {
	designs, err := h.svc.ListDesigns()
	if err != nil {
		slog.Error("list designs", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	// Return empty array instead of null.
	if designs == nil {
		designs = []*models.Design{}
	}
	writeJSON(w, http.StatusOK, designs)
}

// Get handles GET /api/designs/:id.
func (h *DesignHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	d, err := h.svc.GetDesign(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "design not found")
			return
		}
		slog.Error("get design", "err", err, "designID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, d)
}

// Delete handles DELETE /api/designs/:id.
func (h *DesignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.DeleteDesign(id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "design not found")
			return
		}
		slog.Error("delete design", "err", err, "designID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode json response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// parseID extracts a path parameter named key from the URL path using Go 1.22+
// pattern variables and parses it as int64.
func parseID(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	val := r.PathValue(key)
	if val == "" {
		// Fallback: parse the last segment of the path.
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		val = parts[len(parts)-1]
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
