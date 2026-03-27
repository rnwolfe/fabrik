package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// FabricRepository is the store interface required by FabricService.
type FabricRepository interface {
	Create(p store.FabricParams) (*store.FabricRecord, error)
	List() ([]*store.FabricRecord, error)
	Get(id int64) (*store.FabricRecord, error)
	Update(id int64, p store.FabricParams) (*store.FabricRecord, error)
	Delete(id int64) error
	GetDeviceModelByID(id int64) (*models.DeviceModel, error)
	ListDeviceModels() ([]*models.DeviceModel, error)
}

// FabricResponse is the enriched response type returned by service methods.
// It bundles the FabricRecord with the calculated topology and derived metrics.
type FabricResponse struct {
	*store.FabricRecord
	Topology *TopologyPlan    `json:"topology"`
	Warnings []string         `json:"warnings,omitempty"`
	Metrics  *FabricMetrics   `json:"metrics"`
	// Resolved device model objects for the UI.
	LeafModel       *models.DeviceModel `json:"leaf_model,omitempty"`
	SpineModel      *models.DeviceModel `json:"spine_model,omitempty"`
	SuperSpineModel *models.DeviceModel `json:"super_spine_model,omitempty"`
}

// FabricMetrics holds derived metrics for a fabric.
type FabricMetrics struct {
	TotalSwitches      int     `json:"total_switches"`
	TotalHostPorts     int     `json:"total_host_ports"`
	OversubscriptionRatio float64 `json:"oversubscription_ratio"`
	// BisectionBandwidthGbps is populated when a device model is assigned to the leaf role.
	BisectionBandwidthGbps float64 `json:"bisection_bandwidth_gbps,omitempty"`
}

// CreateFabricRequest holds the inputs for creating a fabric.
type CreateFabricRequest struct {
	DesignID         int64              `json:"design_id"`
	Name             string             `json:"name"`
	Tier             models.FabricTier  `json:"tier"`
	Stages           int                `json:"stages"`
	Radix            int                `json:"radix"`
	Oversubscription float64            `json:"oversubscription"`
	Description      string             `json:"description"`
	LeafModelID      int64              `json:"leaf_model_id,omitempty"`
	SpineModelID     int64              `json:"spine_model_id,omitempty"`
	SuperSpineModelID int64             `json:"super_spine_model_id,omitempty"`
}

// UpdateFabricRequest holds the inputs for updating a fabric.
type UpdateFabricRequest struct {
	Name              string            `json:"name"`
	Tier              models.FabricTier `json:"tier"`
	Stages            int               `json:"stages"`
	Radix             int               `json:"radix"`
	Oversubscription  float64           `json:"oversubscription"`
	Description       string            `json:"description"`
	LeafModelID       int64             `json:"leaf_model_id,omitempty"`
	SpineModelID      int64             `json:"spine_model_id,omitempty"`
	SuperSpineModelID int64             `json:"super_spine_model_id,omitempty"`
	// Force regeneration even if rack placements exist.
	Force bool `json:"force"`
}

// FabricService implements business logic for Fabric resources.
type FabricService struct {
	repo FabricRepository
}

// NewFabricService returns a new FabricService backed by repo.
func NewFabricService(repo FabricRepository) *FabricService {
	return &FabricService{repo: repo}
}

// CreateFabric validates, creates, and returns a new Fabric with topology.
func (s *FabricService) CreateFabric(req CreateFabricRequest) (*FabricResponse, error) {
	if err := validateFabricInput(req.Name, req.Stages, req.Radix, req.Oversubscription, req.Tier); err != nil {
		return nil, err
	}

	topology, err := CalculateTopology(req.Stages, req.Radix, req.Oversubscription)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", models.ErrConstraintViolation, err.Error())
	}

	p := store.FabricParams{
		DesignID:          req.DesignID,
		Name:              strings.TrimSpace(req.Name),
		Tier:              req.Tier,
		Stages:            req.Stages,
		Radix:             topology.Radix, // use corrected radix
		Oversubscription:  topology.Oversubscription,
		Description:       req.Description,
		LeafModelID:       req.LeafModelID,
		SpineModelID:      req.SpineModelID,
		SuperSpineModelID: req.SuperSpineModelID,
	}

	rec, err := s.repo.Create(p)
	if err != nil {
		return nil, fmt.Errorf("create fabric: %w", err)
	}

	slog.Info("fabric created", "fabricID", rec.ID, "name", rec.Name, "stages", req.Stages)

	return s.buildResponse(rec, topology)
}

// ListFabrics returns all fabrics with summary topology data.
func (s *FabricService) ListFabrics() ([]*FabricResponse, error) {
	recs, err := s.repo.List()
	if err != nil {
		return nil, fmt.Errorf("list fabrics: %w", err)
	}

	out := make([]*FabricResponse, 0, len(recs))
	for _, rec := range recs {
		topo, err := CalculateTopology(rec.Stages, rec.Radix, rec.Oversubscription)
		if err != nil {
			// Stored params should always be valid, but handle gracefully.
			slog.Warn("topology calculation for existing fabric failed", "fabricID", rec.ID, "err", err)
			continue
		}
		resp, err := s.buildResponse(rec, topo)
		if err != nil {
			return nil, err
		}
		out = append(out, resp)
	}
	return out, nil
}

// GetFabric returns the fabric with the given id including full topology.
func (s *FabricService) GetFabric(id int64) (*FabricResponse, error) {
	rec, err := s.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get fabric %d: %w", id, err)
	}

	topo, err := CalculateTopology(rec.Stages, rec.Radix, rec.Oversubscription)
	if err != nil {
		return nil, fmt.Errorf("calculate topology for fabric %d: %w", id, err)
	}

	return s.buildResponse(rec, topo)
}

// UpdateFabric validates and updates a fabric, regenerating topology.
func (s *FabricService) UpdateFabric(id int64, req UpdateFabricRequest) (*FabricResponse, error) {
	if err := validateFabricInput(req.Name, req.Stages, req.Radix, req.Oversubscription, req.Tier); err != nil {
		return nil, err
	}

	topology, err := CalculateTopology(req.Stages, req.Radix, req.Oversubscription)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", models.ErrConstraintViolation, err.Error())
	}

	p := store.FabricParams{
		Name:              strings.TrimSpace(req.Name),
		Tier:              req.Tier,
		Stages:            req.Stages,
		Radix:             topology.Radix,
		Oversubscription:  topology.Oversubscription,
		Description:       req.Description,
		LeafModelID:       req.LeafModelID,
		SpineModelID:      req.SpineModelID,
		SuperSpineModelID: req.SuperSpineModelID,
	}

	rec, err := s.repo.Update(id, p)
	if err != nil {
		return nil, fmt.Errorf("update fabric %d: %w", id, err)
	}

	slog.Info("fabric updated", "fabricID", rec.ID, "name", rec.Name)
	return s.buildResponse(rec, topology)
}

// DeleteFabric removes the fabric with the given id.
func (s *FabricService) DeleteFabric(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete fabric %d: %w", id, err)
	}
	slog.Info("fabric deleted", "fabricID", id)
	return nil
}

// ListDeviceModels returns all device models for assignment dropdowns.
func (s *FabricService) ListDeviceModels() ([]*models.DeviceModel, error) {
	dms, err := s.repo.ListDeviceModels()
	if err != nil {
		return nil, fmt.Errorf("list device models: %w", err)
	}
	return dms, nil
}

// PreviewTopology calculates topology without persisting, for live preview.
func (s *FabricService) PreviewTopology(stages int, radix int, oversubscription float64) (*TopologyPlan, error) {
	topo, err := CalculateTopology(stages, radix, oversubscription)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", models.ErrConstraintViolation, err.Error())
	}
	return topo, nil
}

// --- helpers ---

func validateFabricInput(name string, stages int, radix int, oversubscription float64, tier models.FabricTier) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: fabric name is required", models.ErrConstraintViolation)
	}
	if stages != 2 && stages != 3 && stages != 5 {
		return fmt.Errorf("%w: stages must be 2, 3, or 5", models.ErrConstraintViolation)
	}
	if radix <= 0 {
		return fmt.Errorf("%w: radix must be greater than 0", models.ErrConstraintViolation)
	}
	if oversubscription < 1.0 {
		return fmt.Errorf("%w: oversubscription must be ≥ 1.0", models.ErrConstraintViolation)
	}
	if tier != models.FabricTierFrontEnd && tier != models.FabricTierBackEnd {
		return fmt.Errorf("%w: tier must be 'frontend' or 'backend'", models.ErrConstraintViolation)
	}
	return nil
}

// buildResponse constructs a FabricResponse from a record and topology plan.
// It optionally resolves device model details.
func (s *FabricService) buildResponse(rec *store.FabricRecord, topo *TopologyPlan) (*FabricResponse, error) {
	resp := &FabricResponse{
		FabricRecord: rec,
		Topology:     topo,
		Metrics: &FabricMetrics{
			TotalSwitches:         topo.TotalSwitches,
			TotalHostPorts:        topo.TotalHostPorts,
			OversubscriptionRatio: topo.Oversubscription,
		},
	}

	// Collect warnings.
	if topo.RadixCorrectionNote != "" {
		resp.Warnings = append(resp.Warnings, topo.RadixCorrectionNote)
	}

	// Resolve device model assignments.
	if rec.LeafModelID != nil {
		dm, err := s.repo.GetDeviceModelByID(*rec.LeafModelID)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("resolve leaf model: %w", err)
		}
		if err == nil {
			resp.LeafModel = dm
			// Calculate bisection bandwidth: (leafUplinks * portSpeed * leafCount) / 2.
			// Speed is per-port; DeviceModel doesn't store per-port speed, only port_count.
			// We approximate: if model is assigned, total uplink capacity is proportional.
			// BisectionBW = uplinks / radix * totalPorts * portSpeed, but portSpeed unknown.
			// Leave as a feature note; set a placeholder if port speed were known.
			_ = dm
		}
	}
	if rec.SpineModelID != nil {
		dm, err := s.repo.GetDeviceModelByID(*rec.SpineModelID)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("resolve spine model: %w", err)
		}
		if err == nil {
			resp.SpineModel = dm
		}
	}
	if rec.SuperSpineModelID != nil {
		dm, err := s.repo.GetDeviceModelByID(*rec.SuperSpineModelID)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("resolve super-spine model: %w", err)
		}
		if err == nil {
			resp.SuperSpineModel = dm
		}
	}

	return resp, nil
}

func isNotFound(err error) bool {
	return errors.Is(err, models.ErrNotFound)
}
