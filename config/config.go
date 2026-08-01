package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type Mode string

const (
	ModeFull    Mode = "full"
	ModeChanged Mode = "changed"
	ModeStaged  Mode = "staged"
)

type Scanner struct {
	Enabled  bool           `json:"enabled"`
	Required bool           `json:"required"`
	Timeout  string         `json:"timeout,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
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
	Workers            int                                 `json:"workers"`
	Scanners           map[string]Scanner                  `json:"scanners"`
	Profiles           map[string][]string                 `json:"profiles,omitempty"`
	Policy             map[finding.Domain]finding.Severity `json:"policy,omitempty"`
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
		Workers:            runtime.GOMAXPROCS(0),
		Profiles: map[string][]string{
			ProfileFast:     {"pattern"},
			ProfileStandard: {"pattern"},
			ProfileFull:     {"pattern"},
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
	for domain, threshold := range c.Policy {
		if !domain.Valid() {
			return fmt.Errorf("invalid policy domain %q", domain)
		}
		if !threshold.Valid() {
			return fmt.Errorf("invalid policy threshold %q for domain %s", threshold, domain)
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
	absRoot, err := filepath.Abs(c.Root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	c.Root = filepath.Clean(absRoot)
	return nil
}
