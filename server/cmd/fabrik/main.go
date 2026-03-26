package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/rnwolfe/fabrik/server/internal/api"
	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/migrations"
	"github.com/rnwolfe/fabrik/server/internal/service"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbPath := "fabrik.db"
	if p := os.Getenv("FABRIK_DB"); p != "" {
		dbPath = p
	}

	db, err := store.Open(dbPath, migrations.FS)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Wire up layers: store → service → handler
	designStore := store.NewDesignStore(db)
	designSvc := service.NewDesignService(designStore)
	designHandler := handlers.NewDesignHandler(designSvc)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Register domain routes
	api.RegisterRoutes(mux, designHandler)

	addr := ":8080"
	if port := os.Getenv("FABRIK_PORT"); port != "" {
		addr = ":" + port
	}

	slog.Info("starting fabrik", "addr", addr, "db", dbPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
