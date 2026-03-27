package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/models"
	"github.com/rnwolfe/fabrik/server/internal/service"
)

// fakeDesignRepo is an in-memory implementation of DesignRepository for tests.
type fakeDesignRepo struct {
	designs map[int64]*models.Design
	nextID  int64
}

func newFakeRepo() *fakeDesignRepo {
	return &fakeDesignRepo{designs: make(map[int64]*models.Design)}
}

func (r *fakeDesignRepo) Create(d *models.Design) (*models.Design, error) {
	r.nextID++
	out := *d
	out.ID = r.nextID
	r.designs[out.ID] = &out
	return &out, nil
}

func (r *fakeDesignRepo) List() ([]*models.Design, error) {
	out := make([]*models.Design, 0, len(r.designs))
	for _, d := range r.designs {
		cp := *d
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeDesignRepo) Get(id int64) (*models.Design, error) {
	d, ok := r.designs[id]
	if !ok {
		return nil, models.ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (r *fakeDesignRepo) Delete(id int64) error {
	if _, ok := r.designs[id]; !ok {
		return models.ErrNotFound
	}
	delete(r.designs, id)
	return nil
}

func TestDesignService_CreateDesign(t *testing.T) {
	tests := []struct {
		name        string
		designName  string
		description string
		wantErr     bool
		wantErrType error
	}{
		{
			name:        "valid creation",
			designName:  "my-design",
			description: "some description",
		},
		{
			name:        "empty name",
			designName:  "",
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "whitespace-only name",
			designName:  "   ",
			wantErr:     true,
			wantErrType: models.ErrConstraintViolation,
		},
		{
			name:        "name with leading and trailing spaces is trimmed",
			designName:  "  my-design  ",
			description: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewDesignService(newFakeRepo())
			d, err := svc.CreateDesign(tc.designName, tc.description)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
					t.Errorf("expected error type %v, got %v", tc.wantErrType, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.ID == 0 {
				t.Error("expected non-zero ID")
			}
			wantName := strings.TrimSpace(tc.designName)
			if d.Name != wantName {
				t.Errorf("expected name %q, got %q", wantName, d.Name)
			}
		})
	}
}

func TestDesignService_ListDesigns(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewDesignService(repo)

	// Empty list.
	designs, err := svc.ListDesigns()
	if err != nil {
		t.Fatalf("ListDesigns: %v", err)
	}
	if len(designs) != 0 {
		t.Errorf("expected 0 designs, got %d", len(designs))
	}

	// Add two designs.
	svc.CreateDesign("d1", "")
	svc.CreateDesign("d2", "")

	designs, err = svc.ListDesigns()
	if err != nil {
		t.Fatalf("ListDesigns after create: %v", err)
	}
	if len(designs) != 2 {
		t.Errorf("expected 2 designs, got %d", len(designs))
	}
}

func TestDesignService_GetDesign(t *testing.T) {
	svc := service.NewDesignService(newFakeRepo())

	created, _ := svc.CreateDesign("get-test", "")

	got, err := svc.GetDesign(created.ID)
	if err != nil {
		t.Fatalf("GetDesign: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, got.ID)
	}

	// Not found.
	_, err = svc.GetDesign(999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDesignService_DeleteDesign(t *testing.T) {
	svc := service.NewDesignService(newFakeRepo())

	created, _ := svc.CreateDesign("delete-test", "")

	if err := svc.DeleteDesign(created.ID); err != nil {
		t.Fatalf("DeleteDesign: %v", err)
	}

	_, err := svc.GetDesign(created.ID)
	if !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}

	// Delete non-existent.
	if err := svc.DeleteDesign(999); !errors.Is(err, models.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
