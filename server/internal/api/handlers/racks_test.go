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

// fakeRackService is an in-memory stub implementing handlers.RackService.
type fakeRackService struct {
	rackTypes    map[int64]*models.RackTemplate
	racks        map[int64]*models.Rack
	devices      map[int64]*models.Device
	nextTypeID   int64
	nextRackID   int64
	nextDeviceID int64
}

func newFakeRackSvc() *fakeRackService {
	return &fakeRackService{
		rackTypes: make(map[int64]*models.RackTemplate),
		racks:     make(map[int64]*models.Rack),
		devices:   make(map[int64]*models.Device),
	}
}

func (s *fakeRackService) CreateRackType(name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: rack type name is required", models.ErrConstraintViolation)
	}
	if heightU <= 0 {
		return nil, fmt.Errorf("%w: height_u must be positive", models.ErrConstraintViolation)
	}
	s.nextTypeID++
	rt := &models.RackTemplate{ID: s.nextTypeID, Name: name, HeightU: heightU, PowerCapacityW: powerCapacityW}
	s.rackTypes[rt.ID] = rt
	return rt, nil
}

func (s *fakeRackService) ListRackTypes() ([]*models.RackTemplate, error) {
	out := make([]*models.RackTemplate, 0, len(s.rackTypes))
	for _, rt := range s.rackTypes {
		cp := *rt
		out = append(out, &cp)
	}
	return out, nil
}

func (s *fakeRackService) GetRackType(id int64) (*models.RackTemplate, error) {
	rt, ok := s.rackTypes[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *rt
	return &cp, nil
}

func (s *fakeRackService) UpdateRackType(id int64, name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	rt, ok := s.rackTypes[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name required", models.ErrConstraintViolation)
	}
	rt.Name = name
	rt.HeightU = heightU
	cp := *rt
	return &cp, nil
}

func (s *fakeRackService) DeleteRackType(id int64) error {
	if _, ok := s.rackTypes[id]; !ok {
		return models.ErrNotFound
	}
	// Simulate conflict: check if any rack uses it.
	for _, r := range s.racks {
		if r.RackTypeID != nil && *r.RackTypeID == id {
			return fmt.Errorf("%w: rack type is referenced", models.ErrConflict)
		}
	}
	delete(s.rackTypes, id)
	return nil
}

func (s *fakeRackService) CreateRack(name, description string, blockID, rackTypeID *int64, heightU, powerCapacityW int) (*models.Rack, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: rack name required", models.ErrConstraintViolation)
	}
	if heightU == 0 {
		heightU = 42
	}
	s.nextRackID++
	r := &models.Rack{ID: s.nextRackID, Name: name, HeightU: heightU, PowerCapacityW: powerCapacityW, BlockID: blockID, RackTypeID: rackTypeID}
	s.racks[r.ID] = r
	return r, nil
}

func (s *fakeRackService) ListRacks(blockID *int64) ([]*models.Rack, error) {
	out := make([]*models.Rack, 0)
	for _, r := range s.racks {
		if blockID != nil {
			if r.BlockID == nil || *r.BlockID != *blockID {
				continue
			}
		}
		cp := *r
		out = append(out, &cp)
	}
	return out, nil
}

func (s *fakeRackService) GetRackSummary(id int64) (*models.RackSummary, error) {
	r, ok := s.racks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *r
	return &models.RackSummary{Rack: cp, Devices: []*models.DeviceSummary{}}, nil
}

func (s *fakeRackService) UpdateRack(id int64, name, description string, blockID *int64) (*models.Rack, error) {
	r, ok := s.racks[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name required", models.ErrConstraintViolation)
	}
	r.Name = name
	cp := *r
	return &cp, nil
}

func (s *fakeRackService) DeleteRack(id int64) error {
	if _, ok := s.racks[id]; !ok {
		return models.ErrNotFound
	}
	delete(s.racks, id)
	return nil
}

func (s *fakeRackService) PlaceDevice(rackID, deviceModelID int64, name, description, role string, position int) (*models.PlaceDeviceResult, error) {
	if _, ok := s.racks[rackID]; !ok {
		return nil, models.ErrNotFound
	}
	s.nextDeviceID++
	d := &models.Device{ID: s.nextDeviceID, RackID: rackID, DeviceModelID: deviceModelID, Name: name, Position: position}
	s.devices[d.ID] = d
	return &models.PlaceDeviceResult{Device: d}, nil
}

func (s *fakeRackService) MoveDeviceInRack(rackID, deviceID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	d, ok := s.devices[deviceID]
	if !ok {
		return nil, models.ErrNotFound
	}
	d.Position = newPosition
	cp := *d
	return &models.PlaceDeviceResult{Device: &cp}, nil
}

func (s *fakeRackService) MoveDeviceCrossRack(srcRackID, deviceID, dstRackID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	d, ok := s.devices[deviceID]
	if !ok {
		return nil, models.ErrNotFound
	}
	d.RackID = dstRackID
	d.Position = newPosition
	cp := *d
	return &models.PlaceDeviceResult{Device: &cp}, nil
}

func (s *fakeRackService) RemoveDevice(rackID, deviceID int64, compact bool) error {
	if _, ok := s.devices[deviceID]; !ok {
		return models.ErrNotFound
	}
	delete(s.devices, deviceID)
	return nil
}

// --- Tests ---

func TestRackHandler_CreateRackType(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "valid", body: `{"name":"42U","height_u":42,"power_capacity_w":10000}`, wantStatus: http.StatusCreated},
		{name: "empty name", body: `{"name":"","height_u":42}`, wantStatus: http.StatusUnprocessableEntity},
		{name: "zero height", body: `{"name":"x","height_u":0}`, wantStatus: http.StatusUnprocessableEntity},
		{name: "invalid json", body: `{bad`, wantStatus: http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewRackHandler(newFakeRackSvc())
			req := httptest.NewRequest(http.MethodPost, "/api/rack-types", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			h.CreateRackType(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRackHandler_ListRackTypes(t *testing.T) {
	svc := newFakeRackSvc()
	svc.CreateRackType("type1", "", 42, 0)
	svc.CreateRackType("type2", "", 24, 0)

	h := handlers.NewRackHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/rack-types", nil)
	rec := httptest.NewRecorder()
	h.ListRackTypes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var rts []models.RackTemplate
	if err := json.NewDecoder(rec.Body).Decode(&rts); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rts) != 2 {
		t.Errorf("expected 2 rack types, got %d", len(rts))
	}
}

func TestRackHandler_GetRackType(t *testing.T) {
	svc := newFakeRackSvc()
	svc.CreateRackType("get-type", "", 42, 0)

	h := handlers.NewRackHandler(svc)

	t.Run("found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/rack-types/{id}", h.GetRackType)
		req := httptest.NewRequest(http.MethodGet, "/api/rack-types/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/rack-types/{id}", h.GetRackType)
		req := httptest.NewRequest(http.MethodGet, "/api/rack-types/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestRackHandler_DeleteRackType_Conflict(t *testing.T) {
	svc := newFakeRackSvc()
	svc.CreateRackType("type1", "", 42, 0)
	typeID := int64(1)
	// Create a rack referencing the type.
	svc.CreateRack("rack1", "", nil, &typeID, 42, 0)

	h := handlers.NewRackHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/rack-types/{id}", h.DeleteRackType)
	req := httptest.NewRequest(http.MethodDelete, "/api/rack-types/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409 conflict, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestRackHandler_CreateRack(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "valid", body: `{"name":"rack-01","height_u":42,"power_capacity_w":5000}`, wantStatus: http.StatusCreated},
		{name: "empty name", body: `{"name":"","height_u":42}`, wantStatus: http.StatusUnprocessableEntity},
		{name: "invalid json", body: `{bad`, wantStatus: http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewRackHandler(newFakeRackSvc())
			req := httptest.NewRequest(http.MethodPost, "/api/racks", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			h.CreateRack(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRackHandler_GetRack(t *testing.T) {
	svc := newFakeRackSvc()
	svc.CreateRack("rack-01", "", nil, nil, 42, 0)

	h := handlers.NewRackHandler(svc)

	t.Run("found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/racks/{id}", h.GetRack)
		req := httptest.NewRequest(http.MethodGet, "/api/racks/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/racks/{id}", h.GetRack)
		req := httptest.NewRequest(http.MethodGet, "/api/racks/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestRackHandler_PlaceDevice_ConstraintErrors(t *testing.T) {
	tests := []struct {
		name       string
		setupSvc   func(*fakeRackService)
		rackID     string
		body       string
		wantStatus int
	}{
		{
			name:       "rack not found",
			setupSvc:   func(s *fakeRackService) {},
			rackID:     "99",
			body:       `{"device_model_id":1,"name":"dev","position":1}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "place in existing rack",
			setupSvc: func(s *fakeRackService) {
				s.CreateRack("rack-1", "", nil, nil, 42, 0)
			},
			rackID:     "1",
			body:       `{"device_model_id":1,"name":"dev","position":1}`,
			wantStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newFakeRackSvc()
			tc.setupSvc(svc)
			h := handlers.NewRackHandler(svc)
			mux := http.NewServeMux()
			mux.HandleFunc("POST /api/racks/{id}/devices", h.PlaceDevice)
			req := httptest.NewRequest(http.MethodPost, "/api/racks/"+tc.rackID+"/devices", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRackHandler_PlaceDevice_ConstraintErrorMapping(t *testing.T) {
	// Test that RU overflow and position overlap return 400.
	errTests := []struct {
		name     string
		svcErr   error
		wantCode int
	}{
		{name: "ru overflow", svcErr: fmt.Errorf("%w: no space", models.ErrRUOverflow), wantCode: http.StatusBadRequest},
		{name: "position overlap", svcErr: fmt.Errorf("%w: overlap", models.ErrPositionOverlap), wantCode: http.StatusBadRequest},
	}

	for _, tc := range errTests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &errPlaceDeviceSvc{err: tc.svcErr}
			h := handlers.NewRackHandler(svc)
			mux := http.NewServeMux()
			mux.HandleFunc("POST /api/racks/{id}/devices", h.PlaceDevice)
			req := httptest.NewRequest(http.MethodPost, "/api/racks/1/devices", bytes.NewBufferString(`{"device_model_id":1,"name":"d","position":1}`))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantCode {
				t.Errorf("expected %d for %s, got %d", tc.wantCode, tc.name, rec.Code)
			}
		})
	}
}

func TestRackHandler_PlaceDevice_WarningInResponse(t *testing.T) {
	svc := &warningPlaceDeviceSvc{}
	h := handlers.NewRackHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/racks/{id}/devices", h.PlaceDevice)
	req := httptest.NewRequest(http.MethodPost, "/api/racks/1/devices", bytes.NewBufferString(`{"device_model_id":1,"name":"d","position":1}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var result models.PlaceDeviceResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Warning == "" {
		t.Error("expected warning in response, got empty")
	}
}

func TestRackHandler_RemoveDevice(t *testing.T) {
	svc := newFakeRackSvc()
	svc.CreateRack("rack-1", "", nil, nil, 42, 0)
	svc.PlaceDevice(1, 1, "dev", "", "leaf", 1)

	h := handlers.NewRackHandler(svc)

	t.Run("success no compact", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/racks/{rack_id}/devices/{device_id}", h.RemoveDevice)
		req := httptest.NewRequest(http.MethodDelete, "/api/racks/1/devices/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/racks/{rack_id}/devices/{device_id}", h.RemoveDevice)
		req := httptest.NewRequest(http.MethodDelete, "/api/racks/1/devices/9999", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestRackHandler_ListRacks_BlockIDFilter(t *testing.T) {
	svc := newFakeRackSvc()
	blockID := int64(5)
	svc.CreateRack("rack-in-block", "", &blockID, nil, 42, 0)
	svc.CreateRack("rack-standalone", "", nil, nil, 42, 0)

	h := handlers.NewRackHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/racks?block_id=5", nil)
	rec := httptest.NewRecorder()
	h.ListRacks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var racks []models.Rack
	if err := json.NewDecoder(rec.Body).Decode(&racks); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(racks) != 1 {
		t.Errorf("expected 1 rack with block_id=5, got %d", len(racks))
	}
}

// --- Helper stubs for error-specific tests ---

type errPlaceDeviceSvc struct {
	*fakeRackService
	err error
}

func (s *errPlaceDeviceSvc) CreateRackType(name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) ListRackTypes() ([]*models.RackTemplate, error) { return nil, nil }
func (s *errPlaceDeviceSvc) GetRackType(id int64) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) UpdateRackType(id int64, name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) DeleteRackType(id int64) error { return nil }
func (s *errPlaceDeviceSvc) CreateRack(name, description string, blockID, rackTypeID *int64, heightU, powerCapacityW int) (*models.Rack, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) ListRacks(blockID *int64) ([]*models.Rack, error) { return nil, nil }
func (s *errPlaceDeviceSvc) GetRackSummary(id int64) (*models.RackSummary, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) UpdateRack(id int64, name, description string, blockID *int64) (*models.Rack, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) DeleteRack(id int64) error { return nil }
func (s *errPlaceDeviceSvc) PlaceDevice(rackID, deviceModelID int64, name, description, role string, position int) (*models.PlaceDeviceResult, error) {
	return nil, s.err
}
func (s *errPlaceDeviceSvc) MoveDeviceInRack(rackID, deviceID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) MoveDeviceCrossRack(srcRackID, deviceID, dstRackID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	return nil, nil
}
func (s *errPlaceDeviceSvc) RemoveDevice(rackID, deviceID int64, compact bool) error {
	return errors.New("not implemented")
}

type warningPlaceDeviceSvc struct{}

func (s *warningPlaceDeviceSvc) CreateRackType(name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) ListRackTypes() ([]*models.RackTemplate, error)  { return nil, nil }
func (s *warningPlaceDeviceSvc) GetRackType(id int64) (*models.RackTemplate, error) { return nil, nil }
func (s *warningPlaceDeviceSvc) UpdateRackType(id int64, name, description string, heightU, powerCapacityW int) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) DeleteRackType(id int64) error { return nil }
func (s *warningPlaceDeviceSvc) CreateRack(name, description string, blockID, rackTypeID *int64, heightU, powerCapacityW int) (*models.Rack, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) ListRacks(blockID *int64) ([]*models.Rack, error) { return nil, nil }
func (s *warningPlaceDeviceSvc) GetRackSummary(id int64) (*models.RackSummary, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) UpdateRack(id int64, name, description string, blockID *int64) (*models.Rack, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) DeleteRack(id int64) error { return nil }
func (s *warningPlaceDeviceSvc) PlaceDevice(rackID, deviceModelID int64, name, description, role string, position int) (*models.PlaceDeviceResult, error) {
	d := &models.Device{ID: 1, RackID: rackID, DeviceModelID: deviceModelID, Name: name, Position: position}
	return &models.PlaceDeviceResult{Device: d, Warning: "power capacity exceeded: 900W used + 500W new = 1400W > 1000W capacity"}, nil
}
func (s *warningPlaceDeviceSvc) MoveDeviceInRack(rackID, deviceID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) MoveDeviceCrossRack(srcRackID, deviceID, dstRackID int64, newPosition int) (*models.PlaceDeviceResult, error) {
	return nil, nil
}
func (s *warningPlaceDeviceSvc) RemoveDevice(rackID, deviceID int64, compact bool) error {
	return nil
}
