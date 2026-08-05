package compatibility

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCurrentContractMatchesGoldenManifest(t *testing.T) {
	actual, err := json.MarshalIndent(Current(), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("testdata/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(expected) > 0 && expected[len(expected)-1] == '\n' {
		expected = expected[:len(expected)-1]
	}
	if string(actual) != string(expected) {
		t.Fatalf("public compatibility contract changed; require migration review\nactual:\n%s\nexpected:\n%s", actual, expected)
	}
}

func TestDecodeRejectsUnknownAndTrailingData(t *testing.T) {
	data, err := json.Marshal(Current())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(data); err != nil {
		t.Fatalf("decode current contract: %v", err)
	}
	if _, err := Decode([]byte(`{"config_schema":1,"unknown":true}`)); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
	if _, err := Decode(append(data, []byte("\n{}")...)); err == nil {
		t.Fatal("expected trailing JSON value error")
	}
}

func TestCompareReportsStableFieldChanges(t *testing.T) {
	previous := Current()
	current := previous
	current.ReportSchema = "security-review/v-next"
	current.CacheKeyVersion = "v-next"
	changes := Compare(previous, current)
	if len(changes) != 2 || changes[0].Field != "report_schema" || changes[1].Field != "cache_key_version" {
		t.Fatalf("unexpected changes: %+v", changes)
	}
}
