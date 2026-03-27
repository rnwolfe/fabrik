package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
)

// fakeMetricsSvc implements handlers.MetricsService for testing.
type fakeMetricsSvc struct {
	metrics map[int64]*models.DesignMetrics
}

func newFakeMetricsSvc() *fakeMetricsSvc {
	return &fakeMetricsSvc{
		metrics: make(map[int64]*models.DesignMetrics),
	}
}

func (f *fakeMetricsSvc) GetDesignMetrics(id int64) (*models.DesignMetrics, error) {
	m, ok := f.metrics[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	return m, nil
}

func TestMetricsHandler_GetDesignMetrics_ok(t *testing.T) {
	svc := newFakeMetricsSvc()
	svc.metrics[1] = &models.DesignMetrics{
		DesignID:      1,
		TotalHosts:    24,
		TotalSwitches: 9,
		Fabrics:       []models.FabricMetricEntry{},
		Power:         models.PowerMetrics{TotalCapacityW: 5000, TotalDrawW: 2500, UtilizationPct: 50},
		Capacity:      models.ResourceCapacity{TotalVCPU: 64},
		PortUtilization: []models.PortUtilizationEntry{},
	}

	h := handlers.NewMetricsHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/designs/1/metrics", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDesignMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var m models.DesignMetrics
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if m.DesignID != 1 {
		t.Errorf("expected design_id=1, got %d", m.DesignID)
	}
	if m.TotalHosts != 24 {
		t.Errorf("expected total_hosts=24, got %d", m.TotalHosts)
	}
	if m.Power.UtilizationPct != 50 {
		t.Errorf("expected power utilization=50, got %f", m.Power.UtilizationPct)
	}
}

func TestMetricsHandler_GetDesignMetrics_notFound(t *testing.T) {
	svc := newFakeMetricsSvc()
	h := handlers.NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/designs/99/metrics", nil)
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()

	h.GetDesignMetrics(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestMetricsHandler_GetDesignMetrics_badID(t *testing.T) {
	svc := newFakeMetricsSvc()
	h := handlers.NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/designs/abc/metrics", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()

	h.GetDesignMetrics(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
