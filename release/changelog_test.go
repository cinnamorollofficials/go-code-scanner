package release

import (
	"os"
	"testing"
)

func TestRepositoryChangelogContract(t *testing.T) {
	data, err := os.ReadFile("../CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateChangelog(data); err != nil {
		t.Fatal(err)
	}
}

func TestValidateChangelogRejectsInvalidOrdering(t *testing.T) {
	data := []byte("# Changelog\n\n## [Unreleased]\n\n## [1.0.0] - 2026-01-01\n\n## [1.1.0] - 2026-02-01\n")
	if err := ValidateChangelog(data); err == nil {
		t.Fatal("out-of-order releases accepted")
	}
}
