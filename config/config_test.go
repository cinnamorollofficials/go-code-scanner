package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if scanners := cfg.Profiles[ProfileFast]; len(scanners) != 1 || scanners[0] != "pattern" {
		t.Fatalf("unexpected fast profile: %v", scanners)
	}
	for _, name := range []string{ProfileStandard, ProfileFull} {
		scanners := cfg.Profiles[name]
		if len(scanners) != 2 || scanners[0] != "pattern" || scanners[1] != "govulncheck" {
			t.Fatalf("unexpected %s profile: %v", name, scanners)
		}
	}
}

func TestDefaultHookProfilesKeepSupplyChainOutOfPreCommit(t *testing.T) {
	cfg := Default()
	if profileContainsForTest(cfg.Profiles[cfg.Hooks.PreCommit.Profile], "govulncheck") {
		t.Fatal("pre-commit profile must remain offline and fast")
	}
	if cfg.Hooks.PrePush.Profile != "" && !profileContainsForTest(cfg.Profiles[cfg.Hooks.PrePush.Profile], "govulncheck") {
		t.Fatal("configured pre-push profile must include supply-chain checks")
	}
}

func profileContainsForTest(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestDefaultPreCommitHookUsesFastStagedProfile(t *testing.T) {
	hook := Default().Hooks.PreCommit
	if !hook.Enabled || !hook.StagedOnly || !hook.NewOnly || hook.Profile != ProfileFast {
		t.Fatalf("unexpected default pre-commit hook: %+v", hook)
	}
}

func TestAdapterConfigurationValidatesKnownNames(t *testing.T) {
	cfg := Default()
	cfg.Root = t.TempDir()
	cfg.Scanners = map[string]Scanner{
		"format": {Enabled: true, Type: "adapter", Adapter: "gofmt", Workspace: "staged"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("known adapter was rejected: %v", err)
	}
	cfg.Scanners["format"] = Scanner{Enabled: true, Type: "adapter", Adapter: "unknown"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("unknown adapter was accepted")
	}
}

func TestSupplyChainPolicyRejectsInvalidOrDuplicatePatterns(t *testing.T) {
	for _, policy := range []SupplyChainPolicy{
		{DependencyAllowlist: []string{"[invalid"}},
		{LicenseDenylist: []string{"GPL-*", "gpl-*"}},
		{DependencyDenylist: []string{""}},
	} {
		cfg := Default()
		cfg.Root = t.TempDir()
		cfg.SupplyChain = policy
		if err := cfg.Validate(); err == nil {
			t.Fatalf("invalid supply-chain policy accepted: %+v", policy)
		}
	}
}

func TestGovernanceRequiredFilesRejectUnsafeAndDuplicatePaths(t *testing.T) {
	for _, files := range [][]string{{"../secret"}, {"/absolute"}, {"SECURITY.md", "security.md"}} {
		cfg := Default()
		cfg.Root = t.TempDir()
		cfg.Governance.RequiredFiles = files
		if err := cfg.Validate(); err == nil {
			t.Fatalf("invalid required files accepted: %v", files)
		}
	}
}

func TestGovernanceRequiredHeaderValidation(t *testing.T) {
	valid := Default()
	valid.Root = t.TempDir()
	valid.Governance.RequiredHeaders = []RequiredHeader{{
		ID: "license-header", Paths: []string{"**/*.go"}, Pattern: `^// Copyright`, MaxLines: 5, Severity: finding.High,
	}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid required header rejected: %v", err)
	}
	invalid := []RequiredHeader{
		{ID: "", Paths: []string{"*.go"}, Pattern: "header"},
		{ID: "bad-glob", Paths: []string{"[bad"}, Pattern: "header"},
		{ID: "bad-regex", Paths: []string{"*.go"}, Pattern: "[bad"},
		{ID: "bad-limit", Paths: []string{"*.go"}, Pattern: "header", MaxLines: -1},
	}
	for _, header := range invalid {
		cfg := Default()
		cfg.Root = t.TempDir()
		cfg.Governance.RequiredHeaders = []RequiredHeader{header}
		if err := cfg.Validate(); err == nil {
			t.Fatalf("invalid required header accepted: %+v", header)
		}
	}
}

func TestGovernanceOwnershipPolicyValidation(t *testing.T) {
	valid := Default()
	valid.Root = t.TempDir()
	valid.Governance.OwnershipFile = ".github/CODEOWNERS"
	valid.Governance.OwnershipRules = []OwnershipRule{{Path: "/internal/auth/**", Owners: []string{"@security", "@identity"}, Severity: finding.High}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid ownership policy rejected: %v", err)
	}

	invalid := []GovernancePolicy{
		{OwnershipFile: "../CODEOWNERS", OwnershipRules: valid.Governance.OwnershipRules},
		{OwnershipRules: []OwnershipRule{{Path: "", Owners: []string{"@security"}}}},
		{OwnershipRules: []OwnershipRule{{Path: "/internal/auth/**", Owners: nil}}},
		{OwnershipRules: []OwnershipRule{{Path: "/internal/auth/**", Owners: []string{"team with spaces"}}}},
		{OwnershipRules: []OwnershipRule{{Path: "/internal/auth/**", Owners: []string{"@security", "@security"}}}},
		{OwnershipRules: []OwnershipRule{{Path: "/internal/auth/**", Owners: []string{"@security"}, Severity: "URGENT"}}},
		{OwnershipRules: []OwnershipRule{{Path: "/same/**", Owners: []string{"@a"}}, {Path: "/same/**", Owners: []string{"@b"}}}},
	}
	for _, governance := range invalid {
		cfg := Default()
		cfg.Root = t.TempDir()
		cfg.Governance = governance
		if err := cfg.Validate(); err == nil {
			t.Fatalf("invalid ownership policy accepted: %+v", governance)
		}
	}
}

func TestGovernanceSuppressionRequirementValidation(t *testing.T) {
	valid := Default()
	valid.Root = t.TempDir()
	valid.Governance.SuppressionRequirements = []SuppressionRequirement{{RuleIDs: []string{"security/*"}, RequireTicket: true, RequireApprover: true}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid suppression requirement rejected: %v", err)
	}
	invalid := [][]SuppressionRequirement{
		{{RequireTicket: true}},
		{{RuleIDs: []string{"security/*"}}},
		{{RuleIDs: []string{"[invalid"}, RequireTicket: true}},
		{{RuleIDs: []string{"security/*", "security/*"}, RequireApprover: true}},
	}
	for _, requirements := range invalid {
		cfg := Default()
		cfg.Root = t.TempDir()
		cfg.Governance.SuppressionRequirements = requirements
		if err := cfg.Validate(); err == nil {
			t.Fatalf("invalid suppression requirements accepted: %+v", requirements)
		}
	}
}

func TestArchitecturePolicyValidation(t *testing.T) {
	cfg := Default()
	cfg.Root = t.TempDir()
	cfg.Architecture = ArchitecturePolicy{
		Layers:                []ArchitectureLayer{{Name: "domain", Paths: []string{"internal/domain/*.go"}}, {Name: "infra", Paths: []string{"internal/infra/*.go"}}},
		ForbiddenDependencies: []ForbiddenDependency{{From: "domain", To: "infra"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid architecture policy rejected: %v", err)
	}
	cfg.Architecture.ForbiddenDependencies[0].To = "missing"
	if err := cfg.Validate(); err == nil {
		t.Fatal("unknown architecture layer accepted")
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

func TestLoadRejectsUnknownField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "security-review.json")
	if err := os.WriteFile(path, []byte(`{"project":"fixture","fail_onn":"HIGH"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "fail_onn") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestLoadRejectsMultipleJSONDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "security-review.json")
	if err := os.WriteFile(path, []byte(`{"project":"first"} {"project":"second"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected multiple document error")
	}
}

func filepathIsAbsolute(path string) bool {
	if len(path) >= 3 && path[1] == ':' {
		return true
	}
	return len(path) > 0 && path[0] == '/'
}
