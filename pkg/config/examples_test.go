package config_test

import (
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
)

func TestValidateConfigurationRecipes(t *testing.T) {
	recipeFiles, err := filepath.Glob("../../examples/config/*.json")
	if err != nil {
		t.Fatalf("failed to glob recipe files: %v", err)
	}
	if len(recipeFiles) == 0 {
		t.Fatalf("expected to find configuration recipes in examples/config/, found 0")
	}

	expectedRecipes := map[string]bool{
		"minimal.json":          false,
		"go-service.json":       false,
		"frontend-app.json":     false,
		"monorepo.json":         false,
		"staged-hook.json":      false,
		"offline.json":          false,
		"strict-ci.json":        false,
		"external-scanner.json": false,
		"gradual-adoption.json": false,
	}

	for _, file := range recipeFiles {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			cfg, err := config.Load(file)
			if err != nil {
				t.Fatalf("failed to strictly decode recipe %s: %v", name, err)
			}
			if cfg.Version != 1 {
				t.Errorf("expected version 1 in %s, got %d", name, cfg.Version)
			}
			expectedRecipes[name] = true
		})
	}

	for name, found := range expectedRecipes {
		if !found {
			t.Errorf("missing expected configuration recipe: %s", name)
		}
	}
}
