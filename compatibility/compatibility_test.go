package compatibility

import (
	"encoding/json"
	"os"
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
