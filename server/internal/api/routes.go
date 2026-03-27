// Package api wires together HTTP handlers and registers all routes.
package api

import (
	"net/http"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
)

// RegisterRoutes registers all API routes on mux.
func RegisterRoutes(mux *http.ServeMux, designs *handlers.DesignHandler, knowledge *handlers.KnowledgeHandler, deviceModels *handlers.DeviceModelHandler, racks *handlers.RackHandler, fabrics *handlers.FabricHandler, blocks *handlers.BlockHandler, management *handlers.ManagementHandler, capacity *handlers.CapacityHandler) {
	// Design CRUD
	mux.HandleFunc("POST /api/designs", designs.Create)
	mux.HandleFunc("GET /api/designs", designs.List)
	mux.HandleFunc("GET /api/designs/{id}", designs.Get)
	mux.HandleFunc("DELETE /api/designs/{id}", designs.Delete)

	mux.HandleFunc("POST /api/catalog/devices", deviceModels.Create)
	mux.HandleFunc("GET /api/catalog/devices", deviceModels.List)
	mux.HandleFunc("GET /api/catalog/devices/{id}", deviceModels.Get)
	mux.HandleFunc("PUT /api/catalog/devices/{id}", deviceModels.Update)
	mux.HandleFunc("DELETE /api/catalog/devices/{id}", deviceModels.Delete)
	mux.HandleFunc("POST /api/catalog/devices/{id}/duplicate", deviceModels.Duplicate)

	mux.HandleFunc("GET /api/knowledge", knowledge.Index)
	mux.HandleFunc("GET /api/knowledge/{path...}", knowledge.Get)

	mux.HandleFunc("POST /api/rack-types", racks.CreateRackType)
	mux.HandleFunc("GET /api/rack-types", racks.ListRackTypes)
	mux.HandleFunc("GET /api/rack-types/{id}", racks.GetRackType)
	mux.HandleFunc("PUT /api/rack-types/{id}", racks.UpdateRackType)
	mux.HandleFunc("DELETE /api/rack-types/{id}", racks.DeleteRackType)

	mux.HandleFunc("POST /api/racks", racks.CreateRack)
	mux.HandleFunc("GET /api/racks", racks.ListRacks)
	mux.HandleFunc("GET /api/racks/{id}", racks.GetRack)
	mux.HandleFunc("PUT /api/racks/{id}", racks.UpdateRack)
	mux.HandleFunc("DELETE /api/racks/{id}", racks.DeleteRack)

	mux.HandleFunc("POST /api/racks/{id}/devices", racks.PlaceDevice)
	mux.HandleFunc("PUT /api/racks/{rack_id}/devices/{device_id}", racks.MoveDeviceInRack)
	mux.HandleFunc("PUT /api/racks/{rack_id}/devices/{device_id}/move", racks.MoveDeviceCrossRack)
	mux.HandleFunc("DELETE /api/racks/{rack_id}/devices/{device_id}", racks.RemoveDevice)

	mux.HandleFunc("POST /api/fabrics/preview", fabrics.Preview)
	mux.HandleFunc("POST /api/fabrics", fabrics.Create)
	mux.HandleFunc("GET /api/fabrics", fabrics.List)
	mux.HandleFunc("GET /api/fabrics/{id}", fabrics.Get)
	mux.HandleFunc("PUT /api/fabrics/{id}", fabrics.Update)
	mux.HandleFunc("DELETE /api/fabrics/{id}", fabrics.Delete)

	mux.HandleFunc("POST /api/blocks", blocks.CreateBlock)
	mux.HandleFunc("GET /api/blocks", blocks.ListBlocks)
	mux.HandleFunc("GET /api/blocks/{id}", blocks.GetBlock)

	mux.HandleFunc("PUT /api/blocks/{id}/aggregations/{plane}", blocks.AssignAggregation)
	mux.HandleFunc("GET /api/blocks/{id}/aggregations/{plane}", blocks.GetAggregation)
	mux.HandleFunc("GET /api/blocks/{id}/aggregations", blocks.ListAggregations)
	mux.HandleFunc("DELETE /api/blocks/{id}/aggregations/{plane}", blocks.DeleteAggregation)

	mux.HandleFunc("GET /api/blocks/{id}/aggregations/{plane}/connections", blocks.ListPortConnections)

	mux.HandleFunc("POST /api/blocks/add-rack", blocks.AddRackToBlock)
	mux.HandleFunc("DELETE /api/blocks/racks/{rack_id}", blocks.RemoveRackFromBlock)

	mux.HandleFunc("PUT /api/blocks/{block_id}/management-agg", management.SetManagementAgg)
	mux.HandleFunc("GET /api/blocks/{block_id}/management-agg", management.GetManagementAgg)
	mux.HandleFunc("DELETE /api/blocks/{block_id}/management-agg", management.RemoveManagementAgg)

	// Capacity aggregation
	mux.HandleFunc("GET /api/designs/{id}/capacity", capacity.GetDesignCapacity)
}
