package scripts

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseBinaryHookLifecycle(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(t.TempDir(), "security-review")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-trimpath", "-buildvcs=false", "-o", binary, "./cmd/security-review")
	build.Dir = ".."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build release binary: %v: %s", err, output)
	}
	runHookGit(t, root, "init")
	hooksDirectory := filepath.Join(root, ".git", "hooks")
	unmanaged := filepath.Join(hooksDirectory, "pre-commit")
	const unmanagedContent = "#!/bin/sh\necho unmanaged\n"
	if err := os.WriteFile(unmanaged, []byte(unmanagedContent), 0o700); err != nil {
		t.Fatal(err)
	}
	if code, _ := runHookCLI(binary, "hook", "install", "pre-commit", "--root", root); code != 3 {
		t.Fatalf("expected unmanaged hook conflict, got exit %d", code)
	}
	content, err := os.ReadFile(unmanaged)
	if err != nil || string(content) != unmanagedContent {
		t.Fatalf("existing hook changed: content=%q err=%v", content, err)
	}
	if err := os.Remove(unmanaged); err != nil {
		t.Fatal(err)
	}

	messageFile := filepath.Join(root, "COMMIT_EDITMSG")
	if err := os.WriteFile(messageFile, []byte("feat: validate release hook\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, event := range []string{"pre-commit", "commit-msg", "pre-push"} {
		t.Run(event, func(t *testing.T) {
			if code, output := runHookCLI(binary, "hook", "install", event, "--root", root); code != 0 {
				t.Fatalf("install exit=%d: %s", code, output)
			}
			if code, output := runHookCLI(binary, "hook", "status", event, "--root", root); code != 0 || !strings.Contains(output, event+": installed") {
				t.Fatalf("status exit=%d: %s", code, output)
			}
			args := []string{"hook", "run", event}
			if event == "commit-msg" {
				args = append(args, "--", messageFile)
			}
			command := exec.Command("git", append([]string{"-C", root}, args...)...)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("execute installed hook: %v: %s", err, output)
			}
			if code, output := runHookCLI(binary, "hook", "uninstall", event, "--root", root); code != 0 {
				t.Fatalf("uninstall exit=%d: %s", code, output)
			}
			if code, output := runHookCLI(binary, "hook", "status", event, "--root", root); code != 0 || !strings.Contains(output, event+": missing") {
				t.Fatalf("missing status exit=%d: %s", code, output)
			}
		})
	}
}

func runHookCLI(binary string, args ...string) (int, string) {
	command := exec.Command(binary, args...)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := command.Run()
	if err == nil {
		return 0, output.String()
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode(), output.String()
	}
	return -1, output.String()
}

func runHookGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
