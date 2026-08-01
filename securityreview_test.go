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
	reviewer, err := New(cfg, WithToolVersion("test-version"))
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
	if report.ToolVersion != "test-version" || report.ConfigHash == "" || report.RuleSetHash == "" {
		t.Fatalf("missing report provenance: %+v", report)
	}
}

func TestReportHashesAreDeterministic(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	firstReviewer, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	secondReviewer, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	first, err := firstReviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := secondReviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first.ConfigHash != second.ConfigHash || first.RuleSetHash != second.RuleSetHash {
		t.Fatalf("report hashes are not deterministic: first=%+v second=%+v", first, second)
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

type describedScanner struct {
	id         string
	descriptor scanner.Descriptor
	result     scanner.Result
	called     bool
}

func (s *describedScanner) ID() string                   { return s.id }
func (s *describedScanner) Describe() scanner.Descriptor { return s.descriptor }
func (s *describedScanner) Scan(context.Context, scanner.Request) scanner.Result {
	s.called = true
	return s.result
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

type cancelScanner struct {
	id      string
	started chan<- struct{}
}

func (s cancelScanner) ID() string { return s.id }

func (s cancelScanner) Scan(ctx context.Context, _ scanner.Request) scanner.Result {
	s.started <- struct{}{}
	<-ctx.Done()
	return scanner.Result{
		State: finding.ScannerFailed, Message: ctx.Err().Error(), Failure: scanner.FailureCanceled,
	}
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
	if report.Scanners[1].FailureKind != string(scanner.FailureTimeout) {
		t.Fatalf("expected timeout failure kind, got %+v", report.Scanners[1])
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
	if report.Scanners[1].FailureKind != string(scanner.FailurePanic) {
		t.Fatalf("expected panic failure kind, got %+v", report.Scanners[1])
	}
}

func TestCallerCancellationIsClassified(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	started := make(chan struct{}, 1)
	reviewer, err := New(cfg, WithScanner(cancelScanner{id: "canceled", started: started}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type runResult struct {
		report *finding.Report
		err    error
	}
	completed := make(chan runResult, 1)
	go func() {
		report, runErr := reviewer.Run(ctx)
		completed <- runResult{report: report, err: runErr}
	}()
	<-started
	cancel()
	result := <-completed
	report, err := result.report, result.err
	if err != nil {
		t.Fatalf("optional cancellation returned an operational error: %v", err)
	}
	status := report.Scanners[1]
	if status.State != finding.ScannerFailed || status.FailureKind != string(scanner.FailureCanceled) {
		t.Fatalf("unexpected cancellation status: %+v", status)
	}
}

func TestPartialScannerGetsStructuredFailureKind(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	partial := &describedScanner{
		id: "partial",
		result: scanner.Result{
			State: finding.ScannerPartial, Message: "one input could not be read",
		},
	}
	reviewer, err := New(cfg, WithScanner(partial))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatalf("optional partial result returned an operational error: %v", err)
	}
	status := report.Scanners[1]
	if status.State != finding.ScannerPartial || status.FailureKind != string(scanner.FailurePartial) || len(report.Warnings) != 1 {
		t.Fatalf("unexpected partial status: %+v warnings=%v", status, report.Warnings)
	}
}

func TestUnsupportedScanModeSkipsScannerWithoutCallingIt(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Mode = "full"
	modeLimited := &describedScanner{
		id: "staged-only",
		descriptor: scanner.Descriptor{
			Domain: finding.Quality, SupportedModes: []string{"staged"},
		},
		result: scanner.Result{State: finding.ScannerClean},
	}
	reviewer, err := New(cfg, WithScanner(modeLimited))
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	status := report.Scanners[1]
	if status.State != finding.ScannerSkipped || modeLimited.called {
		t.Fatalf("mode-limited scanner was not skipped: status=%+v called=%t", status, modeLimited.called)
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

func TestReviewerRegistersConfiguredAdapter(t *testing.T) {
	cfg := config.Default()
	cfg.Root = t.TempDir()
	cfg.Scanners = map[string]config.Scanner{
		"format": {Enabled: false, Type: "adapter", Adapter: "gofmt"},
	}
	reviewer, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	report, err := reviewer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Scanners) != 2 || report.Scanners[1].ID != "format" || report.Scanners[1].State != finding.ScannerSkipped {
		t.Fatalf("configured adapter was not registered: %+v", report.Scanners)
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

func TestFingerprintIsStableAcrossLineMoves(t *testing.T) {
	base := finding.Finding{
		RuleID: "fixture", Tool: "pattern", Domain: finding.Security,
		Severity: finding.High, Description: "fixture finding", Snippet: "dangerousCall(input)",
		Location: finding.Location{File: "app.go", Line: 10},
	}
	moved := base
	moved.Location.Line = 200
	first := normalize([]finding.Finding{base})[0]
	second := normalize([]finding.Finding{moved})[0]
	if first.Fingerprint != second.Fingerprint {
		t.Fatalf("line move changed fingerprint: %s != %s", first.Fingerprint, second.Fingerprint)
	}
}

func TestFingerprintDistinguishesContentAndRepeatedOccurrences(t *testing.T) {
	items := []finding.Finding{
		{RuleID: "fixture", Domain: finding.Security, Severity: finding.High, Description: "finding", Snippet: "first()", Location: finding.Location{File: "app.go", Line: 1}},
		{RuleID: "fixture", Domain: finding.Security, Severity: finding.High, Description: "finding", Snippet: "second()", Location: finding.Location{File: "app.go", Line: 2}},
		{RuleID: "fixture", Domain: finding.Security, Severity: finding.High, Description: "finding", Snippet: "first()", Location: finding.Location{File: "app.go", Line: 3}},
	}
	normalized := normalize(items)
	seen := make(map[string]bool, len(normalized))
	for _, item := range normalized {
		if seen[item.Fingerprint] {
			t.Fatalf("duplicate fingerprint in %+v", normalized)
		}
		seen[item.Fingerprint] = true
	}
}

func TestNormalizeDeduplicatesExactFinding(t *testing.T) {
	item := finding.Finding{
		RuleID: "fixture", Domain: finding.Security, Severity: finding.High,
		Description: "finding", Location: finding.Location{File: "app.go", Line: 4},
	}
	if got := len(normalize([]finding.Finding{item, item})); got != 1 {
		t.Fatalf("expected one finding after deduplication, got %d", got)
	}
}

func TestSummaryCountsFindingsByDomain(t *testing.T) {
	items := []finding.Finding{
		{Domain: finding.Security, Severity: finding.High},
		{Domain: finding.Security, Severity: finding.Medium},
		{Domain: finding.Quality, Severity: finding.Low},
	}
	summary := summarize(items, nil, nil)
	if summary.ByDomain[finding.Security] != 2 || summary.ByDomain[finding.Quality] != 1 {
		t.Fatalf("unexpected domain summary: %+v", summary.ByDomain)
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
