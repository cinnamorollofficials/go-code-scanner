package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestTerminalReportIsActionableAndDoesNotPrintSnippets(t *testing.T) {
	report := &finding.Report{
		Project: "fixture", ScanMode: "staged", Summary: finding.Summary{Total: 1, High: 1},
		Findings: []finding.Finding{{
			RuleID: "rule", Domain: finding.Security, Severity: finding.High,
			BaselineState: finding.BaselineNew, Description: "unsafe behavior",
			Recommendation: "use the safe API", Snippet: "TOP-SECRET-SNIPPET", Fixable: true, Fingerprint: "safe-fingerprint",
			Location: finding.Location{File: "app.go", Line: 12},
		}},
	}
	var output bytes.Buffer
	if err := WriteTerminal(&output, report); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, expected := range []string{
		"[HIGH] [new] security/rule", "app.go:12 unsafe behavior", "Fix: use the safe API",
		"Explain: security-review --explain rule", "Fix: security-review --fix",
		"Suppress: add fingerprint safe-fingerprint with reason and expiry to .security-ignore",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "TOP-SECRET-SNIPPET") {
		t.Fatalf("terminal leaked finding snippet:\n%s", text)
	}
}

func TestTerminalReportPrioritizesNewFindingsAndAppliesLimit(t *testing.T) {
	report := &finding.Report{Findings: []finding.Finding{
		{RuleID: "existing", Domain: finding.Quality, Severity: finding.Critical, BaselineState: finding.BaselineExisting},
		{RuleID: "new", Domain: finding.Security, Severity: finding.High, BaselineState: finding.BaselineNew},
	}}
	var output bytes.Buffer
	if err := WriteTerminalWithOptions(&output, report, TerminalOptions{MaxFindings: 1}); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if !strings.Contains(text, "security/new") || strings.Contains(text, "quality/existing") {
		t.Fatalf("new finding was not prioritized:\n%s", text)
	}
	if !strings.Contains(text, "1 additional findings omitted") {
		t.Fatalf("missing truncation message:\n%s", text)
	}
}
