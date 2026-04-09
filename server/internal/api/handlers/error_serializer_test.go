package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// errorBody is the JSON shape we assert against.
type errorBody struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Refs    []models.ErrRef `json:"refs"`
	Error   string          `json:"error"`
}

// parseErrorBodyFromRecorder decodes the response body into errorBody.
func parseErrorBodyFromRecorder(t *testing.T, rec *httptest.ResponseRecorder) errorBody {
	t.Helper()
	var body errorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body
}

// alwaysErrDesignSvc is a minimal DesignService that always returns a specific error.
type alwaysErrDesignSvc struct {
	err error
}

func (s *alwaysErrDesignSvc) CreateDesign(_, _ string) (*models.Design, error) { return nil, s.err }
func (s *alwaysErrDesignSvc) ListDesigns() ([]*models.Design, error)            { return nil, nil }
func (s *alwaysErrDesignSvc) GetDesign(_ int64) (*models.Design, error)         { return nil, s.err }
func (s *alwaysErrDesignSvc) DeleteDesign(_ int64) error                        { return s.err }
func (s *alwaysErrDesignSvc) GetScaffold(_ int64) (*store.DesignScaffold, error) {
	return nil, s.err
}

// doGetDesign issues GET /api/designs/1 through the DesignHandler and returns the response.
func doGetDesign(t *testing.T, svc handlers.DesignService) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewDesignHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/designs/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/designs/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// doCreateDesign issues POST /api/designs through the DesignHandler.
func doCreateDesign(t *testing.T, svc handlers.DesignService) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewDesignHandler(svc)
	body := bytes.NewBufferString(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/designs", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	return rec
}

func TestWriteDomainError_NotFound(t *testing.T) {
	rec := doGetDesign(t, &alwaysErrDesignSvc{err: models.ErrNotFound})
	body := parseErrorBodyFromRecorder(t, rec)
	if body.Code != "not_found" {
		t.Errorf("code: got %q, want %q", body.Code, "not_found")
	}
	if body.Message == "" {
		t.Error("message should not be empty")
	}
	if body.Error != body.Message {
		t.Errorf("Error field %q != message field %q", body.Error, body.Message)
	}
}

func TestWriteDomainError_ConstraintViolation(t *testing.T) {
	rec := doCreateDesign(t, &alwaysErrDesignSvc{err: models.ErrConstraintViolation})
	body := parseErrorBodyFromRecorder(t, rec)
	if body.Code != "constraint_violation" {
		t.Errorf("code: got %q, want %q", body.Code, "constraint_violation")
	}
}

// alwaysErrDeviceModelSvc implements DeviceModelService returning a fixed error.
type alwaysErrDeviceModelSvc struct {
	err error
}

func (s *alwaysErrDeviceModelSvc) CreateDeviceModel(*models.DeviceModel) (*models.DeviceModel, error) {
	return nil, s.err
}
func (s *alwaysErrDeviceModelSvc) ListDeviceModels(bool) ([]*models.DeviceModel, error) {
	return nil, nil
}
func (s *alwaysErrDeviceModelSvc) GetDeviceModel(_ int64) (*models.DeviceModel, error) {
	return nil, s.err
}
func (s *alwaysErrDeviceModelSvc) UpdateDeviceModel(*models.DeviceModel) (*models.DeviceModel, error) {
	return nil, s.err
}
func (s *alwaysErrDeviceModelSvc) ArchiveDeviceModel(_ int64) error { return s.err }
func (s *alwaysErrDeviceModelSvc) DuplicateDeviceModel(_ int64) (*models.DeviceModel, error) {
	return nil, s.err
}

// doCreateDeviceModel issues POST /api/catalog/devices through DeviceModelHandler.
func doCreateDeviceModel(t *testing.T, svc *alwaysErrDeviceModelSvc) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewDeviceModelHandler(svc)
	body := bytes.NewBufferString(`{"vendor":"A","model":"B","device_model_type":"network","height_u":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/catalog/devices", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	return rec
}

// alwaysErrRackSvc is a minimal RackService that returns a fixed error for PlaceDevice and DeleteRackType.
type alwaysErrRackSvc struct {
	placeErr  error
	deleteErr error
}

func (s *alwaysErrRackSvc) CreateRackType(_, _ string, _, _, _, _ int) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *alwaysErrRackSvc) ListRackTypes() ([]*models.RackTemplate, error)  { return nil, nil }
func (s *alwaysErrRackSvc) GetRackType(_ int64) (*models.RackTemplate, error) { return nil, nil }
func (s *alwaysErrRackSvc) UpdateRackType(_ int64, _, _ string, _, _, _, _ int) (*models.RackTemplate, error) {
	return nil, nil
}
func (s *alwaysErrRackSvc) DeleteRackType(_ int64) error { return s.deleteErr }
func (s *alwaysErrRackSvc) CreateRack(_, _ string, _ *int64, _ *int64, _, _ int, _ models.RackRole) (*models.Rack, error) {
	return nil, nil
}
func (s *alwaysErrRackSvc) ListRacks(_ *int64) ([]*models.Rack, error) { return nil, nil }
func (s *alwaysErrRackSvc) GetRackSummary(_ int64) (*models.RackSummary, error) { return nil, nil }
func (s *alwaysErrRackSvc) UpdateRack(_ int64, _, _ string, _ *int64) (*models.Rack, error) {
	return nil, nil
}
func (s *alwaysErrRackSvc) DeleteRack(_ int64) error { return nil }
func (s *alwaysErrRackSvc) PlaceDevice(_, _ int64, _, _, _ string, _ int) (*models.PlaceDeviceResult, error) {
	return nil, s.placeErr
}
func (s *alwaysErrRackSvc) MoveDeviceInRack(_, _ int64, _ int) (*models.PlaceDeviceResult, error) {
	return nil, nil
}
func (s *alwaysErrRackSvc) MoveDeviceCrossRack(_, _, _ int64, _ int) (*models.PlaceDeviceResult, error) {
	return nil, nil
}
func (s *alwaysErrRackSvc) RemoveDevice(_, _ int64, _ bool) error   { return nil }
func (s *alwaysErrRackSvc) PlaceServerDevices(_, _ int64, _ int) error { return nil }

// doPlaceDevice issues POST /api/racks/1/devices through RackHandler.
func doPlaceDevice(t *testing.T, svc handlers.RackService) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewRackHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/racks/{id}/devices", h.PlaceDevice)
	body := bytes.NewBufferString(`{"device_model_id":1,"position":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/racks/1/devices", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// doDeleteRackType issues DELETE /api/rack-types/1 through RackHandler.
func doDeleteRackType(t *testing.T, svc handlers.RackService) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewRackHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/rack-types/{id}", h.DeleteRackType)
	req := httptest.NewRequest(http.MethodDelete, "/api/rack-types/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// alwaysErrBlockSvc is a minimal BlockService that returns fixed errors.
type alwaysErrBlockSvc struct {
	assignErr  error
	addRackErr error
}

func (s *alwaysErrBlockSvc) CreateBlock(_ int64, _, _ string, _ *int64, _ *int64, _ int) (*models.CreateBlockResult, error) {
	return nil, nil
}
func (s *alwaysErrBlockSvc) GetBlock(_ int64) (*models.Block, error) { return nil, nil }
func (s *alwaysErrBlockSvc) ListBlocks(_ int64) ([]*models.Block, error) { return nil, nil }
func (s *alwaysErrBlockSvc) AssignAggregation(_ int64, _ models.NetworkPlane, _ int64, _ int, _ int) (*models.TierAggregationSummary, error) {
	return nil, s.assignErr
}
func (s *alwaysErrBlockSvc) GetAggregationSummary(_ int64, _ models.NetworkPlane) (*models.TierAggregationSummary, error) {
	return nil, nil
}
func (s *alwaysErrBlockSvc) ListAggregationSummaries(_ int64) ([]*models.TierAggregationSummary, error) {
	return nil, nil
}
func (s *alwaysErrBlockSvc) DeleteAggregation(_ int64, _ models.NetworkPlane) error { return nil }
func (s *alwaysErrBlockSvc) AssignSuperBlockAggregation(_ int64, _ models.NetworkPlane, _ int64, _ int, _ int) (*models.TierAggregationSummary, error) {
	return nil, nil
}
func (s *alwaysErrBlockSvc) GetSuperBlockAggregationSummary(_ int64, _ models.NetworkPlane) (*models.TierAggregationSummary, error) {
	return nil, nil
}
func (s *alwaysErrBlockSvc) AddRackToBlock(_ int64, _ *int64, _ int64) (*models.AddRackToBlockResult, error) {
	return nil, s.addRackErr
}
func (s *alwaysErrBlockSvc) RemoveRackFromBlock(_ int64) error { return nil }
func (s *alwaysErrBlockSvc) ListPortConnections(_ int64, _ models.NetworkPlane) ([]*models.TierPortConnection, error) {
	return nil, nil
}
func (s *alwaysErrBlockSvc) PlaceSpineDevices(_, _ int64, _ int) error         { return nil }
func (s *alwaysErrBlockSvc) DeleteBlock(_ int64) error                         { return nil }
func (s *alwaysErrBlockSvc) ReparentBlock(_, _ int64) (*models.Block, error)   { return nil, nil }

// doAssignAggregation issues PUT /api/blocks/1/aggregations/front_end through BlockHandler.
func doAssignAggregation(t *testing.T, svc handlers.BlockService) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewBlockHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/blocks/{id}/aggregations/{plane}", h.AssignAggregation)
	body := bytes.NewBufferString(`{"device_model_id":1,"spine_count":2}`)
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/1/aggregations/front_end", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// doAddRackToBlock issues POST /api/blocks/rack-assignments through BlockHandler.
func doAddRackToBlock(t *testing.T, svc handlers.BlockService) *httptest.ResponseRecorder {
	t.Helper()
	h := handlers.NewBlockHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/blocks/rack-assignments", h.AddRackToBlock)
	body := bytes.NewBufferString(fmt.Sprintf(`{"rack_id":1,"super_block_id":1}`))
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/rack-assignments", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestWriteDomainError_DomainErrorCodes(t *testing.T) {
	// Table of (domain error, handler exercise fn, expected code).
	tests := []struct {
		name     string
		err      *models.DomainError
		wantCode string
		exercise func(t *testing.T) *httptest.ResponseRecorder
	}{
		{
			name:     "not_found via Get",
			err:      models.ErrNotFound,
			wantCode: "not_found",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doGetDesign(t, &alwaysErrDesignSvc{err: models.ErrNotFound})
			},
		},
		{
			name:     "constraint_violation via Create",
			err:      models.ErrConstraintViolation,
			wantCode: "constraint_violation",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doCreateDesign(t, &alwaysErrDesignSvc{err: models.ErrConstraintViolation})
			},
		},
		{
			name:     "duplicate via CreateDeviceModel",
			err:      models.ErrDuplicate,
			wantCode: "duplicate",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doCreateDeviceModel(t, &alwaysErrDeviceModelSvc{err: models.ErrDuplicate})
			},
		},
		{
			name:     "seed_read_only via UpdateDeviceModel",
			err:      models.ErrSeedReadOnly,
			wantCode: "seed_read_only",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				svc := &alwaysErrDeviceModelSvc{err: models.ErrSeedReadOnly}
				h := handlers.NewDeviceModelHandler(svc)
				mux := http.NewServeMux()
				mux.HandleFunc("PUT /api/catalog/devices/{id}", h.Update)
				body := bytes.NewBufferString(`{"vendor":"A","model":"B","device_model_type":"network","height_u":1}`)
				req := httptest.NewRequest(http.MethodPut, "/api/catalog/devices/1", body)
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)
				return rec
			},
		},
		{
			name:     "ru_overflow via PlaceDevice",
			err:      models.ErrRUOverflow,
			wantCode: "ru_overflow",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doPlaceDevice(t, &alwaysErrRackSvc{placeErr: models.ErrRUOverflow})
			},
		},
		{
			name:     "position_overlap via PlaceDevice",
			err:      models.ErrPositionOverlap,
			wantCode: "position_overlap",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doPlaceDevice(t, &alwaysErrRackSvc{placeErr: models.ErrPositionOverlap})
			},
		},
		{
			name:     "conflict via DeleteRackType",
			err:      models.ErrConflict,
			wantCode: "conflict",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doDeleteRackType(t, &alwaysErrRackSvc{deleteErr: models.ErrConflict})
			},
		},
		{
			name:     "agg_ports_full via AddRackToBlock",
			err:      models.ErrAggPortsFull,
			wantCode: "agg_ports_full",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doAddRackToBlock(t, &alwaysErrBlockSvc{addRackErr: models.ErrAggPortsFull})
			},
		},
		{
			name:     "agg_model_downsize via AssignAggregation",
			err:      models.ErrAggModelDownsize,
			wantCode: "agg_model_downsize",
			exercise: func(t *testing.T) *httptest.ResponseRecorder {
				return doAssignAggregation(t, &alwaysErrBlockSvc{assignErr: models.ErrAggModelDownsize})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := tc.exercise(t)
			body := parseErrorBodyFromRecorder(t, rec)
			if body.Code != tc.wantCode {
				t.Errorf("code: got %q, want %q (status=%d)", body.Code, tc.wantCode, rec.Code)
			}
			if body.Message == "" {
				t.Error("message should not be empty")
			}
			// Error field must equal message for backwards compat.
			if body.Error != body.Message {
				t.Errorf("Error field %q != message field %q", body.Error, body.Message)
			}
		})
	}
}

func TestWriteError_UnknownCode(t *testing.T) {
	// writeError (plain string path) should emit code="unknown".
	h := handlers.NewDesignHandler(newFakeSvc())
	req := httptest.NewRequest(http.MethodGet, "/api/designs/not-a-number", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req) // parseID will fail → writeError("invalid id")

	body := parseErrorBodyFromRecorder(t, rec)
	if body.Code != "unknown" {
		t.Errorf("code: got %q, want %q", body.Code, "unknown")
	}
}

func TestErrorResponse_BackwardsCompat(t *testing.T) {
	// Every error response must include legacy `error`, plus new `code` + `message`.
	rec := doGetDesign(t, &alwaysErrDesignSvc{err: models.ErrNotFound})

	var raw map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"error", "code", "message"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("key %q missing from error response body", key)
		}
	}
}
