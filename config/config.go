package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	pathpkg "path"
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
	RequiresNetwork  bool             `json:"requires_network,omitempty"`
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

type CachePolicy struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Directory string `json:"directory,omitempty"`
	MaxAge    string `json:"max_age,omitempty"`
	MaxBytes  int64  `json:"max_bytes,omitempty"`
}

func (c CachePolicy) MaxAgeDuration() (time.Duration, error) {
	if c.MaxAge == "" {
		return 0, nil
	}
	duration, err := time.ParseDuration(c.MaxAge)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("cache max_age must be a positive duration")
	}
	return duration, nil
}

type GovernancePolicy struct {
	RequiredFiles           []string                 `json:"required_files,omitempty"`
	RequiredHeaders         []RequiredHeader         `json:"required_headers,omitempty"`
	OwnershipFile           string                   `json:"ownership_file,omitempty"`
	OwnershipRules          []OwnershipRule          `json:"ownership_rules,omitempty"`
	SuppressionRequirements []SuppressionRequirement `json:"suppression_requirements,omitempty"`
}

type SuppressionRequirement struct {
	RuleIDs         []string `json:"rule_ids"`
	RequireTicket   bool     `json:"require_ticket,omitempty"`
	RequireApprover bool     `json:"require_approver,omitempty"`
}

type OwnershipRule struct {
	Path     string           `json:"path"`
	Owners   []string         `json:"owners"`
	Severity finding.Severity `json:"severity,omitempty"`
}

type RequiredHeader struct {
	ID             string           `json:"id"`
	Paths          []string         `json:"paths"`
	Pattern        string           `json:"pattern"`
	MaxLines       int              `json:"max_lines,omitempty"`
	Severity       finding.Severity `json:"severity,omitempty"`
	Description    string           `json:"description,omitempty"`
	Recommendation string           `json:"recommendation,omitempty"`
}

type ArchitecturePolicy struct {
	Layers                []ArchitectureLayer   `json:"layers,omitempty"`
	ForbiddenDependencies []ForbiddenDependency `json:"forbidden_dependencies,omitempty"`
	DetectCycles          bool                  `json:"detect_cycles,omitempty"`
}

type ArchitectureLayer struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

type ForbiddenDependency struct {
	From string `json:"from"`
	To   string `json:"to"`
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
		RequiresNetwork: s.RequiresNetwork,
	}
}

const (
	SchemaVersion   = 1
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
	OfflineProfiles      []string                            `json:"offline_profiles,omitempty"`
	Policy               map[finding.Domain]finding.Severity `json:"policy,omitempty"`
	Hooks                Hooks                               `json:"hooks,omitempty"`
	SupplyChain          SupplyChainPolicy                   `json:"supply_chain,omitempty"`
	Governance           GovernancePolicy                    `json:"governance,omitempty"`
	Architecture         ArchitecturePolicy                  `json:"architecture,omitempty"`
	Cache                CachePolicy                         `json:"cache,omitempty"`
	SelectedProfile      string                              `json:"-"`
}

func Default() Config {
	return Config{
		Version:             SchemaVersion,
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
		Cache:               CachePolicy{Directory: ".go-code-scanner-cache", MaxAge: "168h", MaxBytes: 256 * 1024 * 1024},
		Profiles: map[string][]string{
			ProfileFast:     {"pattern"},
			ProfileStandard: {"pattern", "govulncheck"},
			ProfileFull:     {"pattern", "govulncheck"},
		},
		OfflineProfiles: []string{ProfileFast},
		Hooks: Hooks{
			PreCommit: Hook{Enabled: true, Profile: ProfileFast, StagedOnly: true, NewOnly: true},
			CommitMsg: Hook{MaxSubjectLength: 72},
			PrePush:   Hook{Profile: ProfileStandard},
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
	if c.Version != SchemaVersion {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	if c.Project == "" {
		return fmt.Errorf("project is required")
	}
	if _, err := ResolveProjectPath(c.Root, c.Output); err != nil {
		return fmt.Errorf("output: %w", err)
	}
	for index, path := range c.RuleFiles {
		if _, err := ResolveProjectPath(c.Root, path); err != nil {
			return fmt.Errorf("rule_files[%d]: %w", index, err)
		}
	}
	if _, err := ResolveProjectPath(c.Root, c.SuppressionFile); err != nil {
		return fmt.Errorf("suppression_file: %w", err)
	}
	if _, err := ResolveProjectPath(c.Root, c.BaselineFile); err != nil {
		return fmt.Errorf("baseline_file: %w", err)
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
	if c.Cache.Enabled {
		cacheDirectory, err := ResolveProjectPath(c.Root, c.Cache.Directory)
		if err != nil {
			return fmt.Errorf("cache.directory: %w", err)
		}
		if info, statErr := os.Lstat(cacheDirectory); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("cache.directory: symlink directories are not allowed")
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return fmt.Errorf("cache.directory: %w", statErr)
		}
		if _, err := c.Cache.MaxAgeDuration(); err != nil {
			return err
		}
		if c.Cache.MaxBytes < 1 {
			return fmt.Errorf("cache.max_bytes must be at least 1 when cache is enabled")
		}
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
	seenHeaders := make(map[string]struct{}, len(c.Governance.RequiredHeaders))
	for index, header := range c.Governance.RequiredHeaders {
		if strings.TrimSpace(header.ID) == "" {
			return fmt.Errorf("governance.required_headers[%d]: id is required", index)
		}
		if _, ok := seenHeaders[header.ID]; ok {
			return fmt.Errorf("governance.required_headers[%d]: duplicate id %q", index, header.ID)
		}
		seenHeaders[header.ID] = struct{}{}
		if len(header.Paths) == 0 {
			return fmt.Errorf("governance.required_headers[%d]: paths are required", index)
		}
		for _, pattern := range header.Paths {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("governance.required_headers[%d]: empty path pattern", index)
			}
			if _, err := pathpkg.Match(pattern, "fixture"); err != nil {
				return fmt.Errorf("governance.required_headers[%d]: invalid path pattern %q: %w", index, pattern, err)
			}
		}
		if _, err := regexp.Compile(header.Pattern); err != nil || header.Pattern == "" {
			return fmt.Errorf("governance.required_headers[%d]: invalid pattern %q", index, header.Pattern)
		}
		if header.MaxLines < 0 {
			return fmt.Errorf("governance.required_headers[%d]: max_lines cannot be negative", index)
		}
		if header.Severity != "" && !header.Severity.Valid() {
			return fmt.Errorf("governance.required_headers[%d]: invalid severity %q", index, header.Severity)
		}
	}
	ownershipFile := c.Governance.OwnershipFile
	if ownershipFile == "" {
		ownershipFile = "CODEOWNERS"
	}
	cleanOwnershipFile := filepath.ToSlash(filepath.Clean(ownershipFile))
	if filepath.IsAbs(ownershipFile) || cleanOwnershipFile == ".." || strings.HasPrefix(cleanOwnershipFile, "../") {
		return fmt.Errorf("governance.ownership_file contains unsafe path %q", ownershipFile)
	}
	seenOwnershipPaths := make(map[string]struct{}, len(c.Governance.OwnershipRules))
	for index, rule := range c.Governance.OwnershipRules {
		if strings.TrimSpace(rule.Path) == "" || len(rule.Owners) == 0 {
			return fmt.Errorf("governance.ownership_rules[%d]: path and owners are required", index)
		}
		if strings.ContainsAny(rule.Path, "\r\n") {
			return fmt.Errorf("governance.ownership_rules[%d]: path must be one line", index)
		}
		if _, duplicate := seenOwnershipPaths[rule.Path]; duplicate {
			return fmt.Errorf("governance.ownership_rules[%d]: duplicate path %q", index, rule.Path)
		}
		seenOwnershipPaths[rule.Path] = struct{}{}
		seenOwners := make(map[string]struct{}, len(rule.Owners))
		for ownerIndex, owner := range rule.Owners {
			owner = strings.TrimSpace(owner)
			if owner == "" || strings.ContainsAny(owner, " \t\r\n") {
				return fmt.Errorf("governance.ownership_rules[%d].owners[%d]: invalid owner %q", index, ownerIndex, owner)
			}
			if _, duplicate := seenOwners[owner]; duplicate {
				return fmt.Errorf("governance.ownership_rules[%d]: duplicate owner %q", index, owner)
			}
			seenOwners[owner] = struct{}{}
		}
		if rule.Severity != "" && !rule.Severity.Valid() {
			return fmt.Errorf("governance.ownership_rules[%d]: invalid severity %q", index, rule.Severity)
		}
	}
	for index, requirement := range c.Governance.SuppressionRequirements {
		if len(requirement.RuleIDs) == 0 {
			return fmt.Errorf("governance.suppression_requirements[%d]: rule_ids are required", index)
		}
		if !requirement.RequireTicket && !requirement.RequireApprover {
			return fmt.Errorf("governance.suppression_requirements[%d]: at least one requirement must be enabled", index)
		}
		seenPatterns := make(map[string]struct{}, len(requirement.RuleIDs))
		for patternIndex, pattern := range requirement.RuleIDs {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("governance.suppression_requirements[%d].rule_ids[%d]: pattern is required", index, patternIndex)
			}
			if _, err := pathpkg.Match(pattern, "rule-id"); err != nil {
				return fmt.Errorf("governance.suppression_requirements[%d].rule_ids[%d]: invalid pattern %q", index, patternIndex, pattern)
			}
			if _, duplicate := seenPatterns[pattern]; duplicate {
				return fmt.Errorf("governance.suppression_requirements[%d]: duplicate rule pattern %q", index, pattern)
			}
			seenPatterns[pattern] = struct{}{}
		}
	}
	layers := make(map[string]struct{}, len(c.Architecture.Layers))
	for index, layer := range c.Architecture.Layers {
		if strings.TrimSpace(layer.Name) == "" || len(layer.Paths) == 0 {
			return fmt.Errorf("architecture.layers[%d]: name and paths are required", index)
		}
		if _, ok := layers[layer.Name]; ok {
			return fmt.Errorf("architecture.layers[%d]: duplicate name %q", index, layer.Name)
		}
		layers[layer.Name] = struct{}{}
		for _, pattern := range layer.Paths {
			if _, err := pathpkg.Match(pattern, "fixture"); err != nil || strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("architecture.layers[%d]: invalid path pattern %q", index, pattern)
			}
		}
	}
	seenBoundaries := make(map[string]struct{}, len(c.Architecture.ForbiddenDependencies))
	for index, boundary := range c.Architecture.ForbiddenDependencies {
		if _, ok := layers[boundary.From]; !ok {
			return fmt.Errorf("architecture.forbidden_dependencies[%d]: unknown from layer %q", index, boundary.From)
		}
		if _, ok := layers[boundary.To]; !ok {
			return fmt.Errorf("architecture.forbidden_dependencies[%d]: unknown to layer %q", index, boundary.To)
		}
		key := boundary.From + "\x00" + boundary.To
		if _, ok := seenBoundaries[key]; ok {
			return fmt.Errorf("architecture.forbidden_dependencies[%d]: duplicate boundary %s -> %s", index, boundary.From, boundary.To)
		}
		seenBoundaries[key] = struct{}{}
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
	seenOfflineProfiles := make(map[string]struct{}, len(c.OfflineProfiles))
	for _, name := range c.OfflineProfiles {
		if _, exists := c.Profiles[name]; !exists {
			return fmt.Errorf("offline profile %q is not defined", name)
		}
		if _, duplicate := seenOfflineProfiles[name]; duplicate {
			return fmt.Errorf("duplicate offline profile %q", name)
		}
		seenOfflineProfiles[name] = struct{}{}
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
