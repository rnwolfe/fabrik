// Package service contains business logic between handlers and the store layer.
package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// DesignRepository is the store interface required by DesignService.
type DesignRepository interface {
	Create(d *models.Design) (*models.Design, error)
	List() ([]*models.Design, error)
	Get(id int64) (*models.Design, error)
	Delete(id int64) error
	GetOrCreateScaffold(designID int64) (*store.DesignScaffold, error)
}

// DesignService implements business logic for Design resources.
type DesignService struct {
	repo DesignRepository
}

// NewDesignService returns a new DesignService backed by repo.
func NewDesignService(repo DesignRepository) *DesignService {
	return &DesignService{repo: repo}
}

// CreateDesign validates and creates a new Design.
func (s *DesignService) CreateDesign(name, description string) (*models.Design, error) {
	// Trim whitespace so that names like "   " are rejected alongside "".
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: design name is required", models.ErrConstraintViolation)
	}

	d, err := s.repo.Create(&models.Design{Name: name, Description: description})
	if err != nil {
		return nil, fmt.Errorf("create design: %w", err)
	}
	slog.Info("design created", "designID", d.ID, "name", d.Name)
	return d, nil
}

// ListDesigns returns all designs.
func (s *DesignService) ListDesigns() ([]*models.Design, error) {
	designs, err := s.repo.List()
	if err != nil {
		return nil, fmt.Errorf("list designs: %w", err)
	}
	return designs, nil
}

// GetDesign returns the design with the given id.
func (s *DesignService) GetDesign(id int64) (*models.Design, error) {
	d, err := s.repo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get design %d: %w", id, err)
	}
	return d, nil
}

// GetScaffold returns the default site and super-block for a design, creating them if needed.
func (s *DesignService) GetScaffold(designID int64) (*store.DesignScaffold, error) {
	// Verify the design exists.
	if _, err := s.repo.Get(designID); err != nil {
		return nil, fmt.Errorf("get design %d: %w", designID, err)
	}
	scaffold, err := s.repo.GetOrCreateScaffold(designID)
	if err != nil {
		return nil, fmt.Errorf("scaffold design %d: %w", designID, err)
	}
	return scaffold, nil
}

// DeleteDesign removes the design with the given id.
func (s *DesignService) DeleteDesign(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete design %d: %w", id, err)
	}
	slog.Info("design deleted", "designID", id)
	return nil
}
