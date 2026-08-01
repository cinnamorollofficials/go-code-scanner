package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/gitrepo"
)

func TestMaterializeIndexUsesStagedContent(t *testing.T) {
	repository := initRepository(t)
	path := filepath.Join(repository.Root(), "nested", "app.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("staged\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, repository.Root(), "add", "nested/app.go")
	if err := os.WriteFile(path, []byte("unstaged\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	snapshot, err := MaterializeIndex(context.Background(), repository, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	snapshotRoot := snapshot.Root()
	defer snapshot.Close()
	content, err := os.ReadFile(filepath.Join(snapshotRoot, "nested", "app.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "staged\n" {
		t.Fatalf("expected staged content, got %q", content)
	}
	if _, err := os.Stat(filepath.Join(snapshotRoot, ".git")); !os.IsNotExist(err) {
		t.Fatal("snapshot unexpectedly contains Git metadata")
	}
	if err := snapshot.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(snapshotRoot); !os.IsNotExist(err) {
		t.Fatal("snapshot was not removed")
	}
}

func TestMaterializeIndexEnforcesLimits(t *testing.T) {
	repository := initRepository(t)
	for _, name := range []string{"one.go", "two.go"} {
		if err := os.WriteFile(filepath.Join(repository.Root(), name), []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, repository.Root(), "add", "one.go", "two.go")

	_, err := MaterializeIndex(context.Background(), repository, Limits{MaxFiles: 1, MaxBytes: 100})
	if err == nil || !strings.Contains(err.Error(), "file limit") {
		t.Fatalf("expected file limit error, got %v", err)
	}
}

func TestMaterializeIndexRejectsEscapingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink behavior requires additional privileges on Windows")
	}
	repository := initRepository(t)
	path := filepath.Join(repository.Root(), "escape")
	if err := os.Symlink("../outside", path); err != nil {
		t.Fatal(err)
	}
	runGit(t, repository.Root(), "add", "escape")

	_, err := MaterializeIndex(context.Background(), repository, DefaultLimits())
	if err == nil || !strings.Contains(err.Error(), "escapes staged workspace") {
		t.Fatalf("expected escaping symlink error, got %v", err)
	}
}

func initRepository(t *testing.T) *gitrepo.Repository {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	repository, err := gitrepo.Open(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
