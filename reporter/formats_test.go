package reporter

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestArtifactReportsDoNotExposeSnippets(t *testing.T) {
	report := formatFixtureReport()
	tests := []struct {
		name  string
		ext   string
		write func(string, *finding.Report) error
	}{
		{name: "json", ext: ".json", write: WriteJSON},
		{name: "sarif", ext: ".sarif", write: WriteSARIF},
		{name: "junit", ext: ".xml", write: WriteJUnit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "report"+test.ext)
			if err := test.write(path, report); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), "TOP-SECRET-SNIPPET") {
				t.Fatalf("%s output leaked snippet: %s", test.name, data)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0o600 {
				t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
			}
		})
	}
}

func TestSARIFStructure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.sarif")
	if err := WriteSARIF(path, formatFixtureReport()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document sarifDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != "2.1.0" || len(document.Runs) != 1 || len(document.Runs[0].Results) != 1 {
		t.Fatalf("unexpected SARIF structure: %+v", document)
	}
	if document.Runs[0].Results[0].BaselineState != "new" {
		t.Fatalf("baseline state was not mapped: %+v", document.Runs[0].Results[0])
	}
}

func TestJUnitStructure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.xml")
	if err := WriteJUnit(path, formatFixtureReport()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document junitTestSuites
	if err := xml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Suites) != 1 || document.Suites[0].Tests != 1 || document.Suites[0].Failures != 1 {
		t.Fatalf("unexpected JUnit structure: %+v", document)
	}
}

func TestJUnitSeparatesPolicyFailuresAndOperationalErrors(t *testing.T) {
	report := formatFixtureReport()
	report.Scanners = []finding.ScannerStatus{
		{ID: "pattern", State: finding.ScannerFindings},
		{ID: "external", State: finding.ScannerFailed, FailureKind: "timeout", Message: "TOP-SECRET-SNIPPET"},
	}
	path := filepath.Join(t.TempDir(), "report.xml")
	if err := WriteJUnit(path, report); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document junitTestSuites
	if err := xml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	suite := document.Suites[0]
	if suite.Tests != 2 || suite.Failures != 1 || suite.Errors != 1 || suite.Cases[1].Error == nil || suite.Cases[1].Error.Type != "timeout" {
		t.Fatalf("unexpected JUnit policy/operational mapping: %+v", suite)
	}
	if strings.Contains(string(data), "TOP-SECRET-SNIPPET") {
		t.Fatalf("JUnit leaked operational scanner message: %s", data)
	}
}

func formatFixtureReport() *finding.Report {
	return &finding.Report{
		Project: "fixture", ToolVersion: "test", FingerprintVersion: "2",
		Findings: []finding.Finding{{
			Fingerprint: "abcdef", RuleID: "rule", Domain: finding.Security,
			Category: "injection", Severity: finding.High, Description: "unsafe behavior",
			Recommendation: "use safe behavior", Snippet: "TOP-SECRET-SNIPPET",
			BaselineState: finding.BaselineNew,
			Location:      finding.Location{File: "app.go", Line: 12},
		}},
	}
}
