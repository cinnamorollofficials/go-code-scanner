package adapters

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner/command"
)

func TestGoAdapterPresets(t *testing.T) {
	tests := []struct {
		name   string
		domain finding.Domain
	}{
		{name: Gofmt, domain: finding.Quality},
		{name: GoVet, domain: finding.Reliability},
		{name: GoTest, domain: finding.Reliability},
		{name: Govulncheck, domain: finding.SupplyChain},
		{name: Gosec, domain: finding.Security},
		{name: Gitleaks, domain: finding.Security},
		{name: Trivy, domain: finding.SupplyChain},
		{name: OSVScanner, domain: finding.SupplyChain},
		{name: Semgrep, domain: finding.Security},
	}
	for _, test := range tests {
		spec, err := Spec(test.name, test.name, Options{Workspace: command.WorkspaceStaged})
		if err != nil {
			t.Fatal(err)
		}
		if spec.Domain != test.domain || spec.Workspace != command.WorkspaceStaged || len(spec.Command) < 2 {
			t.Fatalf("unexpected %s adapter spec: %+v", test.name, spec)
		}
	}
}

func TestSecurityAdapterExitCodes(t *testing.T) {
	want := map[string]int{
		Govulncheck: 3, Gosec: 1, Gitleaks: 1, Trivy: 1, OSVScanner: 1, Semgrep: 1,
	}
	for name, exitCode := range want {
		spec, err := Spec(name, name, Options{})
		if err != nil {
			t.Fatal(err)
		}
		if len(spec.FindingExitCodes) != 1 || spec.FindingExitCodes[0] != exitCode {
			t.Fatalf("unexpected %s finding exit codes: %v", name, spec.FindingExitCodes)
		}
	}
}

func TestAdapterRejectsUnknownName(t *testing.T) {
	if _, err := New("fixture", "unknown", Options{}); err == nil {
		t.Fatal("expected unknown adapter error")
	}
}
