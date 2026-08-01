package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
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

func TestHookInstallStatusAndUninstall(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"hook", "install", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("install exit=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"hook", "status", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("status exit=%d stderr=%s", code, stderr.String())
	}
	if stdout.String() != "pre-commit: installed\n" {
		t.Fatalf("unexpected status output %q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"hook", "uninstall", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("uninstall exit=%d stderr=%s", code, stderr.String())
	}
}

func TestPreCommitHookScansIndexInsteadOfWorkingTree(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	path := filepath.Join(root, "app.ts")
	if err := os.WriteFile(path, []byte("google-mock-jwt-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", root, "add", "app.ts").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, output)
	}
	if err := os.WriteFile(path, []byte("const safe = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	args := []string{"hook", "run", "pre-commit", "--root", root}
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("staged finding did not block hook: exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	if output, err := exec.Command("git", "-C", root, "add", "app.ts").CombinedOutput(); err != nil {
		t.Fatalf("git add safe content: %v: %s", err, output)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("safe staged content blocked hook: exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "security_findings.json")); err != nil {
		t.Fatalf("hook did not write report: %v", err)
	}
}

func TestPreCommitHookCanBeDisabledByConfig(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	data := []byte(`{"project":"disabled-hook","hooks":{"pre_commit":{"enabled":false}}}`)
	if err := os.WriteFile(filepath.Join(root, "security-review.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	args := []string{"hook", "run", "pre-commit", "--root", root}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("disabled hook exit=%d stderr=%s", code, stderr.String())
	}
	if stdout.String() != "pre-commit: disabled\n" {
		t.Fatalf("unexpected disabled hook output %q", stdout.String())
	}
}
