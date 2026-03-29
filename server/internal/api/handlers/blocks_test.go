package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
)

// --- fakeBlockService ---

type fakeBlockService struct {
	blocks map[int64]*models.Block
	aggs   map[string]*models.TierAggregationSummary // key: "blockID:plane"
	conns  map[string][]*models.TierPortConnection
	nextID int64
}

func newFakeBlockSvc() *fakeBlockService {
	return &fakeBlockService{
		blocks: make(map[int64]*models.Block),
		aggs:   make(map[string]*models.TierAggregationSummary),
		conns:  make(map[string][]*models.TierPortConnection),
	}
}

func (s *fakeBlockService) aggKey(blockID int64, plane models.NetworkPlane) string {
	return fmt.Sprintf("%d:%s", blockID, plane)
}

func (s *fakeBlockService) CreateBlock(superBlockID int64, name, description string, leafModelID, spineModelID *int64, spineCount int) (*models.CreateBlockResult, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: block name is required", models.ErrConstraintViolation)
	}
	s.nextID++
	b := &models.Block{ID: s.nextID, SuperBlockID: superBlockID, Name: name, Description: description}
	s.blocks[b.ID] = b
	return &models.CreateBlockResult{Block: b}, nil
}

func (s *fakeBlockService) GetBlock(id int64) (*models.Block, error) {
	b, ok := s.blocks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (s *fakeBlockService) ListBlocks(superBlockID int64) ([]*models.Block, error) {
	var out []*models.Block
	for _, b := range s.blocks {
		if b.SuperBlockID == superBlockID {
			cp := *b
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *fakeBlockService) AssignAggregation(blockID int64, plane models.NetworkPlane, deviceModelID int64, spineCount int) (*models.TierAggregationSummary, error) {
	if _, ok := s.blocks[blockID]; !ok {
		return nil, models.ErrNotFound
	}
	if deviceModelID == 999 {
		return nil, fmt.Errorf("%w: downsize not allowed", models.ErrAggModelDownsize)
	}
	summary := &models.TierAggregationSummary{
		TierAggregation: models.TierAggregation{
			ScopeType:     models.ScopeBlock,
			ScopeID:       blockID,
			Plane:         plane,
			DeviceModelID: deviceModelID,
			SpineCount:    spineCount,
		},
		TotalPorts:     32,
		AvailablePorts: 32,
	}
	s.aggs[s.aggKey(blockID, plane)] = summary
	return summary, nil
}

func (s *fakeBlockService) GetAggregationSummary(blockID int64, plane models.NetworkPlane) (*models.TierAggregationSummary, error) {
	sum, ok := s.aggs[s.aggKey(blockID, plane)]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *sum
	return &cp, nil
}

func (s *fakeBlockService) ListAggregationSummaries(blockID int64) ([]*models.TierAggregationSummary, error) {
	var out []*models.TierAggregationSummary
	for _, sum := range s.aggs {
		if sum.ScopeID == blockID {
			cp := *sum
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *fakeBlockService) DeleteAggregation(blockID int64, plane models.NetworkPlane) error {
	key := s.aggKey(blockID, plane)
	if _, ok := s.aggs[key]; !ok {
		return models.ErrNotFound
	}
	delete(s.aggs, key)
	return nil
}

func (s *fakeBlockService) AddRackToBlock(rackID int64, blockID *int64, superBlockID int64) (*models.AddRackToBlockResult, error) {
	if rackID == 999 {
		return nil, fmt.Errorf("%w: agg is full", models.ErrAggPortsFull)
	}
	if blockID == nil && superBlockID <= 0 {
		return nil, fmt.Errorf("%w: block_id or super_block_id required", models.ErrConstraintViolation)
	}
	var bid int64 = 1
	if blockID != nil {
		bid = *blockID
	}
	return &models.AddRackToBlockResult{
		Rack:        &models.Rack{ID: rackID, BlockID: &bid},
		Connections: []*models.TierPortConnection{},
	}, nil
}

func (s *fakeBlockService) RemoveRackFromBlock(rackID int64) error {
	if rackID == 999 {
		return models.ErrNotFound
	}
	return nil
}

func (s *fakeBlockService) ListPortConnections(blockID int64, plane models.NetworkPlane) ([]*models.TierPortConnection, error) {
	key := s.aggKey(blockID, plane)
	if _, ok := s.aggs[key]; !ok {
		return nil, models.ErrNotFound
	}
	return s.conns[key], nil
}

// --- helpers ---

func blockRequest(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}
	r := httptest.NewRequest(method, url, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func blockResponse(t *testing.T, h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// --- Tests ---

func TestBlockHandler_CreateBlock(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)

	t.Run("valid request", func(t *testing.T) {
		body := map[string]any{"super_block_id": 1, "name": "row-A"}
		r := blockRequest(t, "POST", "/api/blocks", body)
		w := blockResponse(t, http.HandlerFunc(h.CreateBlock), r)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var result models.CreateBlockResult
		json.NewDecoder(w.Body).Decode(&result)
		if result.Block.Name != "row-A" {
			t.Errorf("expected name 'row-A', got %q", result.Block.Name)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		body := map[string]any{"super_block_id": 1, "name": ""}
		r := blockRequest(t, "POST", "/api/blocks", body)
		w := blockResponse(t, http.HandlerFunc(h.CreateBlock), r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d", w.Code)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/api/blocks", bytes.NewReader([]byte("not json")))
		w := blockResponse(t, http.HandlerFunc(h.CreateBlock), r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestBlockHandler_GetBlock(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(1, "row-A", "", nil, nil, 0)

	t.Run("found", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/1", nil)
		r.SetPathValue("id", "1")
		w := blockResponse(t, http.HandlerFunc(h.GetBlock), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/9999", nil)
		r.SetPathValue("id", "9999")
		w := blockResponse(t, http.HandlerFunc(h.GetBlock), r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/abc", nil)
		r.SetPathValue("id", "abc")
		w := blockResponse(t, http.HandlerFunc(h.GetBlock), r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestBlockHandler_ListBlocks(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(5, "row-A", "", nil, nil, 0)
	svc.CreateBlock(5, "row-B", "", nil, nil, 0)

	t.Run("valid super_block_id", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks?super_block_id=5", nil)
		w := blockResponse(t, http.HandlerFunc(h.ListBlocks), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		var blocks []*models.Block
		json.NewDecoder(w.Body).Decode(&blocks)
		if len(blocks) != 2 {
			t.Errorf("expected 2 blocks, got %d", len(blocks))
		}
	})

	t.Run("missing super_block_id", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks", nil)
		w := blockResponse(t, http.HandlerFunc(h.ListBlocks), r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestBlockHandler_AssignAggregation(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(1, "row-A", "", nil, nil, 0)

	t.Run("valid assignment", func(t *testing.T) {
		body := map[string]any{"device_model_id": 10, "spine_count": 4}
		r := blockRequest(t, "PUT", "/api/blocks/1/aggregations/front_end", body)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "front_end")
		w := blockResponse(t, http.HandlerFunc(h.AssignAggregation), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid plane", func(t *testing.T) {
		body := map[string]any{"device_model_id": 10}
		r := blockRequest(t, "PUT", "/api/blocks/1/aggregations/badplane", body)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "badplane")
		w := blockResponse(t, http.HandlerFunc(h.AssignAggregation), r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("downsize rejected", func(t *testing.T) {
		body := map[string]any{"device_model_id": 999}
		r := blockRequest(t, "PUT", "/api/blocks/1/aggregations/front_end", body)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "front_end")
		w := blockResponse(t, http.HandlerFunc(h.AssignAggregation), r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("block not found", func(t *testing.T) {
		body := map[string]any{"device_model_id": 10}
		r := blockRequest(t, "PUT", "/api/blocks/9999/aggregations/front_end", body)
		r.SetPathValue("id", "9999")
		r.SetPathValue("plane", "front_end")
		w := blockResponse(t, http.HandlerFunc(h.AssignAggregation), r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestBlockHandler_GetAggregation(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(1, "row-A", "", nil, nil, 0)
	svc.AssignAggregation(1, models.NetworkPlaneFrontEnd, 10, 0)

	t.Run("found", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/1/aggregations/front_end", nil)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "front_end")
		w := blockResponse(t, http.HandlerFunc(h.GetAggregation), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/1/aggregations/management", nil)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "management")
		w := blockResponse(t, http.HandlerFunc(h.GetAggregation), r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestBlockHandler_DeleteAggregation(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(1, "row-A", "", nil, nil, 0)
	svc.AssignAggregation(1, models.NetworkPlaneFrontEnd, 10, 0)

	t.Run("deletes successfully", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/blocks/1/aggregations/front_end", nil)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "front_end")
		w := blockResponse(t, http.HandlerFunc(h.DeleteAggregation), r)
		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/blocks/1/aggregations/management", nil)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "management")
		w := blockResponse(t, http.HandlerFunc(h.DeleteAggregation), r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestBlockHandler_AddRackToBlock(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(1, "row-A", "", nil, nil, 0)

	t.Run("valid placement with block_id", func(t *testing.T) {
		body := map[string]any{"rack_id": 5, "block_id": 1}
		r := blockRequest(t, "POST", "/api/blocks/add-rack", body)
		w := blockResponse(t, http.HandlerFunc(h.AddRackToBlock), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("agg ports full", func(t *testing.T) {
		body := map[string]any{"rack_id": 999, "block_id": 1}
		r := blockRequest(t, "POST", "/api/blocks/add-rack", body)
		w := blockResponse(t, http.HandlerFunc(h.AddRackToBlock), r)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing rack_id", func(t *testing.T) {
		body := map[string]any{"block_id": 1}
		r := blockRequest(t, "POST", "/api/blocks/add-rack", body)
		w := blockResponse(t, http.HandlerFunc(h.AddRackToBlock), r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("default block auto-creation with super_block_id", func(t *testing.T) {
		body := map[string]any{"rack_id": 5, "super_block_id": 10}
		r := blockRequest(t, "POST", "/api/blocks/add-rack", body)
		w := blockResponse(t, http.HandlerFunc(h.AddRackToBlock), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestBlockHandler_RemoveRackFromBlock(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)

	t.Run("success", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/blocks/racks/5", nil)
		r.SetPathValue("rack_id", "5")
		w := blockResponse(t, http.HandlerFunc(h.RemoveRackFromBlock), r)
		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/blocks/racks/999", nil)
		r.SetPathValue("rack_id", "999")
		w := blockResponse(t, http.HandlerFunc(h.RemoveRackFromBlock), r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestBlockHandler_ListPortConnections(t *testing.T) {
	svc := newFakeBlockSvc()
	h := handlers.NewBlockHandler(svc)
	svc.CreateBlock(1, "row-A", "", nil, nil, 0)
	svc.AssignAggregation(1, models.NetworkPlaneFrontEnd, 10, 0)

	t.Run("returns empty list", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/1/aggregations/front_end/connections", nil)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "front_end")
		w := blockResponse(t, http.HandlerFunc(h.ListPortConnections), r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		var conns []*models.TierPortConnection
		json.NewDecoder(w.Body).Decode(&conns)
		if conns == nil {
			t.Error("expected non-nil connections slice")
		}
	})

	t.Run("agg not found", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/blocks/1/aggregations/management/connections", nil)
		r.SetPathValue("id", "1")
		r.SetPathValue("plane", "management")
		w := blockResponse(t, http.HandlerFunc(h.ListPortConnections), r)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

// Verify fakeBlockService satisfies the handlers.BlockService interface.
var _ handlers.BlockService = (*fakeBlockService)(nil)

// Silence "declared and not used" for errors import.
var _ = errors.New
