package securityreview

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestBuiltInScannersUseOnlyStagedContent(t *testing.T) {
	t.Run("pattern", func(t *testing.T) {
		for _, test := range []struct {
			name, staged, working string
			wantFinding           bool
		}{
			{name: "safe staged unsafe working", staged: "package fixture\n", working: "const token = \"google-mock-jwt-token\"\n", wantFinding: false},
			{name: "unsafe staged safe working", staged: "const token = \"google-mock-jwt-token\"\n", working: "package fixture\n", wantFinding: true},
		} {
			t.Run(test.name, func(t *testing.T) {
				root := stagedFixtureRepository(t, "app.go", test.staged, test.working)
				cfg := config.Default()
				cfg.Root, cfg.Mode = root, config.ModeStaged
				reviewer, err := New(cfg)
				if err != nil {
					t.Fatal(err)
				}
				report, err := reviewer.Run(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				got := hasRule(report.Findings, "mock-token")
				if got != test.wantFinding {
					t.Fatalf("pattern staged isolation finding=%t want=%t: %+v", got, test.wantFinding, report.Findings)
				}
			})
		}
	})

	t.Run("architecture", func(t *testing.T) {
		unsafe := "package domain\nimport \"example.test/app/internal/infra\"\n"
		safe := "package domain\n"
		for _, test := range []struct {
			name, staged, working string
			wantFinding           bool
		}{
			{name: "safe staged unsafe working", staged: safe, working: unsafe, wantFinding: false},
			{name: "unsafe staged safe working", staged: unsafe, working: safe, wantFinding: true},
		} {
			t.Run(test.name, func(t *testing.T) {
				root := t.TempDir()
				resourceRunGit(t, root, "init")
				writeStagedFixture(t, root, "go.mod", "module example.test/app\n", "module example.test/app\n")
				writeStagedFixture(t, root, "internal/infra/store.go", "package infra\n", "package infra\n")
				writeStagedFixture(t, root, "internal/domain/service.go", test.staged, test.working)
				cfg := config.Default()
				cfg.Root, cfg.Mode = root, config.ModeStaged
				cfg.Architecture = config.ArchitecturePolicy{
					Layers: []config.ArchitectureLayer{
						{Name: "domain", Paths: []string{"internal/domain/*.go"}},
						{Name: "infra", Paths: []string{"internal/infra/*.go"}},
					},
					ForbiddenDependencies: []config.ForbiddenDependency{{From: "domain", To: "infra"}},
				}
				reviewer, err := New(cfg)
				if err != nil {
					t.Fatal(err)
				}
				report, err := reviewer.Run(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				got := hasRule(report.Findings, "architecture/forbidden-dependency")
				if got != test.wantFinding {
					t.Fatalf("architecture staged isolation finding=%t want=%t: %+v", got, test.wantFinding, report.Findings)
				}
			})
		}
	})
}

func stagedFixtureRepository(t *testing.T, name, staged, working string) string {
	t.Helper()
	root := t.TempDir()
	resourceRunGit(t, root, "init")
	writeStagedFixture(t, root, name, staged, working)
	return root
}

func writeStagedFixture(t *testing.T, root, name, staged, working string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(staged), 0o600); err != nil {
		t.Fatal(err)
	}
	resourceRunGit(t, root, "add", name)
	if err := os.WriteFile(path, []byte(working), 0o600); err != nil {
		t.Fatal(err)
	}
}

func hasRule(findings []finding.Finding, ruleID string) bool {
	for _, item := range findings {
		if item.RuleID == ruleID {
			return true
		}
	}
	return false
}
