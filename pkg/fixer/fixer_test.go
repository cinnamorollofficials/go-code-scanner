package fixer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestApplyTrailingWhitespaceDryRunAndWrite(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.go")
	original := []byte("package app  \r\n\r\nfunc run() {}\t\n")
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatal(err)
	}
	findings := []finding.Finding{
		{RuleID: "trailing-whitespace", Fixable: true, Location: finding.Location{File: "app.go", Line: 1}},
		{RuleID: "trailing-whitespace", Fixable: true, Location: finding.Location{File: "app.go", Line: 3}},
	}
	changes, err := Apply(root, findings, true)
	if err != nil || len(changes) != 2 {
		t.Fatalf("dry-run changes=%v err=%v", changes, err)
	}
	unchanged, _ := os.ReadFile(path)
	if string(unchanged) != string(original) {
		t.Fatal("dry-run modified file")
	}
	changes, err = Apply(root, findings, false)
	if err != nil || len(changes) != 2 {
		t.Fatalf("apply changes=%v err=%v", changes, err)
	}
	updated, _ := os.ReadFile(path)
	if string(updated) != "package app\r\n\r\nfunc run() {}\n" {
		t.Fatalf("unexpected fixed content %q", updated)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("file mode changed to %o", info.Mode().Perm())
	}
}

func TestApplyRejectsEscapingPath(t *testing.T) {
	_, err := Apply(t.TempDir(), []finding.Finding{{
		RuleID: "trailing-whitespace", Fixable: true, Location: finding.Location{File: "../outside.go", Line: 1},
	}}, false)
	if err == nil {
		t.Fatal("expected escaping path error")
	}
}

func TestApplyRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(target, []byte("package outside  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "link.go")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := Apply(root, []finding.Finding{{
		RuleID: "trailing-whitespace", Fixable: true, Location: finding.Location{File: "link.go", Line: 1},
	}}, true)
	if err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestApplySQLInjectionFix(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "repo.go")
	original := []byte("package repo\n\nfunc findUser(id string) string {\n\tquery := \"SELECT * FROM users WHERE id = \" + id\n\treturn query\n}\n")
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatal(err)
	}

	findings := []finding.Finding{
		{RuleID: "SQLI-001", Fixable: true, Location: finding.Location{File: "repo.go", Line: 4}},
	}
	changes, err := Apply(root, findings, false)
	if err != nil || len(changes) != 1 {
		t.Fatalf("apply changes=%v err=%v", changes, err)
	}

	updated, _ := os.ReadFile(path)
	if !strings.Contains(string(updated), "$1") {
		t.Fatalf("expected $1 in fixed content, got: %s", string(updated))
	}
}

