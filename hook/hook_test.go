package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/gitrepo"
)

func TestManagerLifecycle(t *testing.T) {
	manager, repository := newManager(t)
	ctx := context.Background()

	state, err := manager.Status(ctx, PreCommit)
	if err != nil {
		t.Fatal(err)
	}
	if state != Missing {
		t.Fatalf("expected missing hook, got %s", state)
	}
	if err := manager.Install(ctx, PreCommit); err != nil {
		t.Fatal(err)
	}
	if err := manager.Install(ctx, PreCommit); err != nil {
		t.Fatalf("idempotent install failed: %v", err)
	}
	state, err = manager.Status(ctx, PreCommit)
	if err != nil {
		t.Fatal(err)
	}
	if state != Installed {
		t.Fatalf("expected installed hook, got %s", state)
	}

	hooksDir, err := repository.HooksDir(ctx)
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(hooksDir, PreCommit))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), marker) || !strings.Contains(string(content), "hook run pre-commit") {
		t.Fatalf("unexpected hook content: %s", content)
	}

	if err := manager.Uninstall(ctx, PreCommit); err != nil {
		t.Fatal(err)
	}
	if err := manager.Uninstall(ctx, PreCommit); err != nil {
		t.Fatalf("idempotent uninstall failed: %v", err)
	}
}

func TestManagerRefusesUnmanagedHook(t *testing.T) {
	manager, repository := newManager(t)
	ctx := context.Background()
	hooksDir, err := repository.HooksDir(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(hooksDir, PreCommit)
	if err := os.WriteFile(target, []byte("#!/bin/sh\necho user-hook\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := manager.Install(ctx, PreCommit); err == nil {
		t.Fatal("expected install conflict")
	}
	if err := manager.Uninstall(ctx, PreCommit); err == nil {
		t.Fatal("expected uninstall conflict")
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "user-hook") {
		t.Fatal("unmanaged hook was modified")
	}
}

func newManager(t *testing.T) (*Manager, *gitrepo.Repository) {
	t.Helper()
	root := t.TempDir()
	if output, err := exec.Command("git", "-C", root, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	repository, err := gitrepo.Open(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := NewManager(repository, filepath.Join(root, "bin", "security-review"))
	if err != nil {
		t.Fatal(err)
	}
	return manager, repository
}
