package finding

import (
	"encoding/json"
	"testing"
)

func TestConfidenceValidation(t *testing.T) {
	if !ConfidenceHigh.Valid() || !ConfidenceMedium.Valid() || !ConfidenceLow.Valid() || !ConfidenceUnknown.Valid() {
		t.Fatal("expected standard confidence levels to be valid")
	}
	if Confidence("invalid").Valid() {
		t.Fatal("expected arbitrary string to be invalid confidence")
	}
}

func TestExploitabilityValidation(t *testing.T) {
	if !ExploitabilityLikely.Valid() || !ExploitabilityUnlikely.Valid() || !ExploitabilityUnknown.Valid() {
		t.Fatal("expected standard exploitability levels to be valid")
	}
	if Exploitability("impossible").Valid() {
		t.Fatal("expected arbitrary string to be invalid exploitability")
	}
}

func TestFindingStateValidation(t *testing.T) {
	validStates := []FindingState{
		FindingCandidate,
		FindingProbable,
		FindingConfirmed,
		FindingNeedsContext,
		FindingDismissedWithEvidence,
		FindingFixedVerified,
		FindingFixedNotVerified,
	}
	for _, state := range validStates {
		if !state.Valid() {
			t.Fatalf("expected state %q to be valid", state)
		}
	}
	if FindingState("unknown_state").Valid() {
		t.Fatal("expected unknown state to be invalid")
	}
}

func TestFindingSerializationWithDataflow(t *testing.T) {
	f := Finding{
		ID:             "SQLI-001-TEST",
		Tool:           "sqltaint",
		Domain:         Security,
		Category:       "injection",
		Severity:       High,
		Confidence:     ConfidenceHigh,
		Exploitability: ExploitabilityLikely,
		FindingState:   FindingConfirmed,
		Description:    "Untrusted query parameter reaches SQL sink",
		Location: Location{
			File: "internal/db.go",
			Line: 42,
		},
		Dataflow: []DataflowStep{
			{
				Type:        StepSource,
				Location:    Location{File: "internal/handler.go", Line: 15},
				Label:       "req.URL.Query().Get(\"id\")",
				Explanation: "User input received from HTTP query parameter",
			},
			{
				Type:        StepPropagator,
				Location:    Location{File: "internal/handler.go", Line: 18},
				Label:       "query := \"SELECT * FROM users WHERE id = \" + id",
				Explanation: "Tainted string concatenated into SQL query string",
			},
			{
				Type:        StepSink,
				Location:    Location{File: "internal/db.go", Line: 42},
				Label:       "db.Query(query)",
				Explanation: "Executable SQL string executed by driver sink",
			},
		},
	}

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("failed to marshal finding: %v", err)
	}

	var unmarshaled Finding
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal finding: %v", err)
	}

	if unmarshaled.Confidence != ConfidenceHigh {
		t.Errorf("got confidence %q, want %q", unmarshaled.Confidence, ConfidenceHigh)
	}
	if unmarshaled.Exploitability != ExploitabilityLikely {
		t.Errorf("got exploitability %q, want %q", unmarshaled.Exploitability, ExploitabilityLikely)
	}
	if unmarshaled.FindingState != FindingConfirmed {
		t.Errorf("got finding_state %q, want %q", unmarshaled.FindingState, FindingConfirmed)
	}
	if len(unmarshaled.Dataflow) != 3 {
		t.Fatalf("got %d dataflow steps, want 3", len(unmarshaled.Dataflow))
	}
	if unmarshaled.Dataflow[0].Type != StepSource {
		t.Errorf("got step 0 type %q, want %q", unmarshaled.Dataflow[0].Type, StepSource)
	}
	if unmarshaled.Dataflow[2].Type != StepSink {
		t.Errorf("got step 2 type %q, want %q", unmarshaled.Dataflow[2].Type, StepSink)
	}
}
