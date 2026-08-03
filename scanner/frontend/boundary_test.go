package frontend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

func TestBoundaryAndServerOnlyEnforcement(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tempDir, "src/components"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, "src/server"), 0755)

	clientCode := `"use client"; import db from '../server/db'; export default function Component() {}`
	serverCode := `"use server"; export const db = { query: () => {} };`

	_ = os.WriteFile(filepath.Join(tempDir, "src/components/Card.tsx"), []byte(clientCode), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "src/server/db.ts"), []byte(serverCode), 0644)

	cfg := config.Default()
	cfg.Root = tempDir
	cfg.Frontend.Enabled = true
	cfg.Frontend.ServerRoots = []string{"src/server"}

	checker := NewBoundaryChecker(cfg)
	src := mockSource(filepath.Join(tempDir, "src/components/Card.tsx"), clientCode)

	findings, err := checker.Check(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) == 0 {
		t.Fatalf("expected boundary violation finding, got 0")
	}

	if findings[0].RuleID != "frontend/client-server-boundary-violation" {
		t.Errorf("expected ruleID frontend/client-server-boundary-violation, got %s", findings[0].RuleID)
	}
}
