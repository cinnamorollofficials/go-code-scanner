package reporter

import (
	"bytes"
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
	if document.Schema != "https://json.schemastore.org/sarif-2.1.0.json" || document.Runs[0].Tool.Driver.Name != "go-code-scanner" {
		t.Fatalf("unexpected SARIF public schema identity: %+v", document)
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
	if document.XMLName.Local != "testsuites" || document.Suites[0].Name != "fixture" {
		t.Fatalf("unexpected JUnit public schema identity: %+v", document)
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

func TestFrontendFindingsPublicReportFormatsContract(t *testing.T) {
	report := &finding.Report{
		Project: "frontend-fixture", ToolVersion: "v1.0.0", SchemaVersion: "1.0", FingerprintVersion: "3",
		Findings: []finding.Finding{
			{
				ID: "F-FE-001", Fingerprint: "fe-fingerprint-1", RuleID: "frontend/dangerously-set-inner-html",
				Tool: "frontend", Domain: finding.Security, Category: "xss", Severity: finding.High,
				Description: "Unsanitized HTML injection via dangerouslySetInnerHTML", Recommendation: "Use DOMPurify or safe JSX text",
				Snippet: "TOP-SECRET-FRONTEND-SNIPPET", Location: finding.Location{File: "Component.tsx", Line: 42},
				BaselineState: finding.BaselineNew, Metadata: map[string]string{"sink": "dangerouslySetInnerHTML"},
			},
			{
				ID: "F-FE-002", Fingerprint: "fe-fingerprint-2", RuleID: "frontend/missing-subresource-integrity",
				Tool: "frontend", Domain: finding.Security, Category: "integrity", Severity: finding.High,
				Description: "Cross-origin script loaded without Subresource Integrity", Recommendation: "Add integrity and crossorigin attributes",
				Snippet: "<script src=\"https://cdn.example.com/lib.js\"></script>", Location: finding.Location{File: "index.html", Line: 10},
				BaselineState: finding.BaselineNew, Metadata: map[string]string{"url": "https://cdn.example.com/lib.js"},
			},
		},
	}

	// 1. Terminal output check
	var termBuf bytes.Buffer
	if err := WriteTerminal(&termBuf, report); err != nil {
		t.Fatalf("WriteTerminal failed: %v", err)
	}
	termStr := termBuf.String()
	if strings.Contains(termStr, "TOP-SECRET-FRONTEND-SNIPPET") {
		t.Errorf("Terminal report leaked snippet")
	}
	if !strings.Contains(termStr, "frontend/dangerously-set-inner-html") || !strings.Contains(termStr, "frontend/missing-subresource-integrity") {
		t.Errorf("Terminal report missing stable rule IDs")
	}

	// 2. JSON, SARIF, JUnit check
	writers := map[string]func(string, *finding.Report) error{
		"json":  WriteJSON,
		"sarif": WriteSARIF,
		"junit": WriteJUnit,
	}
	for name, write := range writers {
		path := filepath.Join(t.TempDir(), "report."+name)
		if err := write(path, report); err != nil {
			t.Fatalf("%s writer failed: %v", name, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "TOP-SECRET-FRONTEND-SNIPPET") {
			t.Errorf("%s report leaked snippet", name)
		}
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
