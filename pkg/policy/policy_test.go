package policy

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestViolationsByDomainUsesOverridesAndFallback(t *testing.T) {
	report := &finding.Report{Findings: []finding.Finding{
		{RuleID: "quality", Domain: finding.Quality, Severity: finding.Medium},
		{RuleID: "security", Domain: finding.Security, Severity: finding.High},
		{RuleID: "hardening", Domain: finding.Hardening, Severity: finding.Medium},
	}}
	overrides := map[finding.Domain]finding.Severity{
		finding.Quality:  finding.Low,
		finding.Security: finding.Critical,
	}

	violations := ViolationsByDomain(report, finding.High, overrides)
	if len(violations) != 1 || violations[0].RuleID != "quality" {
		t.Fatalf("unexpected violations: %+v", violations)
	}
}

func TestViolationsRetainsGlobalThresholdBehavior(t *testing.T) {
	report := &finding.Report{Findings: []finding.Finding{
		{Severity: finding.High},
		{Severity: finding.Medium},
	}}
	if got := len(Violations(report, finding.High)); got != 1 {
		t.Fatalf("expected one violation, got %d", got)
	}
}

func TestNewViolationsIgnoreExistingFindings(t *testing.T) {
	report := &finding.Report{Findings: []finding.Finding{
		{Severity: finding.Critical, BaselineState: finding.BaselineExisting},
		{Severity: finding.High, BaselineState: finding.BaselineNew},
	}}
	violations := NewViolationsByDomain(report, finding.High, nil)
	if len(violations) != 1 || violations[0].BaselineState != finding.BaselineNew {
		t.Fatalf("unexpected new violations: %+v", violations)
	}
}

func TestEvaluateReturnsStructuredDecision(t *testing.T) {
	report := &finding.Report{Findings: []finding.Finding{
		{Domain: finding.Security, Severity: finding.High},
		{Domain: finding.Quality, Severity: finding.Medium},
		{Domain: finding.Security, Severity: finding.Critical},
	}}
	decision := Evaluate(report, finding.High, map[finding.Domain]finding.Severity{
		finding.Quality: finding.Medium,
	}, false)
	if decision.Allowed || len(decision.Violations) != 3 || len(decision.Reasons) != 2 {
		t.Fatalf("unexpected decision: %+v", decision)
	}
	if decision.Reasons[0].Domain != finding.Quality || decision.Reasons[0].Threshold != finding.Medium || decision.Reasons[0].Count != 1 {
		t.Fatalf("unexpected deterministic reasons: %+v", decision.Reasons)
	}
}
