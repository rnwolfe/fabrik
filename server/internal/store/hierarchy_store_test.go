package store_test

import (
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func TestHierarchyStore_Empty(t *testing.T) {
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	hs := store.NewHierarchyStore(db)

	d, err := ds.Create(&models.Design{Name: "empty-design"})
	if err != nil {
		t.Fatalf("Create design: %v", err)
	}

	h, err := hs.GetDesignHierarchy(d.ID)
	if err != nil {
		t.Fatalf("GetDesignHierarchy: %v", err)
	}
	if h.DesignID != d.ID {
		t.Errorf("expected designID %d, got %d", d.ID, h.DesignID)
	}
	if len(h.Sites) != 0 {
		t.Errorf("expected 0 sites, got %d", len(h.Sites))
	}
}

func TestHierarchyStore_SingleSite(t *testing.T) {
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	ss := store.NewSiteStore(db)
	sbs := store.NewSuperBlockStore(db)
	bs := store.NewBlockStore(db)
	hs := store.NewHierarchyStore(db)

	d, _ := ds.Create(&models.Design{Name: "single-site"})
	site, _ := ss.CreateSite(&models.Site{DesignID: d.ID, Name: "Site 1"})
	sb, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: site.ID, Name: "Hall 1"})
	bs.CreateBlock(&models.Block{SuperBlockID: sb.ID, Name: "Block 1"})
	bs.CreateBlock(&models.Block{SuperBlockID: sb.ID, Name: "Block 2"})

	h, err := hs.GetDesignHierarchy(d.ID)
	if err != nil {
		t.Fatalf("GetDesignHierarchy: %v", err)
	}

	if len(h.Sites) != 1 {
		t.Fatalf("expected 1 site, got %d", len(h.Sites))
	}

	gotSite := h.Sites[0]
	if gotSite.Name != "Site 1" {
		t.Errorf("expected site name %q, got %q", "Site 1", gotSite.Name)
	}
	if len(gotSite.SuperBlocks) != 1 {
		t.Fatalf("expected 1 super-block, got %d", len(gotSite.SuperBlocks))
	}

	gotSB := gotSite.SuperBlocks[0]
	if gotSB.Name != "Hall 1" {
		t.Errorf("expected super-block name %q, got %q", "Hall 1", gotSB.Name)
	}
	if len(gotSB.Blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(gotSB.Blocks))
	}
}

func TestHierarchyStore_MultiSite(t *testing.T) {
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	ss := store.NewSiteStore(db)
	sbs := store.NewSuperBlockStore(db)
	bs := store.NewBlockStore(db)
	hs := store.NewHierarchyStore(db)

	d, _ := ds.Create(&models.Design{Name: "multi-site"})

	// Site A with 2 super-blocks, 1 block each.
	siteA, _ := ss.CreateSite(&models.Site{DesignID: d.ID, Name: "Site A"})
	sbA1, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: siteA.ID, Name: "SB A1"})
	sbA2, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: siteA.ID, Name: "SB A2"})
	bs.CreateBlock(&models.Block{SuperBlockID: sbA1.ID, Name: "Block A1-1"})
	bs.CreateBlock(&models.Block{SuperBlockID: sbA2.ID, Name: "Block A2-1"})

	// Site B with 1 super-block, 0 blocks.
	siteB, _ := ss.CreateSite(&models.Site{DesignID: d.ID, Name: "Site B"})
	sbs.CreateSuperBlock(&models.SuperBlock{SiteID: siteB.ID, Name: "SB B1"})

	h, err := hs.GetDesignHierarchy(d.ID)
	if err != nil {
		t.Fatalf("GetDesignHierarchy: %v", err)
	}

	if len(h.Sites) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(h.Sites))
	}

	siteARes := h.Sites[0]
	if siteARes.Name != "Site A" {
		t.Errorf("expected first site %q, got %q", "Site A", siteARes.Name)
	}
	if len(siteARes.SuperBlocks) != 2 {
		t.Errorf("expected 2 super-blocks in Site A, got %d", len(siteARes.SuperBlocks))
	}

	siteBRes := h.Sites[1]
	if siteBRes.Name != "Site B" {
		t.Errorf("expected second site %q, got %q", "Site B", siteBRes.Name)
	}
	if len(siteBRes.SuperBlocks) != 1 {
		t.Errorf("expected 1 super-block in Site B, got %d", len(siteBRes.SuperBlocks))
	}
	if len(siteBRes.SuperBlocks[0].Blocks) != 0 {
		t.Errorf("expected 0 blocks in SB B1, got %d", len(siteBRes.SuperBlocks[0].Blocks))
	}
}

func TestHierarchyStore_BlocksNotInHierarchyIgnored(t *testing.T) {
	// Blocks that are in super-blocks NOT belonging to any site in the design
	// should not appear. This tests isolation: a second design's hierarchy
	// should not bleed into the first.
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	ss := store.NewSiteStore(db)
	sbs := store.NewSuperBlockStore(db)
	bs := store.NewBlockStore(db)
	hs := store.NewHierarchyStore(db)

	d1, _ := ds.Create(&models.Design{Name: "design-1"})
	d2, _ := ds.Create(&models.Design{Name: "design-2"})

	site1, _ := ss.CreateSite(&models.Site{DesignID: d1.ID, Name: "Site 1"})
	sb1, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: site1.ID, Name: "SB1"})
	bs.CreateBlock(&models.Block{SuperBlockID: sb1.ID, Name: "Block-D1"})

	site2, _ := ss.CreateSite(&models.Site{DesignID: d2.ID, Name: "Site 2"})
	sb2, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: site2.ID, Name: "SB2"})
	bs.CreateBlock(&models.Block{SuperBlockID: sb2.ID, Name: "Block-D2"})

	h1, err := hs.GetDesignHierarchy(d1.ID)
	if err != nil {
		t.Fatalf("GetDesignHierarchy d1: %v", err)
	}

	if len(h1.Sites) != 1 {
		t.Fatalf("d1: expected 1 site, got %d", len(h1.Sites))
	}
	if len(h1.Sites[0].SuperBlocks[0].Blocks) != 1 {
		t.Errorf("d1: expected 1 block, got %d", len(h1.Sites[0].SuperBlocks[0].Blocks))
	}
	if h1.Sites[0].SuperBlocks[0].Blocks[0].Name != "Block-D1" {
		t.Errorf("d1: expected block name %q, got %q", "Block-D1", h1.Sites[0].SuperBlocks[0].Blocks[0].Name)
	}
}
