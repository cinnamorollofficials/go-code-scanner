package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/gitrepo"
)

func TestManagerLifecycle(t *testing.T) {
	for _, event := range []string{PreCommit, CommitMsg, PrePush} {
		t.Run(event, func(t *testing.T) {
			testManagerLifecycle(t, event)
		})
	}
}

func testManagerLifecycle(t *testing.T, event string) {
	manager, repository := newManager(t)
	ctx := context.Background()

	state, err := manager.Status(ctx, event)
	if err != nil {
		t.Fatal(err)
	}
	if state != Missing {
		t.Fatalf("expected missing hook, got %s", state)
	}
	if err := manager.Install(ctx, event); err != nil {
		t.Fatal(err)
	}
	if err := manager.Install(ctx, event); err != nil {
		t.Fatalf("idempotent install failed: %v", err)
	}
	state, err = manager.Status(ctx, event)
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
	content, err := os.ReadFile(filepath.Join(hooksDir, event))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), marker) || !strings.Contains(string(content), "hook run "+event) {
		t.Fatalf("unexpected hook content: %s", content)
	}
	if event == CommitMsg && !strings.Contains(string(content), `--file "$1"`) {
		t.Fatalf("commit-msg hook does not forward the message file: %s", content)
	}

	if err := manager.Uninstall(ctx, event); err != nil {
		t.Fatal(err)
	}
	if err := manager.Uninstall(ctx, event); err != nil {
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

func TestManagerRefusesHookSymlinkWithoutTouchingTarget(t *testing.T) {
	manager, repository := newManager(t)
	ctx := context.Background()
	hooksDir, err := repository.HooksDir(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside-hook")
	content := []byte("#!/bin/sh\necho outside\n")
	if err := os.WriteFile(outside, content, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(hooksDir, PreCommit)
	if err := os.Symlink(outside, target); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if state, err := manager.Status(ctx, PreCommit); err != nil || state != Conflict {
		t.Fatalf("symlink status=%s err=%v", state, err)
	}
	if err := manager.Install(ctx, PreCommit); err == nil {
		t.Fatal("hook install accepted symlink target")
	}
	if err := manager.Uninstall(ctx, PreCommit); err == nil {
		t.Fatal("hook uninstall accepted symlink target")
	}
	got, err := os.ReadFile(outside)
	if err != nil || string(got) != string(content) {
		t.Fatalf("outside hook changed: content=%q err=%v", got, err)
	}
}

func newManager(t *testing.T) (*Manager, *gitrepo.Repository) {
	t.Helper()
	root := t.TempDir()
	cmd := exec.Command("git", "-C", root, "init")
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test Runner",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test Runner",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
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
