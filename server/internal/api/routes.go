// Package api wires together HTTP handlers and registers all routes.
package api

import (
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
)

// RegisterRoutes registers all API routes on mux.
func RegisterRoutes(mux *http.ServeMux, designs *handlers.DesignHandler, knowledge *handlers.KnowledgeHandler, deviceModels *handlers.DeviceModelHandler) {
	// Design CRUD
	mux.HandleFunc("POST /api/designs", designs.Create)
	mux.HandleFunc("GET /api/designs", designs.List)
	mux.HandleFunc("GET /api/designs/{id}", designs.Get)
	mux.HandleFunc("DELETE /api/designs/{id}", designs.Delete)

	// Device model catalog
	mux.HandleFunc("POST /api/catalog/devices", deviceModels.Create)
	mux.HandleFunc("GET /api/catalog/devices", deviceModels.List)
	mux.HandleFunc("GET /api/catalog/devices/{id}", deviceModels.Get)
	mux.HandleFunc("PUT /api/catalog/devices/{id}", deviceModels.Update)
	mux.HandleFunc("DELETE /api/catalog/devices/{id}", deviceModels.Delete)
	mux.HandleFunc("POST /api/catalog/devices/{id}/duplicate", deviceModels.Duplicate)

	// Knowledge base
	mux.HandleFunc("GET /api/knowledge", knowledge.Index)
	mux.HandleFunc("GET /api/knowledge/{path...}", knowledge.Get)
}
