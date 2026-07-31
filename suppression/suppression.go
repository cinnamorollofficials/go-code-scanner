package suppression

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type Rule struct {
	RuleID      string `json:"rule_id,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Reason      string `json:"reason"`
	ApprovedBy  string `json:"approved_by,omitempty"`
	Expires     string `json:"expires"`
	Ticket      string `json:"ticket,omitempty"`
}

type File struct {
	Version      any    `json:"version"`
	Suppressions []Rule `json:"suppressions"`
}

func Load(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode suppressions: %w", err)
	}
	for index, rule := range file.Suppressions {
		if err := validate(rule); err != nil {
			return nil, fmt.Errorf("suppression %d: %w", index+1, err)
		}
	}
	return file.Suppressions, nil
}

func Apply(findings []finding.Finding, rules []Rule, now time.Time) (active, suppressed []finding.Finding, stale []string) {
	for _, item := range findings {
		matched := false
		for _, rule := range rules {
			if !matches(item, rule) {
				continue
			}
			if expired(rule, now) {
				stale = appendUnique(stale, item.Location.File)
				continue
			}
			item.Suppressed = true
			item.SuppressionReason = rule.Reason
			suppressed = append(suppressed, item)
			matched = true
			break
		}
		if !matched {
			active = append(active, item)
		}
	}
	return active, suppressed, stale
}

func matches(item finding.Finding, rule Rule) bool {
	want := normalizePath(rule.File)
	got := normalizePath(item.Location.File)
	fileMatches := got == want
	ruleMatches := rule.RuleID == "" || rule.RuleID == item.RuleID
	fingerprintMatches := rule.Fingerprint == "" || rule.Fingerprint == item.Fingerprint
	lineMatches := rule.Line == -1 || rule.Line == item.Location.Line
	return fileMatches && ruleMatches && fingerprintMatches && lineMatches
}

func expired(rule Rule, now time.Time) bool {
	if rule.Expires == "" {
		return false
	}
	expires, err := time.Parse("2006-01-02", rule.Expires)
	return err == nil && now.After(expires.Add(24*time.Hour-time.Nanosecond))
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func validate(rule Rule) error {
	if strings.TrimSpace(rule.File) == "" {
		return fmt.Errorf("file is required")
	}
	if strings.TrimSpace(rule.Reason) == "" {
		return fmt.Errorf("reason is required")
	}
	if strings.TrimSpace(rule.Expires) == "" {
		return fmt.Errorf("expires is required")
	}
	if _, err := time.Parse("2006-01-02", rule.Expires); err != nil {
		return fmt.Errorf("invalid expires %q: %w", rule.Expires, err)
	}
	return nil
}

func normalizePath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(path)), "./")
}
