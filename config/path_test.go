package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectPathRejectsTraversalAndSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	if got, err := ResolveProjectPath(root, "reports/result.json"); err != nil || got != filepath.Join(root, "reports/result.json") {
		t.Fatalf("safe path rejected: got=%q err=%v", got, err)
	}
	if _, err := ResolveProjectPath(root, "../outside.json"); err == nil {
		t.Fatal("lexical traversal accepted")
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "reports")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := ResolveProjectPath(root, "reports/result.json"); err == nil {
		t.Fatal("symlink escape accepted")
	}
}
