package gitrepo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOpenFromNestedDirectory(t *testing.T) {
	root := initRepository(t)
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	repository, err := Open(context.Background(), nested)
	if err != nil {
		t.Fatal(err)
	}
	if repository.Root() != root {
		t.Fatalf("expected root %q, got %q", root, repository.Root())
	}
}

func TestHooksDirRespectsCoreHooksPath(t *testing.T) {
	root := initRepository(t)
	runGit(t, root, "config", "core.hooksPath", ".custom-hooks")
	repository, err := Open(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	hooksDir, err := repository.HooksDir(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, ".custom-hooks")
	if hooksDir != want {
		t.Fatalf("expected hooks directory %q, got %q", want, hooksDir)
	}
}

func TestGitPathRejectsEmptyName(t *testing.T) {
	repository := &Repository{root: t.TempDir()}
	if _, err := repository.GitPath(context.Background(), ""); err == nil {
		t.Fatal("expected empty name error")
	}
}

func initRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
