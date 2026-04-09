package store_test

import (
	"errors"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

// seedDesign creates a design and returns its ID.
func seedDesign(t *testing.T, ds *store.DesignStore) int64 {
	t.Helper()
	d, err := ds.Create(&models.Design{Name: "test-design"})
	if err != nil {
		t.Fatalf("seedDesign: %v", err)
	}
	return d.ID
}

func TestSiteStore_CRUD(t *testing.T) {
	db := openTestDB(t)
	ds := store.NewDesignStore(db)
	ss := store.NewSiteStore(db)

	designID := seedDesign(t, ds)

	t.Run("create", func(t *testing.T) {
		site, err := ss.CreateSite(&models.Site{
			DesignID:    designID,
			Name:        "Site A",
			Description: "primary DC",
		})
		if err != nil {
			t.Fatalf("CreateSite: %v", err)
		}
		if site.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if site.Name != "Site A" {
			t.Errorf("expected name %q, got %q", "Site A", site.Name)
		}
		if site.DesignID != designID {
			t.Errorf("expected designID %d, got %d", designID, site.DesignID)
		}
		if site.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("get", func(t *testing.T) {
		created, _ := ss.CreateSite(&models.Site{DesignID: designID, Name: "Site B"})
		got, err := ss.GetSite(created.ID)
		if err != nil {
			t.Fatalf("GetSite: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := ss.GetSite(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("list", func(t *testing.T) {
		db2 := openTestDB(t)
		ds2 := store.NewDesignStore(db2)
		ss2 := store.NewSiteStore(db2)
		did2 := seedDesign(t, ds2)
		ss2.CreateSite(&models.Site{DesignID: did2, Name: "X"})
		ss2.CreateSite(&models.Site{DesignID: did2, Name: "Y"})
		sites, err := ss2.ListSites(did2)
		if err != nil {
			t.Fatalf("ListSites: %v", err)
		}
		if len(sites) != 2 {
			t.Errorf("expected 2 sites, got %d", len(sites))
		}
	})

	t.Run("update", func(t *testing.T) {
		created, _ := ss.CreateSite(&models.Site{DesignID: designID, Name: "Site C"})
		updated, err := ss.UpdateSite(created.ID, "Site C-updated", "new desc")
		if err != nil {
			t.Fatalf("UpdateSite: %v", err)
		}
		if updated.Name != "Site C-updated" {
			t.Errorf("expected %q, got %q", "Site C-updated", updated.Name)
		}
		if updated.Description != "new desc" {
			t.Errorf("expected %q, got %q", "new desc", updated.Description)
		}
	})

	t.Run("update not found", func(t *testing.T) {
		_, err := ss.UpdateSite(999999, "x", "")
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		created, _ := ss.CreateSite(&models.Site{DesignID: designID, Name: "delete-me"})
		if err := ss.DeleteSite(created.ID); err != nil {
			t.Fatalf("DeleteSite: %v", err)
		}
		_, err := ss.GetSite(created.ID)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		err := ss.DeleteSite(999999)
		if !errors.Is(err, models.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("count super-blocks in site", func(t *testing.T) {
		db2 := openTestDB(t)
		ds2 := store.NewDesignStore(db2)
		ss2 := store.NewSiteStore(db2)
		sbs := store.NewSuperBlockStore(db2)
		did2 := seedDesign(t, ds2)

		site, _ := ss2.CreateSite(&models.Site{DesignID: did2, Name: "Counted Site"})

		count0, _ := ss2.CountSuperBlocksInSite(site.ID)
		if count0 != 0 {
			t.Errorf("expected 0 super-blocks, got %d", count0)
		}

		sbs.CreateSuperBlock(&models.SuperBlock{SiteID: site.ID, Name: "SB1"})
		sbs.CreateSuperBlock(&models.SuperBlock{SiteID: site.ID, Name: "SB2"})

		count2, err := ss2.CountSuperBlocksInSite(site.ID)
		if err != nil {
			t.Fatalf("CountSuperBlocksInSite: %v", err)
		}
		if count2 != 2 {
			t.Errorf("expected 2 super-blocks, got %d", count2)
		}
	})
}
