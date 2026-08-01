package rules

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestCompileDefaultsLegacyRuleToSecurityDomain(t *testing.T) {
	compiled, err := Compile([]Rule{{
		ID: "legacy", Pattern: "secret", Severity: finding.High,
		Category: "secret_leak", Description: "legacy rule",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if compiled[0].Domain != finding.Security {
		t.Fatalf("expected security domain, got %q", compiled[0].Domain)
	}
}

func TestCompileRejectsInvalidDomain(t *testing.T) {
	_, err := Compile([]Rule{{
		ID: "invalid", Pattern: "value", Severity: finding.Low,
		Domain: finding.Domain("performance"), Category: "style", Description: "invalid domain",
	}})
	if err == nil {
		t.Fatal("expected invalid domain error")
	}
}
