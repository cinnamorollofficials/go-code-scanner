package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if !filepathIsAbsolute(cfg.Root) {
		t.Fatalf("expected absolute root, got %q", cfg.Root)
	}
}

func TestConfigRejectsInvalidWorkerCount(t *testing.T) {
	cfg := Default()
	cfg.Workers = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected worker validation error")
	}
}

func TestDefaultProfilesUseBuiltInPatternScanner(t *testing.T) {
	cfg := Default()
	for _, name := range []string{ProfileFast, ProfileStandard, ProfileFull} {
		scanners := cfg.Profiles[name]
		if len(scanners) != 1 || scanners[0] != "pattern" {
			t.Fatalf("unexpected %s profile: %v", name, scanners)
		}
	}
}

func TestDefaultPreCommitHookUsesFastStagedProfile(t *testing.T) {
	hook := Default().Hooks.PreCommit
	if !hook.Enabled || !hook.StagedOnly || !hook.NewOnly || hook.Profile != ProfileFast {
		t.Fatalf("unexpected default pre-commit hook: %+v", hook)
	}
}

func TestDefaultBaselinePath(t *testing.T) {
	if got := Default().BaselineFile; got != ".security-baseline.json" {
		t.Fatalf("unexpected default baseline path %q", got)
	}
}

func TestConfigRejectsUnknownHookProfile(t *testing.T) {
	cfg := Default()
	cfg.Hooks.PreCommit.Profile = "missing"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unknown hook profile error")
	}
}

func TestConfigRejectsUnknownSelectedProfile(t *testing.T) {
	cfg := Default()
	cfg.SelectedProfile = "missing"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unknown selected profile error")
	}
}

func TestThresholdUsesDomainOverrideAndGlobalFallback(t *testing.T) {
	cfg := Default()
	cfg.FailOn = finding.Critical
	cfg.Policy = map[finding.Domain]finding.Severity{
		finding.Quality: finding.Medium,
	}

	if got := cfg.Threshold(finding.Quality); got != finding.Medium {
		t.Fatalf("expected quality threshold %q, got %q", finding.Medium, got)
	}
	if got := cfg.Threshold(finding.Security); got != finding.Critical {
		t.Fatalf("expected global fallback %q, got %q", finding.Critical, got)
	}
}

func TestConfigRejectsInvalidPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy map[finding.Domain]finding.Severity
	}{
		{name: "domain", policy: map[finding.Domain]finding.Severity{"performance": finding.High}},
		{name: "threshold", policy: map[finding.Domain]finding.Severity{finding.Security: "BLOCKER"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := Default()
			cfg.Policy = test.policy
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected policy validation error")
			}
		})
	}
}

func TestConfigRejectsDuplicateScannerInProfile(t *testing.T) {
	cfg := Default()
	cfg.Profiles[ProfileFast] = []string{"pattern", "pattern"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate scanner validation error")
	}
}

func TestConfigValidatesScannerTimeout(t *testing.T) {
	tests := []string{"not-a-duration", "0s", "-1s"}
	for _, timeout := range tests {
		t.Run(timeout, func(t *testing.T) {
			cfg := Default()
			cfg.Scanners = map[string]Scanner{
				"pattern": {Enabled: true, Timeout: timeout},
			}
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected timeout validation error")
			}
		})
	}
}

func TestScannerTimeoutDuration(t *testing.T) {
	duration, err := (Scanner{Timeout: "250ms"}).TimeoutDuration()
	if err != nil {
		t.Fatal(err)
	}
	if duration != 250*time.Millisecond {
		t.Fatalf("expected 250ms, got %s", duration)
	}
}

func TestConfigValidatesCommandScanner(t *testing.T) {
	cfg := Default()
	cfg.Scanners = map[string]Scanner{
		"quality-tool": {
			Enabled: true, Type: "command", Domain: finding.Quality,
			Command: []string{"quality-tool", "check"}, Workspace: "staged", OnMissing: "skip",
			FindingExitCodes: []int{1}, Severity: finding.High,
			Category: "lint", Description: "Quality tool reported findings",
			OutputFormat: "json-lines", Environment: []string{"CI"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigRejectsInvalidCommandScanner(t *testing.T) {
	cfg := Default()
	cfg.Scanners = map[string]Scanner{
		"invalid": {Enabled: true, Type: "command", Domain: finding.Quality},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid command scanner error")
	}
}

func TestLoadLegacyConfigWithoutProfilesOrPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "security-review.json")
	data := []byte(`{
		"version": 1,
		"project": "legacy",
		"root": ".",
		"mode": "full",
		"output": "report.json",
		"fail_on": "HIGH",
		"include_extensions": [".go"],
		"exclude_directories": [".git"],
		"exclude_files": [],
		"rule_files": [],
		"suppression_file": ".security-ignore",
		"workers": 1,
		"scanners": {}
	}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Threshold(finding.Security); got != finding.High {
		t.Fatalf("expected legacy fail_on threshold %q, got %q", finding.High, got)
	}
	if len(cfg.Profiles[ProfileFast]) != 1 {
		t.Fatalf("expected default profiles to survive legacy load: %v", cfg.Profiles)
	}
}

func filepathIsAbsolute(path string) bool {
	if len(path) >= 3 && path[1] == ':' {
		return true
	}
	return len(path) > 0 && path[0] == '/'
}
