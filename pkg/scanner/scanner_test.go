package scanner

import (
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestFailureKindValidity(t *testing.T) {
	for _, kind := range []FailureKind{
		FailureCanceled, FailureExecution, FailureMissing,
		FailurePanic, FailurePartial, FailureTimeout,
	} {
		if !kind.Valid() {
			t.Errorf("expected %q to be valid", kind)
		}
	}
	if FailureKind("unknown").Valid() || FailureKind("").Valid() {
		t.Fatal("unknown and empty failure kinds must be invalid")
	}
}

func TestScannerStateValidity(t *testing.T) {
	for _, state := range []finding.ScannerState{
		finding.ScannerClean, finding.ScannerFindings, finding.ScannerPartial,
		finding.ScannerSkipped, finding.ScannerFailed,
	} {
		if !state.Valid() {
			t.Errorf("expected %q to be valid", state)
		}
	}
	if finding.ScannerState("unknown").Valid() || finding.ScannerState("").Valid() {
		t.Fatal("unknown and empty scanner states must be invalid")
	}
}
