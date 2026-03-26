package knowledge_test

import (
	"testing"
	"testing/fstest"

	"github.com/rnwolfe/fabrik/server/internal/knowledge"
)

func newTestFS() fstest.MapFS {
	return fstest.MapFS{
		"networking/clos.md": {
			Data: []byte(`---
title: Clos Basics
category: networking
tags: [clos, fabric]
---

# Clos Basics

Body text here.
`),
		},
		"networking/ecmp.md": {
			Data: []byte(`---
title: ECMP
category: networking
tags: [ecmp, multipath]
---

ECMP content.
`),
		},
		"infrastructure/rack.md": {
			Data: []byte(`# No Frontmatter Article

Just a body.
`),
		},
		"notmarkdown.txt": {Data: []byte("ignored")},
	}
}

func TestLoadIndex(t *testing.T) {
	loader := knowledge.NewLoader(newTestFS())
	idx, err := loader.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}

	if len(idx.Articles) != 3 {
		t.Errorf("expected 3 articles, got %d", len(idx.Articles))
	}

	// Index entries must not include content.
	for _, a := range idx.Articles {
		if a.Content != "" {
			t.Errorf("article %q should not have content in index", a.Path)
		}
	}
}

func TestLoadArticle_WithFrontmatter(t *testing.T) {
	loader := knowledge.NewLoader(newTestFS())
	a, err := loader.LoadArticle("networking/clos")
	if err != nil {
		t.Fatalf("LoadArticle: %v", err)
	}

	if a.Title != "Clos Basics" {
		t.Errorf("title: got %q, want %q", a.Title, "Clos Basics")
	}
	if a.Category != "networking" {
		t.Errorf("category: got %q, want %q", a.Category, "networking")
	}
	if len(a.Tags) != 2 {
		t.Errorf("tags: got %d, want 2", len(a.Tags))
	}
	if a.Content == "" {
		t.Error("content should be non-empty for full article load")
	}
}

func TestLoadArticle_NoFrontmatter(t *testing.T) {
	loader := knowledge.NewLoader(newTestFS())
	a, err := loader.LoadArticle("infrastructure/rack")
	if err != nil {
		t.Fatalf("LoadArticle: %v", err)
	}

	// Title falls back to filename.
	if a.Title == "" {
		t.Error("title should fall back to filename stem")
	}
	// Category falls back to uncategorized.
	if a.Category != "uncategorized" {
		t.Errorf("category: got %q, want %q", a.Category, "uncategorized")
	}
	// Tags defaults to empty slice.
	if a.Tags == nil {
		t.Error("tags should be empty slice, not nil")
	}
	if a.Content == "" {
		t.Error("content should be non-empty")
	}
}

func TestLoadArticle_NotFound(t *testing.T) {
	loader := knowledge.NewLoader(newTestFS())
	_, err := loader.LoadArticle("does/not/exist")
	if err == nil {
		t.Error("expected error for missing article")
	}
}

func TestLoadArticle_Path(t *testing.T) {
	loader := knowledge.NewLoader(newTestFS())
	a, err := loader.LoadArticle("networking/ecmp")
	if err != nil {
		t.Fatalf("LoadArticle: %v", err)
	}
	if a.Path != "networking/ecmp" {
		t.Errorf("path: got %q, want %q", a.Path, "networking/ecmp")
	}
}
