package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// DeriveFabricService is the interface the handler calls.
type DeriveFabricService interface {
	DeriveFabric(designID int64, plane models.NetworkPlane) (*service.DerivedFabric, error)
}

// DeriveFabricHandler handles derived-fabric endpoints.
type DeriveFabricHandler struct {
	svc DeriveFabricService
}

// NewDeriveFabricHandler returns a new DeriveFabricHandler.
func NewDeriveFabricHandler(svc DeriveFabricService) *DeriveFabricHandler {
	return &DeriveFabricHandler{svc: svc}
}

// GetDerivedFabric handles GET /api/designs/{id}/fabric.
// Returns the derived Clos topology for a design's front_end plane.
func (h *DeriveFabricHandler) GetDerivedFabric(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	df, err := h.svc.DeriveFabric(id, models.PlaneFrontEnd)
	if errors.Is(err, models.ErrNotFound) {
		writeError(w, http.StatusNotFound, "design not found")
		return
	}
	if err != nil {
		slog.Error("derive fabric", "err", err, "designID", id)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, df)
}
