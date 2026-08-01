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

func TestConfigRejectsRuleFilesOutsideProject(t *testing.T) {
	cfg := Default()
	cfg.Root = t.TempDir()
	cfg.RuleFiles = []string{"../outside-rules.json"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("outside-project rule file accepted")
	}

	out := t.TempDir()
	if err := os.Symlink(out, filepath.Join(cfg.Root, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	cfg.RuleFiles = []string{"linked/rules.json"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("symlink-escaped rule file accepted")
	}
}

func TestConfigRejectsSuppressionFileOutsideProject(t *testing.T) {
	cfg := Default()
	cfg.Root = t.TempDir()
	cfg.SuppressionFile = filepath.Join(t.TempDir(), "suppressions.json")
	if err := cfg.Validate(); err == nil {
		t.Fatal("outside-project suppression file accepted")
	}

	out := t.TempDir()
	if err := os.Symlink(out, filepath.Join(cfg.Root, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	cfg.SuppressionFile = "linked/suppressions.json"
	if err := cfg.Validate(); err == nil {
		t.Fatal("symlink-escaped suppression file accepted")
	}
}

func TestConfigRejectsBaselineFileOutsideProject(t *testing.T) {
	cfg := Default()
	cfg.Root = t.TempDir()
	cfg.BaselineFile = "../baseline.json"
	if err := cfg.Validate(); err == nil {
		t.Fatal("outside-project baseline file accepted")
	}

	out := t.TempDir()
	if err := os.Symlink(out, filepath.Join(cfg.Root, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	cfg.BaselineFile = "linked/baseline.json"
	if err := cfg.Validate(); err == nil {
		t.Fatal("symlink-escaped baseline file accepted")
	}
}
