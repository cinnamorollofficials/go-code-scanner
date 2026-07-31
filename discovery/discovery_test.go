package discovery

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

func TestFullDiscoveryExcludesDependencies(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "main.go"))
	write(t, filepath.Join(root, "node_modules", "ignored.js"))
	write(t, filepath.Join(root, "notes.txt"))
	cfg := config.Default()
	cfg.Root = root
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	sources, err := Sources(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || filepath.Base(sources[0].Path) != "main.go" {
		t.Fatalf("unexpected sources: %v", sources)
	}
}

func TestStagedSourceReadsGitIndexInsteadOfWorkingTree(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	path := filepath.Join(root, "app.ts")
	if err := os.WriteFile(path, []byte("staged secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "app.ts")
	if err := os.WriteFile(path, []byte("safe working tree\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Root = root
	cfg.Mode = config.ModeStaged
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	sources, err := Sources(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 {
		t.Fatalf("got %d sources, want 1", len(sources))
	}
	reader, err := sources[0].Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "staged secret\n" {
		t.Fatalf("read %q, want staged content", content)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", root}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}

func write(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
}
