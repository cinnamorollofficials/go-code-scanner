package reporter

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestJSONReportSchemaGolden(t *testing.T) {
	report := &finding.Report{
		SchemaVersion: "1.0", FingerprintVersion: "2", ToolVersion: "v1.2.3",
		ConfigHash: "config-hash", RuleSetHash: "rules-hash", Timestamp: time.Unix(0, 0).UTC(), Duration: 42,
		ScanMode: "staged", Project: "fixture",
		Summary: finding.Summary{Total: 1, High: 1, New: 1, ByDomain: map[finding.Domain]int{finding.Security: 1}},
		Findings: []finding.Finding{{
			ID: "finding-1", Fingerprint: "fingerprint-1", RuleID: "security/example", Tool: "pattern", Domain: finding.Security,
			Category: "example", Severity: finding.High, Description: "example finding", Recommendation: "fix it",
			Tags: []string{"example"}, Location: finding.Location{File: "app.go", Line: 7}, BaselineState: finding.BaselineNew,
			Metadata: map[string]string{"symbol": "main"},
		}},
		Scanners: []finding.ScannerStatus{{ID: "pattern", State: finding.ScannerFindings, Required: true, Duration: 12, Domain: finding.Security, Capabilities: []string{"line-patterns"}, SupportedModes: []string{"staged"}}},
	}
	path := filepath.Join(t.TempDir(), "report.json")
	if err := WriteJSON(path, report); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("testdata/report.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("JSON report schema changed; review and update the golden contract intentionally\nactual:\n%s\nexpected:\n%s", actual, expected)
	}
}

func TestWriteJSONReplacesExistingReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := &finding.Report{SchemaVersion: "1.0", Timestamp: time.Unix(0, 0).UTC()}
	if err := WriteJSON(path, report); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "old" {
		t.Fatal("report was not replaced")
	}
	if _, err := os.Stat(path + ".previous"); !os.IsNotExist(err) {
		t.Fatal("backup should be removed after successful replacement")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected report mode 0600, got %o", info.Mode().Perm())
	}
}
