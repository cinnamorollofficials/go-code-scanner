package discovery

import (
	"context"
	"os"
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
	files, err := Files(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "main.go" {
		t.Fatalf("unexpected files: %v", files)
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
