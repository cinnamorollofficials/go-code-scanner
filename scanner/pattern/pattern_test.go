package pattern

import (
	"context"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

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

func TestRedactionUsesRuleClassification(t *testing.T) {
	compiled, err := rules.Compile([]rules.Rule{
		{ID: "authorization", Pattern: "Authorization", Severity: finding.High, Category: "authorization", Description: "header"},
		{ID: "tagged", Pattern: "private", Severity: finding.High, Category: "logging", Description: "private data", Tags: []string{"sensitive"}},
		{ID: "tokenizer", Pattern: "tokenizer", Severity: finding.Low, Category: "quality", Description: "safe term"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := redact(compiled[0], `Authorization: Bearer actual-value`); got != "[REDACTED: potentially sensitive source line]" {
		t.Fatalf("authorization header leaked: %q", got)
	}
	if got := redact(compiled[1], `log.Info("private", customerRecord)`); got != "[REDACTED: potentially sensitive source line]" {
		t.Fatalf("tagged sensitive line leaked: %q", got)
	}
	if got := redact(compiled[2], `tokenizer := strings.NewReader(input)`); got != `tokenizer := strings.NewReader(input)` {
		t.Fatalf("safe token substring was over-redacted: %q", got)
	}
}

func TestRedactionTruncatesLongUnicodeSnippetSafely(t *testing.T) {
	compiled, err := rules.Compile([]rules.Rule{{
		ID: "long-line", Pattern: "x", Severity: finding.Low, Category: "quality", Description: "long line",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := redact(compiled[0], strings.Repeat("界", 250))
	if !strings.HasSuffix(got, "…") || len([]rune(got)) != 201 || !utf8.ValidString(got) {
		t.Fatalf("snippet was not safely truncated: rune_count=%d suffix=%q", len([]rune(got)), got[len(got)-3:])
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

func TestScanAppliesFileLevelPolicies(t *testing.T) {
	files := []scanner.Source{
		memorySource("/repo/debug.dump", "binary-ish fixture"),
		memorySource("/repo/Dockerfile", "FROM alpine\nUSER root\n"),
		memorySource("/repo/generated.go", "// Code generated by fixture. DO NOT EDIT.\npackage fixture\n"),
	}
	result := New(nil, 1).Scan(context.Background(), scanner.Request{Root: "/repo", Mode: "staged", Files: files})
	if result.State != finding.ScannerFindings || len(result.Findings) != 3 {
		t.Fatalf("unexpected file policy result: %+v", result)
	}
	want := map[string]finding.Domain{
		"temporary-artifact":     finding.Quality,
		"docker-root-user":       finding.Hardening,
		"generated-file-changed": finding.Quality,
	}
	for _, item := range result.Findings {
		if domain, ok := want[item.RuleID]; !ok || domain != item.Domain {
			t.Fatalf("unexpected file finding: %+v", item)
		}
		delete(want, item.RuleID)
	}
	if len(want) != 0 {
		t.Fatalf("missing file policy findings: %v", want)
	}
}

func memorySource(path, content string) scanner.Source {
	return scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(content)), nil
	}}
}
