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
