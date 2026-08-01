package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
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
	OutputFormat     string           `json:"output_format,omitempty"`
	Environment      []string         `json:"environment,omitempty"`
	Options          map[string]any   `json:"options,omitempty"`
}

type Hook struct {
	Enabled    bool   `json:"enabled"`
	Profile    string `json:"profile,omitempty"`
	StagedOnly bool   `json:"staged_only,omitempty"`
	NewOnly    bool   `json:"new_only,omitempty"`
}

type Hooks struct {
	PreCommit Hook `json:"pre_commit"`
	CommitMsg Hook `json:"commit_msg"`
	PrePush   Hook `json:"pre_push"`
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
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
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
	Version            int                                 `json:"version"`
	Project            string                              `json:"project"`
	Root               string                              `json:"root"`
	Mode               Mode                                `json:"mode"`
	Output             string                              `json:"output"`
	FailOn             finding.Severity                    `json:"fail_on"`
	IncludeExtensions  []string                            `json:"include_extensions"`
	ExcludeDirectories []string                            `json:"exclude_directories"`
	ExcludeFiles       []string                            `json:"exclude_files"`
	RuleFiles          []string                            `json:"rule_files"`
	SuppressionFile    string                              `json:"suppression_file"`
	BaselineFile       string                              `json:"baseline_file,omitempty"`
	Workers            int                                 `json:"workers"`
	Scanners           map[string]Scanner                  `json:"scanners"`
	Profiles           map[string][]string                 `json:"profiles,omitempty"`
	Policy             map[finding.Domain]finding.Severity `json:"policy,omitempty"`
	Hooks              Hooks                               `json:"hooks,omitempty"`
	SelectedProfile    string                              `json:"-"`
}

func Default() Config {
	return Config{
		Version:            1,
		Project:            "security-review",
		Root:               ".",
		Mode:               ModeFull,
		Output:             "security_findings.json",
		FailOn:             finding.Critical,
		IncludeExtensions:  []string{".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"},
		ExcludeDirectories: []string{".git", "node_modules", "vendor", "dist", "build", ".next", "out", "bin"},
		ExcludeFiles:       []string{"security_findings.json", "package-lock.json"},
		SuppressionFile:    ".security-ignore",
		BaselineFile:       ".security-baseline.json",
		Workers:            runtime.GOMAXPROCS(0),
		Profiles: map[string][]string{
			ProfileFast:     {"pattern"},
			ProfileStandard: {"pattern"},
			ProfileFull:     {"pattern"},
		},
		Hooks: Hooks{PreCommit: Hook{
			Enabled: true, Profile: ProfileFast, StagedOnly: true, NewOnly: true,
		}},
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

func (c Config) validateHook(name string, hook Hook) error {
	if !hook.Enabled {
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
