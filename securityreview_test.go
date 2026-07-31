package securityreview

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
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

type failingScanner struct{ id string }

func (f failingScanner) ID() string { return f.id }

func (f failingScanner) Scan(context.Context, scanner.Request) scanner.Result {
	return scanner.Result{State: finding.ScannerFailed, Message: "fixture failure", Duration: time.Millisecond}
}

func TestOptionalScannerFailureReturnsReportAndWarning(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	reviewer, err := New(cfg, WithScanner(failingScanner{id: "optional"}))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatalf("optional failure returned error: %v", err)
	}
	if report == nil || len(report.Warnings) != 1 {
		t.Fatalf("expected report warning, got %+v", report)
	}
}

func TestRequiredScannerFailureReturnsReportAndError(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	reviewer, err := New(cfg, WithRequiredScanner(failingScanner{id: "required"}))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err == nil || report == nil {
		t.Fatalf("expected report and operational error, report=%v err=%v", report, err)
	}
}
