package adapters

import "testing"

func TestStructuredAdapterParsers(t *testing.T) {
	t.Run("gosec", func(t *testing.T) {
		data := []byte(`{"Issues":[{"severity":"HIGH","confidence":"MEDIUM","rule_id":"G204","details":"Subprocess launched with variable","file":"cmd/main.go","line":"12-13"}]}`)
		items, err := parseGosec(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "G204" || items[0].Line != 12 {
			t.Fatalf("unexpected gosec parse: items=%+v err=%v", items, err)
		}
	})
	t.Run("trivy", func(t *testing.T) {
		data := []byte(`{"Results":[{"Target":"go.mod","Vulnerabilities":[{"VulnerabilityID":"CVE-2026-1","PkgName":"example/module","InstalledVersion":"v1.0.0","FixedVersion":"v1.0.1","Severity":"CRITICAL","Title":"Example issue","PrimaryURL":"https://example.invalid/CVE-2026-1"}]}]}`)
		items, err := parseTrivy(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "CVE-2026-1" || items[0].Metadata["fixed_version"] != "v1.0.1" {
			t.Fatalf("unexpected trivy parse: items=%+v err=%v", items, err)
		}
	})
	t.Run("semgrep", func(t *testing.T) {
		data := []byte(`{"results":[{"check_id":"go.lang.security.audit","path":"app.go","start":{"line":7},"extra":{"message":"Potential issue","severity":"ERROR"}}]}`)
		items, err := parseSemgrep(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "go.lang.security.audit" || items[0].Line != 7 {
			t.Fatalf("unexpected semgrep parse: items=%+v err=%v", items, err)
		}
	})
}
