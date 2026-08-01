package pattern

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func TestRedactAlwaysHidesSecretLeakSource(t *testing.T) {
	compiled, err := rules.Compile([]rules.Rule{{
		ID: "credential", Pattern: "credential", Severity: finding.High,
		Category: "secret_leak", Description: "credential found",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := redact(compiled[0], `credential = "value-without-sensitive-keyword"`)
	if got != "[REDACTED: credential]" {
		t.Fatalf("got %q", got)
	}
}

func TestRulesAreIndexedByExtension(t *testing.T) {
	compiled, err := rules.Compile([]rules.Rule{
		{ID: "generic", Pattern: "generic", Severity: finding.Low, Category: "test", Description: "generic"},
		{ID: "go-only", Pattern: "go", Severity: finding.Low, Category: "test", Description: "go", Extensions: []string{".go"}},
		{ID: "ts-only", Pattern: "ts", Severity: finding.Low, Category: "test", Description: "ts", Extensions: []string{".ts"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := New(compiled, 2)
	got := s.rulesFor(".go")
	if len(got) != 2 || got[0].ID != "generic" || got[1].ID != "go-only" {
		t.Fatalf("unexpected indexed rules: %+v", got)
	}
}

func TestScanPropagatesRuleMetadata(t *testing.T) {
	compiled, err := rules.Compile([]rules.Rule{{
		ID: "metadata", Pattern: "unsafe", Severity: finding.High, Domain: finding.Quality,
		Category: "style", Description: "unsafe code", Documentation: "https://example.com/rule",
		Tags: []string{"style"}, Fixable: true, Extensions: []string{".go"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	source := scanner.Source{Path: "/repo/app.go", Open: func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("unsafe\n")), nil
	}}
	result := New(compiled, 1).Scan(context.Background(), scanner.Request{Root: "/repo", Sources: []scanner.Source{source}})
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %+v", result)
	}
	item := result.Findings[0]
	if item.Documentation == "" || len(item.Tags) != 1 || !item.Fixable {
		t.Fatalf("finding metadata was not propagated: %+v", item)
	}
}

func TestScanReturnsPartialForConfiguredInputLimits(t *testing.T) {
	compiled, err := rules.Compile([]rules.Rule{{
		ID: "fixture", Pattern: "unsafe", Severity: finding.High, Category: "test", Description: "fixture",
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		content string
		limits  Limits
		message string
	}{
		{name: "file", content: "safe\nsafe\nsafe\n", limits: Limits{MaxFileBytes: 5, MaxLineBytes: 100}, message: "pattern_max_file_bytes"},
		{name: "line", content: strings.Repeat("x", 20), limits: Limits{MaxFileBytes: 100, MaxLineBytes: 10}, message: "pattern_max_line_bytes"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := scanner.Source{Path: "/repo/app.go", Open: func(context.Context) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(test.content)), nil
			}}
			result := New(compiled, 1, test.limits).Scan(context.Background(), scanner.Request{Root: "/repo", Sources: []scanner.Source{source}})
			if result.State != finding.ScannerPartial || result.Failure != scanner.FailurePartial || !strings.Contains(result.Message, test.message) {
				t.Fatalf("unexpected limited scan result: %+v", result)
			}
		})
	}
}
