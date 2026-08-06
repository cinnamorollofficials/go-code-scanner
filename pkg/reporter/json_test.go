package reporter

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
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
	actualNormalized := bytes.ReplaceAll(actual, []byte("\r\n"), []byte("\n"))
	expectedNormalized := bytes.ReplaceAll(expected, []byte("\r\n"), []byte("\n"))
	if !bytes.Equal(actualNormalized, expectedNormalized) {
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
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("expected report mode 0600, got %o", info.Mode().Perm())
	}
}

func TestWriteJSONRejectsTargetAndBackupSymlinks(t *testing.T) {
	for _, backup := range []bool{false, true} {
		directory := t.TempDir()
		target := filepath.Join(directory, "outside.json")
		if err := os.WriteFile(target, []byte("outside"), 0o600); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(directory, "report.json")
		link := path
		if backup {
			if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
				t.Fatal(err)
			}
			link = path + ".previous"
		}
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		report := &finding.Report{SchemaVersion: "1.0", Timestamp: time.Unix(0, 0).UTC()}
		if err := WriteJSON(path, report); err == nil {
			t.Fatalf("backup=%t: symlink report path accepted", backup)
		}
		content, err := os.ReadFile(target)
		if err != nil || string(content) != "outside" {
			t.Fatalf("backup=%t: symlink target changed: content=%q err=%v", backup, content, err)
		}
	}
}
