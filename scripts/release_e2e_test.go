package scripts

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReleasePipelineEndToEnd(t *testing.T) {
	root := t.TempDir()
	dist := filepath.Join(root, "dist")
	if err := os.MkdirAll(dist, 0o700); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(root, "security-review")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-trimpath", "-buildvcs=false", "-o", binary, "./cmd/security-review")
	build.Dir = ".."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build release CLI: %v: %s", err, output)
	}

	extension := ".tar.gz"
	if runtime.GOOS == "windows" {
		extension = ".zip"
	}
	archive := filepath.Join(dist, "security-review_v1.2.3_"+runtime.GOOS+"_"+runtime.GOARCH+extension)
	runReleaseCLI(t, 0, binary, "release", "archive", "--binary", binary, "--output", archive, "--timestamp", "2026-01-02T03:04:05Z")
	originalArchive, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}

	checksums := exec.Command("sh", "checksums.sh")
	checksums.Env = append(os.Environ(), "DIST_DIR="+dist)
	if output, err := checksums.CombinedOutput(); err != nil {
		t.Fatalf("generate checksums: %v: %s", err, output)
	}
	manifest := filepath.Join(dist, "SHA256SUMS")
	runReleaseCLI(t, 0, binary, "release", "checksums", "verify", "--manifest", manifest, "--directory", dist)

	privateKeyPath, publicKeyPath := writeE2EReleaseKeys(t, root)
	provenance := filepath.Join(dist, "provenance.json")
	signature := filepath.Join(dist, "provenance.sig")
	runReleaseCLI(t, 0, binary,
		"release", "provenance", "generate",
		"--directory", dist,
		"--output", provenance,
		"--version", "v1.2.3",
		"--commit", "0123456789abcdef",
		"--build-date", "2026-01-02T03:04:05Z",
		"--builder", "test/e2e",
		"--private-key", privateKeyPath,
		"--signature", signature,
	)
	verifyArgs := []string{"release", "verify", "--provenance", provenance, "--signature", signature, "--public-key", publicKeyPath, "--directory", dist}
	runReleaseCLI(t, 0, binary, verifyArgs...)

	if err := os.WriteFile(archive, []byte("tampered artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
	runReleaseCLI(t, 1, binary, "release", "checksums", "verify", "--manifest", manifest, "--directory", dist)
	runReleaseCLI(t, 1, binary, verifyArgs...)

	if err := os.WriteFile(archive, originalArchive, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(signature, []byte("tampered signature\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runReleaseCLI(t, 1, binary, verifyArgs...)
}

func runReleaseCLI(t *testing.T, expectedExit int, binary string, args ...string) {
	t.Helper()
	command := exec.Command(binary, args...)
	output, err := command.CombinedOutput()
	actualExit := 0
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run %v: %v: %s", args, err, output)
		}
		actualExit = exitError.ExitCode()
	}
	if actualExit != expectedExit {
		t.Fatalf("run %v: expected exit %d, got %d: %s", args, expectedExit, actualExit, output)
	}
}

func writeE2EReleaseKeys(t *testing.T, directory string) (string, string) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	privatePath := filepath.Join(directory, "private.pem")
	if err := os.WriteFile(privatePath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}), 0o600); err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	publicPath := filepath.Join(directory, "public.pem")
	if err := os.WriteFile(publicPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0o600); err != nil {
		t.Fatal(err)
	}
	return privatePath, publicPath
}
