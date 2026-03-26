// Package knowledge provides article loading, frontmatter parsing, and indexing
// for the embedded knowledge base. Articles are Markdown files stored in
// docs/knowledge/ and embedded in the binary at build time.
package knowledge

import (
	"bufio"
	"io/fs"
	"log/slog"
	"path"
	"strings"
)

// Article represents a single knowledge base article.
type Article struct {
	// Path is the article's URL-safe identifier relative to the knowledge root,
	// e.g. "networking/clos-fabric-fundamentals".
	Path string `json:"path"`

	// Title is from frontmatter; falls back to filename if absent.
	Title string `json:"title"`

	// Category is from frontmatter; falls back to "uncategorized".
	Category string `json:"category"`

	// Tags is from frontmatter; empty slice if absent.
	Tags []string `json:"tags"`

	// Content is the raw Markdown body (frontmatter stripped).
	Content string `json:"content,omitempty"`
}

// Index holds a catalogue of all articles without their full content.
type Index struct {
	Articles []*Article `json:"articles"`
}

// Loader reads articles from an embedded filesystem.
type Loader struct {
	fs fs.FS
}

// NewLoader creates a Loader that reads from fsys.
func NewLoader(fsys fs.FS) *Loader {
	return &Loader{fs: fsys}
}

// LoadIndex walks the filesystem and returns metadata for all .md files.
// Content is not included in index entries.
func (l *Loader) LoadIndex() (*Index, error) {
	var articles []*Article

	err := fs.WalkDir(l.fs, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if path.Ext(p) != ".md" {
			return nil
		}

		article, loadErr := l.loadArticle(p, false)
		if loadErr != nil {
			slog.Warn("failed to load knowledge article", "path", p, "err", loadErr)
			return nil // skip broken articles
		}
		articles = append(articles, article)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Index{Articles: articles}, nil
}

// LoadArticle loads a single article by its path (without .md extension).
// Returns nil if the article is not found.
func (l *Loader) LoadArticle(articlePath string) (*Article, error) {
	filePath := articlePath + ".md"
	return l.loadArticle(filePath, true)
}

// loadArticle reads the file at filePath and parses its frontmatter.
// If includeContent is true, the stripped Markdown body is included.
func (l *Loader) loadArticle(filePath string, includeContent bool) (*Article, error) {
	data, err := fs.ReadFile(l.fs, filePath)
	if err != nil {
		return nil, err
	}

	article := &Article{}

	// Derive the article path: strip leading directory prefix and .md extension.
	article.Path = strings.TrimSuffix(filePath, ".md")

	fm, body := splitFrontmatter(string(data))
	parseFrontmatter(fm, article)

	// Fallback title: filename stem (last path segment, underscores → spaces).
	if article.Title == "" {
		stem := path.Base(article.Path)
		article.Title = titlify(stem)
	}
	if article.Category == "" {
		article.Category = "uncategorized"
	}
	if article.Tags == nil {
		article.Tags = []string{}
	}

	if includeContent {
		article.Content = body
	}

	return article, nil
}

// splitFrontmatter splits YAML frontmatter (delimited by ---) from body text.
// Returns ("", full content) if no frontmatter is present.
func splitFrontmatter(content string) (frontmatter, body string) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", content
	}

	// Find closing ---
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}

	if end == -1 {
		return "", content
	}

	fm := strings.Join(lines[1:end], "\n")
	body = strings.Join(lines[end+1:], "\n")
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}
	return fm, body
}

// parseFrontmatter parses a minimal YAML subset (key: value, key: [a, b, c])
// into the article metadata fields.
func parseFrontmatter(fm string, a *Article) {
	scanner := bufio.NewScanner(strings.NewReader(fm))
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		switch key {
		case "title":
			a.Title = unquote(val)
		case "category":
			a.Category = unquote(val)
		case "tags":
			a.Tags = parseTags(val)
		}
	}
}

// parseTags parses an inline YAML sequence: [tag1, tag2, tag3]
func parseTags(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	if tags == nil {
		return []string{}
	}
	return tags
}

// unquote removes surrounding single or double quotes from a string.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// titlify converts a hyphenated filename stem into a title-cased string.
func titlify(s string) string {
	parts := strings.Split(s, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
