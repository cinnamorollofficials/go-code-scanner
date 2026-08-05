package suppression

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

const SchemaVersion = 1

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
	Version      int    `json:"version"`
	Suppressions []Rule `json:"suppressions"`
}

type Requirement struct {
	RuleIDs         []string
	RequireTicket   bool
	RequireApprover bool
}

func Load(path string) ([]Rule, error) {
	return LoadWithRequirements(path, nil)
}

// Add validates and appends a suppression using an atomic, least-privilege write.
// When dryRun is true, the resulting file is returned without modifying path.
func Add(path string, rule Rule, dryRun bool) (*File, error) {
	if err := validate(rule); err != nil {
		return nil, err
	}
	file := &File{Version: SchemaVersion}
	data, err := os.ReadFile(path)
	if err == nil {
		decoded, decodeErr := decodeFile(data)
		if decodeErr != nil {
			return nil, decodeErr
		}
		file = decoded
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	for _, existing := range file.Suppressions {
		if existing.RuleID == rule.RuleID && existing.Fingerprint == rule.Fingerprint && normalizePath(existing.File) == normalizePath(rule.File) && existing.Line == rule.Line {
			return nil, fmt.Errorf("matching suppression already exists")
		}
	}
	file.Suppressions = append(file.Suppressions, rule)
	if dryRun {
		return file, nil
	}
	encoded, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode suppressions: %w", err)
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create suppression directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".suppression-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary suppression file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return nil, err
	}
	if _, err := temporary.Write(append(encoded, '\n')); err != nil {
		_ = temporary.Close()
		return nil, err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return nil, err
	}
	if err := temporary.Close(); err != nil {
		return nil, err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return nil, fmt.Errorf("replace suppressions: %w", err)
	}
	return file, nil
}

func LoadWithRequirements(path string, requirements []Requirement) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	file, err := decodeFile(data)
	if err != nil {
		return nil, err
	}
	for index, rule := range file.Suppressions {
		if err := validate(rule); err != nil {
			return nil, fmt.Errorf("suppression %d: %w", index+1, err)
		}
		if err := validateRequirements(rule, requirements); err != nil {
			return nil, fmt.Errorf("suppression %d: %w", index+1, err)
		}
	}
	return file.Suppressions, nil
}

func decodeFile(data []byte) (*File, error) {
	var file File
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("decode suppressions: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode suppressions: multiple JSON documents are not allowed")
		}
		return nil, fmt.Errorf("decode suppressions: trailing data: %w", err)
	}
	if file.Version != SchemaVersion {
		return nil, fmt.Errorf("unsupported suppression version %d", file.Version)
	}
	return &file, nil
}

func validateRequirements(rule Rule, requirements []Requirement) error {
	for _, requirement := range requirements {
		matched := false
		for _, pattern := range requirement.RuleIDs {
			if ok, _ := pathpkg.Match(pattern, rule.RuleID); ok {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		if requirement.RequireTicket && strings.TrimSpace(rule.Ticket) == "" {
			return fmt.Errorf("ticket is required for rule %q", rule.RuleID)
		}
		if requirement.RequireApprover && strings.TrimSpace(rule.ApprovedBy) == "" {
			return fmt.Errorf("approved_by is required for rule %q", rule.RuleID)
		}
	}
	return nil
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
