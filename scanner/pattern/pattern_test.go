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

func TestScanAppliesOptionalQualitySizePolicies(t *testing.T) {
	limits := Limits{
		MaxFileBytes: 1024, MaxLineBytes: 1024,
		QualityMaxFileBytes: 12, QualityMaxLineLength: 8,
	}
	source := memorySource("/repo/app.go", "package fixture\n")
	result := New(nil, 1, limits).Scan(context.Background(), scanner.Request{
		Root: "/repo", Mode: "full", Sources: []scanner.Source{source},
	})
	if result.State != finding.ScannerFindings || len(result.Findings) != 2 {
		t.Fatalf("unexpected quality size policy result: %+v", result)
	}
	ids := map[string]bool{}
	for _, item := range result.Findings {
		ids[item.RuleID] = true
	}
	if !ids["line-length"] || !ids["source-file-size"] {
		t.Fatalf("missing quality policy findings: %v", ids)
	}
}

func TestScanAppliesOfflineSupplyChainPolicies(t *testing.T) {
	files := []scanner.Source{
		memorySource("/repo/package.json", `{"dependencies":{"unsafe":"latest","safe":"^1.2.3"}}`),
		memorySource("/repo/go.mod", "module example.test/app\nreplace example.test/lib => ../lib\n"),
		memorySource("/repo/Dockerfile", "FROM alpine:latest\nUSER app\n"),
		memorySource("/repo/.github/workflows/ci.yml", "steps:\n  - uses: actions/checkout@v4\n"),
	}
	result := New(nil, 1).Scan(context.Background(), scanner.Request{Root: "/repo", Mode: "staged", Files: files})
	if result.State != finding.ScannerFindings || len(result.Findings) != 5 {
		t.Fatalf("unexpected supply-chain result: %+v", result)
	}
	want := map[string]bool{
		"manifest-without-lockfile": true, "javascript-unpinned-dependency": true,
		"go-local-replace": true, "docker-latest-tag": true, "github-action-mutable-ref": true,
	}
	for _, item := range result.Findings {
		if item.Domain != finding.SupplyChain || !want[item.RuleID] {
			t.Fatalf("unexpected supply-chain finding: %+v", item)
		}
		delete(want, item.RuleID)
	}
	if len(want) != 0 {
		t.Fatalf("missing supply-chain findings: %v", want)
	}
}

func TestOfflineSupplyChainPoliciesAcceptPinnedInputs(t *testing.T) {
	sha := strings.Repeat("a", 40)
	files := []scanner.Source{
		memorySource("/repo/package.json", `{"dependencies":{"safe":"^1.2.3"}}`),
		memorySource("/repo/package-lock.json", `{}`),
		memorySource("/repo/go.mod", "module example.test/app\nrequire example.test/lib v1.2.3\n"),
		memorySource("/repo/Dockerfile", "FROM alpine:3.22\nUSER app\n"),
		memorySource("/repo/.github/workflows/ci.yml", "steps:\n  - uses: actions/checkout@"+sha+"\n"),
	}
	result := New(nil, 1).Scan(context.Background(), scanner.Request{Root: "/repo", Mode: "staged", Files: files})
	if result.State != finding.ScannerClean || len(result.Findings) != 0 {
		t.Fatalf("pinned inputs produced findings: %+v", result)
	}
}

func TestConfigurableDependencyAndLicensePolicies(t *testing.T) {
	limits := Limits{
		MaxFileBytes: 1024 * 1024, MaxLineBytes: 1024 * 1024,
		DependencyAllowlist: []string{"@approved/*"}, DependencyDenylist: []string{"unsafe-*"},
		LicenseAllowlist: []string{"MIT", "Apache-*"}, LicenseDenylist: []string{"GPL-*"},
	}
	files := []scanner.Source{
		memorySource("/repo/package.json", `{"dependencies":{"@approved/core":"1.0.0","unsafe-lib":"1.0.0","unreviewed":"1.0.0"}}`),
		memorySource("/repo/package-lock.json", `{"packages":{"node_modules/ok":{"name":"ok","license":"MIT"},"node_modules/bad":{"name":"bad","license":"GPL-3.0"},"node_modules/unknown":{"name":"unknown","license":"BSD-3-Clause"}}}`),
	}
	result := New(nil, 1, limits).Scan(context.Background(), scanner.Request{Root: "/repo", Mode: "full", Files: files})
	counts := map[string]int{}
	for _, item := range result.Findings {
		counts[item.RuleID]++
	}
	if counts["dependency-policy"] != 2 || counts["dependency-license-policy"] != 2 {
		t.Fatalf("unexpected configurable policy findings: %+v", result.Findings)
	}
}

func TestRequiredRepositoryFilePolicyUsesCompleteInventory(t *testing.T) {
	limits := Limits{MaxFileBytes: 1024, MaxLineBytes: 1024, RequiredFiles: []string{"SECURITY.md", "CODEOWNERS"}}
	repositoryFiles := []scanner.Source{memorySource("/repo/SECURITY.md", "policy")}
	result := New(nil, 1, limits).Scan(context.Background(), scanner.Request{
		Root: "/repo", Mode: "staged", RepositoryFiles: repositoryFiles,
	})
	if result.State != finding.ScannerFindings || len(result.Findings) != 1 {
		t.Fatalf("unexpected required-file result: %+v", result)
	}
	item := result.Findings[0]
	if item.RuleID != "required-file-missing" || item.Domain != finding.Governance || item.Location.File != "CODEOWNERS" {
		t.Fatalf("unexpected governance finding: %+v", item)
	}
}

func memorySource(path, content string) scanner.Source {
	return scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(content)), nil
	}}
}
