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
	if report.Findings[0].Domain != finding.Security {
		t.Fatalf("expected security domain, got %q", report.Findings[0].Domain)
	}
}

type failingScanner struct{ id string }

func (f failingScanner) ID() string { return f.id }

func (f failingScanner) Scan(context.Context, scanner.Request) scanner.Result {
	return scanner.Result{State: finding.ScannerFailed, Message: "fixture failure", Duration: time.Millisecond}
}

type panicScanner struct{ id string }

func (s panicScanner) ID() string { return s.id }

func (s panicScanner) Scan(context.Context, scanner.Request) scanner.Result {
	panic("disabled scanner was executed")
}

func TestDisabledScannerIsSkipped(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Scanners = map[string]config.Scanner{
		"disabled": {Enabled: false, Required: true},
	}
	reviewer, err := New(cfg, WithRequiredScanner(panicScanner{id: "disabled"}))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Scanners) != 2 || report.Scanners[1].State != finding.ScannerSkipped {
		t.Fatalf("expected skipped scanner status, got %+v", report.Scanners)
	}
}

type slowScanner struct{ id string }

func (s slowScanner) ID() string { return s.id }

func (s slowScanner) Scan(ctx context.Context, _ scanner.Request) scanner.Result {
	<-ctx.Done()
	time.Sleep(10 * time.Millisecond)
	return scanner.Result{State: finding.ScannerFailed, Message: ctx.Err().Error()}
}

func TestScannerTimeoutReturnsPartialReport(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Scanners = map[string]config.Scanner{
		"slow": {Enabled: true, Timeout: "5ms"},
	}
	reviewer, err := New(cfg, WithScanner(slowScanner{id: "slow"}))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatalf("optional timeout returned an operational error: %v", err)
	}
	if report.Scanners[1].State != finding.ScannerFailed || len(report.Warnings) != 1 {
		t.Fatalf("expected timeout failure and warning, got %+v", report)
	}
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
