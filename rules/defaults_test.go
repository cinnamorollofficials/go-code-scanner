package rules

import (
	"regexp"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestDefaultRulePacks(t *testing.T) {
	all := Default()
	if len(all) != 31 {
		t.Fatalf("expected 31 default rules, got %d", len(all))
	}

	wantCounts := map[finding.Domain]int{
		finding.Quality:     5,
		finding.Security:    14,
		finding.Hardening:   6,
		finding.Reliability: 6,
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

type ruleExample struct {
	positive string
	negative string
}

func testRuleExamples(t *testing.T, rules []Rule, examples map[string]ruleExample) {
	t.Helper()
	if len(rules) != len(examples) {
		t.Fatalf("expected examples for all %d rules, got %d", len(rules), len(examples))
	}
	for _, rule := range rules {
		example, ok := examples[rule.ID]
		if !ok {
			t.Errorf("missing example for rule %q", rule.ID)
			continue
		}
		expression, err := regexp.Compile("(?i)" + rule.Pattern)
		if err != nil {
			t.Errorf("compile rule %q: %v", rule.ID, err)
			continue
		}
		if !expression.MatchString(example.positive) {
			t.Errorf("rule %q did not match positive example %q", rule.ID, example.positive)
		}
		if expression.MatchString(example.negative) {
			t.Errorf("rule %q matched negative example %q", rule.ID, example.negative)
		}
	}
}

func TestDefaultRulesCompile(t *testing.T) {
	if _, err := Compile(Default()); err != nil {
		t.Fatal(err)
	}
}
