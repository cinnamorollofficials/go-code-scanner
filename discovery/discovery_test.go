package discovery

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
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

func TestFilesIncludesMetadataCandidatesWithoutScanningDependencies(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "Dockerfile"))
	write(t, filepath.Join(root, "debug.dump"))
	write(t, filepath.Join(root, "node_modules", "ignored.tmp"))
	cfg := config.Default()
	cfg.Root = root
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	files, err := Files(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := sourceBasenames(files); len(got) != 2 || got[0] != "Dockerfile" || got[1] != "debug.dump" {
		t.Fatalf("unexpected metadata candidates: %v", got)
	}
	sources, err := Sources(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 {
		t.Fatalf("metadata-only files leaked into regex sources: %v", sourceBasenames(sources))
	}
}

func TestFilesIncludesRegexExcludedLockfileForMetadataPolicy(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "package-lock.json"))
	cfg := config.Default()
	cfg.Root = root
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	files, err := Files(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := sourceBasenames(files); len(got) != 1 || got[0] != "package-lock.json" {
		t.Fatalf("lockfile missing from metadata candidates: %v", got)
	}
	sources, err := Sources(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 {
		t.Fatalf("lockfile leaked into regex sources: %v", sourceBasenames(sources))
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

func TestStagedDiscoveryPreservesSpecialCharactersInPath(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	name := "odd name\tline\napp.ts"
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("staged content\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "--", name)

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
	if sources[0].Path != path {
		t.Fatalf("path was not preserved: got %q want %q", sources[0].Path, path)
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
	if string(content) != "staged content\n" {
		t.Fatalf("unexpected staged content %q", content)
	}
}

func TestStagedDiscoveryHandlesGitChangeKinds(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "scanner@example.invalid")
	runGit(t, root, "config", "user.name", "Scanner Test")
	for _, name := range []string{"modified.go", "renamed.go", "copied.go", "deleted.go"} {
		writeContent(t, filepath.Join(root, name), "package fixture\n")
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "test: initial fixture")

	writeContent(t, filepath.Join(root, "added.go"), "package added\n")
	writeContent(t, filepath.Join(root, "modified.go"), "package modified\n")
	runGit(t, root, "mv", "renamed.go", "renamed_new.go")
	copyContent, err := os.ReadFile(filepath.Join(root, "copied.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "copied_new.go"), copyContent, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "deleted.go")); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "-A")

	sources := stagedSources(t, root)
	got := sourceBasenames(sources)
	want := []string{"added.go", "copied_new.go", "modified.go", "renamed_new.go"}
	if len(got) != len(want) {
		t.Fatalf("staged sources=%v want=%v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("staged sources=%v want=%v", got, want)
		}
	}
}

func TestChangedDiscoveryWorksWithoutHEAD(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	writeContent(t, filepath.Join(root, "first.go"), "package first\n")
	runGit(t, root, "add", "first.go")
	cfg := config.Default()
	cfg.Root = root
	cfg.Mode = config.ModeChanged
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	sources, err := Sources(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := sourceBasenames(sources); len(got) != 1 || got[0] != "first.go" {
		t.Fatalf("unexpected unborn-HEAD sources: %v", got)
	}
}

func TestStagedSymlinkDoesNotReadOutsideRepository(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	outside := filepath.Join(t.TempDir(), "secret.ts")
	writeContent(t, outside, "google-mock-jwt-token\n")
	link := filepath.Join(root, "linked.ts")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	runGit(t, root, "add", "linked.ts")
	sources := stagedSources(t, root)
	if len(sources) != 1 {
		t.Fatalf("got %d symlink sources, want 1", len(sources))
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
	if string(content) != outside {
		t.Fatalf("staged symlink followed its target: got %q want link payload %q", content, outside)
	}
}

func stagedSources(t *testing.T, root string) []scanner.Source {
	t.Helper()
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
	return sources
}

func sourceBasenames(sources []scanner.Source) []string {
	names := make([]string, len(sources))
	for index, source := range sources {
		names[index] = filepath.Base(source.Path)
	}
	sort.Strings(names)
	return names
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

func writeContent(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
