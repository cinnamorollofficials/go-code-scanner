package scripts

import (
	"os"
	"strings"
	"testing"
)

func TestReleaseCandidateGateCoversRequiredChecks(t *testing.T) {
	data, err := os.ReadFile("release-candidate.sh")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)
	for _, required := range []string{
		"go test ./...",
		"go test -race ./...",
		"go vet ./...",
		"git diff --check",
		"TestRuntimeCachePreservesResultsAndInvalidatesContent",
		"TestReleaseBuildIsReproducible",
		"TestReleasePipelineEndToEnd",
		"TestReleaseBinaryHookLifecycle",
		"Golden|Contract|Structure|Checksums",
		"./scripts/fuzz-smoke.sh",
		"./scripts/vulnerability-scan.sh --if-available",
		"./scripts/performance-budget.sh",
		"go run ./cmd/security-review scan",
	} {
		if !strings.Contains(contents, required) {
			t.Fatalf("release-candidate gate is missing %q", required)
		}
	}
}
