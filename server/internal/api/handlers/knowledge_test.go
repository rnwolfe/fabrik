package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/rnwolfe/fabrik/server/internal/api/handlers"
	"github.com/rnwolfe/fabrik/server/internal/knowledge"
)

func newTestKnowledgeLoader() *knowledge.Loader {
	fsys := fstest.MapFS{
		"networking/clos.md": {
			Data: []byte(`---
title: Clos Basics
category: networking
tags: [clos, fabric]
---

Body text here.
`),
		},
		"infrastructure/rack.md": {
			Data: []byte(`# Rack Design

Just a body.
`),
		},
	}
	return knowledge.NewLoader(fsys)
}

func TestKnowledgeHandler_Index(t *testing.T) {
	h := handlers.NewKnowledgeHandler(newTestKnowledgeLoader())

	req := httptest.NewRequest(http.MethodGet, "/api/knowledge", nil)
	rec := httptest.NewRecorder()
	h.Index(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	var idx knowledge.Index
	if err := json.NewDecoder(rec.Body).Decode(&idx); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(idx.Articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(idx.Articles))
	}

	for _, a := range idx.Articles {
		if a.Content != "" {
			t.Errorf("index article %q should not have content", a.Path)
		}
	}
}

func TestKnowledgeHandler_Get_Found(t *testing.T) {
	h := handlers.NewKnowledgeHandler(newTestKnowledgeLoader())

	req := httptest.NewRequest(http.MethodGet, "/api/knowledge/networking/clos", nil)
	req.SetPathValue("path", "networking/clos")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	var a knowledge.Article
	if err := json.NewDecoder(rec.Body).Decode(&a); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if a.Title != "Clos Basics" {
		t.Errorf("title: got %q, want %q", a.Title, "Clos Basics")
	}
	if a.Content == "" {
		t.Error("full article should include content")
	}
}

func TestKnowledgeHandler_Get_NotFound(t *testing.T) {
	h := handlers.NewKnowledgeHandler(newTestKnowledgeLoader())

	req := httptest.NewRequest(http.MethodGet, "/api/knowledge/does/not/exist", nil)
	req.SetPathValue("path", "does/not/exist")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestKnowledgeHandler_Get_MissingPath(t *testing.T) {
	h := handlers.NewKnowledgeHandler(newTestKnowledgeLoader())

	req := httptest.NewRequest(http.MethodGet, "/api/knowledge/", nil)
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestKnowledgeHandler_Get_InvalidPath(t *testing.T) {
	h := handlers.NewKnowledgeHandler(newTestKnowledgeLoader())

	tests := []struct {
		name string
		path string
	}{
		{"path traversal", "../etc/passwd"},
		{"absolute path", "/etc/passwd"},
		{"double dot segment", "valid/../etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/knowledge/"+tt.path, nil)
			req.SetPathValue("path", tt.path)
			rec := httptest.NewRecorder()
			h.Get(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("path %q: status got %d, want %d", tt.path, rec.Code, http.StatusBadRequest)
			}
		})
	}
}

// ensure knowledge.Loader satisfies KnowledgeLoader at compile time
var _ interface {
	LoadIndex() (*knowledge.Index, error)
	LoadArticle(string) (*knowledge.Article, error)
} = (*knowledge.Loader)(nil)
