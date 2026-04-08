package store_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// seedSiteDB creates a design + site using concrete store types and returns (designID, siteID).
func seedSiteDB(t *testing.T, ds *store.DesignStore, ss *store.SiteStore) (int64, int64) {
	t.Helper()
	did := seedDesign(t, ds)
	site, err := ss.CreateSite(&models.Site{DesignID: did, Name: "test-site"})
	if err != nil {
		t.Fatalf("seedSiteDB: CreateSite: %v", err)
	}
	return did, site.ID
}

func TestSuperBlockStore_CRUD(t *testing.T) {
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	ss := store.NewSiteStore(db)
	sbs := store.NewSuperBlockStore(db)

	_, siteID := seedSiteDB(t, ds, ss)

	t.Run("create", func(t *testing.T) {
		sb, err := sbs.CreateSuperBlock(&models.SuperBlock{
			SiteID:      siteID,
			Name:        "Hall A",
			Description: "main hall",
		})
		if err != nil {
			t.Fatalf("CreateSuperBlock: %v", err)
		}
		if sb.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if sb.Name != "Hall A" {
			t.Errorf("expected name %q, got %q", "Hall A", sb.Name)
		}
		if sb.SiteID != siteID {
			t.Errorf("expected siteID %d, got %d", siteID, sb.SiteID)
		}
		if sb.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("get", func(t *testing.T) {
		created, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: siteID, Name: "Hall B"})
		got, err := sbs.GetSuperBlock(created.ID)
		if err != nil {
			t.Fatalf("GetSuperBlock: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := sbs.GetSuperBlock(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("list", func(t *testing.T) {
		db2 := openTestDB(t)
		ds2 := store.NewDesignStore(db2)
		ss2 := store.NewSiteStore(db2)
		sbs2 := store.NewSuperBlockStore(db2)
		_, sid := seedSiteDB(t, ds2, ss2)
		sbs2.CreateSuperBlock(&models.SuperBlock{SiteID: sid, Name: "P"})
		sbs2.CreateSuperBlock(&models.SuperBlock{SiteID: sid, Name: "Q"})
		result, err := sbs2.ListSuperBlocks(sid)
		if err != nil {
			t.Fatalf("ListSuperBlocks: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 super-blocks, got %d", len(result))
		}
	})

	t.Run("update", func(t *testing.T) {
		created, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: siteID, Name: "Hall C"})
		updated, err := sbs.UpdateSuperBlock(created.ID, "Hall C-updated", "new desc")
		if err != nil {
			t.Fatalf("UpdateSuperBlock: %v", err)
		}
		if updated.Name != "Hall C-updated" {
			t.Errorf("expected %q, got %q", "Hall C-updated", updated.Name)
		}
	})

	t.Run("update not found", func(t *testing.T) {
		_, err := sbs.UpdateSuperBlock(999999, "x", "")
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		created, _ := sbs.CreateSuperBlock(&models.SuperBlock{SiteID: siteID, Name: "delete-me"})
		if err := sbs.DeleteSuperBlock(created.ID); err != nil {
			t.Fatalf("DeleteSuperBlock: %v", err)
		}
		_, err := sbs.GetSuperBlock(created.ID)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		err := sbs.DeleteSuperBlock(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("count blocks in super-block", func(t *testing.T) {
		db2 := openTestDB(t)
		ds2 := store.NewDesignStore(db2)
		ss2 := store.NewSiteStore(db2)
		sbs2 := store.NewSuperBlockStore(db2)
		bs := store.NewBlockStore(db2)
		_, sid := seedSiteDB(t, ds2, ss2)

		sb, _ := sbs2.CreateSuperBlock(&models.SuperBlock{SiteID: sid, Name: "Counted SB"})

		count0, _ := sbs2.CountBlocksInSuperBlock(sb.ID)
		if count0 != 0 {
			t.Errorf("expected 0 blocks, got %d", count0)
		}

		bs.CreateBlock(&models.Block{SuperBlockID: sb.ID, Name: "B1"})
		bs.CreateBlock(&models.Block{SuperBlockID: sb.ID, Name: "B2"})

		count2, err := sbs2.CountBlocksInSuperBlock(sb.ID)
		if err != nil {
			t.Fatalf("CountBlocksInSuperBlock: %v", err)
		}
		if count2 != 2 {
			t.Errorf("expected 2 blocks, got %d", count2)
		}
	})

	t.Run("constraint violation: delete with blocks", func(t *testing.T) {
		db2 := openTestDB(t)
		ds2 := store.NewDesignStore(db2)
		ss2 := store.NewSiteStore(db2)
		sbs2 := store.NewSuperBlockStore(db2)
		bs := store.NewBlockStore(db2)
		_, sid := seedSiteDB(t, ds2, ss2)

		sb, _ := sbs2.CreateSuperBlock(&models.SuperBlock{SiteID: sid, Name: "Has Blocks"})
		bs.CreateBlock(&models.Block{SuperBlockID: sb.ID, Name: "child-block"})

		// CountBlocksInSuperBlock should return 1
		count, _ := sbs2.CountBlocksInSuperBlock(sb.ID)
		if count != 1 {
			t.Errorf("expected 1 block, got %d", count)
		}
		// The store itself doesn't enforce the constraint — the service layer does.
		// But we verify count returns the right value so the service can act on it.
	})
}
