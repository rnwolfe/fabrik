package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// fakeFabricService is an in-memory implementation of FabricService for handler tests.
type fakeFabricService struct {
	fabrics map[int64]*service.FabricResponse
	nextID  int64
}

func newFakeFabricSvc() *fakeFabricService {
	return &fakeFabricService{fabrics: make(map[int64]*service.FabricResponse)}
}

func makeFabricResp(id int64, name string, stages, radix int, os float64) *service.FabricResponse {
	return &service.FabricResponse{
		FabricRecord: &store.FabricRecord{
			Fabric: models.Fabric{
				ID:   id,
				Name: name,
				Tier: models.FabricTierFrontEnd,
			},
			Stages:           stages,
			Radix:            radix,
			Oversubscription: os,
		},
		Topology: &service.TopologyPlan{
			Stages: stages, Radix: radix, Oversubscription: os,
			SpineCount: 32, LeafCount: 1, LeafUplinks: 32, LeafDownlinks: 32,
			TotalSwitches: 33, TotalHostPorts: 32,
		},
		Metrics: &service.FabricMetrics{
			LeafSpineOversubscription: os,
		},
	}
}

func (s *fakeFabricService) CreateFabric(req service.CreateFabricRequest) (*service.FabricResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("fabric name is required"))
	}
	if req.Stages != 2 && req.Stages != 3 && req.Stages != 5 {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("stages must be 2, 3, or 5"))
	}
	if req.Radix <= 0 {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("radix must be > 0"))
	}
	if req.Oversubscription < 1.0 {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("oversubscription must be >= 1.0"))
	}
	s.nextID++
	resp := makeFabricResp(s.nextID, req.Name, req.Stages, req.Radix, req.Oversubscription)
	s.fabrics[resp.ID] = resp
	return resp, nil
}

func (s *fakeFabricService) ListFabrics() ([]*service.FabricResponse, error) {
	out := make([]*service.FabricResponse, 0, len(s.fabrics))
	for _, f := range s.fabrics {
		cp := *f
		out = append(out, &cp)
	}
	return out, nil
}

func (s *fakeFabricService) GetFabric(id int64) (*service.FabricResponse, error) {
	f, ok := s.fabrics[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *f
	return &cp, nil
}

func (s *fakeFabricService) UpdateFabric(id int64, req service.UpdateFabricRequest) (*service.FabricResponse, error) {
	if _, ok := s.fabrics[id]; !ok {
		return nil, models.ErrNotFound
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("fabric name is required"))
	}
	resp := makeFabricResp(id, req.Name, req.Stages, req.Radix, req.Oversubscription)
	s.fabrics[id] = resp
	return resp, nil
}

func (s *fakeFabricService) DeleteFabric(id int64) error {
	if _, ok := s.fabrics[id]; !ok {
		return models.ErrNotFound
	}
	delete(s.fabrics, id)
	return nil
}

func (s *fakeFabricService) PreviewTopology(req service.PreviewTopologyRequest) (*service.TopologyPlan, error) {
	if req.Stages != 2 && req.Stages != 3 && req.Stages != 5 {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("stages must be 2, 3, or 5"))
	}
	if req.Radix <= 0 {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("radix must be > 0"))
	}
	if req.Oversubscription < 1.0 {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("oversubscription must be >= 1.0"))
	}
	return &service.TopologyPlan{
		Stages: req.Stages, Radix: req.Radix, Oversubscription: req.Oversubscription,
		SpineCount: 32, LeafCount: 64, LeafUplinks: 32, LeafDownlinks: 32,
		TotalSwitches: 96, TotalHostPorts: 2048,
	}, nil
}

func (s *fakeFabricService) ListDeviceModels() ([]*models.DeviceModel, error) {
	return []*models.DeviceModel{}, nil
}

// --- Tests ---

func TestFabricHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid",
			body:       `{"name":"test-fabric","tier":"frontend","stages":2,"radix":64,"oversubscription":1.0}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty name",
			body:       `{"name":"","tier":"frontend","stages":2,"radix":64,"oversubscription":1.0}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid stages",
			body:       `{"name":"test","tier":"frontend","stages":4,"radix":64,"oversubscription":1.0}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "zero radix",
			body:       `{"name":"test","tier":"frontend","stages":2,"radix":0,"oversubscription":1.0}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "oversubscription below 1",
			body:       `{"name":"test","tier":"frontend","stages":2,"radix":64,"oversubscription":0.5}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid json",
			body:       `{bad json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "body exceeds 1MB",
			body:       `{"name":"` + strings.Repeat("a", (1<<20)+1) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewFabricHandler(newFakeFabricSvc())
			req := httptest.NewRequest(http.MethodPost, "/api/fabrics", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Create(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestFabricHandler_List(t *testing.T) {
	svc := newFakeFabricSvc()
	svc.CreateFabric(service.CreateFabricRequest{
		Name: "f1", Tier: models.FabricTierFrontEnd, Stages: 2, Radix: 64, Oversubscription: 1.0,
	})
	svc.CreateFabric(service.CreateFabricRequest{
		Name: "f2", Tier: models.FabricTierBackEnd, Stages: 3, Radix: 48, Oversubscription: 2.0,
	})

	h := handlers.NewFabricHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/fabrics", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var fabrics []json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&fabrics); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(fabrics) != 2 {
		t.Errorf("expected 2 fabrics, got %d", len(fabrics))
	}
}

func TestFabricHandler_Get(t *testing.T) {
	svc := newFakeFabricSvc()
	svc.CreateFabric(service.CreateFabricRequest{
		Name: "get-fabric", Tier: models.FabricTierFrontEnd, Stages: 2, Radix: 64, Oversubscription: 1.0,
	})

	h := handlers.NewFabricHandler(svc)

	t.Run("found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/fabrics/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/fabrics/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/fabrics/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/fabrics/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/fabrics/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/fabrics/not-a-number", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestFabricHandler_Update(t *testing.T) {
	svc := newFakeFabricSvc()
	svc.CreateFabric(service.CreateFabricRequest{
		Name: "update-me", Tier: models.FabricTierFrontEnd, Stages: 2, Radix: 64, Oversubscription: 1.0,
	})

	h := handlers.NewFabricHandler(svc)

	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/fabrics/{id}", h.Update)
		body := `{"name":"updated","tier":"backend","stages":3,"radix":48,"oversubscription":2.0}`
		req := httptest.NewRequest(http.MethodPut, "/api/fabrics/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/fabrics/{id}", h.Update)
		body := `{"name":"updated","tier":"frontend","stages":2,"radix":64,"oversubscription":1.0}`
		req := httptest.NewRequest(http.MethodPut, "/api/fabrics/9999", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/fabrics/{id}", h.Update)
		req := httptest.NewRequest(http.MethodPut, "/api/fabrics/1", bytes.NewBufferString("{bad"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestFabricHandler_Delete(t *testing.T) {
	svc := newFakeFabricSvc()
	svc.CreateFabric(service.CreateFabricRequest{
		Name: "delete-me", Tier: models.FabricTierFrontEnd, Stages: 2, Radix: 64, Oversubscription: 1.0,
	})

	h := handlers.NewFabricHandler(svc)

	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/fabrics/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/fabrics/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/fabrics/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/fabrics/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestFabricHandler_Preview(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid 2-stage",
			body:       `{"stages":2,"radix":64,"oversubscription":1.0}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid 3-stage",
			body:       `{"stages":3,"radix":48,"oversubscription":2.0}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid stages",
			body:       `{"stages":4,"radix":64,"oversubscription":1.0}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewFabricHandler(newFakeFabricSvc())
			req := httptest.NewRequest(http.MethodPost, "/api/fabrics/preview", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Preview(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
