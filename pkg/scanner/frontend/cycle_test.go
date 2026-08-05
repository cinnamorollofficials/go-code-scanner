package frontend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
)

func TestImportCycleDetection(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tempDir, "src"), 0755)

	// Create circular import between src/a.ts and src/b.ts
	codeA := `import { b } from './b'; export const a = 1;`
	codeB := `import { a } from './a'; export const b = 2;`

	_ = os.WriteFile(filepath.Join(tempDir, "src/a.ts"), []byte(codeA), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "src/b.ts"), []byte(codeB), 0644)

	cfg := config.Default()
	cfg.Root = tempDir
	cfg.Frontend.Enabled = true

	checker := NewCycleChecker(cfg)
	srcA := mockSource(filepath.Join(tempDir, "src/a.ts"), codeA)

	findings, err := checker.Check(context.Background(), srcA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) == 0 {
		t.Fatalf("expected cycle finding for src/a.ts, got 0")
	}

	if findings[0].RuleID != "frontend/import-cycle" {
		t.Errorf("expected ruleID frontend/import-cycle, got %s", findings[0].RuleID)
	}
}
