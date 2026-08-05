package securityreview

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	cachepkg "github.com/cinnamorollofficials/go-code-scanner/pkg/cache"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/gitrepo"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
	commandscanner "github.com/cinnamorollofficials/go-code-scanner/pkg/scanner/command"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/workspace"
)

func TestResourceBoundariesEndToEnd(t *testing.T) {
	t.Run("oversized source", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "large.go"), []byte(strings.Repeat("x", 128)), 0o600); err != nil {
			t.Fatal(err)
		}
		cfg := config.Default()
		cfg.Root = root
		cfg.PatternMaxFileBytes = 16
		reviewer, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		report, runErr := reviewer.Run(context.Background())
		if report == nil || runErr == nil || report.Scanners[0].FailureKind != string(scanner.FailurePartial) {
			t.Fatalf("oversized source did not produce a bounded partial report: report=%+v err=%v", report, runErr)
		}
	})

	t.Run("snapshot file count and cleanup", func(t *testing.T) {
		temporaryRoot := t.TempDir()
		t.Setenv("TMPDIR", temporaryRoot)
		root := t.TempDir()
		resourceRunGit(t, root, "init")
		for _, name := range []string{"one.go", "two.go"} {
			if err := os.WriteFile(filepath.Join(root, name), []byte("package fixture\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			resourceRunGit(t, root, "add", name)
		}
		repository, err := gitrepo.Open(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := workspace.MaterializeIndex(context.Background(), repository, workspace.Limits{MaxFiles: 1, MaxBytes: 1024}); err == nil || !strings.Contains(err.Error(), "file limit") {
			t.Fatalf("expected staged file-count boundary, got %v", err)
		}
		entries, err := os.ReadDir(temporaryRoot)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "security-review-index-") {
				t.Fatalf("failed snapshot leaked %s", entry.Name())
			}
		}
	})

	t.Run("external output limit", func(t *testing.T) {
		t.Setenv("RESOURCE_BOUNDARY_HELPER", "output")
		source, err := commandscanner.New(commandscanner.Spec{
			ID: "bounded", Domain: finding.Security,
			Command:  []string{os.Args[0], "-test.run=TestResourceBoundaryHelperProcess"},
			Severity: finding.High, Category: "boundary", Description: "boundary fixture",
			OutputFormat: commandscanner.OutputJSONLines, MaxOutputBytes: 16,
			Environment: []string{"RESOURCE_BOUNDARY_HELPER"},
		})
		if err != nil {
			t.Fatal(err)
		}
		result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir()})
		if result.State != finding.ScannerFailed || !strings.Contains(result.Message, "exceeded configured limit") {
			t.Fatalf("command output boundary was not enforced: %+v", result)
		}
	})

	t.Run("timeout and cancellation", func(t *testing.T) {
		cfg := config.Default()
		cfg.Root = t.TempDir()
		cfg.Scanners = map[string]config.Scanner{"slow": {Enabled: true, Timeout: "5ms"}}
		reviewer, err := New(cfg, WithScanner(slowScanner{id: "slow"}))
		if err != nil {
			t.Fatal(err)
		}
		report, err := reviewer.Run(context.Background())
		if err != nil || report.Scanners[1].FailureKind != string(scanner.FailureTimeout) {
			t.Fatalf("timeout boundary was not classified: report=%+v err=%v", report, err)
		}

		started := make(chan struct{}, 1)
		cfg.Scanners = nil
		reviewer, err = New(cfg, WithScanner(cancelScanner{id: "cancel", started: started}))
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		completed := make(chan *finding.Report, 1)
		go func() {
			result, _ := reviewer.Run(ctx)
			completed <- result
		}()
		<-started
		cancel()
		report = <-completed
		if report.Scanners[1].FailureKind != string(scanner.FailureCanceled) {
			t.Fatalf("cancellation boundary was not classified: %+v", report.Scanners[1])
		}
	})

	t.Run("cache retention exact byte boundary", func(t *testing.T) {
		directory := t.TempDir()
		store := cachepkg.Store{Directory: directory}
		for _, version := range []string{"1", "2"} {
			key, err := cachepkg.Key(cachepkg.KeyInput{ScannerID: "boundary", ScannerVersion: version})
			if err != nil {
				t.Fatal(err)
			}
			if err := store.Put(key, scanner.Result{Message: strings.Repeat(version, 32)}); err != nil {
				t.Fatal(err)
			}
		}
		stats, err := store.Stats()
		if err != nil {
			t.Fatal(err)
		}
		if removed, err := store.Prune(0, stats.Bytes); err != nil || removed != 0 {
			t.Fatalf("exact cache boundary removed entries: removed=%d err=%v", removed, err)
		}
		if removed, err := store.Prune(0, stats.Bytes-1); err != nil || removed != 1 {
			t.Fatalf("over-boundary cache did not prune oldest entry: removed=%d err=%v", removed, err)
		}
	})
}

func TestResourceBoundaryHelperProcess(t *testing.T) {
	if os.Getenv("RESOURCE_BOUNDARY_HELPER") != "output" {
		return
	}
	_, _ = os.Stdout.WriteString(strings.Repeat("x", 128))
	os.Exit(1)
}

func resourceRunGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
