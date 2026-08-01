package command

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

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
	case "json-lines":
		_, _ = os.Stdout.WriteString(`{"rule_id":"external-rule","category":"lint","severity":"MEDIUM","description":"external finding","file":"src/app.go","line":9}` + "\n")
		os.Exit(10)
	case "json-escape":
		_, _ = os.Stdout.WriteString(`{"rule_id":"escape","file":"../secret","line":1}` + "\n")
		os.Exit(10)
	case "path-output":
		_, _ = os.Stdout.WriteString("a.go\nnested/b.go\na.go\n")
		os.Exit(0)
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
		child := exec.Command(os.Args[0], "-test.run=TestCommandHelperProcess", "--", "wait-child")
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
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
