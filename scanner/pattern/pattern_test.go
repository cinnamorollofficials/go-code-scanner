package pattern

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
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
