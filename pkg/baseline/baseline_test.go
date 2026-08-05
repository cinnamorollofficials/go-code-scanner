package baseline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestLoadRejectsUnknownFieldsAndTrailingDocuments(t *testing.T) {
	for _, content := range []string{
		`{"version":1,"fingerprint_version":"3","generated_at":"1970-01-01T00:00:00Z","entries":[],"unexpected":true}`,
		`{"version":1,"fingerprint_version":"3","generated_at":"1970-01-01T00:00:00Z","entries":[]} {}`,
	} {
		path := filepath.Join(t.TempDir(), "baseline.json")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("invalid baseline accepted: %s", content)
		}
	}
}

func TestBaselineRoundTripAndComparison(t *testing.T) {
	report := reportWith("v2", finding.Finding{
		ID: "F-0001", Fingerprint: "existing", RuleID: "rule-one",
		Domain: finding.Security, Location: finding.Location{File: "b.go", Line: 4},
	})
	file, err := FromReport(report, time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "baseline.json")
	if err := Write(path, file); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	current := reportWith("v2",
		finding.Finding{Fingerprint: "existing", RuleID: "rule-one", Domain: finding.Security, Location: finding.Location{File: "b.go", Line: 40}},
		finding.Finding{Fingerprint: "new", RuleID: "rule-two", Domain: finding.Quality, Location: finding.Location{File: "a.go", Line: 2}},
	)
	comparison, err := Compare(current, loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(comparison.New) != 1 || len(comparison.Existing) != 1 || len(comparison.Resolved) != 0 {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
	if current.Findings[0].BaselineState != finding.BaselineExisting || current.Findings[1].BaselineState != finding.BaselineNew {
		t.Fatalf("report was not classified: %+v", current.Findings)
	}
	if current.Summary.New != 1 || current.Summary.Existing != 1 || current.Summary.Resolved != 0 {
		t.Fatalf("summary was not classified: %+v", current.Summary)
	}
}

func TestCompareReportsResolvedEntries(t *testing.T) {
	file := &File{
		Version: Version, FingerprintVersion: "v2",
		Entries: []Entry{{Fingerprint: "resolved", RuleID: "rule", Domain: finding.Hardening, File: "app.go"}},
	}
	report := reportWith("v2")
	comparison, err := Compare(report, file)
	if err != nil {
		t.Fatal(err)
	}
	if len(comparison.Resolved) != 1 || comparison.Resolved[0].Fingerprint != "resolved" {
		t.Fatalf("unexpected resolved entries: %+v", comparison.Resolved)
	}
	if report.Summary.Resolved != 1 {
		t.Fatalf("resolved summary was not updated: %+v", report.Summary)
	}
}

func TestCompareRejectsFingerprintVersionMismatch(t *testing.T) {
	file := &File{Version: Version, FingerprintVersion: "old"}
	if _, err := Compare(reportWith("new"), file); err == nil {
		t.Fatal("expected fingerprint version mismatch")
	}
}

func reportWith(version string, items ...finding.Finding) *finding.Report {
	return &finding.Report{FingerprintVersion: version, Findings: items}
}
