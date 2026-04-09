package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// SiteRepository is the store interface required by SiteService.
type SiteRepository interface {
	CreateSite(site *models.Site) (*models.Site, error)
	GetSite(id int64) (*models.Site, error)
	ListSites(designID int64) ([]*models.Site, error)
	UpdateSite(id int64, name, description string) (*models.Site, error)
	DeleteSite(id int64) error
	CountSuperBlocksInSite(siteID int64) (int, error)
}

// SiteService implements business logic for Site resources.
type SiteService struct {
	repo      SiteRepository
	hierarchy HierarchyRepository
}

// HierarchyRepository provides the hierarchy tree query.
type HierarchyRepository interface {
	GetDesignHierarchy(designID int64) (*store.DesignHierarchy, error)
}

// NewSiteService returns a new SiteService backed by repo and hierarchy.
func NewSiteService(repo SiteRepository, hierarchy HierarchyRepository) *SiteService {
	return &SiteService{repo: repo, hierarchy: hierarchy}
}

// CreateSite validates and creates a new site under a design.
func (s *SiteService) CreateSite(designID int64, name, description string) (*models.Site, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: site name is required", models.ErrConstraintViolation)
	}
	site, err := s.repo.CreateSite(&models.Site{
		DesignID:    designID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create site: %w", err)
	}
	slog.Info("site created", "siteID", site.ID, "designID", designID, "name", site.Name)
	return site, nil
}

// GetSite returns the site with the given id.
func (s *SiteService) GetSite(id int64) (*models.Site, error) {
	site, err := s.repo.GetSite(id)
	if err != nil {
		return nil, fmt.Errorf("get site %d: %w", id, err)
	}
	return site, nil
}

// ListSites returns all sites for a design.
func (s *SiteService) ListSites(designID int64) ([]*models.Site, error) {
	sites, err := s.repo.ListSites(designID)
	if err != nil {
		return nil, fmt.Errorf("list sites: %w", err)
	}
	return sites, nil
}

// UpdateSite validates and updates a site's name and description.
func (s *SiteService) UpdateSite(id int64, name, description string) (*models.Site, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: site name is required", models.ErrConstraintViolation)
	}
	site, err := s.repo.UpdateSite(id, name, description)
	if err != nil {
		return nil, fmt.Errorf("update site %d: %w", id, err)
	}
	slog.Info("site updated", "siteID", id, "name", name)
	return site, nil
}

// DeleteSite removes a site. Returns ErrConstraintViolation when super-blocks still exist.
func (s *SiteService) DeleteSite(id int64) error {
	count, err := s.repo.CountSuperBlocksInSite(id)
	if err != nil {
		return fmt.Errorf("count super-blocks: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%w: site has %d super-block(s); delete them first", models.ErrConstraintViolation, count)
	}
	if err := s.repo.DeleteSite(id); err != nil {
		return fmt.Errorf("delete site %d: %w", id, err)
	}
	slog.Info("site deleted", "siteID", id)
	return nil
}

// GetDesignHierarchy returns the full nested tree for a design.
func (s *SiteService) GetDesignHierarchy(designID int64) (*store.DesignHierarchy, error) {
	h, err := s.hierarchy.GetDesignHierarchy(designID)
	if err != nil {
		return nil, fmt.Errorf("get hierarchy for design %d: %w", designID, err)
	}
	return h, nil
}
