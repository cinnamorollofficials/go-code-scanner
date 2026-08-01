package release

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestArchiveBinaryIsDeterministic(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "security-review")
	if err := os.WriteFile(binary, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1_700_000_000, 0)
	for _, extension := range []string{".tar.gz", ".zip"} {
		first := filepath.Join(directory, "first"+extension)
		second := filepath.Join(directory, "second"+extension)
		if err := ArchiveBinary(binary, first, stamp); err != nil {
			t.Fatal(err)
		}
		if err := ArchiveBinary(binary, second, stamp); err != nil {
			t.Fatal(err)
		}
		firstData, _ := os.ReadFile(first)
		secondData, _ := os.ReadFile(second)
		if !bytes.Equal(firstData, secondData) {
			t.Fatalf("%s archive is not deterministic", extension)
		}
	}
}

func TestArchiveBinaryRejectsSymlink(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(target, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := ArchiveBinary(link, filepath.Join(directory, "release.zip"), time.Now()); err == nil {
		t.Fatal("expected symlink rejection")
	}
}
