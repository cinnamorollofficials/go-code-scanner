package suppression

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadRejectsInvalidSuppression(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	content := `{"version":1,"suppressions":[{"file":"src/a.go","line":1,"reason":"","expires":"not-a-date"}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestMatchesRequiresExactPath(t *testing.T) {
	item := finding.Finding{RuleID: "one", Location: finding.Location{File: "service/src/a.go", Line: 4}}
	rule := Rule{RuleID: "one", File: "src/a.go", Line: 4, Reason: "accepted", Expires: "2030-01-01"}
	active, suppressed, _ := Apply([]finding.Finding{item}, []Rule{rule}, time.Now())
	if len(active) != 1 || len(suppressed) != 0 {
		t.Fatal("suffix-only path must not suppress a finding")
	}
}

func TestLoadEnforcesTargetedGovernanceRequirements(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	requirements := []Requirement{{RuleIDs: []string{"security/*"}, RequireTicket: true, RequireApprover: true}}
	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"version":1,"suppressions":[{"rule_id":"security/hardcoded-secret","file":"src/a.go","line":1,"reason":"migration","expires":"2030-01-01"}]}`)
	if _, err := LoadWithRequirements(path, requirements); err == nil || !strings.Contains(err.Error(), "ticket is required") {
		t.Fatalf("missing ticket was not rejected: %v", err)
	}
	write(`{"version":1,"suppressions":[{"rule_id":"security/hardcoded-secret","file":"src/a.go","line":1,"reason":"migration","ticket":"SEC-123","expires":"2030-01-01"}]}`)
	if _, err := LoadWithRequirements(path, requirements); err == nil || !strings.Contains(err.Error(), "approved_by is required") {
		t.Fatalf("missing approver was not rejected: %v", err)
	}
	write(`{"version":1,"suppressions":[{"rule_id":"security/hardcoded-secret","file":"src/a.go","line":1,"reason":"migration","ticket":"SEC-123","approved_by":"security-team","expires":"2030-01-01"}]}`)
	if _, err := LoadWithRequirements(path, requirements); err != nil {
		t.Fatalf("complete governed suppression rejected: %v", err)
	}
}

func TestAddSupportsDryRunAndAtomicWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".security-ignore")
	rule := Rule{RuleID: "security/example", File: "app.go", Line: 7, Reason: "reviewed exception", Expires: "2030-01-01"}
	if result, err := Add(path, rule, true); err != nil || len(result.Suppressions) != 1 {
		t.Fatalf("dry-run failed: result=%+v err=%v", result, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote suppression file")
	}
	if _, err := Add(path, rule, false); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected suppression permissions: %o", info.Mode().Perm())
	}
	if _, err := Add(path, rule, false); err == nil {
		t.Fatal("duplicate suppression was accepted")
	}
}
