package policy

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
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
