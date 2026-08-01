package scripts

import (
	"os"
	"os/exec"
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
	for _, target := range []string{"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64", "windows_amd64.exe", "windows_arm64.exe"} {
		if !strings.Contains(text, target) {
			t.Fatalf("release matrix is missing %s: %s", target, text)
		}
	}
}
