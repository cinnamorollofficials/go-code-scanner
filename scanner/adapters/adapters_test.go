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

func TestAdapterRejectsUnknownName(t *testing.T) {
	if _, err := New("fixture", "unknown", Options{}); err == nil {
		t.Fatal("expected unknown adapter error")
	}
}
