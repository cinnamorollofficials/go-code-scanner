//go:generate go run ../../cmd/gen-config-doc/main.go

package config

import (
	"encoding/json"
	"fmt"
	"sort"
)

type FieldMetadata struct {
	Path          string   `json:"path"`
	Type          string   `json:"type"`
	Default       any      `json:"default"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	Required      bool     `json:"required"`
	Description   string   `json:"description"`
	Version       int      `json:"version"`
}

type MetadataSchema struct {
	Version int             `json:"version"`
	Fields  []FieldMetadata `json:"fields"`
}

func GenerateMetadata() (MetadataSchema, error) {
	fields := []FieldMetadata{
		{Path: "version", Type: "int", Default: 1, Required: true, Description: "Configuration schema version (must be 1).", Version: 1},
		{Path: "project", Type: "string", Default: "security-review", Required: false, Description: "Project or repository identifier.", Version: 1},
		{Path: "root", Type: "string", Default: ".", Required: false, Description: "Target root workspace directory.", Version: 1},
		{Path: "mode", Type: "string", Default: "full", AllowedValues: []string{"full", "changed", "staged"}, Required: false, Description: "Default scan discovery mode.", Version: 1},
		{Path: "output", Type: "string", Default: "security_findings.json", Required: false, Description: "Default report output filename.", Version: 1},
		{Path: "fail_on", Type: "string", Default: "CRITICAL", AllowedValues: []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"}, Required: false, Description: "Global severity threshold that triggers exit code 1.", Version: 1},
		{Path: "include_extensions", Type: "[]string", Default: []string{".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts", ".html", ".vue", ".svelte", ".yaml", ".yml", ".json"}, Required: false, Description: "List of file extensions included during discovery.", Version: 1},
		{Path: "exclude_directories", Type: "[]string", Default: []string{".git", "node_modules", "vendor", "dist", "build", ".next", ".nuxt", ".svelte-kit", ".output", "out", "bin"}, Required: false, Description: "Directory names or paths excluded during discovery.", Version: 1},
		{Path: "exclude_files", Type: "[]string", Default: []string{"security_findings.json", "package-lock.json"}, Required: false, Description: "Filenames excluded during discovery.", Version: 1},
		{Path: "rule_files", Type: "[]string", Default: []string{}, Required: false, Description: "External JSON rule file paths.", Version: 1},
		{Path: "suppression_file", Type: "string", Default: "", Required: false, Description: "Path to suppressions JSON file.", Version: 1},
		{Path: "baseline_file", Type: "string", Default: "", Required: false, Description: "Path to baseline snapshot JSON file.", Version: 1},
		{Path: "workers", Type: "int", Default: 4, Required: false, Description: "Maximum concurrent worker goroutines.", Version: 1},
		{Path: "pattern_max_file_bytes", Type: "int64", Default: 1048576, Required: false, Description: "Maximum file size in bytes for pattern scanner.", Version: 1},
		{Path: "pattern_max_line_bytes", Type: "int", Default: 4096, Required: false, Description: "Maximum line buffer length in bytes for pattern scanner.", Version: 1},
		{Path: "scanners", Type: "map[string]Scanner", Default: map[string]any{}, Required: false, Description: "Map of declared scanner definitions.", Version: 1},
		{Path: "profiles", Type: "map[string][]string", Default: map[string]any{}, Required: false, Description: "Custom named profile scanner mappings.", Version: 1},
		{Path: "offline_profiles", Type: "[]string", Default: []string{}, Required: false, Description: "List of profiles runnable offline without network access.", Version: 1},
		{Path: "hooks.pre_commit.enabled", Type: "bool", Default: false, Required: false, Description: "Enables pre-commit git hook execution.", Version: 1},
		{Path: "hooks.pre_commit.profile", Type: "string", Default: "fast", Required: false, Description: "Performance profile used by pre-commit hook.", Version: 1},
		{Path: "hooks.pre_commit.staged_only", Type: "bool", Default: false, Required: false, Description: "Restricts pre-commit scan to git staged index snapshot.", Version: 1},
		{Path: "hooks.pre_commit.new_only", Type: "bool", Default: false, Required: false, Description: "Evaluates pre-commit findings against baseline snapshot.", Version: 1},
		{Path: "frontend.enabled", Type: "bool", Default: false, Required: false, Description: "Enables native frontend AST and pattern scanning.", Version: 1},
		{Path: "frontend.frameworks", Type: "[]string", Default: []string{}, Required: false, Description: "Framework identifiers for frontend detection.", Version: 1},
		{Path: "frontend.client_roots", Type: "[]string", Default: []string{}, Required: false, Description: "Client-side root directory paths.", Version: 1},
		{Path: "frontend.server_roots", Type: "[]string", Default: []string{}, Required: false, Description: "Server-side root directory paths.", Version: 1},
		{Path: "cache.enabled", Type: "bool", Default: false, Required: false, Description: "Enables local AST and scan caching.", Version: 1},
		{Path: "cache.directory", Type: "string", Default: "", Required: false, Description: "Path to local cache storage directory.", Version: 1},
		{Path: "cache.max_age", Type: "string", Default: "", Required: false, Description: "Maximum cache entry retention duration string.", Version: 1},
		{Path: "cache.max_bytes", Type: "int64", Default: 0, Required: false, Description: "Maximum byte size threshold for local cache directory.", Version: 1},
	}

	seen := make(map[string]bool)
	for _, f := range fields {
		if seen[f.Path] {
			return MetadataSchema{}, fmt.Errorf("duplicate configuration field path in metadata: %s", f.Path)
		}
		seen[f.Path] = true
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})

	return MetadataSchema{
		Version: SchemaVersion,
		Fields:  fields,
	}, nil
}

func MetadataJSON() ([]byte, error) {
	meta, err := GenerateMetadata()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(meta, "", "  ")
}
