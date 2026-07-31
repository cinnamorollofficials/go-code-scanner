package securityreview

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

func TestReviewerRunsDefaultPatternScanner(t *testing.T) {
	root := t.TempDir()
	source := []byte("const token = 'google-mock-jwt-token'\n")
	if err := os.WriteFile(filepath.Join(root, "app.ts"), source, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = root
	cfg.Project = "fixture"
	reviewer, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Critical != 1 || len(report.Findings) != 1 {
		t.Fatalf("unexpected report: %+v", report.Summary)
	}
	if report.Findings[0].Snippet != "[REDACTED: mock-token]" {
		t.Fatalf("sensitive snippet was not redacted: %q", report.Findings[0].Snippet)
	}
}
