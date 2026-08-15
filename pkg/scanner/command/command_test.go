package command

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestCommandScannerStates(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		state finding.ScannerState
		count int
	}{
		{name: "clean", mode: "clean", state: finding.ScannerClean},
		{name: "findings", mode: "findings", state: finding.ScannerFindings, count: 1},
		{name: "failure", mode: "failure", state: finding.ScannerFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := helperScanner(t, test.mode, WorkspaceRoot)
			result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
			if result.State != test.state || len(result.Findings) != test.count {
				t.Fatalf("unexpected result: %+v", result)
			}
			if test.count == 1 && result.Findings[0].Domain != finding.Quality {
				t.Fatalf("unexpected finding: %+v", result.Findings[0])
			}
		})
	}
}

func TestCommandScannerDescriptor(t *testing.T) {
	source := helperScanner(t, "clean", WorkspaceStaged)
	descriptor := source.Describe()
	if descriptor.Domain != finding.Quality || len(descriptor.SupportedModes) != 1 || descriptor.SupportedModes[0] != "staged" {
		t.Fatalf("unexpected descriptor: %+v", descriptor)
	}
}

func TestMissingCommandCanSkipOrFail(t *testing.T) {
	for _, test := range []struct {
		behavior string
		state    finding.ScannerState
	}{
		{behavior: OnMissingSkip, state: finding.ScannerSkipped},
		{behavior: OnMissingFail, state: finding.ScannerFailed},
	} {
		source, err := New(Spec{
			ID: "missing", Domain: finding.Quality, Command: []string{"definitely-not-a-real-security-review-command"},
			OnMissing: test.behavior, Severity: finding.High, Category: "tool", Description: "missing tool",
		})
		if err != nil {
			t.Fatal(err)
		}
		result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
		if result.State != test.state {
			t.Fatalf("on_missing=%s: expected %s, got %+v", test.behavior, test.state, result)
		}
	}
}

func TestCommandScannerRejectsRelativeExecutablePaths(t *testing.T) {
	for _, executable := range []string{"./tool", "../tool", "tools/tool", `tools\tool`} {
		_, err := New(Spec{ID: "unsafe", Domain: finding.Quality, Command: []string{executable}, Severity: finding.High, Category: "tool", Description: "unsafe tool"})
		if err == nil {
			t.Fatalf("relative executable path %q accepted", executable)
		}
	}
}

func TestAbsoluteExecutableSymlinkIsNotExecuted(t *testing.T) {
	target, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "tool")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	source, err := New(Spec{ID: "symlink", Domain: finding.Quality, Command: []string{link}, OnMissing: OnMissingFail, Severity: finding.High, Category: "tool", Description: "symlink tool"})
	if err != nil {
		t.Fatal(err)
	}
	result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerFailed || result.Failure != scanner.FailureMissing {
		t.Fatalf("absolute executable symlink was not rejected: %+v", result)
	}
}

func TestCommandScannerUsesStagedWorkspace(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	path := filepath.Join(root, "value.txt")
	if err := os.WriteFile(path, []byte("staged"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "value.txt")
	if err := os.WriteFile(path, []byte("unstaged"), 0o600); err != nil {
		t.Fatal(err)
	}

	source := helperScanner(t, "check-staged", WorkspaceStaged)
	result := source.Scan(context.Background(), scanner.Request{Root: root, Mode: "staged"})
	if result.State != finding.ScannerClean {
		t.Fatalf("expected staged snapshot content, got %+v", result)
	}
}

func TestCommandScannerStagedIsolationBothDirections(t *testing.T) {
	for _, test := range []struct {
		name, staged, working string
		wantState             finding.ScannerState
	}{
		{name: "safe staged unsafe working", staged: "safe", working: "unsafe", wantState: finding.ScannerClean},
		{name: "unsafe staged safe working", staged: "unsafe", working: "safe", wantState: finding.ScannerFindings},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			runGit(t, root, "init")
			path := filepath.Join(root, "value.txt")
			if err := os.WriteFile(path, []byte(test.staged), 0o600); err != nil {
				t.Fatal(err)
			}
			runGit(t, root, "add", "value.txt")
			if err := os.WriteFile(path, []byte(test.working), 0o600); err != nil {
				t.Fatal(err)
			}
			source := helperScanner(t, "find-staged-unsafe", WorkspaceStaged)
			result := source.Scan(context.Background(), scanner.Request{Root: root, Mode: "staged"})
			if result.State != test.wantState {
				t.Fatalf("staged isolation state=%s want=%s: %+v", result.State, test.wantState, result)
			}
		})
	}
}

func TestCommandScannerEnforcesConfiguredSnapshotLimits(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	for _, name := range []string{"one.txt", "two.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, root, "add", "one.txt", "two.txt")
	source := helperScanner(t, "clean", WorkspaceStaged)
	source.spec.SnapshotMaxFiles = 1
	result := source.Scan(context.Background(), scanner.Request{Root: root, Mode: "staged"})
	if result.State != finding.ScannerFailed || !strings.Contains(result.Message, "file limit") {
		t.Fatalf("expected snapshot limit failure, got %+v", result)
	}
}

func TestCommandScannerDoesNotExposeExternalDiagnostics(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(root, "scanner")
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-specific")
	}
	if err := os.WriteFile(executable, []byte("#!/bin/sh\necho 'CANARY-SECRET-DO-NOT-LEAK' >&2\nexit 2\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	source, err := New(Spec{
		ID: "external", Domain: finding.Security, Command: []string{executable},
		Severity: finding.High, Category: "external", Description: "external finding",
	})
	if err != nil {
		t.Fatal(err)
	}
	result := source.Scan(context.Background(), scanner.Request{Root: root})
	if result.State != finding.ScannerFailed {
		t.Fatalf("expected scanner failure, got %+v", result)
	}
	if strings.Contains(result.Message, "CANARY-SECRET-DO-NOT-LEAK") {
		t.Fatalf("external diagnostic leaked into scanner result: %s", result.Message)
	}
}

func TestCommandScannerParsesJSONLines(t *testing.T) {
	source := helperScanner(t, "json-lines", WorkspaceRoot)
	source.spec.OutputFormat = OutputJSONLines
	result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerFindings || len(result.Findings) != 1 {
		t.Fatalf("unexpected structured result: %+v", result)
	}
	item := result.Findings[0]
	if item.RuleID != "external-rule" || item.Tool != "fixture" || item.Domain != finding.Quality || item.Location.File != "src/app.go" || item.Location.Line != 9 {
		t.Fatalf("unexpected structured finding: %+v", item)
	}
}

func TestCommandScannerParsesSuccessfulPathOutput(t *testing.T) {
	source := helperScanner(t, "path-output", WorkspaceRoot)
	source.spec.OutputFormat = OutputPaths
	source.spec.FindingsOnOutput = true
	result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerFindings || len(result.Findings) != 2 {
		t.Fatalf("unexpected path output result: %+v", result)
	}
	if result.Findings[0].Location.File != "a.go" || result.Findings[1].Location.File != "nested/b.go" {
		t.Fatalf("unexpected path findings: %+v", result.Findings)
	}
}

func TestCommandScannerParsesAndCleansOutputFile(t *testing.T) {
	temporaryRoot := t.TempDir()
	t.Setenv("TMPDIR", temporaryRoot)
	source := helperScanner(t, "output-file", WorkspaceRoot)
	source.spec.Command = append(source.spec.Command, "{output}")
	source.spec.OutputFile = true
	source.spec.Parser = func(data []byte) ([]ParsedFinding, error) {
		if string(data) != `{"rule":"fixture"}` {
			t.Fatalf("unexpected output file data: %q", data)
		}
		return []ParsedFinding{{RuleID: "fixture-rule", File: "app.go", Line: 2}}, nil
	}
	result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerFindings || len(result.Findings) != 1 || result.Findings[0].RuleID != "fixture-rule" {
		t.Fatalf("unexpected output-file result: %+v", result)
	}
	entries, err := os.ReadDir(temporaryRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "security-review-output-") {
			t.Fatalf("temporary adapter output was not removed: %s", entry.Name())
		}
	}
}

func TestCommandScannerRejectsEscapingStructuredPath(t *testing.T) {
	source := helperScanner(t, "json-escape", WorkspaceRoot)
	source.spec.OutputFormat = OutputJSONLines
	result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerFailed {
		t.Fatalf("expected invalid output path failure, got %+v", result)
	}
}

func TestCommandScannerRejectsInvalidEnvironmentName(t *testing.T) {
	_, err := New(Spec{
		ID: "invalid-environment", Domain: finding.Quality, Command: []string{"tool"},
		Severity: finding.High, Category: "fixture", Description: "fixture",
		Environment: []string{"lowercase-secret"},
	})
	if err == nil {
		t.Fatal("expected invalid environment name error")
	}
}

func TestCommandScannerFiltersEnvironment(t *testing.T) {
	t.Setenv("COMMAND_SCANNER_SECRET", "must-not-leak")
	source := helperScanner(t, "environment-filtered", WorkspaceRoot)
	result := source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerClean {
		t.Fatalf("unexpected filtered environment result: %+v", result)
	}

	source = helperScanner(t, "environment-allowed", WorkspaceRoot)
	source.spec.Environment = []string{"COMMAND_SCANNER_SECRET"}
	result = source.Scan(context.Background(), scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerClean {
		t.Fatalf("unexpected allowed environment result: %+v", result)
	}
}

func helperScanner(t *testing.T, mode, workspaceMode string) *Scanner {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	source, err := New(Spec{
		ID: "fixture", Domain: finding.Quality,
		Command:   []string{exe, "-test.run=TestCommandHelperProcess", "--", mode},
		Workspace: workspaceMode, FindingExitCodes: []int{10},
		Severity: finding.High, Category: "fixture", Description: "fixture finding",
	})
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestCommandHelperProcess(t *testing.T) {
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		return
	}
	switch os.Args[separator+1] {
	case "clean":
		os.Exit(0)
	case "findings":
		os.Exit(10)
	case "failure":
		os.Exit(11)
	case "check-staged":
		content, err := os.ReadFile("value.txt")
		if err != nil || string(content) != "staged" {
			os.Exit(12)
		}
		os.Exit(0)
	case "find-staged-unsafe":
		content, err := os.ReadFile("value.txt")
		if err != nil {
			os.Exit(12)
		}
		if string(content) == "unsafe" {
			os.Exit(10)
		}
		os.Exit(0)
	case "json-lines":
		_, _ = os.Stdout.WriteString(`{"rule_id":"external-rule","category":"lint","severity":"MEDIUM","description":"external finding","file":"src/app.go","line":9}` + "\n")
		os.Exit(10)
	case "json-escape":
		_, _ = os.Stdout.WriteString(`{"rule_id":"escape","file":"../secret","line":1}` + "\n")
		os.Exit(10)
	case "path-output":
		_, _ = os.Stdout.WriteString("a.go\nnested/b.go\na.go\n")
		os.Exit(0)
	case "output-file":
		if separator+2 >= len(os.Args) {
			os.Exit(12)
		}
		if err := os.WriteFile(os.Args[separator+2], []byte(`{"rule":"fixture"}`), 0o600); err != nil {
			os.Exit(12)
		}
		os.Exit(10)
	case "environment-filtered":
		if os.Getenv("COMMAND_SCANNER_SECRET") != "" {
			os.Exit(12)
		}
		os.Exit(0)
	case "environment-allowed":
		if os.Getenv("COMMAND_SCANNER_SECRET") != "must-not-leak" {
			os.Exit(12)
		}
		os.Exit(0)
	case "spawn-child":
		if separator+2 >= len(os.Args) {
			os.Exit(12)
		}
		exe, err := os.Executable()
		if err != nil {
			exe = os.Args[0]
		}
		child := exec.Command(exe, "-test.run=TestCommandHelperProcess", "--", "wait-child")
		if err := child.Start(); err != nil {
			os.Exit(12)
		}
		if err := os.WriteFile(os.Args[separator+2], []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(12)
		}
		_ = child.Wait()
		os.Exit(0)
	case "wait-child":
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	cmd := exec.Command("git", commandArgs...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test Runner",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test Runner",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
