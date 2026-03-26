// Package api wires together HTTP handlers and registers all routes.
package api

import (
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
)

// RegisterRoutes registers all API routes on mux.
func RegisterRoutes(mux *http.ServeMux, designs *handlers.DesignHandler) {
	// Design CRUD
	mux.HandleFunc("POST /api/designs", designs.Create)
	mux.HandleFunc("GET /api/designs", designs.List)
	mux.HandleFunc("GET /api/designs/{id}", designs.Get)
	mux.HandleFunc("DELETE /api/designs/{id}", designs.Delete)
}
