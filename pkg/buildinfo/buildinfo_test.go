package buildinfo

import "testing"

func TestStringIncludesInjectedReleaseMetadata(t *testing.T) {
	originalVersion, originalCommit, originalDate := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = originalVersion, originalCommit, originalDate })
	Version, Commit, Date = "v1.2.3", "abc123", "2026-08-02T00:00:00Z"
	if got, want := String(), "v1.2.3 commit=abc123 built=2026-08-02T00:00:00Z"; got != want {
		t.Fatalf("build metadata=%q want=%q", got, want)
	}
}
