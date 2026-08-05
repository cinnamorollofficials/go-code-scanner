package release

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyChecksums(t *testing.T) {
	directory := t.TempDir()
	artifact := filepath.Join(directory, "security-review.tar.gz")
	content := []byte("artifact")
	if err := os.WriteFile(artifact, content, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(directory, "SHA256SUMS")
	digest := sha256.Sum256(content)
	if err := os.WriteFile(manifest, []byte(fmt.Sprintf("%x  security-review.tar.gz\n", digest)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyChecksums(manifest, directory); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyChecksums(manifest, directory); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("expected mismatch, got %v", err)
	}
}

func TestVerifyChecksumsRejectsUnsafeManifest(t *testing.T) {
	directory := t.TempDir()
	manifest := filepath.Join(directory, "SHA256SUMS")
	digest := strings.Repeat("0", 64)
	for _, entry := range []string{
		digest + "  ../artifact\n",
		digest + "  artifact\n" + digest + "  artifact\n",
		"not-a-digest  artifact\n",
	} {
		if err := os.WriteFile(manifest, []byte(entry), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := VerifyChecksums(manifest, directory); err == nil {
			t.Fatalf("expected unsafe manifest rejection for %q", entry)
		}
	}
}
