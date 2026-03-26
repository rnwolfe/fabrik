package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
)

// fakeDesignService is an in-memory implementation of DesignService for tests.
type fakeDesignService struct {
	designs map[int64]*models.Design
	nextID  int64
}

func newFakeSvc() *fakeDesignService {
	return &fakeDesignService{designs: make(map[int64]*models.Design)}
}

func (s *fakeDesignService) CreateDesign(name, description string) (*models.Design, error) {
	if name == "" {
		return nil, errors.Join(models.ErrConstraintViolation, errors.New("design name is required"))
	}
	s.nextID++
	d := &models.Design{ID: s.nextID, Name: name, Description: description}
	s.designs[d.ID] = d
	return d, nil
}

func (s *fakeDesignService) ListDesigns() ([]*models.Design, error) {
	out := make([]*models.Design, 0, len(s.designs))
	for _, d := range s.designs {
		cp := *d
		out = append(out, &cp)
	}
	return out, nil
}

func (s *fakeDesignService) GetDesign(id int64) (*models.Design, error) {
	d, ok := s.designs[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (s *fakeDesignService) DeleteDesign(id int64) error {
	if _, ok := s.designs[id]; !ok {
		return models.ErrNotFound
	}
	delete(s.designs, id)
	return nil
}

func TestDesignHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid",
			body:       `{"name":"test","description":"desc"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty name",
			body:       `{"name":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid json",
			body:       `{bad json`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewDesignHandler(newFakeSvc())
			req := httptest.NewRequest(http.MethodPost, "/api/designs", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Create(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDesignHandler_List(t *testing.T) {
	svc := newFakeSvc()
	svc.CreateDesign("d1", "")
	svc.CreateDesign("d2", "")

	h := handlers.NewDesignHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/designs", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var designs []models.Design
	if err := json.NewDecoder(rec.Body).Decode(&designs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(designs) != 2 {
		t.Errorf("expected 2 designs, got %d", len(designs))
	}
}

func TestDesignHandler_Get(t *testing.T) {
	svc := newFakeSvc()
	svc.CreateDesign("get-design", "")

	h := handlers.NewDesignHandler(svc)

	t.Run("found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/designs/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/designs/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/designs/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/designs/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("bad id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/designs/{id}", h.Get)
		req := httptest.NewRequest(http.MethodGet, "/api/designs/not-a-number", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestDesignHandler_Delete(t *testing.T) {
	svc := newFakeSvc()
	svc.CreateDesign("delete-me", "")

	h := handlers.NewDesignHandler(svc)

	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/designs/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/designs/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/designs/{id}", h.Delete)
		req := httptest.NewRequest(http.MethodDelete, "/api/designs/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}
