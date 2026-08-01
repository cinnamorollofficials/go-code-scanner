package scripts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReleaseBuildIsReproducible(t *testing.T) {
	first := buildReleaseForReproducibility(t, filepath.Join(t.TempDir(), "first"))
	second := buildReleaseForReproducibility(t, filepath.Join(t.TempDir(), "second"))
	firstDigest := sha256.Sum256(first)
	secondDigest := sha256.Sum256(second)
	if !bytes.Equal(firstDigest[:], secondDigest[:]) {
		t.Fatalf("release archives are not reproducible: first=%x second=%x", firstDigest, secondDigest)
	}
}

func buildReleaseForReproducibility(t *testing.T, directory string) []byte {
	t.Helper()
	command := exec.Command("sh", "build-release.sh")
	command.Env = append(os.Environ(),
		"VERSION=v1.2.3",
		"COMMIT=0123456789abcdef",
		"BUILD_DATE=2026-01-02T03:04:05Z",
		"DIST_DIR="+directory,
		"TARGETS="+runtime.GOOS+"/"+runtime.GOARCH,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("release build failed: %v: %s", err, output)
	}
	extension := ".tar.gz"
	if runtime.GOOS == "windows" {
		extension = ".zip"
	}
	path := filepath.Join(directory, fmt.Sprintf("security-review_v1.2.3_%s_%s%s", runtime.GOOS, runtime.GOARCH, extension))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
