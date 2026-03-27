package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
)

// fakeDeviceModelService is an in-memory implementation for handler tests.
type fakeDeviceModelService struct {
	dms    map[int64]*models.DeviceModel
	nextID int64
}

func newFakeDMSvc() *fakeDeviceModelService {
	return &fakeDeviceModelService{dms: make(map[int64]*models.DeviceModel)}
}

func (s *fakeDeviceModelService) CreateDeviceModel(dm *models.DeviceModel) (*models.DeviceModel, error) {
	if dm.Vendor == "" {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("vendor is required"))
	}
	if dm.Model == "" {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("model is required"))
	}
	for _, existing := range s.dms {
		if existing.Vendor == dm.Vendor && existing.Model == dm.Model {
			return nil, models.ErrDuplicate
		}
	}
	s.nextID++
	out := *dm
	out.ID = s.nextID
	out.CreatedAt = time.Now()
	out.UpdatedAt = time.Now()
	s.dms[out.ID] = &out
	return &out, nil
}

func (s *fakeDeviceModelService) ListDeviceModels(includeArchived bool) ([]*models.DeviceModel, error) {
	out := make([]*models.DeviceModel, 0)
	for _, dm := range s.dms {
		if !includeArchived && dm.ArchivedAt != nil {
			continue
		}
		cp := *dm
		out = append(out, &cp)
	}
	return out, nil
}

func (s *fakeDeviceModelService) GetDeviceModel(id int64) (*models.DeviceModel, error) {
	dm, ok := s.dms[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *dm
	return &cp, nil
}

func (s *fakeDeviceModelService) UpdateDeviceModel(dm *models.DeviceModel) (*models.DeviceModel, error) {
	existing, ok := s.dms[dm.ID]
	if !ok {
		return nil, models.ErrNotFound
	}
	if existing.IsSeed {
		return nil, models.ErrSeedReadOnly
	}
	if dm.Vendor == "" {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("vendor is required"))
	}
	out := *dm
	s.dms[dm.ID] = &out
	return &out, nil
}

func (s *fakeDeviceModelService) ArchiveDeviceModel(id int64) error {
	dm, ok := s.dms[id]
	if !ok {
		return models.ErrNotFound
	}
	if dm.IsSeed {
		return models.ErrSeedReadOnly
	}
	now := time.Now()
	dm.ArchivedAt = &now
	return nil
}

func (s *fakeDeviceModelService) DuplicateDeviceModel(id int64) (*models.DeviceModel, error) {
	src, ok := s.dms[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	s.nextID++
	cp := *src
	cp.ID = s.nextID
	cp.Model = src.Model + " (copy)"
	cp.IsSeed = false
	cp.ArchivedAt = nil
	s.dms[cp.ID] = &cp
	return &cp, nil
}

// --- tests ---

func TestDeviceModelHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid",
			body:       `{"vendor":"Acme","model":"Switch X","port_count":48,"height_u":1,"power_watts":300}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty vendor",
			body:       `{"vendor":"","model":"Model X","height_u":1}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty model",
			body:       `{"vendor":"Acme","model":"","height_u":1}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `{bad json`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewDeviceModelHandler(newFakeDMSvc())
			req := httptest.NewRequest(http.MethodPost, "/api/catalog/devices", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Create(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("want %d got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDeviceModelHandler_CreateDuplicate(t *testing.T) {
	svc := newFakeDMSvc()
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "Acme", Model: "SW", HeightU: 1})

	h := handlers.NewDeviceModelHandler(svc)
	body := `{"vendor":"Acme","model":"SW","height_u":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/catalog/devices", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestDeviceModelHandler_List(t *testing.T) {
	svc := newFakeDMSvc()
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "A", Model: "M1", HeightU: 1})
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "B", Model: "M2", HeightU: 1})

	h := handlers.NewDeviceModelHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/catalog/devices", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	var out []*models.DeviceModel
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("want 2 models, got %d", len(out))
	}
}

func TestDeviceModelHandler_List_EmptyIsArray(t *testing.T) {
	h := handlers.NewDeviceModelHandler(newFakeDMSvc())
	req := httptest.NewRequest(http.MethodGet, "/api/catalog/devices", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var out interface{}
	json.NewDecoder(rec.Body).Decode(&out)
	if _, ok := out.([]interface{}); !ok {
		t.Error("expected JSON array for empty result, got", out)
	}
}

func TestDeviceModelHandler_Get(t *testing.T) {
	svc := newFakeDMSvc()
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "M", HeightU: 1})

	h := handlers.NewDeviceModelHandler(svc)

	t.Run("found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/catalog/devices/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/catalog/devices/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("want 200, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/catalog/devices/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/catalog/devices/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d", rec.Code)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/catalog/devices/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/catalog/devices/nan", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})
}

func TestDeviceModelHandler_Update(t *testing.T) {
	svc := newFakeDMSvc()
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "Orig", HeightU: 1})

	// Add a seed model directly
	svc.nextID++
	seedID := svc.nextID
	svc.dms[seedID] = &models.DeviceModel{ID: seedID, Vendor: "Seed", Model: "SW", HeightU: 1, IsSeed: true}

	h := handlers.NewDeviceModelHandler(svc)

	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/catalog/devices/{id}", h.Update)
		body := `{"vendor":"V","model":"Updated","height_u":2}`
		req := httptest.NewRequest(http.MethodPut, "/api/catalog/devices/1", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("seed model returns 403", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/catalog/devices/{id}", h.Update)
		body := `{"vendor":"X","model":"Y","height_u":1}`
		req := httptest.NewRequest(http.MethodPut, "/api/catalog/devices/2", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("want 403, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/catalog/devices/{id}", h.Update)
		body := `{"vendor":"V","model":"M","height_u":1}`
		req := httptest.NewRequest(http.MethodPut, "/api/catalog/devices/9999", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d", rec.Code)
		}
	})
}

func TestDeviceModelHandler_Delete(t *testing.T) {
	svc := newFakeDMSvc()
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "M", HeightU: 1})

	// Seed model
	svc.nextID++
	seedID := svc.nextID
	svc.dms[seedID] = &models.DeviceModel{ID: seedID, Vendor: "Seed", Model: "SW", HeightU: 1, IsSeed: true}

	h := handlers.NewDeviceModelHandler(svc)

	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/catalog/devices/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/catalog/devices/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Errorf("want 204, got %d", rec.Code)
		}
	})

	t.Run("seed model returns 403", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/catalog/devices/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/catalog/devices/2", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("want 403, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/catalog/devices/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/catalog/devices/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d", rec.Code)
		}
	})
}

func TestDeviceModelHandler_Duplicate(t *testing.T) {
	svc := newFakeDMSvc()
	svc.CreateDeviceModel(&models.DeviceModel{Vendor: "V", Model: "Source", HeightU: 1})

	h := handlers.NewDeviceModelHandler(svc)

	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/catalog/devices/{id}/duplicate", h.Duplicate)
		req := httptest.NewRequest(http.MethodPost, "/api/catalog/devices/1/duplicate", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("want 201, got %d (body: %s)", rec.Code, rec.Body.String())
		}
		var dm models.DeviceModel
		json.NewDecoder(rec.Body).Decode(&dm)
		if dm.IsSeed {
			t.Error("duplicate should not be a seed")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/catalog/devices/{id}/duplicate", h.Duplicate)
		req := httptest.NewRequest(http.MethodPost, "/api/catalog/devices/9999/duplicate", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d", rec.Code)
		}
	})
}
