package store

import (
	"database/sql"
	"fmt"

	"github.com/rnwolfe/fabrik/server/internal/models"
)

// HierarchyBlock is a block node in the hierarchy tree.
type HierarchyBlock struct {
	models.Block
}

// HierarchySuperBlock is a super-block node with its nested blocks.
type HierarchySuperBlock struct {
	models.SuperBlock
	Blocks []*HierarchyBlock `json:"blocks"`
}

// HierarchySite is a site node with its nested super-blocks.
type HierarchySite struct {
	models.Site
	SuperBlocks []*HierarchySuperBlock `json:"super_blocks"`
}

// DesignHierarchy is the full nested tree for a design.
type DesignHierarchy struct {
	DesignID int64            `json:"design_id"`
	Sites    []*HierarchySite `json:"sites"`
}

// HierarchyStore builds the full hierarchy tree for a design in one pass.
type HierarchyStore struct {
	db *sql.DB
}

// NewHierarchyStore returns a new HierarchyStore backed by db.
func NewHierarchyStore(db *sql.DB) *HierarchyStore {
	return &HierarchyStore{db: db}
}

// GetDesignHierarchy returns the full site → super_block → block tree for designID.
func (s *HierarchyStore) GetDesignHierarchy(designID int64) (*DesignHierarchy, error) {
	// Load all sites for the design.
	siteRows, err := s.db.Query(`
		SELECT id, design_id, name, description, created_at, updated_at
		FROM sites WHERE design_id = ? ORDER BY id`, designID)
	if err != nil {
		return nil, fmt.Errorf("hierarchy: list sites: %w", err)
	}
	defer siteRows.Close()

	siteMap := map[int64]*HierarchySite{}
	var siteOrder []int64
	for siteRows.Next() {
		site := &HierarchySite{SuperBlocks: []*HierarchySuperBlock{}}
		if err := siteRows.Scan(&site.ID, &site.DesignID, &site.Name, &site.Description, &site.CreatedAt, &site.UpdatedAt); err != nil {
			return nil, fmt.Errorf("hierarchy: scan site: %w", err)
		}
		siteMap[site.ID] = site
		siteOrder = append(siteOrder, site.ID)
	}
	if err := siteRows.Err(); err != nil {
		return nil, fmt.Errorf("hierarchy: iterate sites: %w", err)
	}

	if len(siteOrder) == 0 {
		return &DesignHierarchy{DesignID: designID, Sites: []*HierarchySite{}}, nil
	}

	// Load all super-blocks for these sites.
	sbRows, err := s.db.Query(`
		SELECT id, site_id, name, description, created_at, updated_at
		FROM super_blocks WHERE site_id IN (
			SELECT id FROM sites WHERE design_id = ?
		) ORDER BY id`, designID)
	if err != nil {
		return nil, fmt.Errorf("hierarchy: list super-blocks: %w", err)
	}
	defer sbRows.Close()

	sbMap := map[int64]*HierarchySuperBlock{}
	for sbRows.Next() {
		sb := &HierarchySuperBlock{Blocks: []*HierarchyBlock{}}
		if err := sbRows.Scan(&sb.ID, &sb.SiteID, &sb.Name, &sb.Description, &sb.CreatedAt, &sb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("hierarchy: scan super-block: %w", err)
		}
		sbMap[sb.ID] = sb
		if site, ok := siteMap[sb.SiteID]; ok {
			site.SuperBlocks = append(site.SuperBlocks, sb)
		}
	}
	if err := sbRows.Err(); err != nil {
		return nil, fmt.Errorf("hierarchy: iterate super-blocks: %w", err)
	}

	// Load all blocks for the super-blocks.
	blockRows, err := s.db.Query(`
		SELECT id, super_block_id, name, description, created_at, updated_at
		FROM blocks WHERE super_block_id IN (
			SELECT id FROM super_blocks WHERE site_id IN (
				SELECT id FROM sites WHERE design_id = ?
			)
		) ORDER BY id`, designID)
	if err != nil {
		return nil, fmt.Errorf("hierarchy: list blocks: %w", err)
	}
	defer blockRows.Close()

	for blockRows.Next() {
		b := &HierarchyBlock{}
		if err := blockRows.Scan(&b.ID, &b.SuperBlockID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("hierarchy: scan block: %w", err)
		}
		if sb, ok := sbMap[b.SuperBlockID]; ok {
			sb.Blocks = append(sb.Blocks, b)
		}
	}
	if err := blockRows.Err(); err != nil {
		return nil, fmt.Errorf("hierarchy: iterate blocks: %w", err)
	}

	sites := make([]*HierarchySite, 0, len(siteOrder))
	for _, id := range siteOrder {
		sites = append(sites, siteMap[id])
	}
	return &DesignHierarchy{DesignID: designID, Sites: sites}, nil
}
