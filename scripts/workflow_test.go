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

func TestReleaseWorkflowBuildsAndVerifiesTaggedArtifacts(t *testing.T) {
	workflow, err := os.ReadFile("../.github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(workflow)
	actions := regexp.MustCompile(`(?m)^\s*uses:\s+[^@\s]+@([0-9a-f]{40})(?:\s|$)`).FindAllStringSubmatch(contents, -1)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions pinned to full commit SHAs, found %d", len(actions))
	}
	for _, command := range []string{
		"release changelog validate",
		"git show -s --format=%cI",
		"./scripts/build-release.sh",
		"./scripts/checksums.sh",
		"sha256sum --check SHA256SUMS",
		"release provenance generate",
		`--directory dist`,
		`--output dist/provenance.json`,
		`--version "${GITHUB_REF_NAME}"`,
		`--commit "${GITHUB_SHA}"`,
		`--build-date "${BUILD_DATE}"`,
		`--builder "github.com/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"`,
	} {
		if !strings.Contains(contents, command) {
			t.Fatalf("release workflow is missing %q", command)
		}
	}
	if !strings.Contains(contents, "contents: read") || !strings.Contains(contents, "persist-credentials: false") {
		t.Fatal("release verification must use read-only permissions without persisted credentials")
	}
}
