package suppression

import (
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestApplySeparatesActiveSuppressedAndStale(t *testing.T) {
	items := []finding.Finding{
		{RuleID: "one", Location: finding.Location{File: "src/a.go", Line: 4}},
		{RuleID: "two", Location: finding.Location{File: "src/b.go", Line: 5}},
	}
	rules := []Rule{
		{RuleID: "one", File: "src/a.go", Line: 4, Reason: "accepted", Expires: "2030-01-01"},
		{RuleID: "two", File: "src/b.go", Line: 5, Reason: "old", Expires: "2020-01-01"},
	}
	active, suppressed, stale := Apply(items, rules, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))
	if len(active) != 1 || len(suppressed) != 1 || len(stale) != 1 {
		t.Fatalf("active=%d suppressed=%d stale=%d", len(active), len(suppressed), len(stale))
	}
}
