package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/rnwolfe/fabrik/server/internal/api"
	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/knowledge"
	"github.com/rnwolfe/fabrik/server/internal/migrations"
	"github.com/rnwolfe/fabrik/server/internal/service"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

//go:embed all:docs/knowledge
var docsFS embed.FS

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

	deviceModelStore := store.NewDeviceModelStore(db)
	deviceModelSvc := service.NewDeviceModelService(deviceModelStore)
	deviceModelHandler := handlers.NewDeviceModelHandler(deviceModelSvc)

	// Wire up rack types and racks
	rackTypeStore := store.NewRackTypeStore(db)
	rackStore := store.NewRackStore(db)
	rackSvc := service.NewRackService(rackTypeStore, rackStore)
	rackHandler := handlers.NewRackHandler(rackSvc)

<<<<<<< HEAD
	fabricStore := store.NewFabricStore(db)
	fabricSvc := service.NewFabricService(fabricStore)
	fabricHandler := handlers.NewFabricHandler(fabricSvc)
=======
	// Wire up blocks and block aggregation
	blockStore := store.NewBlockStore(db)
	blockSvc := service.NewBlockService(blockStore)
	blockHandler := handlers.NewBlockHandler(blockSvc)
>>>>>>> ab59967 (feat: block aggregation and rack-to-aggregation connectivity)

	// Wire up knowledge base
	knowledgeSub, err := fs.Sub(docsFS, "docs/knowledge")
	if err != nil {
		slog.Error("failed to sub knowledge FS", "err", err)
		os.Exit(1)
	}
	knowledgeLoader := knowledge.NewLoader(knowledgeSub)
	knowledgeHandler := handlers.NewKnowledgeHandler(knowledgeLoader)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Register domain routes
<<<<<<< HEAD
	api.RegisterRoutes(mux, designHandler, knowledgeHandler, deviceModelHandler, rackHandler, fabricHandler)
=======
	api.RegisterRoutes(mux, designHandler, knowledgeHandler, deviceModelHandler, rackHandler, blockHandler)
>>>>>>> ab59967 (feat: block aggregation and rack-to-aggregation connectivity)

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
