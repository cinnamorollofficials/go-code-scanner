package scripts

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestBuildReleaseDryRunCoversSupportedMatrix(t *testing.T) {
	command := exec.Command("sh", "build-release.sh", "--dry-run")
	command.Env = append(os.Environ(), "VERSION=v1.2.3")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("release dry-run failed: %v: %s", err, output)
	}
	text := string(output)
	for _, target := range []string{"linux_amd64.tar.gz", "linux_arm64.tar.gz", "darwin_amd64.tar.gz", "darwin_arm64.tar.gz", "windows_amd64.zip", "windows_arm64.zip"} {
		if !strings.Contains(text, target) {
			t.Fatalf("release matrix is missing %s: %s", target, text)
		}
	}
	if strings.Contains(text, "dist/security-review_v1.2.3_windows_amd64.exe ") {
		t.Fatalf("dry run includes a raw distribution binary: %s", text)
	}
}

func TestBuildReleaseProducesOnlyArchives(t *testing.T) {
	directory := t.TempDir()
	command := exec.Command("sh", "build-release.sh")
	command.Env = append(os.Environ(),
		"VERSION=v1.2.3",
		"COMMIT=abc123",
		"BUILD_DATE=2026-01-02T03:04:05Z",
		"DIST_DIR="+directory,
		"TARGETS="+runtime.GOOS+"/"+runtime.GOARCH,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("release build failed: %v: %s", err, output)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one distribution archive, got %v", entries)
	}
	extension := ".tar.gz"
	if runtime.GOOS == "windows" {
		extension = ".zip"
	}
	if !strings.HasSuffix(entries[0].Name(), extension) {
		t.Fatalf("unexpected raw distribution artifact %q", entries[0].Name())
	}
}
