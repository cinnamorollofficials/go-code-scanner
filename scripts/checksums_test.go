package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestChecksumsAreDeterministicAndExcludeManifest(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "z-binary"), []byte("z"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "a-binary"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	run := func() string {
		t.Helper()
		command := exec.Command("sh", "checksums.sh")
		command.Env = append(os.Environ(), "DIST_DIR="+directory)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("checksum generation failed: %v: %s", err, output)
		}
		data, err := os.ReadFile(filepath.Join(directory, "SHA256SUMS"))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	first, second := run(), run()
	if first != second {
		t.Fatalf("checksum manifest changed between runs:\n%s\n%s", first, second)
	}
	if strings.Index(first, "a-binary") > strings.Index(first, "z-binary") || strings.Contains(first, "SHA256SUMS") {
		t.Fatalf("checksum manifest is not sorted or contains itself: %s", first)
	}
}
