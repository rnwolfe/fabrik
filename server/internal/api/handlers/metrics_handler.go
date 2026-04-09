package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// MetricsService is the business logic interface required by MetricsHandler.
type MetricsService interface {
	GetDesignMetrics(designID int64) (*models.DesignMetrics, error)
}

// MetricsHandler handles HTTP requests for design metrics.
type MetricsHandler struct {
	svc MetricsService
}

// NewMetricsHandler returns a new MetricsHandler using svc.
func NewMetricsHandler(svc MetricsService) *MetricsHandler {
	return &MetricsHandler{svc: svc}
}

// GetDesignMetrics handles GET /api/designs/{id}/metrics.
// It returns all computed metrics for the given design.
func (h *MetricsHandler) GetDesignMetrics(w http.ResponseWriter, r *http.Request) {
	designID, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	m, err := h.svc.GetDesignMetrics(designID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			writeError(w, http.StatusNotFound, "design not found")
			return
		}
		slog.Error("get design metrics", "err", err, "designID", designID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, m)
}
