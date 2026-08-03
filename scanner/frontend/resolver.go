package frontend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

type Resolver struct {
	rootDir   string
	baseURL   string
	aliases   map[string][]string
	workspaces map[string]string
}

type TSConfig struct {
	CompilerOptions struct {
		BaseURL string              `json:"baseUrl"`
		Paths   map[string][]string `json:"paths"`
	} `json:"compilerOptions"`
}

type PackageJSON struct {
	Name       string      `json:"name"`
	Workspaces interface{} `json:"workspaces"`
}

func NewResolver(cfg config.Config) *Resolver {
	root := cfg.Root
	if root == "" {
		root = "."
	}
	r := &Resolver{
		rootDir:    filepath.Clean(root),
		aliases:    make(map[string][]string),
		workspaces: make(map[string]string),
	}
	r.loadTSConfig()
	r.loadWorkspaces()
	return r
}

func (r *Resolver) loadTSConfig() {
	for _, name := range []string{"tsconfig.json", "jsconfig.json"} {
		p := filepath.Join(r.rootDir, name)
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var tsCfg TSConfig
		if err := json.Unmarshal(stripJSONComments(data), &tsCfg); err == nil {
			r.baseURL = tsCfg.CompilerOptions.BaseURL
			for pattern, targets := range tsCfg.CompilerOptions.Paths {
				prefix := strings.TrimSuffix(pattern, "*")
				var cleanTargets []string
				for _, target := range targets {
					cleanTargets = append(cleanTargets, strings.TrimSuffix(target, "*"))
				}
				r.aliases[prefix] = cleanTargets
			}
			break
		}
	}
}

func (r *Resolver) loadWorkspaces() {
	p := filepath.Join(r.rootDir, "package.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	var pkg PackageJSON
	_ = json.Unmarshal(data, &pkg)
}

func (r *Resolver) Resolve(fromFile, specifier string) (string, bool) {
	if specifier == "" {
		return "", false
	}

	supportedExts := []string{"", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts", ".vue", ".svelte", ".html"}

	if strings.HasPrefix(specifier, ".") {
		dir := filepath.Dir(fromFile)
		base := filepath.Join(dir, specifier)
		return r.resolveFileOrIndex(base, supportedExts)
	}

	for prefix, targets := range r.aliases {
		if strings.HasPrefix(specifier, prefix) {
			rest := strings.TrimPrefix(specifier, prefix)
			for _, target := range targets {
				baseDir := r.rootDir
				if r.baseURL != "" {
					baseDir = filepath.Join(r.rootDir, r.baseURL)
				}
				base := filepath.Join(baseDir, target, rest)
				if resolved, ok := r.resolveFileOrIndex(base, supportedExts); ok {
					return resolved, true
				}
			}
		}
	}

	if strings.HasPrefix(specifier, "@/") || strings.HasPrefix(specifier, "~/") {
		rest := specifier[2:]
		for _, baseDirName := range []string{"src", "."} {
			base := filepath.Join(r.rootDir, baseDirName, rest)
			if resolved, ok := r.resolveFileOrIndex(base, supportedExts); ok {
				return resolved, true
			}
		}
	}

	if strings.HasPrefix(specifier, "$lib/") {
		rest := specifier[5:]
		base := filepath.Join(r.rootDir, "src/lib", rest)
		if resolved, ok := r.resolveFileOrIndex(base, supportedExts); ok {
			return resolved, true
		}
	}

	return "", false
}

func (r *Resolver) resolveFileOrIndex(base string, exts []string) (string, bool) {
	cleanBase := filepath.Clean(base)

	rel, err := filepath.Rel(r.rootDir, cleanBase)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", false
	}

	for _, ext := range exts {
		candidate := cleanBase + ext
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			if rRel, err := filepath.Rel(r.rootDir, candidate); err == nil {
				return filepath.ToSlash(filepath.Clean(rRel)), true
			}
		}
	}

	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx"} {
		candidate := filepath.Join(cleanBase, "index"+ext)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			if rRel, err := filepath.Rel(r.rootDir, candidate); err == nil {
				return filepath.ToSlash(filepath.Clean(rRel)), true
			}
		}
	}

	return "", false
}

func stripJSONComments(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	var clean []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		clean = append(clean, l)
	}
	return []byte(strings.Join(clean, "\n"))
}
