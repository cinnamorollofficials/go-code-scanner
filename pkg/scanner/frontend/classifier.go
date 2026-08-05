package frontend

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type Scope string

const (
	ScopeClient  Scope = "client"
	ScopeServer  Scope = "server"
	ScopeShared  Scope = "shared"
	ScopeUnknown Scope = "unknown"
)

func (s Scope) Valid() bool {
	return s == ScopeClient || s == ScopeServer || s == ScopeShared || s == ScopeUnknown
}

// Classifier classifies a source file into a Scope based on config, file conventions, and directives.
type Classifier struct {
	cfg config.Config
}

func NewClassifier(cfg config.Config) *Classifier {
	return &Classifier{cfg: cfg}
}

func (c *Classifier) Classify(ctx context.Context, src scanner.Source) Scope {
	if ctx == nil || ctx.Err() != nil {
		return ScopeUnknown
	}

	relPath := src.Path
	if c.cfg.Root != "" {
		if rel, err := filepath.Rel(c.cfg.Root, src.Path); err == nil {
			relPath = rel
		}
	}
	cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))

	// 1. Explicit configured roots take precedence over automatic conventions
	if scope := c.classifyByConfiguredRoots(cleanRelPath); scope != ScopeUnknown {
		return scope
	}

	// 2. Framework file conventions
	if scope := c.classifyByFileConventions(cleanRelPath); scope != ScopeUnknown {
		return scope
	}

	// 3. Content inspection for directives ("use client", "use server")
	if scope := c.classifyByContent(ctx, src); scope != ScopeUnknown {
		return scope
	}

	// 4. Extension defaults
	ext := strings.ToLower(filepath.Ext(cleanRelPath))
	switch ext {
	case ".html", ".vue", ".svelte":
		return ScopeClient
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".mts", ".cts":
		return ScopeClient
	}

	return ScopeUnknown
}

func (c *Classifier) classifyByConfiguredRoots(cleanPath string) Scope {
	fp := c.cfg.Frontend
	for _, root := range fp.ClientRoots {
		if pathMatchesRoot(cleanPath, root) {
			return ScopeClient
		}
	}
	for _, root := range fp.ServerRoots {
		if pathMatchesRoot(cleanPath, root) {
			return ScopeServer
		}
	}
	for _, root := range fp.SharedRoots {
		if pathMatchesRoot(cleanPath, root) {
			return ScopeShared
		}
	}
	return ScopeUnknown
}

func pathMatchesRoot(path, root string) bool {
	cleanRoot := filepath.ToSlash(filepath.Clean(root))
	if cleanRoot == "." || cleanRoot == "" {
		return true
	}
	if path == cleanRoot || strings.HasPrefix(path, cleanRoot+"/") {
		return true
	}
	return false
}

func (c *Classifier) classifyByFileConventions(cleanPath string) Scope {
	base := filepath.Base(cleanPath)
	lowerBase := strings.ToLower(base)
	lowerPath := strings.ToLower(cleanPath)

	// File suffix conventions (.client.* vs .server.*)
	if strings.Contains(lowerBase, ".client.") {
		return ScopeClient
	}
	if strings.Contains(lowerBase, ".server.") {
		return ScopeServer
	}

	// SvelteKit conventions
	if strings.HasPrefix(lowerBase, "+page.server.") ||
		strings.HasPrefix(lowerBase, "+layout.server.") ||
		strings.HasPrefix(lowerBase, "+server.") {
		return ScopeServer
	}
	if strings.HasPrefix(lowerBase, "+page.") ||
		strings.HasPrefix(lowerBase, "+layout.") ||
		strings.HasPrefix(lowerBase, "+error.") {
		return ScopeClient
	}

	// Next.js & Nuxt server API route conventions
	if strings.Contains(lowerPath, "app/api/") ||
		strings.Contains(lowerPath, "pages/api/") ||
		strings.Contains(lowerPath, "server/api/") ||
		strings.Contains(lowerPath, "server/routes/") ||
		strings.Contains(lowerPath, "server/middleware/") {
		return ScopeServer
	}

	return ScopeUnknown
}

func (c *Classifier) classifyByContent(ctx context.Context, src scanner.Source) Scope {
	if src.Open == nil {
		return ScopeUnknown
	}
	rc, err := src.Open(ctx)
	if err != nil {
		return ScopeUnknown
	}
	defer rc.Close()

	buf := make([]byte, 4096)
	n, _ := io.ReadFull(rc, buf)
	header := buf[:n]

	if bytes.Contains(header, []byte(`"use client"`)) || bytes.Contains(header, []byte(`'use client'`)) {
		return ScopeClient
	}
	if bytes.Contains(header, []byte(`"use server"`)) || bytes.Contains(header, []byte(`'use server'`)) {
		return ScopeServer
	}

	return ScopeUnknown
}
