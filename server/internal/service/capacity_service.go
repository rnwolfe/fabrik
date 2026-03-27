package service

import (
	"fmt"
	"log/slog"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// CapacityRepository defines the database queries needed by CapacityService.
// All queries are implemented against the raw *sql.DB to avoid duplicating store
// abstractions — capacity aggregation spans multiple tables with complex joins.
type CapacityRepository interface {
	QueryRackCapacity(rackID int64) (*models.CapacitySummary, error)
	QueryBlockCapacity(blockID int64) (*models.CapacitySummary, error)
	QuerySuperBlockCapacity(superBlockID int64) (*models.CapacitySummary, error)
	QuerySiteCapacity(siteID int64) (*models.CapacitySummary, error)
	QueryDesignCapacity(designID int64) (*models.CapacitySummary, error)
}

// CapacityService aggregates power and resource capacity across the design hierarchy.
type CapacityService struct {
	repo CapacityRepository
}

// NewCapacityService returns a new CapacityService backed by repo.
func NewCapacityService(repo CapacityRepository) *CapacityService {
	return &CapacityService{repo: repo}
}

// GetRackCapacity returns aggregated capacity for a single rack.
func (s *CapacityService) GetRackCapacity(rackID int64) (*models.CapacitySummary, error) {
	c, err := s.repo.QueryRackCapacity(rackID)
	if err != nil {
		return nil, fmt.Errorf("get rack capacity %d: %w", rackID, err)
	}
	slog.Debug("rack capacity computed", "rackID", rackID)
	return c, nil
}

// GetBlockCapacity returns aggregated capacity for all racks in a block.
func (s *CapacityService) GetBlockCapacity(blockID int64) (*models.CapacitySummary, error) {
	c, err := s.repo.QueryBlockCapacity(blockID)
	if err != nil {
		return nil, fmt.Errorf("get block capacity %d: %w", blockID, err)
	}
	slog.Debug("block capacity computed", "blockID", blockID)
	return c, nil
}

// GetSuperBlockCapacity returns aggregated capacity for all blocks in a super-block.
func (s *CapacityService) GetSuperBlockCapacity(superBlockID int64) (*models.CapacitySummary, error) {
	c, err := s.repo.QuerySuperBlockCapacity(superBlockID)
	if err != nil {
		return nil, fmt.Errorf("get super-block capacity %d: %w", superBlockID, err)
	}
	slog.Debug("super-block capacity computed", "superBlockID", superBlockID)
	return c, nil
}

// GetSiteCapacity returns aggregated capacity for all super-blocks in a site.
func (s *CapacityService) GetSiteCapacity(siteID int64) (*models.CapacitySummary, error) {
	c, err := s.repo.QuerySiteCapacity(siteID)
	if err != nil {
		return nil, fmt.Errorf("get site capacity %d: %w", siteID, err)
	}
	slog.Debug("site capacity computed", "siteID", siteID)
	return c, nil
}

// GetDesignCapacity returns aggregated capacity for all sites in a design.
func (s *CapacityService) GetDesignCapacity(designID int64) (*models.CapacitySummary, error) {
	c, err := s.repo.QueryDesignCapacity(designID)
	if err != nil {
		return nil, fmt.Errorf("get design capacity %d: %w", designID, err)
	}
	slog.Debug("design capacity computed", "designID", designID)
	return c, nil
}
