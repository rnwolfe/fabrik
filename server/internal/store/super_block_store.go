package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// SuperBlockStore provides CRUD operations for SuperBlock records.
type SuperBlockStore struct {
	db *sql.DB
}

// NewSuperBlockStore returns a new SuperBlockStore backed by db.
func NewSuperBlockStore(db *sql.DB) *SuperBlockStore {
	return &SuperBlockStore{db: db}
}

// CreateSuperBlock inserts a new SuperBlock and returns the saved record.
func (s *SuperBlockStore) CreateSuperBlock(sb *models.SuperBlock) (*models.SuperBlock, error) {
	const q = `
		INSERT INTO super_blocks (site_id, name, description)
		VALUES (?, ?, ?)
		RETURNING id, site_id, name, description, created_at, updated_at`

	out := &models.SuperBlock{}
	err := s.db.QueryRow(q, sb.SiteID, sb.Name, sb.Description).
		Scan(&out.ID, &out.SiteID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create super-block: %w", err)
	}
	return out, nil
}

// GetSuperBlock returns the SuperBlock with the given id, or models.ErrNotFound.
func (s *SuperBlockStore) GetSuperBlock(id int64) (*models.SuperBlock, error) {
	const q = `
		SELECT id, site_id, name, description, created_at, updated_at
		FROM super_blocks WHERE id = ?`

	out := &models.SuperBlock{}
	err := s.db.QueryRow(q, id).
		Scan(&out.ID, &out.SiteID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get super-block %d: %w", id, err)
	}
	return out, nil
}

// ListSuperBlocks returns all super-blocks for a site, ordered by id.
func (s *SuperBlockStore) ListSuperBlocks(siteID int64) ([]*models.SuperBlock, error) {
	const q = `
		SELECT id, site_id, name, description, created_at, updated_at
		FROM super_blocks WHERE site_id = ? ORDER BY id`

	rows, err := s.db.Query(q, siteID)
	if err != nil {
		return nil, fmt.Errorf("list super-blocks: %w", err)
	}
	defer rows.Close()

	var out []*models.SuperBlock
	for rows.Next() {
		sb := &models.SuperBlock{}
		if err := rows.Scan(&sb.ID, &sb.SiteID, &sb.Name, &sb.Description, &sb.CreatedAt, &sb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan super-block: %w", err)
		}
		out = append(out, sb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate super-blocks: %w", err)
	}
	return out, nil
}

// UpdateSuperBlock updates the name and description of a super-block.
func (s *SuperBlockStore) UpdateSuperBlock(id int64, name, description string) (*models.SuperBlock, error) {
	const q = `
		UPDATE super_blocks SET name = ?, description = ?,
		updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		RETURNING id, site_id, name, description, created_at, updated_at`

	out := &models.SuperBlock{}
	err := s.db.QueryRow(q, name, description, id).
		Scan(&out.ID, &out.SiteID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update super-block %d: %w", id, err)
	}
	return out, nil
}

// DeleteSuperBlock deletes a super-block by id (CASCADE deletes blocks).
func (s *SuperBlockStore) DeleteSuperBlock(id int64) error {
	res, err := s.db.Exec(`DELETE FROM super_blocks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete super-block %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete super-block %d rows affected: %w", id, err)
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

// CountBlocksInSuperBlock returns the number of blocks under a super-block.
func (s *SuperBlockStore) CountBlocksInSuperBlock(superBlockID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM blocks WHERE super_block_id = ?`, superBlockID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count blocks in super-block %d: %w", superBlockID, err)
	}
	return count, nil
}
