package frontend

import (
	"testing"
)

func TestRuleRegistryLookupAndDefinitions(t *testing.T) {
	rule, ok := LookupRule("frontend/dom-injection")
	if !ok {
		t.Fatal("expected frontend/dom-injection rule to exist in registry")
	}
	if rule.ID != "frontend/dom-injection" {
		t.Fatalf("unexpected rule ID: %s", rule.ID)
	}
	if !rule.Domain.Valid() {
		t.Fatalf("invalid domain in rule metadata: %v", rule.Domain)
	}

	defs := RuleDefinitions()
	if len(defs) != len(Registry) {
		t.Fatalf("expected %d rules, got %d", len(Registry), len(defs))
	}
}
