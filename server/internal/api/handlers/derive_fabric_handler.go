package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid design id", http.StatusBadRequest)
		return
	}

	df, err := h.svc.DeriveFabric(id, models.PlaneFrontEnd)
	if errors.Is(err, models.ErrNotFound) {
		http.Error(w, "design not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(df)
}
