package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanExitCodes(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "safe.go"), []byte("package safe\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"scan", "--root", root, "--ci"}, &stdout, &stderr); code != 0 {
		t.Fatalf("safe scan exit=%d stderr=%s", code, stderr.String())
	}
	if err := os.WriteFile(filepath.Join(root, "bad.ts"), []byte("google-mock-jwt-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"scan", "--root", root, "--ci"}, &stdout, &stderr); code != 1 {
		t.Fatalf("unsafe scan exit=%d stderr=%s", code, stderr.String())
	}
}
