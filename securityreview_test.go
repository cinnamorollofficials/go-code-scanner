package securityreview

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

type barrierScanner struct {
	id      string
	started chan<- string
	release <-chan struct{}
}

func (s barrierScanner) ID() string { return s.id }

func (s barrierScanner) Scan(ctx context.Context, _ scanner.Request) scanner.Result {
	s.started <- s.id
	select {
	case <-s.release:
		return scanner.Result{State: finding.ScannerClean}
	case <-ctx.Done():
		return scanner.Result{State: finding.ScannerFailed, Message: ctx.Err().Error()}
	}
}

func TestScannersRunConcurrentlyAndKeepRegistrationOrder(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Workers = 2
	started := make(chan string, 2)
	release := make(chan struct{})
	reviewer, err := New(cfg,
		WithScanner(barrierScanner{id: "first", started: started, release: release}),
		WithScanner(barrierScanner{id: "second", started: started, release: release}),
	)
	if err != nil {
		t.Fatal(err)
	}

	type runResult struct {
		report *finding.Report
		err    error
	}
	completed := make(chan runResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() {
		report, runErr := reviewer.Run(ctx)
		completed <- runResult{report: report, err: runErr}
	}()

	seen := map[string]bool{}
	for range 2 {
		select {
		case id := <-started:
			seen[id] = true
		case <-ctx.Done():
			t.Fatal("scanners did not reach the barrier concurrently")
		}
	}
	close(release)
	result := <-completed
	if result.err != nil {
		t.Fatal(result.err)
	}
	if !seen["first"] || !seen["second"] {
		t.Fatalf("unexpected scanners reached barrier: %v", seen)
	}
	if got := []string{result.report.Scanners[1].ID, result.report.Scanners[2].ID}; got[0] != "first" || got[1] != "second" {
		t.Fatalf("scanner statuses are not deterministic: %v", got)
	}
}

func TestScannerPanicBecomesOptionalFailure(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	reviewer, err := New(cfg, WithScanner(panicScanner{id: "panic"}))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatalf("optional panic returned operational error: %v", err)
	}
	if report.Scanners[1].State != finding.ScannerFailed || len(report.Warnings) != 1 {
		t.Fatalf("expected failed scanner and warning, got %+v", report)
	}
	if !strings.Contains(report.Scanners[1].Message, "panic") {
		t.Fatalf("expected panic message, got %q", report.Scanners[1].Message)
	}
}

func TestSelectedProfileSkipsScannerOutsideProfile(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.SelectedProfile = config.ProfileFast
	reviewer, err := New(cfg, WithRequiredScanner(panicScanner{id: "not-fast"}))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.Scanners[1].State != finding.ScannerSkipped {
		t.Fatalf("expected scanner outside profile to be skipped: %+v", report.Scanners)
	}
}

func TestReviewerRegistersConfiguredCommandScannersInIDOrder(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Scanners = map[string]config.Scanner{
		"z-command": configuredCommandScanner("clean"),
		"a-command": configuredCommandScanner("findings"),
	}
	reviewer, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Scanners) != 3 || report.Scanners[1].ID != "a-command" || report.Scanners[2].ID != "z-command" {
		t.Fatalf("configured scanner order is not deterministic: %+v", report.Scanners)
	}
	if len(report.Findings) != 1 || report.Findings[0].Tool != "a-command" {
		t.Fatalf("expected normalized command finding, got %+v", report.Findings)
	}
}

func configuredCommandScanner(mode string) config.Scanner {
	return config.Scanner{
		Enabled: true, Type: "command", Domain: finding.Quality,
		Command:          []string{os.Args[0], "-test.run=TestConfiguredCommandHelperProcess", "--", mode},
		FindingExitCodes: []int{10}, Severity: finding.High,
		Category: "fixture", Description: "configured command finding",
	}
}

func TestConfiguredCommandHelperProcess(t *testing.T) {
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		return
	}
	if os.Args[separator+1] == "findings" {
		os.Exit(10)
	}
	os.Exit(0)
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
