package securityreview

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/baseline"
	cachepkg "github.com/cinnamorollofficials/go-code-scanner/cache"
	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/discovery"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
	patternscanner "github.com/cinnamorollofficials/go-code-scanner/scanner/pattern"
)

func BenchmarkDiscovery(b *testing.B) {
	root := b.TempDir()
	for index := range 100 {
		path := filepath.Join(root, fmt.Sprintf("pkg%03d", index), "file.go")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package fixture\n"), 0o600); err != nil {
			b.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.Root = root
	b.ReportAllocs()
	for b.Loop() {
		if sources, err := discovery.Sources(context.Background(), cfg); err != nil || len(sources) != 100 {
			b.Fatalf("discovery sources=%d err=%v", len(sources), err)
		}
	}
}

func BenchmarkPatternScanning(b *testing.B) {
	compiled, err := rules.Compile([]rules.Rule{{
		ID: "budget", Pattern: "unsafe", Severity: finding.High, Category: "budget", Description: "budget fixture",
	}})
	if err != nil {
		b.Fatal(err)
	}
	sources := make([]scanner.Source, 100)
	content := strings.Repeat("safe line\n", 100)
	for index := range sources {
		path := fmt.Sprintf("/repo/pkg%03d/file.go", index)
		sources[index] = scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(content)), nil
		}}
	}
	source := patternscanner.New(compiled, 4)
	request := scanner.Request{Root: "/repo", Mode: "full", Sources: sources}
	b.ReportAllocs()
	for b.Loop() {
		if result := source.Scan(context.Background(), request); result.State != finding.ScannerClean {
			b.Fatalf("unexpected pattern result: %+v", result)
		}
	}
}

func BenchmarkBaselineComparison(b *testing.B) {
	findings := make([]finding.Finding, 1000)
	entries := make([]baseline.Entry, 1000)
	for index := range findings {
		fingerprint := fmt.Sprintf("fingerprint-%04d", index)
		findings[index] = finding.Finding{Fingerprint: fingerprint, RuleID: "budget", Domain: finding.Quality, Location: finding.Location{File: fmt.Sprintf("file-%04d.go", index)}}
		entries[index] = baseline.Entry{Fingerprint: fingerprint, RuleID: "budget", Domain: finding.Quality, File: findings[index].Location.File}
	}
	file := &baseline.File{Version: baseline.Version, FingerprintVersion: "3", Entries: entries}
	b.ReportAllocs()
	for b.Loop() {
		report := &finding.Report{FingerprintVersion: "3", Findings: append([]finding.Finding(nil), findings...)}
		if _, err := baseline.Compare(report, file); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheHit(b *testing.B) {
	store := cachepkg.Store{Directory: b.TempDir()}
	key, err := cachepkg.Key(cachepkg.KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err != nil {
		b.Fatal(err)
	}
	if err := store.Put(key, scanner.Result{State: finding.ScannerClean}); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, found, err := store.Get(key); err != nil || !found {
			b.Fatalf("cache hit found=%t err=%v", found, err)
		}
	}
}

func BenchmarkFastPreCommit(b *testing.B) {
	root := b.TempDir()
	performanceRunGit(b, root, "init")
	for index := range 20 {
		name := fmt.Sprintf("pkg/file%02d.go", index)
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package fixture\n"), 0o600); err != nil {
			b.Fatal(err)
		}
		performanceRunGit(b, root, "add", name)
	}
	cfg := config.Default()
	cfg.Root, cfg.Mode, cfg.SelectedProfile = root, config.ModeStaged, config.ProfileFast
	reviewer, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if report, err := reviewer.Run(context.Background()); err != nil || report == nil {
			b.Fatalf("fast pre-commit report=%v err=%v", report, err)
		}
	}
}

func performanceRunGit(tb testing.TB, root string, args ...string) {
	tb.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		tb.Fatalf("git %v: %v: %s", args, err, output)
	}
}
