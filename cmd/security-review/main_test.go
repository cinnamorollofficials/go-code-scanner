package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/baseline"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
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

func TestHookLifecycleSupportsEveryEvent(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	for _, event := range []string{"pre-commit", "commit-msg", "pre-push"} {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), []string{"hook", "install", event, "--root", root}, &stdout, &stderr); code != 0 {
			t.Fatalf("install %s exit=%d stderr=%s", event, code, stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := run(context.Background(), []string{"hook", "status", event, "--root", root}, &stdout, &stderr); code != 0 {
			t.Fatalf("status %s exit=%d stderr=%s", event, code, stderr.String())
		}
		if stdout.String() != event+": installed\n" {
			t.Fatalf("unexpected %s status output %q", event, stdout.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := run(context.Background(), []string{"hook", "uninstall", event, "--root", root}, &stdout, &stderr); code != 0 {
			t.Fatalf("uninstall %s exit=%d stderr=%s", event, code, stderr.String())
		}
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

func TestCommitMsgHookValidatesConfiguredPolicy(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	data := []byte(`{"project":"message-hook","hooks":{"commit_msg":{"enabled":true,"max_subject_length":72}}}`)
	if err := os.WriteFile(filepath.Join(root, "security-review.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	messagePath := filepath.Join(root, "COMMIT_EDITMSG")
	args := []string{"hook", "run", "commit-msg", "--root", root, "--file", messagePath}
	var stdout, stderr bytes.Buffer
	if err := os.WriteFile(messagePath, []byte("not conventional\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("invalid message exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := os.WriteFile(messagePath, []byte("feat(hook): validate messages\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("valid message exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestPrePushHookRunsConfiguredFullWorkspaceProfile(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	data := []byte(`{
		"project":"pre-push-hook",
		"hooks":{"pre_push":{"enabled":true,"profile":"standard"}}
	}`)
	if err := os.WriteFile(filepath.Join(root, "security-review.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	unsafePath := filepath.Join(root, "unstaged.ts")
	if err := os.WriteFile(unsafePath, []byte("google-mock-jwt-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"hook", "run", "pre-push", "--root", root}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("unsafe full workspace did not block pre-push: exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := os.WriteFile(unsafePath, []byte("const safe = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("safe full workspace blocked pre-push: exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestBaselineCommandsAndNewOnlyPolicy(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "existing.go"), []byte("change-me-in-production\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(root, "security_findings.json")
	baselinePath := filepath.Join(root, ".security-baseline.json")
	var stdout, stderr bytes.Buffer

	if code := run(context.Background(), []string{"scan", "--root", root, "--quiet"}, &stdout, &stderr); code != 0 {
		t.Fatalf("initial scan exit=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	createArgs := []string{"baseline", "create", "--report", reportPath, "--baseline", baselinePath}
	if code := run(context.Background(), createArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("baseline create exit=%d stderr=%s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	scanArgs := []string{"scan", "--root", root, "--baseline", baselinePath, "--new-only", "--ci", "--quiet"}
	if code := run(context.Background(), scanArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("existing finding blocked new-only policy: exit=%d stderr=%s", code, stderr.String())
	}

	if err := os.WriteFile(filepath.Join(root, "new.ts"), []byte("google-mock-jwt-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), scanArgs, &stdout, &stderr); code != 1 {
		t.Fatalf("new finding did not block policy: exit=%d stderr=%s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	statusArgs := []string{"baseline", "status", "--report", reportPath, "--baseline", baselinePath}
	if code := run(context.Background(), statusArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("baseline status exit=%d stderr=%s", code, stderr.String())
	}
	if stdout.String() != "Baseline: new=1 existing=1 resolved=0\n" {
		t.Fatalf("unexpected baseline status %q", stdout.String())
	}
}

func TestBaselineUpdateRequiresResolvedFindingApproval(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, "report.json")
	baselinePath := filepath.Join(root, "baseline.json")
	writeBaselineReport(t, reportPath, "first", "second")
	var stdout, stderr bytes.Buffer
	create := []string{"baseline", "create", "--report", reportPath, "--baseline", baselinePath}
	if code := run(context.Background(), create, &stdout, &stderr); code != 0 {
		t.Fatalf("create exit=%d stderr=%s", code, stderr.String())
	}
	original, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	writeBaselineReport(t, reportPath, "first")

	stdout.Reset()
	stderr.Reset()
	update := []string{"baseline", "update", "--report", reportPath, "--baseline", baselinePath}
	if code := run(context.Background(), append(update, "--dry-run"), &stdout, &stderr); code != 0 {
		t.Fatalf("dry-run exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "resolved=1") || !strings.Contains(stdout.String(), "dry-run") {
		t.Fatalf("unexpected dry-run output: %q", stdout.String())
	}
	unchanged, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, unchanged) {
		t.Fatal("dry-run modified baseline")
	}

	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), update, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "--accept-resolved") {
		t.Fatalf("unapproved update exit=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), append(update, "--accept-resolved"), &stdout, &stderr); code != 0 {
		t.Fatalf("approved update exit=%d stderr=%s", code, stderr.String())
	}
	updated, err := baseline.Load(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Entries) != 1 {
		t.Fatalf("approved update wrote %d entries", len(updated.Entries))
	}
}

func writeBaselineReport(t *testing.T, path string, fingerprints ...string) {
	t.Helper()
	report := finding.Report{FingerprintVersion: securityreview.FingerprintVersion}
	for _, fingerprint := range fingerprints {
		report.Findings = append(report.Findings, finding.Finding{
			ID: fingerprint, Fingerprint: fingerprint, RuleID: "fixture-rule", Domain: finding.Quality,
			Location: finding.Location{File: fingerprint + ".go", Line: 1},
		})
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestScanWritesSelectedReportFormat(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad.ts"), []byte("google-mock-jwt-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"sarif", "junit"} {
		t.Run(format, func(t *testing.T) {
			output := filepath.Join(root, "report."+format)
			var stdout, stderr bytes.Buffer
			args := []string{"scan", "--root", root, "--format", format, "--output", output, "--quiet"}
			if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
				t.Fatalf("scan exit=%d stderr=%s", code, stderr.String())
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) == 0 {
				t.Fatal("report is empty")
			}
		})
	}
}
