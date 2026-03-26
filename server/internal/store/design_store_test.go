package store_test

import (
	"sync"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/store"
)

func TestDesignStore_CRUD(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDesignStore(db)

	t.Run("create", func(t *testing.T) {
		d, err := s.Create(&models.Design{Name: "test-design", Description: "desc"})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if d.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if d.Name != "test-design" {
			t.Errorf("expected name %q, got %q", "test-design", d.Name)
		}
		if d.Description != "desc" {
			t.Errorf("expected description %q, got %q", "desc", d.Description)
		}
		if d.CreatedAt.IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("list", func(t *testing.T) {
		designs, err := s.List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(designs) == 0 {
			t.Error("expected at least one design")
		}
	})

	t.Run("get", func(t *testing.T) {
		created, _ := s.Create(&models.Design{Name: "get-test"})
		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, got.ID)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := s.Get(999999)
		if err == nil {
			t.Fatal("expected error for missing design")
		}
		if !isNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		created, _ := s.Create(&models.Design{Name: "delete-test"})
		if err := s.Delete(created.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := s.Get(created.ID)
		if !isNotFound(err) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("delete not found", func(t *testing.T) {
		err := s.Delete(999999)
		if !isNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDesignStore_ConcurrentReads(t *testing.T) {
	db := openTestDB(t)
	s := store.NewDesignStore(db)

	// Seed some data.
	for i := 0; i < 5; i++ {
		if _, err := s.Create(&models.Design{Name: "concurrent-design"}); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.List(); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent read error: %v", err)
	}
}

// isNotFound reports whether err wraps models.ErrNotFound.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == models.ErrNotFound.Error() ||
		containsError(err, models.ErrNotFound)
}

func containsError(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
