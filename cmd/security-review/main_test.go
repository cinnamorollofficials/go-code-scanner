package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/baseline"
	cachepkg "github.com/cinnamorollofficials/go-code-scanner/cache"
	compatibilitypkg "github.com/cinnamorollofficials/go-code-scanner/compatibility"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	releasepkg "github.com/cinnamorollofficials/go-code-scanner/release"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
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

func TestScanExplainRuleAndVerboseStatus(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"scan", "--root", root, "--explain", "merge-conflict-marker"}, &stdout, &stderr); code != 0 {
		t.Fatalf("explain exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Rule: merge-conflict-marker") || !strings.Contains(stdout.String(), "Domain: quality") {
		t.Fatalf("unexpected rule explanation: %q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"scan", "--root", root, "--verbose"}, &stdout, &stderr); code != 0 {
		t.Fatalf("verbose scan exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "required=true") || !strings.Contains(stdout.String(), "capabilities=") {
		t.Fatalf("verbose scanner metadata missing: %q", stdout.String())
	}
}

func TestScanFixDryRunAndApply(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.go")
	if err := os.WriteFile(path, []byte("package app  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"scan", "--root", root, "--fix", "--dry-run", "--quiet"}, &stdout, &stderr); code != 0 {
		t.Fatalf("fix dry-run exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Would fix app.go:1") {
		t.Fatalf("missing dry-run preview: %q", stdout.String())
	}
	content, _ := os.ReadFile(path)
	if string(content) != "package app  \n" {
		t.Fatal("fix dry-run changed source")
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"scan", "--root", root, "--fix", "--quiet"}, &stdout, &stderr); code != 0 {
		t.Fatalf("fix exit=%d stderr=%s", code, stderr.String())
	}
	content, _ = os.ReadFile(path)
	if string(content) != "package app\n" {
		t.Fatalf("source was not fixed: %q", content)
	}
	var report finding.Report
	reportData, _ := os.ReadFile(filepath.Join(root, "security_findings.json"))
	if err := json.Unmarshal(reportData, &report); err != nil {
		t.Fatal(err)
	}
	for _, item := range report.Findings {
		if item.RuleID == "trailing-whitespace" {
			t.Fatal("post-fix report retained fixed finding")
		}
	}
}

func TestScanColorModes(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte("package bad  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		mode      string
		wantColor bool
	}{
		{mode: "always", wantColor: true},
		{mode: "never", wantColor: false},
		{mode: "auto", wantColor: false},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), []string{"scan", "--root", root, "--color", test.mode}, &stdout, &stderr); code != 0 {
			t.Fatalf("color=%s exit=%d stderr=%s", test.mode, code, stderr.String())
		}
		hasColor := strings.Contains(stdout.String(), "\x1b[")
		if hasColor != test.wantColor {
			t.Fatalf("color=%s hasColor=%t output=%q", test.mode, hasColor, stdout.String())
		}
	}
}

func TestSuppressAddDryRunAndWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".security-ignore")
	args := []string{"suppress", "add", "--suppression-file", path, "--file", "app.go", "--line", "7", "--rule", "security/example", "--reason", "reviewed", "--expires", "2030-01-01"}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), append(args, "--dry-run"), &stdout, &stderr); code != 0 {
		t.Fatalf("dry-run exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Would add") {
		t.Fatalf("missing dry-run preview: %q", stdout.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote suppression file")
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("write exit=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestCacheStatsAndClean(t *testing.T) {
	directory := t.TempDir()
	store := cachepkg.Store{Directory: directory}
	key, _ := cachepkg.Key(cachepkg.KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err := store.Put(key, scanner.Result{}); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(directory, "README")
	if err := os.WriteFile(foreign, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"cache", "stats", "--dir", directory}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "entries=1") {
		t.Fatalf("unexpected cache stats: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"cache", "clean", "--dir", directory}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "removed=1") {
		t.Fatalf("unexpected cache clean: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatal("cache clean removed foreign file")
	}
}

func TestReleaseVerifyCommand(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	provenance := filepath.Join(directory, "provenance.json")
	if err := os.WriteFile(provenance, []byte("provenance"), 0o600); err != nil {
		t.Fatal(err)
	}
	privateDER, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	privatePath := filepath.Join(directory, "private.pem")
	if err := os.WriteFile(privatePath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}), 0o600); err != nil {
		t.Fatal(err)
	}
	signature, err := releasepkg.SignFile(provenance, privatePath)
	if err != nil {
		t.Fatal(err)
	}
	signaturePath := filepath.Join(directory, "provenance.sig")
	if err := os.WriteFile(signaturePath, []byte(signature), 0o600); err != nil {
		t.Fatal(err)
	}
	publicDER, _ := x509.MarshalPKIXPublicKey(publicKey)
	publicPath := filepath.Join(directory, "public.pem")
	if err := os.WriteFile(publicPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	args := []string{"release", "verify", "--provenance", provenance, "--signature", signaturePath, "--public-key", publicPath}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "verified") {
		t.Fatalf("verify code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(provenance, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("tampered provenance exit=%d", code)
	}
}

func TestReleaseVerifyCommandValidatesSubjectsWhenRequested(t *testing.T) {
	directory := t.TempDir()
	artifact := filepath.Join(directory, "release.tar.gz")
	if err := os.WriteFile(artifact, []byte("artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
	provenance := filepath.Join(directory, "provenance.json")
	options := releasepkg.ProvenanceOptions{Version: "v1.2.3", Commit: "abc123", BuildDate: time.Unix(1, 0), Builder: "test"}
	if err := releasepkg.WriteProvenance(directory, provenance, options); err != nil {
		t.Fatal(err)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privatePath, publicPath := writeReleaseKeys(t, directory, privateKey, publicKey)
	signature, err := releasepkg.SignFile(provenance, privatePath)
	if err != nil {
		t.Fatal(err)
	}
	signaturePath := filepath.Join(directory, "provenance.sig")
	if err := os.WriteFile(signaturePath, []byte(signature+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"release", "verify", "--provenance", provenance, "--signature", signaturePath, "--public-key", publicPath, "--directory", directory}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("combined verification failed with %d: %s", code, stderr.String())
	}
	if err := os.WriteFile(artifact, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("expected subject mismatch exit 1, got %d: %s", code, stderr.String())
	}
}

func writeReleaseKeys(t *testing.T, directory string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (string, string) {
	t.Helper()
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

func TestReleaseArchiveCommand(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "security-review")
	if err := os.WriteFile(binary, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(directory, "security-review.tar.gz")
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"release", "archive", "--binary", binary, "--output", output, "--timestamp", "2026-01-02T03:04:05Z"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("archive failed with %d: %s", code, stderr.String())
	}
	if info, err := os.Stat(output); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("archive was not created: %v", err)
	}
}

func TestReleaseArchiveCommandRejectsInvalidInputs(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "security-review")
	if err := os.WriteFile(binary, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "binary-link")
	if err := os.Symlink(binary, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	cases := [][]string{
		{"release", "archive", "--binary", binary, "--output", filepath.Join(directory, "release.tar.gz")},
		{"release", "archive", "--binary", binary, "--output", filepath.Join(directory, "release.bin"), "--timestamp", "2026-01-02T03:04:05Z"},
		{"release", "archive", "--binary", binary, "--output", filepath.Join(directory, "release.zip"), "--timestamp", "invalid"},
		{"release", "archive", "--binary", link, "--output", filepath.Join(directory, "release.zip"), "--timestamp", "2026-01-02T03:04:05Z"},
		{"release", "archive", "--binary", binary, "--output", filepath.Join(directory, "release.zip"), "--timestamp", "2026-01-02T03:04:05Z", "extra"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, &stdout, &stderr); code != 2 {
			t.Fatalf("expected invalid arguments for %v, got %d", args, code)
		}
	}
}

func TestReleaseChecksumsVerifyCommand(t *testing.T) {
	directory := t.TempDir()
	artifact := filepath.Join(directory, "release.tar.gz")
	content := []byte("artifact")
	if err := os.WriteFile(artifact, content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	manifest := filepath.Join(directory, "SHA256SUMS")
	if err := os.WriteFile(manifest, []byte(fmt.Sprintf("%x  release.tar.gz\n", digest)), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"release", "checksums", "verify", "--manifest", manifest, "--directory", directory}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 {
		t.Fatalf("checksum verification failed with %d: %s", code, stderr.String())
	}
	if err := os.WriteFile(artifact, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("expected mismatch exit 1, got %d: %s", code, stderr.String())
	}
}

func TestReleaseChecksumsVerifyCommandRejectsInvalidInput(t *testing.T) {
	cases := [][]string{
		{"release", "checksums", "verify"},
		{"release", "checksums", "verify", "--manifest", "missing", "--directory", "."},
		{"release", "checksums", "verify", "--manifest", "missing", "--directory", ".", "extra"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, &stdout, &stderr); code != 2 {
			t.Fatalf("expected invalid input exit 2 for %v, got %d", args, code)
		}
	}
}

func TestReleaseChangelogValidateCommand(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "CHANGELOG.md")
	valid := "# Changelog\n\n## [Unreleased]\n\n## [1.0.0] - 2026-01-01\n"
	if err := os.WriteFile(path, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	args := []string{"release", "changelog", "validate", "--file", path}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "valid") {
		t.Fatalf("valid changelog exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(path, []byte("invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 {
		t.Fatalf("invalid changelog exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestUpgradeCheckCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"upgrade", "check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("print contract exit=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"config_schema": 1`) {
		t.Fatalf("current contract missing from output: %q", stdout.String())
	}

	contract := compatibilitypkg.Current()
	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "contract.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	args := []string{"upgrade", "check", "--contract", path}
	if code := run(context.Background(), args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "unchanged") {
		t.Fatalf("unchanged contract exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	contract.ReportSchema = "security-review/v-next"
	data, _ = json.Marshal(contract)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), args, &stdout, &stderr); code != 1 || !strings.Contains(stdout.String(), "migration required") {
		t.Fatalf("changed contract exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
