package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
)

// fakeCapacitySvc implements handlers.CapacityService for testing.
type fakeCapacitySvc struct {
	designs     map[int64]*models.CapacitySummary
	sites       map[int64]*models.CapacitySummary
	superBlocks map[int64]*models.CapacitySummary
	blocks      map[int64]*models.CapacitySummary
	racks       map[int64]*models.CapacitySummary
}

func newFakeCapacitySvc() *fakeCapacitySvc {
	return &fakeCapacitySvc{
		designs:     make(map[int64]*models.CapacitySummary),
		sites:       make(map[int64]*models.CapacitySummary),
		superBlocks: make(map[int64]*models.CapacitySummary),
		blocks:      make(map[int64]*models.CapacitySummary),
		racks:       make(map[int64]*models.CapacitySummary),
	}
}

func (f *fakeCapacitySvc) GetRackCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := f.racks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return c, nil
}

func (f *fakeCapacitySvc) GetBlockCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := f.blocks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return c, nil
}

func (f *fakeCapacitySvc) GetSuperBlockCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := f.superBlocks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return c, nil
}

func (f *fakeCapacitySvc) GetSiteCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := f.sites[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return c, nil
}

func (f *fakeCapacitySvc) GetDesignCapacity(id int64) (*models.CapacitySummary, error) {
	c, ok := f.designs[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return c, nil
}

func TestCapacityHandler_GetDesignCapacity_DesignLevel(t *testing.T) {
	svc := newFakeCapacitySvc()
	svc.designs[1] = &models.CapacitySummary{
		Level:             models.CapacityLevelDesign,
		ID:                1,
		Name:              "my-design",
		PowerWattsTypical: 5000,
		DeviceCount:       20,
	}

	h := handlers.NewCapacityHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/designs/{id}/capacity", h.GetDesignCapacity)

	req := httptest.NewRequest("GET", "/api/designs/1/capacity", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: want 200 got %d — body: %s", w.Code, w.Body.String())
	}

	var c models.CapacitySummary
	if err := json.NewDecoder(w.Body).Decode(&c); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if c.PowerWattsTypical != 5000 {
		t.Errorf("power_watts_typical: want 5000 got %d", c.PowerWattsTypical)
	}
	if c.DeviceCount != 20 {
		t.Errorf("device_count: want 20 got %d", c.DeviceCount)
	}
}

func TestCapacityHandler_GetDesignCapacity_RackLevel(t *testing.T) {
	svc := newFakeCapacitySvc()
	svc.racks[7] = &models.CapacitySummary{
		Level:             models.CapacityLevelRack,
		ID:                7,
		Name:              "rack-7",
		PowerWattsTypical: 800,
	}

	h := handlers.NewCapacityHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/designs/{id}/capacity", h.GetDesignCapacity)

	req := httptest.NewRequest("GET", "/api/designs/1/capacity?level=rack&entity_id=7", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: want 200 got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCapacityHandler_GetDesignCapacity_NotFound(t *testing.T) {
	svc := newFakeCapacitySvc()

	h := handlers.NewCapacityHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/designs/{id}/capacity", h.GetDesignCapacity)

	req := httptest.NewRequest("GET", "/api/designs/999/capacity", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: want 404 got %d", w.Code)
	}
}

func TestCapacityHandler_GetDesignCapacity_InvalidLevel(t *testing.T) {
	svc := newFakeCapacitySvc()

	h := handlers.NewCapacityHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/designs/{id}/capacity", h.GetDesignCapacity)

	req := httptest.NewRequest("GET", "/api/designs/1/capacity?level=invalid", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: want 400 got %d", w.Code)
	}
}

func TestCapacityHandler_GetDesignCapacity_MissingEntityID(t *testing.T) {
	svc := newFakeCapacitySvc()

	h := handlers.NewCapacityHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/designs/{id}/capacity", h.GetDesignCapacity)

	// Rack level requires entity_id
	req := httptest.NewRequest("GET", "/api/designs/1/capacity?level=rack", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: want 400 got %d", w.Code)
	}
}
