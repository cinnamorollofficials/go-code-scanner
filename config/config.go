package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner/adapters"
	commandscanner "github.com/cinnamorollofficials/go-code-scanner/scanner/command"
)

type Mode string

const (
	ModeFull    Mode = "full"
	ModeChanged Mode = "changed"
	ModeStaged  Mode = "staged"
)

type Scanner struct {
	Enabled          bool             `json:"enabled"`
	Required         bool             `json:"required"`
	Timeout          string           `json:"timeout,omitempty"`
	Type             string           `json:"type,omitempty"`
	Adapter          string           `json:"adapter,omitempty"`
	Args             []string         `json:"args,omitempty"`
	Domain           finding.Domain   `json:"domain,omitempty"`
	Command          []string         `json:"command,omitempty"`
	Workspace        string           `json:"workspace,omitempty"`
	OnMissing        string           `json:"on_missing,omitempty"`
	FindingExitCodes []int            `json:"finding_exit_codes,omitempty"`
	Severity         finding.Severity `json:"severity,omitempty"`
	Category         string           `json:"category,omitempty"`
	Description      string           `json:"description,omitempty"`
	Version          string           `json:"version,omitempty"`
	MaxOutputBytes   int              `json:"max_output_bytes,omitempty"`
	SnapshotMaxFiles int64            `json:"snapshot_max_files,omitempty"`
	SnapshotMaxBytes int64            `json:"snapshot_max_bytes,omitempty"`
	OutputFormat     string           `json:"output_format,omitempty"`
	Environment      []string         `json:"environment,omitempty"`
	Options          map[string]any   `json:"options,omitempty"`
}

type Hook struct {
	Enabled          bool   `json:"enabled"`
	Profile          string `json:"profile,omitempty"`
	StagedOnly       bool   `json:"staged_only,omitempty"`
	NewOnly          bool   `json:"new_only,omitempty"`
	MessagePattern   string `json:"message_pattern,omitempty"`
	MaxSubjectLength int    `json:"max_subject_length,omitempty"`
}

type Hooks struct {
	PreCommit Hook `json:"pre_commit"`
	CommitMsg Hook `json:"commit_msg"`
	PrePush   Hook `json:"pre_push"`
}

type SupplyChainPolicy struct {
	DependencyAllowlist []string `json:"dependency_allowlist,omitempty"`
	DependencyDenylist  []string `json:"dependency_denylist,omitempty"`
	LicenseAllowlist    []string `json:"license_allowlist,omitempty"`
	LicenseDenylist     []string `json:"license_denylist,omitempty"`
}

type GovernancePolicy struct {
	RequiredFiles []string `json:"required_files,omitempty"`
}

func (s Scanner) TimeoutDuration() (time.Duration, error) {
	if s.Timeout == "" {
		return 0, nil
	}
	duration, err := time.ParseDuration(s.Timeout)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", s.Timeout, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("timeout must be greater than zero")
	}
	return duration, nil
}

func (s Scanner) CommandSpec(id string) commandscanner.Spec {
	return commandscanner.Spec{
		ID: id, Domain: s.Domain, Command: s.Command, Workspace: s.Workspace,
		OnMissing: s.OnMissing, FindingExitCodes: s.FindingExitCodes,
		Severity: s.Severity, Category: s.Category, Description: s.Description,
		Version: s.Version, MaxOutputBytes: s.MaxOutputBytes,
		SnapshotMaxFiles: s.SnapshotMaxFiles, SnapshotMaxBytes: s.SnapshotMaxBytes,
		OutputFormat: s.OutputFormat, Environment: s.Environment,
	}
}

const (
	ProfileFast     = "fast"
	ProfileStandard = "standard"
	ProfileFull     = "full"
)

func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("decode config: multiple JSON documents are not allowed")
		}
		return Config{}, fmt.Errorf("decode config: trailing data: %w", err)
	}
	if cfg.Root == "." || cfg.Root == "" {
		cfg.Root = filepath.Dir(path)
	} else if !filepath.IsAbs(cfg.Root) {
		cfg.Root = filepath.Join(filepath.Dir(path), cfg.Root)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

type Config struct {
	Version              int                                 `json:"version"`
	Project              string                              `json:"project"`
	Root                 string                              `json:"root"`
	Mode                 Mode                                `json:"mode"`
	Output               string                              `json:"output"`
	FailOn               finding.Severity                    `json:"fail_on"`
	IncludeExtensions    []string                            `json:"include_extensions"`
	ExcludeDirectories   []string                            `json:"exclude_directories"`
	ExcludeFiles         []string                            `json:"exclude_files"`
	RuleFiles            []string                            `json:"rule_files"`
	SuppressionFile      string                              `json:"suppression_file"`
	BaselineFile         string                              `json:"baseline_file,omitempty"`
	Workers              int                                 `json:"workers"`
	PatternMaxFileBytes  int64                               `json:"pattern_max_file_bytes"`
	PatternMaxLineBytes  int                                 `json:"pattern_max_line_bytes"`
	QualityMaxFileBytes  int64                               `json:"quality_max_file_bytes,omitempty"`
	QualityMaxLineLength int                                 `json:"quality_max_line_length,omitempty"`
	Scanners             map[string]Scanner                  `json:"scanners"`
	Profiles             map[string][]string                 `json:"profiles,omitempty"`
	Policy               map[finding.Domain]finding.Severity `json:"policy,omitempty"`
	Hooks                Hooks                               `json:"hooks,omitempty"`
	SupplyChain          SupplyChainPolicy                   `json:"supply_chain,omitempty"`
	Governance           GovernancePolicy                    `json:"governance,omitempty"`
	SelectedProfile      string                              `json:"-"`
}

func Default() Config {
	return Config{
		Version:             1,
		Project:             "security-review",
		Root:                ".",
		Mode:                ModeFull,
		Output:              "security_findings.json",
		FailOn:              finding.Critical,
		IncludeExtensions:   []string{".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"},
		ExcludeDirectories:  []string{".git", "node_modules", "vendor", "dist", "build", ".next", "out", "bin"},
		ExcludeFiles:        []string{"security_findings.json", "package-lock.json"},
		SuppressionFile:     ".security-ignore",
		BaselineFile:        ".security-baseline.json",
		Workers:             runtime.GOMAXPROCS(0),
		PatternMaxFileBytes: 2 * 1024 * 1024,
		PatternMaxLineBytes: 1024 * 1024,
		Profiles: map[string][]string{
			ProfileFast:     {"pattern"},
			ProfileStandard: {"pattern"},
			ProfileFull:     {"pattern"},
		},
		Hooks: Hooks{
			PreCommit: Hook{Enabled: true, Profile: ProfileFast, StagedOnly: true, NewOnly: true},
			CommitMsg: Hook{MaxSubjectLength: 72},
		},
	}
}

// Threshold returns the domain-specific policy threshold when configured and
// otherwise falls back to the legacy global fail_on value.
func (c Config) Threshold(domain finding.Domain) finding.Severity {
	if threshold, ok := c.Policy[domain]; ok {
		return threshold
	}
	return c.FailOn
}

func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	if c.Project == "" {
		return fmt.Errorf("project is required")
	}
	if c.Mode != ModeFull && c.Mode != ModeChanged && c.Mode != ModeStaged {
		return fmt.Errorf("invalid scan mode %q", c.Mode)
	}
	if !c.FailOn.Valid() {
		return fmt.Errorf("invalid fail_on severity %q", c.FailOn)
	}
	if c.Workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}
	if c.PatternMaxFileBytes < 1 || c.PatternMaxLineBytes < 1 {
		return fmt.Errorf("pattern file and line byte limits must be at least 1")
	}
	if c.QualityMaxFileBytes < 0 || c.QualityMaxLineLength < 0 {
		return fmt.Errorf("quality file and line limits cannot be negative")
	}
	for name, patterns := range map[string][]string{
		"dependency_allowlist": c.SupplyChain.DependencyAllowlist,
		"dependency_denylist":  c.SupplyChain.DependencyDenylist,
		"license_allowlist":    c.SupplyChain.LicenseAllowlist,
		"license_denylist":     c.SupplyChain.LicenseDenylist,
	} {
		seen := make(map[string]struct{}, len(patterns))
		for _, pattern := range patterns {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("supply_chain.%s contains an empty pattern", name)
			}
			if _, err := filepath.Match(pattern, "fixture"); err != nil {
				return fmt.Errorf("supply_chain.%s has invalid pattern %q: %w", name, pattern, err)
			}
			key := strings.ToLower(pattern)
			if _, ok := seen[key]; ok {
				return fmt.Errorf("supply_chain.%s contains duplicate pattern %q", name, pattern)
			}
			seen[key] = struct{}{}
		}
	}
	seenRequiredFiles := make(map[string]struct{}, len(c.Governance.RequiredFiles))
	for _, required := range c.Governance.RequiredFiles {
		required = filepath.ToSlash(filepath.Clean(required))
		if required == "." || filepath.IsAbs(required) || required == ".." || strings.HasPrefix(required, "../") {
			return fmt.Errorf("governance.required_files contains unsafe path %q", required)
		}
		key := strings.ToLower(required)
		if _, ok := seenRequiredFiles[key]; ok {
			return fmt.Errorf("governance.required_files contains duplicate path %q", required)
		}
		seenRequiredFiles[key] = struct{}{}
	}
	for domain, threshold := range c.Policy {
		if !domain.Valid() {
			return fmt.Errorf("invalid policy domain %q", domain)
		}
		if !threshold.Valid() {
			return fmt.Errorf("invalid policy threshold %q for domain %s", threshold, domain)
		}
	}
	for id, configured := range c.Scanners {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("scanner id is required")
		}
		if _, err := configured.TimeoutDuration(); err != nil {
			return fmt.Errorf("scanner %s: %w", id, err)
		}
		switch configured.Type {
		case "", "pattern":
			if configured.Type == "pattern" && id != "pattern" {
				return fmt.Errorf("scanner %s: pattern type is reserved for scanner %q", id, "pattern")
			}
		case "command":
			if id == "pattern" {
				return fmt.Errorf("scanner %s: command cannot replace built-in pattern scanner", id)
			}
			if _, err := commandscanner.New(configured.CommandSpec(id)); err != nil {
				return err
			}
		case "adapter":
			if id == "pattern" {
				return fmt.Errorf("scanner %s: adapter cannot replace built-in pattern scanner", id)
			}
			if _, err := adapters.New(id, configured.Adapter, configured.AdapterOptions()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("scanner %s: unsupported type %q", id, configured.Type)
		}
	}
	for name, scanners := range c.Profiles {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("profile name is required")
		}
		seen := make(map[string]struct{}, len(scanners))
		for _, scannerID := range scanners {
			scannerID = strings.TrimSpace(scannerID)
			if scannerID == "" {
				return fmt.Errorf("profile %s: scanner id is required", name)
			}
			if _, ok := seen[scannerID]; ok {
				return fmt.Errorf("profile %s: duplicate scanner %q", name, scannerID)
			}
			seen[scannerID] = struct{}{}
		}
	}
	if err := c.validateHook("pre_commit", c.Hooks.PreCommit); err != nil {
		return err
	}
	if err := c.validateHook("commit_msg", c.Hooks.CommitMsg); err != nil {
		return err
	}
	if err := c.validateHook("pre_push", c.Hooks.PrePush); err != nil {
		return err
	}
	if c.SelectedProfile != "" {
		if _, ok := c.Profiles[c.SelectedProfile]; !ok {
			return fmt.Errorf("unknown selected profile %q", c.SelectedProfile)
		}
	}
	absRoot, err := filepath.Abs(c.Root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	c.Root = filepath.Clean(absRoot)
	return nil
}

func (s Scanner) AdapterOptions() adapters.Options {
	return adapters.Options{
		Args: s.Args, Workspace: s.Workspace, OnMissing: s.OnMissing, Environment: s.Environment,
		MaxOutputBytes: s.MaxOutputBytes, SnapshotMaxFiles: s.SnapshotMaxFiles, SnapshotMaxBytes: s.SnapshotMaxBytes,
	}
}

func (c Config) validateHook(name string, hook Hook) error {
	if !hook.Enabled {
		return nil
	}
	if name == "commit_msg" {
		if hook.MaxSubjectLength < 0 {
			return fmt.Errorf("hook %s: max_subject_length cannot be negative", name)
		}
		if hook.MessagePattern != "" {
			if _, err := regexp.Compile(hook.MessagePattern); err != nil {
				return fmt.Errorf("hook %s: invalid message_pattern: %w", name, err)
			}
		}
		return nil
	}
	if strings.TrimSpace(hook.Profile) == "" {
		return fmt.Errorf("hook %s: profile is required", name)
	}
	if _, ok := c.Profiles[hook.Profile]; !ok {
		return fmt.Errorf("hook %s: unknown profile %q", name, hook.Profile)
	}
	return nil
}
