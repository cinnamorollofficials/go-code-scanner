package rules

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestDefaultRulePacks(t *testing.T) {
	all := Default()
	if len(all) != 14 {
		t.Fatalf("expected 14 default rules, got %d", len(all))
	}

	wantCounts := map[finding.Domain]int{
		finding.Security:    12,
		finding.Hardening:   1,
		finding.Reliability: 1,
	}
	counts := make(map[finding.Domain]int)
	ids := make(map[string]struct{}, len(all))
	for _, rule := range all {
		if !rule.Domain.Valid() {
			t.Fatalf("rule %q has invalid domain %q", rule.ID, rule.Domain)
		}
		if _, exists := ids[rule.ID]; exists {
			t.Fatalf("duplicate default rule ID %q", rule.ID)
		}
		ids[rule.ID] = struct{}{}
		counts[rule.Domain]++
	}

	for domain, want := range wantCounts {
		if counts[domain] != want {
			t.Fatalf("expected %d %s rules, got %d", want, domain, counts[domain])
		}
	}
}

func TestDefaultRulesCompile(t *testing.T) {
	if _, err := Compile(Default()); err != nil {
		t.Fatal(err)
	}
}
