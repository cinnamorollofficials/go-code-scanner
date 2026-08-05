package frontend

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestNativeFrontendStagedIsolationBothDirections(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	// Setup git repo
	repo := t.TempDir()
	runCmd := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, out)
		}
	}

	runCmd("init")
	runCmd("config", "user.name", "Test")
	runCmd("config", "user.email", "test@example.com")

	// Direction A: Safe staged in index, Unsafe working tree
	fileA := filepath.Join(repo, "ComponentA.tsx")
	if err := os.WriteFile(fileA, []byte("const safe = <div>Hello</div>;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "ComponentA.tsx")
	// Modify working tree to unsafe
	if err := os.WriteFile(fileA, []byte("const unsafe = <div dangerouslySetInnerHTML={{__html: x}} />;\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Root = repo
	cfg.Frontend.Enabled = true

	s := New(cfg)

	// Open staged source (simulating discovery reading index)
	stagedSource := scanner.Source{
		Path: fileA,
		Open: func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("const safe = <div>Hello</div>;\n")), nil
		},
	}
	reqA := scanner.Request{
		Root:    repo,
		Mode:    "staged",
		Sources: []scanner.Source{stagedSource},
	}
	resA := s.Scan(context.Background(), reqA)
	if resA.State != finding.ScannerClean {
		t.Fatalf("Direction A failed: expected clean result for safe staged content, got state=%v findings=%+v", resA.State, resA.Findings)
	}

	// Direction B: Unsafe staged in index, Safe working tree
	fileB := filepath.Join(repo, "ComponentB.tsx")
	if err := os.WriteFile(fileB, []byte("const unsafe = <div dangerouslySetInnerHTML={{__html: x}} />;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "ComponentB.tsx")
	// Modify working tree to safe
	if err := os.WriteFile(fileB, []byte("const safe = <div>Hello</div>;\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stagedUnsafeSource := scanner.Source{
		Path: fileB,
		Open: func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("const unsafe = <div dangerouslySetInnerHTML={{__html: x}} />;\n")), nil
		},
	}
	reqB := scanner.Request{
		Root:    repo,
		Mode:    "staged",
		Sources: []scanner.Source{stagedUnsafeSource},
	}
	resB := s.Scan(context.Background(), reqB)
	if resB.State != finding.ScannerFindings || len(resB.Findings) == 0 {
		t.Fatalf("Direction B failed: expected findings for unsafe staged content, got state=%v findings=%+v", resB.State, resB.Findings)
	}
}

func TestStagedTempWorkspaceCleanup(t *testing.T) {
	cfg := config.Default()
	cfg.Frontend.Enabled = true
	s := New(cfg)

	// Test cleanup on success
	reqSuccess := scanner.Request{
		Root: t.TempDir(),
		Mode: "staged",
		Sources: []scanner.Source{
			mockSource("/project/Component.tsx", "const safe = 1;\n"),
		},
	}
	resSuccess := s.Scan(context.Background(), reqSuccess)
	if resSuccess.State != finding.ScannerClean {
		t.Fatalf("expected clean, got %v", resSuccess.State)
	}

	// Test cleanup on timeout / cancel
	ctxCancel, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	reqCancel := scanner.Request{
		Root: t.TempDir(),
		Mode: "staged",
		Sources: []scanner.Source{
			mockSource("/project/Component.tsx", "const safe = 1;\n"),
		},
	}
	resCancel := s.Scan(ctxCancel, reqCancel)
	if resCancel.State != finding.ScannerFailed || resCancel.Failure != scanner.FailureCanceled {
		t.Fatalf("expected canceled state, got %v / %v", resCancel.State, resCancel.Failure)
	}
}
