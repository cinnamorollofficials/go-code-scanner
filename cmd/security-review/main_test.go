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

func TestScanUsesDomainPolicyAndGlobalCLIOverride(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.ts"), []byte("const API_URL = 'http://localhost'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "security-review.json")
	configData := []byte(`{
		"project": "domain-policy",
		"output": "report.json",
		"policy": {"hardening": "HIGH"}
	}`)
	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	args := []string{"scan", "--config", configPath, "--ci", "--quiet"}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("domain policy scan exit=%d stderr=%s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	args = append(args, "--fail-on", "medium")
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("global override scan exit=%d stderr=%s", code, stderr.String())
	}
}
