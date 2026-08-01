package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestLoadRejectsUnknownFieldsAndTrailingDocuments(t *testing.T) {
	for _, content := range []string{
		`{"version":1,"rules":[],"unexpected":true}`,
		`{"version":1,"rules":[]} {"version":1,"rules":[]}`,
	} {
		path := filepath.Join(t.TempDir(), "rules.json")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load([]string{path}); err == nil {
			t.Fatalf("invalid rule file accepted: %s", content)
		}
	}
}

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

func TestCompilePreservesFindingMetadata(t *testing.T) {
	compiled, err := Compile([]Rule{{
		ID: "metadata", Pattern: "value", Severity: finding.Low, Domain: finding.Quality,
		Category: "style", Description: "metadata rule", Documentation: "https://example.com/rule",
		Tags: []string{"style", "automatic"}, Fixable: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	rule := compiled[0]
	if rule.Documentation == "" || len(rule.Tags) != 2 || !rule.Fixable {
		t.Fatalf("metadata was not preserved: %+v", rule)
	}
}

func TestCompileRejectsInvalidTags(t *testing.T) {
	for _, tags := range [][]string{{""}, {"same", "same"}} {
		_, err := Compile([]Rule{{
			ID: "tags", Pattern: "value", Severity: finding.Low, Domain: finding.Quality,
			Category: "style", Description: "tags rule", Tags: tags,
		}})
		if err == nil {
			t.Fatalf("expected invalid tags error for %v", tags)
		}
	}
}
