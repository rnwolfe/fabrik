package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
)

// fakeManagementService is an in-memory implementation for handler tests.
type fakeManagementService struct {
	aggs map[int64]*models.TierAggregation
}

func newFakeMgmtSvc() *fakeManagementService {
	return &fakeManagementService{aggs: make(map[int64]*models.TierAggregation)}
}

func (s *fakeManagementService) SetManagementAgg(blockID int64, deviceModelID int64) (*models.TierAggregation, error) {
	if deviceModelID <= 0 {
		return nil, models.ErrConstraintViolation
	}
	agg := &models.TierAggregation{
		ID:            1,
		ScopeType:     models.ScopeBlock,
		ScopeID:       blockID,
		Plane:         models.PlaneManagement,
		DeviceModelID: deviceModelID,
	}
	s.aggs[blockID] = agg
	return agg, nil
}

func (s *fakeManagementService) GetManagementAgg(blockID int64) (*models.TierAggregation, error) {
	agg, ok := s.aggs[blockID]
	if !ok {
		return nil, models.ErrNotFound
	}
	return agg, nil
}

func (s *fakeManagementService) RemoveManagementAgg(blockID int64) error {
	if _, ok := s.aggs[blockID]; !ok {
		return models.ErrNotFound
	}
	delete(s.aggs, blockID)
	return nil
}

func (s *fakeManagementService) ListBlockAggregations(blockID int64) ([]*models.TierAggregation, error) {
	var out []*models.TierAggregation
	for _, a := range s.aggs {
		if a.ScopeID == blockID {
			out = append(out, a)
		}
	}
	if out == nil {
		out = []*models.TierAggregation{}
	}
	return out, nil
}

// Helpers to build requests with path values for Go 1.22+ pattern-based mux.
func newManagementRequest(method, path string, body interface{}) *http.Request {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	return req
}

func setManagementPathValue(r *http.Request, key, value string) *http.Request {
	r.SetPathValue(key, value)
	return r
}

func TestManagementHandler_SetManagementAgg(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	body := map[string]interface{}{
		"device_model_id": 5,
	}
	r := setManagementPathValue(newManagementRequest(http.MethodPut, "/api/blocks/1/management-agg", body), "block_id", "1")
	w := httptest.NewRecorder()
	h.SetManagementAgg(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var agg models.TierAggregation
	if err := json.NewDecoder(w.Body).Decode(&agg); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if agg.ScopeID != 1 {
		t.Errorf("ScopeID = %d, want 1", agg.ScopeID)
	}
	if agg.DeviceModelID != 5 {
		t.Errorf("DeviceModelID = %d, want 5", agg.DeviceModelID)
	}
}

func TestManagementHandler_SetManagementAgg_InvalidBlockID(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	r := setManagementPathValue(newManagementRequest(http.MethodPut, "/api/blocks/abc/management-agg", map[string]int{}), "block_id", "abc")
	w := httptest.NewRecorder()
	h.SetManagementAgg(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestManagementHandler_GetManagementAgg(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	svc.aggs[2] = &models.TierAggregation{ID: 1, ScopeType: models.ScopeBlock, ScopeID: 2, Plane: models.PlaneManagement, DeviceModelID: 10}

	r := setManagementPathValue(newManagementRequest(http.MethodGet, "/api/blocks/2/management-agg", nil), "block_id", "2")
	w := httptest.NewRecorder()
	h.GetManagementAgg(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestManagementHandler_GetManagementAgg_NotFound(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	r := setManagementPathValue(newManagementRequest(http.MethodGet, "/api/blocks/99/management-agg", nil), "block_id", "99")
	w := httptest.NewRecorder()
	h.GetManagementAgg(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestManagementHandler_RemoveManagementAgg(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	svc.aggs[3] = &models.TierAggregation{ID: 1, ScopeType: models.ScopeBlock, ScopeID: 3, Plane: models.PlaneManagement, DeviceModelID: 7}

	r := setManagementPathValue(newManagementRequest(http.MethodDelete, "/api/blocks/3/management-agg", nil), "block_id", "3")
	w := httptest.NewRecorder()
	h.RemoveManagementAgg(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestManagementHandler_RemoveManagementAgg_NotFound(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	r := setManagementPathValue(newManagementRequest(http.MethodDelete, "/api/blocks/99/management-agg", nil), "block_id", "99")
	w := httptest.NewRecorder()
	h.RemoveManagementAgg(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestManagementHandler_ListBlockAggregations(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	svc.aggs[4] = &models.TierAggregation{ID: 1, ScopeType: models.ScopeBlock, ScopeID: 4, Plane: models.PlaneManagement, DeviceModelID: 3}

	r := setManagementPathValue(newManagementRequest(http.MethodGet, "/api/blocks/4/aggregations", nil), "block_id", "4")
	w := httptest.NewRecorder()
	h.ListBlockAggregations(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var aggs []models.TierAggregation
	if err := json.NewDecoder(w.Body).Decode(&aggs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(aggs) != 1 {
		t.Errorf("expected 1 aggregation, got %d", len(aggs))
	}
}

func TestManagementHandler_ListBlockAggregations_Empty(t *testing.T) {
	svc := newFakeMgmtSvc()
	h := handlers.NewManagementHandler(svc)

	r := setManagementPathValue(newManagementRequest(http.MethodGet, "/api/blocks/99/aggregations", nil), "block_id", "99")
	w := httptest.NewRecorder()
	h.ListBlockAggregations(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var aggs []models.TierAggregation
	if err := json.NewDecoder(w.Body).Decode(&aggs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(aggs) != 0 {
		t.Errorf("expected 0 aggregations, got %d", len(aggs))
	}
}
