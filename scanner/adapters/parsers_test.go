package adapters

import (
	"fmt"
	"strings"
	"testing"
)

func TestStructuredAdapterParsers(t *testing.T) {
	t.Run("gosec", func(t *testing.T) {
		data := []byte(`{"Issues":[{"severity":"HIGH","confidence":"MEDIUM","rule_id":"G204","details":"Subprocess launched with variable","file":"cmd/main.go","line":"12-13"}]}`)
		items, err := parseGosec(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "G204" || items[0].Line != 12 {
			t.Fatalf("unexpected gosec parse: items=%+v err=%v", items, err)
		}
	})
	t.Run("gitleaks redacts secret fields", func(t *testing.T) {
		data := []byte(`[{"RuleID":"generic-api-key","Description":"Generic API Key","File":"config.go","StartLine":4,"Fingerprint":"abc:config.go:generic-api-key:4","Commit":"abc","Secret":"must-not-leak","Match":"token=must-not-leak"}]`)
		items, err := parseGitleaks(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "generic-api-key" || items[0].Line != 4 {
			t.Fatalf("unexpected Gitleaks parse: items=%+v err=%v", items, err)
		}
		encoded := fmt.Sprintf("%+v", items[0])
		if strings.Contains(encoded, "must-not-leak") {
			t.Fatalf("Gitleaks secret leaked into parsed finding: %s", encoded)
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
	t.Run("govulncheck stream", func(t *testing.T) {
		data := []byte("{\"config\":{\"protocol_version\":\"v1.0.0\"}}\n" +
			"{\"finding\":{\"osv\":\"GO-2026-0001\",\"fixed_version\":\"v1.2.3\",\"trace\":[{\"module\":\"example/mod\",\"version\":\"v1.0.0\",\"package\":\"example/mod/pkg\",\"function\":\"Run\",\"position\":{\"filename\":\"main.go\",\"line\":14}}]}}\n")
		items, err := parseGovulncheck(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "GO-2026-0001" || items[0].Line != 14 || items[0].Metadata["fixed_version"] != "v1.2.3" {
			t.Fatalf("unexpected govulncheck parse: items=%+v err=%v", items, err)
		}
	})
	t.Run("osv-scanner", func(t *testing.T) {
		data := []byte(`{"results":[{"source":{"path":"go.mod"},"packages":[{"package":{"name":"example/mod","version":"v1.0.0","ecosystem":"Go"},"vulnerabilities":[{"id":"GO-2026-0001","summary":"Example vulnerability","database_specific":{"severity":"HIGH"}}]}]}]}`)
		items, err := parseOSVScanner(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "GO-2026-0001" || items[0].File != "go.mod" || items[0].Metadata["package"] != "example/mod" {
			t.Fatalf("unexpected OSV-Scanner parse: items=%+v err=%v", items, err)
		}
	})
	t.Run("eslint", func(t *testing.T) {
		data := []byte(`[{"filePath":"src/app.js","messages":[{"ruleId":"no-unused-vars","severity":2,"line":10,"column":5}]}]`)
		items, err := parseESLint(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "no-unused-vars" || items[0].Line != 10 {
			t.Fatalf("unexpected eslint parse: items=%+v err=%v", items, err)
		}
	})
	t.Run("tsc", func(t *testing.T) {
		data := []byte("src/index.ts(10,5): error TS2322: Type 'string' is not assignable to type 'number'.\n")
		items, err := parseTSC(data)
		if err != nil || len(items) != 1 || items[0].RuleID != "TS2322" || items[0].File != "src/index.ts" || items[0].Line != 10 {
			t.Fatalf("unexpected tsc parse: items=%+v err=%v", items, err)
		}
	})
}
