package frontend

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
)

func TestResolverAndAliasResolution(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory structure
	_ = os.MkdirAll(filepath.Join(tempDir, "src/components"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "src/components/Card.tsx"), []byte("export const Card = {};"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "src/components/index.ts"), []byte("export * from './Card';"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "src/utils.ts"), []byte("export const util = 1;"), 0644)

	// Create tsconfig.json
	tsconfig := `{
		"compilerOptions": {
			"baseUrl": ".",
			"paths": {
				"@/*": ["src/*"]
			}
		}
	}`
	_ = os.WriteFile(filepath.Join(tempDir, "tsconfig.json"), []byte(tsconfig), 0644)

	cfg := config.Default()
	cfg.Root = tempDir

	resolver := NewResolver(cfg)

	// Test 1: Relative file import
	res, ok := resolver.Resolve(filepath.Join(tempDir, "src/app.ts"), "./utils")
	if !ok || res != "src/utils.ts" {
		t.Errorf("expected src/utils.ts, got %s (ok=%v)", res, ok)
	}

	// Test 2: Alias import via @/
	res, ok = resolver.Resolve(filepath.Join(tempDir, "src/app.ts"), "@/components/Card")
	if !ok || res != "src/components/Card.tsx" {
		t.Errorf("expected src/components/Card.tsx, got %s (ok=%v)", res, ok)
	}

	// Test 3: Index file resolution
	res, ok = resolver.Resolve(filepath.Join(tempDir, "src/app.ts"), "./components")
	if !ok || res != "src/components/index.ts" {
		t.Errorf("expected src/components/index.ts, got %s (ok=%v)", res, ok)
	}

	// Test 4: Unsafe path outside root rejected
	_, ok = resolver.Resolve(filepath.Join(tempDir, "src/app.ts"), "../../../etc/passwd")
	if ok {
		t.Errorf("expected unsafe path to be rejected")
	}
}
