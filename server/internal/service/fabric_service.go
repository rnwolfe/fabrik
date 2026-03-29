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
	Topology *TopologyPlan  `json:"topology"`
	Warnings []string       `json:"warnings,omitempty"`
	Metrics  *FabricMetrics `json:"metrics"`
	// Resolved device model objects for the UI.
	LeafModel       *models.DeviceModel `json:"leaf_model,omitempty"`
	SpineModel      *models.DeviceModel `json:"spine_model,omitempty"`
	SuperSpineModel *models.DeviceModel `json:"super_spine_model,omitempty"`
}

// FabricMetrics holds derived oversubscription metrics for a fabric.
type FabricMetrics struct {
	// LeafSpineOversubscription is the downlink:uplink ratio at the leaf-spine boundary.
	LeafSpineOversubscription float64 `json:"leaf_spine_oversubscription"`
	// SpineSuperSpineOversubscription is the ratio at the spine-to-super-spine boundary
	// (only present for 3-stage and 5-stage fabrics).
	SpineSuperSpineOversubscription *float64 `json:"spine_super_spine_oversubscription,omitempty"`
}

// CreateFabricRequest holds the inputs for creating a fabric.
type CreateFabricRequest struct {
	DesignID         int64             `json:"design_id"`
	Name             string            `json:"name"`
	Tier             models.FabricTier `json:"tier"`
	Stages           int               `json:"stages"`
	Radix            int               `json:"radix"`
	Oversubscription float64           `json:"oversubscription"`
	// LeafCount is the number of leaf switches to include. 0 means full fabric
	// (populate all spine ports); 1 means minimum viable (single leaf).
	LeafCount         int   `json:"leaf_count,omitempty"`
	Description       string `json:"description"`
	LeafModelID       int64 `json:"leaf_model_id,omitempty"`
	SpineModelID      int64 `json:"spine_model_id,omitempty"`
	SuperSpineModelID int64 `json:"super_spine_model_id,omitempty"`
}

// UpdateFabricRequest holds the inputs for updating a fabric.
type UpdateFabricRequest struct {
	Name             string            `json:"name"`
	Tier             models.FabricTier `json:"tier"`
	Stages           int               `json:"stages"`
	Radix            int               `json:"radix"`
	Oversubscription float64           `json:"oversubscription"`
	// LeafCount is the number of leaf switches to include. 0 means full fabric;
	// 1 means minimum viable.
	LeafCount         int   `json:"leaf_count,omitempty"`
	Description       string `json:"description"`
	LeafModelID       int64 `json:"leaf_model_id,omitempty"`
	SpineModelID      int64 `json:"spine_model_id,omitempty"`
	SuperSpineModelID int64 `json:"super_spine_model_id,omitempty"`
	// Force regeneration even if rack placements exist.
	Force bool `json:"force"`
}

// PreviewTopologyRequest holds parameters for a topology preview (no persistence).
type PreviewTopologyRequest struct {
	Stages           int     `json:"stages"`
	Radix            int     `json:"radix"`
	Oversubscription float64 `json:"oversubscription"`
	LeafCount        int     `json:"leaf_count,omitempty"`
	LeafModelID      int64   `json:"leaf_model_id,omitempty"`
	SpineModelID     int64   `json:"spine_model_id,omitempty"`
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

	hints, err := s.buildHints(req.LeafModelID, req.SpineModelID, req.Radix, req.LeafCount)
	if err != nil {
		return nil, err
	}

	topology, err := CalculateTopology(req.Stages, hints.leafRadix, req.Oversubscription, &TopologyHints{
		SpineRadix: hints.spineRadix,
		LeafCount:  req.LeafCount,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", models.ErrConstraintViolation, err.Error())
	}

	p := store.FabricParams{
		DesignID:          req.DesignID,
		Name:              strings.TrimSpace(req.Name),
		Tier:              req.Tier,
		Stages:            req.Stages,
		Radix:             topology.Radix, // use corrected leaf radix
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

	slog.Info("fabric created", "fabricID", rec.ID, "name", rec.Name, "stages", req.Stages,
		"leafCount", topology.LeafCount, "spineCount", topology.SpineCount)

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
		hints, err := s.buildHintsFromRecord(rec)
		if err != nil {
			slog.Warn("resolve device models for fabric failed", "fabricID", rec.ID, "err", err)
		}
		topo, err := CalculateTopology(rec.Stages, hints.leafRadix, rec.Oversubscription, &TopologyHints{
			SpineRadix: hints.spineRadix,
		})
		if err != nil {
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

	hints, err := s.buildHintsFromRecord(rec)
	if err != nil {
		slog.Warn("resolve device models for fabric", "fabricID", id, "err", err)
	}
	topo, err := CalculateTopology(rec.Stages, hints.leafRadix, rec.Oversubscription, &TopologyHints{
		SpineRadix: hints.spineRadix,
	})
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

	hints, err := s.buildHints(req.LeafModelID, req.SpineModelID, req.Radix, req.LeafCount)
	if err != nil {
		return nil, err
	}

	topology, err := CalculateTopology(req.Stages, hints.leafRadix, req.Oversubscription, &TopologyHints{
		SpineRadix: hints.spineRadix,
		LeafCount:  req.LeafCount,
	})
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
func (s *FabricService) PreviewTopology(req PreviewTopologyRequest) (*TopologyPlan, error) {
	hints, err := s.buildHints(req.LeafModelID, req.SpineModelID, req.Radix, req.LeafCount)
	if err != nil {
		return nil, err
	}
	topo, err := CalculateTopology(req.Stages, hints.leafRadix, req.Oversubscription, &TopologyHints{
		SpineRadix: hints.spineRadix,
		LeafCount:  req.LeafCount,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", models.ErrConstraintViolation, err.Error())
	}
	return topo, nil
}

// --- helpers ---

// radixHints holds resolved leaf and spine radix values for topology calculation.
type radixHints struct {
	leafRadix  int
	spineRadix int
	// leafPortGroups holds the port groups from the leaf device model, if any.
	// The fabric/block context decides which group is uplink vs downlink.
	leafPortGroups []models.PortGroup
}

// buildHints resolves leaf and spine radix from device model IDs (when provided).
// fallbackRadix is used when no leaf model is assigned.
func (s *FabricService) buildHints(leafModelID, spineModelID int64, fallbackRadix, leafCount int) (radixHints, error) {
	h := radixHints{
		leafRadix:  fallbackRadix,
		spineRadix: 0, // 0 → topology_calc falls back to leafRadix
	}
	if leafModelID > 0 {
		dm, err := s.repo.GetDeviceModelByID(leafModelID)
		if err != nil && !isNotFound(err) {
			return h, fmt.Errorf("resolve leaf model: %w", err)
		}
		if err == nil && dm.PortCount > 0 {
			h.leafRadix = dm.PortCount
		}
		if err == nil {
			h.leafPortGroups = dm.PortGroups
		}
	}
	if spineModelID > 0 {
		dm, err := s.repo.GetDeviceModelByID(spineModelID)
		if err != nil && !isNotFound(err) {
			return h, fmt.Errorf("resolve spine model: %w", err)
		}
		if err == nil && dm.PortCount > 0 {
			h.spineRadix = dm.PortCount
		}
	}
	return h, nil
}

// buildHintsFromRecord resolves radix hints from an existing FabricRecord.
func (s *FabricService) buildHintsFromRecord(rec *store.FabricRecord) (radixHints, error) {
	leafModelID := int64(0)
	spineModelID := int64(0)
	if rec.LeafModelID != nil {
		leafModelID = *rec.LeafModelID
	}
	if rec.SpineModelID != nil {
		spineModelID = *rec.SpineModelID
	}
	return s.buildHints(leafModelID, spineModelID, rec.Radix, 0)
}

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
func (s *FabricService) buildResponse(rec *store.FabricRecord, topo *TopologyPlan) (*FabricResponse, error) {
	metrics := &FabricMetrics{
		LeafSpineOversubscription: topo.Oversubscription,
	}
	if topo.Stages >= 3 && topo.SuperSpineCount > 0 {
		ratio := topo.Oversubscription
		metrics.SpineSuperSpineOversubscription = &ratio
	}

	resp := &FabricResponse{
		FabricRecord: rec,
		Topology:     topo,
		Metrics:      metrics,
	}

	if topo.RadixCorrectionNote != "" {
		resp.Warnings = append(resp.Warnings, topo.RadixCorrectionNote)
	}

	// Resolve device model assignments (models already looked up for topology;
	// re-fetch here for the response payload).
	if rec.LeafModelID != nil {
		dm, err := s.repo.GetDeviceModelByID(*rec.LeafModelID)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("resolve leaf model: %w", err)
		}
		if err == nil {
			resp.LeafModel = dm
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
