package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// SiteStore provides CRUD operations for Site records.
type SiteStore struct {
	db *sql.DB
}

// NewSiteStore returns a new SiteStore backed by db.
func NewSiteStore(db *sql.DB) *SiteStore {
	return &SiteStore{db: db}
}

// CreateSite inserts a new Site and returns the saved record.
func (s *SiteStore) CreateSite(site *models.Site) (*models.Site, error) {
	const q = `
		INSERT INTO sites (design_id, name, description)
		VALUES (?, ?, ?)
		RETURNING id, design_id, name, description, created_at, updated_at`

	out := &models.Site{}
	err := s.db.QueryRow(q, site.DesignID, site.Name, site.Description).
		Scan(&out.ID, &out.DesignID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create site: %w", err)
	}
	return out, nil
}

// GetSite returns the Site with the given id, or models.ErrNotFound.
func (s *SiteStore) GetSite(id int64) (*models.Site, error) {
	const q = `
		SELECT id, design_id, name, description, created_at, updated_at
		FROM sites WHERE id = ?`

	out := &models.Site{}
	err := s.db.QueryRow(q, id).
		Scan(&out.ID, &out.DesignID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get site %d: %w", id, err)
	}
	return out, nil
}

// ListSites returns all sites for a design, ordered by id.
func (s *SiteStore) ListSites(designID int64) ([]*models.Site, error) {
	const q = `
		SELECT id, design_id, name, description, created_at, updated_at
		FROM sites WHERE design_id = ? ORDER BY id`

	rows, err := s.db.Query(q, designID)
	if err != nil {
		return nil, fmt.Errorf("list sites: %w", err)
	}
	defer rows.Close()

	var out []*models.Site
	for rows.Next() {
		site := &models.Site{}
		if err := rows.Scan(&site.ID, &site.DesignID, &site.Name, &site.Description, &site.CreatedAt, &site.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan site: %w", err)
		}
		out = append(out, site)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sites: %w", err)
	}
	return out, nil
}

// UpdateSite updates the name and description of a site.
func (s *SiteStore) UpdateSite(id int64, name, description string) (*models.Site, error) {
	const q = `
		UPDATE sites SET name = ?, description = ?,
		updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, design_id, name, description, created_at, updated_at`

	out := &models.Site{}
	err := s.db.QueryRow(q, name, description, id).
		Scan(&out.ID, &out.DesignID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update site %d: %w", id, err)
	}
	return out, nil
}

// DeleteSite deletes a site by id (CASCADE deletes super_blocks and blocks).
func (s *SiteStore) DeleteSite(id int64) error {
	res, err := s.db.Exec(`DELETE FROM sites WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete site %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete site %d rows affected: %w", id, err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// CountSuperBlocksInSite returns the number of super-blocks under a site.
func (s *SiteStore) CountSuperBlocksInSite(siteID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM super_blocks WHERE site_id = ?`, siteID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count super-blocks in site %d: %w", siteID, err)
	}
	return count, nil
}
