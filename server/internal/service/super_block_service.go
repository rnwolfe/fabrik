package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// SuperBlockRepository is the store interface required by SuperBlockService.
type SuperBlockRepository interface {
	CreateSuperBlock(sb *models.SuperBlock) (*models.SuperBlock, error)
	GetSuperBlock(id int64) (*models.SuperBlock, error)
	ListSuperBlocks(siteID int64) ([]*models.SuperBlock, error)
	UpdateSuperBlock(id int64, name, description string) (*models.SuperBlock, error)
	DeleteSuperBlock(id int64) error
	CountBlocksInSuperBlock(superBlockID int64) (int, error)
}

// SuperBlockService implements business logic for SuperBlock resources.
type SuperBlockService struct {
	repo SuperBlockRepository
}

// NewSuperBlockService returns a new SuperBlockService backed by repo.
func NewSuperBlockService(repo SuperBlockRepository) *SuperBlockService {
	return &SuperBlockService{repo: repo}
}

// CreateSuperBlock validates and creates a new super-block under a site.
func (s *SuperBlockService) CreateSuperBlock(siteID int64, name, description string) (*models.SuperBlock, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: super-block name is required", models.ErrConstraintViolation)
	}
	sb, err := s.repo.CreateSuperBlock(&models.SuperBlock{
		SiteID:      siteID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create super-block: %w", err)
	}
	slog.Info("super-block created", "superBlockID", sb.ID, "siteID", siteID, "name", sb.Name)
	return sb, nil
}

// GetSuperBlock returns the super-block with the given id.
func (s *SuperBlockService) GetSuperBlock(id int64) (*models.SuperBlock, error) {
	sb, err := s.repo.GetSuperBlock(id)
	if err != nil {
		return nil, fmt.Errorf("get super-block %d: %w", id, err)
	}
	return sb, nil
}

// ListSuperBlocks returns all super-blocks for a site.
func (s *SuperBlockService) ListSuperBlocks(siteID int64) ([]*models.SuperBlock, error) {
	sbs, err := s.repo.ListSuperBlocks(siteID)
	if err != nil {
		return nil, fmt.Errorf("list super-blocks: %w", err)
	}
	return sbs, nil
}

// UpdateSuperBlock validates and updates a super-block's name and description.
func (s *SuperBlockService) UpdateSuperBlock(id int64, name, description string) (*models.SuperBlock, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: super-block name is required", models.ErrConstraintViolation)
	}
	sb, err := s.repo.UpdateSuperBlock(id, name, description)
	if err != nil {
		return nil, fmt.Errorf("update super-block %d: %w", id, err)
	}
	slog.Info("super-block updated", "superBlockID", id, "name", name)
	return sb, nil
}

// DeleteSuperBlock removes a super-block. Returns ErrConstraintViolation when blocks still exist.
func (s *SuperBlockService) DeleteSuperBlock(id int64) error {
	count, err := s.repo.CountBlocksInSuperBlock(id)
	if err != nil {
		return fmt.Errorf("count blocks: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%w: super-block has %d block(s); delete them first", models.ErrConstraintViolation, count)
	}
	if err := s.repo.DeleteSuperBlock(id); err != nil {
		return fmt.Errorf("delete super-block %d: %w", id, err)
	}
	slog.Info("super-block deleted", "superBlockID", id)
	return nil
}
