package scripts_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestCIWorkflowUsesPinnedActionsAndVerificationScript(t *testing.T) {
	workflow, err := os.ReadFile("../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(workflow)
	actions := regexp.MustCompile(`(?m)^\s*uses:\s+[^@\s]+@([0-9a-f]{40})(?:\s|$)`).FindAllStringSubmatch(contents, -1)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions pinned to full commit SHAs, found %d", len(actions))
	}
	if !strings.Contains(contents, "run: ./scripts/verify.sh") {
		t.Fatal("CI workflow must run the canonical verification script")
	}
	if !strings.Contains(contents, "persist-credentials: false") {
		t.Fatal("checkout credentials must not persist after checkout")
	}
}
