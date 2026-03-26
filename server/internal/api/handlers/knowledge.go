package handlers

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/rnwolfe/fabrik/server/internal/knowledge"
)

// KnowledgeLoader is the interface required by KnowledgeHandler.
type KnowledgeLoader interface {
	LoadIndex() (*knowledge.Index, error)
	LoadArticle(articlePath string) (*knowledge.Article, error)
}

// KnowledgeHandler handles HTTP requests for the knowledge base.
type KnowledgeHandler struct {
	loader KnowledgeLoader
}

// NewKnowledgeHandler returns a new KnowledgeHandler.
func NewKnowledgeHandler(loader KnowledgeLoader) *KnowledgeHandler {
	return &KnowledgeHandler{loader: loader}
}

// Index handles GET /api/knowledge.
// Returns the article index (metadata without content).
func (h *KnowledgeHandler) Index(w http.ResponseWriter, r *http.Request) {
	idx, err := h.loader.LoadIndex()
	if err != nil {
		slog.Error("load knowledge index", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, idx)
}

// Get handles GET /api/knowledge/{path...}.
// Returns the full article including Markdown content.
func (h *KnowledgeHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Extract the article path from the URL.
	// Pattern: /api/knowledge/{path} where path may contain slashes.
	rawPath := r.PathValue("path")
	if rawPath == "" {
		// Fallback: strip the fixed prefix manually.
		rawPath = strings.TrimPrefix(r.URL.Path, "/api/knowledge/")
	}
	rawPath = strings.TrimSuffix(rawPath, "/")

	if rawPath == "" {
		writeError(w, http.StatusBadRequest, "article path is required")
		return
	}

	// Reject paths containing ".." segments or starting with "/" to prevent
	// directory traversal. fs.ValidPath also enforces the io/fs path contract.
	if strings.HasPrefix(rawPath, "/") || strings.Contains(rawPath, "..") || !fs.ValidPath(rawPath) {
		writeError(w, http.StatusBadRequest, "invalid article path")
		return
	}

	article, err := h.loader.LoadArticle(rawPath)
	if err != nil {
		slog.Error("load knowledge article", "err", err, "path", rawPath)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if article == nil {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}

	writeJSON(w, http.StatusOK, article)
}
