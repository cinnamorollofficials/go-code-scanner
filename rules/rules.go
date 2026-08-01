package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type Rule struct {
	ID             string           `json:"id"`
	Pattern        string           `json:"pattern"`
	Severity       finding.Severity `json:"severity"`
	Domain         finding.Domain   `json:"domain,omitempty"`
	Category       string           `json:"category"`
	Description    string           `json:"description"`
	Recommendation string           `json:"recommendation,omitempty"`
	Extensions     []string         `json:"extensions,omitempty"`
	Enabled        *bool            `json:"enabled,omitempty"`
}

type Set struct {
	Version int    `json:"version"`
	Rules   []Rule `json:"rules"`
}

type Compiled struct {
	Rule
	Regex *regexp.Regexp
}

func Load(paths []string) ([]Compiled, error) {
	all := Default()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read rules %s: %w", path, err)
		}
		var set Set
		if err := json.Unmarshal(data, &set); err != nil {
			return nil, fmt.Errorf("decode rules %s: %w", path, err)
		}
		if set.Version != 1 {
			return nil, fmt.Errorf("rules %s: unsupported version %d", path, set.Version)
		}
		all = append(all, set.Rules...)
	}
	return Compile(all)
}

func Compile(input []Rule) ([]Compiled, error) {
	seen := make(map[string]struct{}, len(input))
	compiled := make([]Compiled, 0, len(input))
	for _, rule := range input {
		if rule.Enabled != nil && !*rule.Enabled {
			continue
		}
		if rule.Domain == "" {
			rule.Domain = finding.Security
		}
		if rule.ID == "" || rule.Category == "" || rule.Description == "" {
			return nil, fmt.Errorf("rule id, category, and description are required")
		}
		if !rule.Domain.Valid() {
			return nil, fmt.Errorf("rule %s: invalid domain %q", rule.ID, rule.Domain)
		}
		if _, ok := seen[rule.ID]; ok {
			return nil, fmt.Errorf("duplicate rule id %q", rule.ID)
		}
		seen[rule.ID] = struct{}{}
		if !rule.Severity.Valid() {
			return nil, fmt.Errorf("rule %s: invalid severity %q", rule.ID, rule.Severity)
		}
		re, err := regexp.Compile("(?i)" + rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", rule.ID, err)
		}
		compiled = append(compiled, Compiled{Rule: rule, Regex: re})
	}
	return compiled, nil
}

func MatchesExtension(rule Compiled, extension string) bool {
	if len(rule.Extensions) == 0 {
		return true
	}
	for _, allowed := range rule.Extensions {
		if strings.EqualFold(allowed, extension) {
			return true
		}
	}
	return false
}
