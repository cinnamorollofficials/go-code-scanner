package command

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
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

func helperScanner(t *testing.T, mode, workspaceMode string) *Scanner {
	t.Helper()
	source, err := New(Spec{
		ID: "fixture", Domain: finding.Quality,
		Command:   []string{os.Args[0], "-test.run=TestCommandHelperProcess", "--", mode},
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
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
