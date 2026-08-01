package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

const Version = 1

type Entry struct {
	Fingerprint string         `json:"fingerprint"`
	RuleID      string         `json:"rule_id"`
	Domain      finding.Domain `json:"domain"`
	File        string         `json:"file"`
}

type File struct {
	Version            int       `json:"version"`
	FingerprintVersion string    `json:"fingerprint_version"`
	GeneratedAt        time.Time `json:"generated_at"`
	Entries            []Entry   `json:"entries"`
}

type Comparison struct {
	New      []finding.Finding
	Existing []finding.Finding
	Resolved []Entry
}

func FromReport(report *finding.Report, now time.Time) (*File, error) {
	if report == nil {
		return nil, fmt.Errorf("report is required")
	}
	if report.FingerprintVersion == "" {
		return nil, fmt.Errorf("report fingerprint version is required")
	}
	entries := make([]Entry, 0, len(report.Findings))
	seen := make(map[string]struct{}, len(report.Findings))
	for _, item := range report.Findings {
		if item.Fingerprint == "" {
			return nil, fmt.Errorf("finding %s has no fingerprint", item.ID)
		}
		if _, ok := seen[item.Fingerprint]; ok {
			return nil, fmt.Errorf("duplicate finding fingerprint %q", item.Fingerprint)
		}
		seen[item.Fingerprint] = struct{}{}
		entries = append(entries, Entry{
			Fingerprint: item.Fingerprint, RuleID: item.RuleID,
			Domain: item.Domain, File: filepath.ToSlash(item.Location.File),
		})
	}
	sortEntries(entries)
	return &File{
		Version: Version, FingerprintVersion: report.FingerprintVersion,
		GeneratedAt: now.UTC(), Entries: entries,
	}, nil
}

func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode baseline: %w", err)
	}
	if err := file.Validate(); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *File) Validate() error {
	if f == nil {
		return fmt.Errorf("baseline is required")
	}
	if f.Version != Version {
		return fmt.Errorf("unsupported baseline version %d", f.Version)
	}
	if f.FingerprintVersion == "" {
		return fmt.Errorf("baseline fingerprint version is required")
	}
	seen := make(map[string]struct{}, len(f.Entries))
	for index, entry := range f.Entries {
		if entry.Fingerprint == "" {
			return fmt.Errorf("baseline entry %d: fingerprint is required", index+1)
		}
		if _, ok := seen[entry.Fingerprint]; ok {
			return fmt.Errorf("baseline entry %d: duplicate fingerprint %q", index+1, entry.Fingerprint)
		}
		seen[entry.Fingerprint] = struct{}{}
		if !entry.Domain.Valid() {
			return fmt.Errorf("baseline entry %d: invalid domain %q", index+1, entry.Domain)
		}
		if entry.RuleID == "" {
			return fmt.Errorf("baseline entry %d: rule_id is required", index+1)
		}
		if entry.File == "" {
			return fmt.Errorf("baseline entry %d: file is required", index+1)
		}
	}
	return nil
}

func Write(path string, file *File) error {
	if err := file.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("encode baseline: %w", err)
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create baseline directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".security-baseline-*.json")
	if err != nil {
		return fmt.Errorf("create temporary baseline: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("secure temporary baseline: %w", err)
	}
	if _, err := temporary.Write(append(data, '\n')); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary baseline: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary baseline: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary baseline: %w", err)
	}
	if err := replace(path, temporaryPath); err != nil {
		return err
	}
	return nil
}

func replace(path, temporaryPath string) error {
	backup := path + ".previous"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Rename(temporaryPath, path); err != nil {
			return fmt.Errorf("install baseline: %w", err)
		}
		return nil
	}
	_ = os.Remove(backup)
	if err := os.Rename(path, backup); err != nil {
		return fmt.Errorf("backup existing baseline: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if restoreErr := os.Rename(backup, path); restoreErr != nil {
			return fmt.Errorf("install baseline: %v; restore previous baseline: %w", err, restoreErr)
		}
		return fmt.Errorf("install baseline: %w", err)
	}
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("remove baseline backup: %w", err)
	}
	return nil
}

func Compare(report *finding.Report, file *File) (Comparison, error) {
	if report == nil {
		return Comparison{}, fmt.Errorf("report is required")
	}
	if err := file.Validate(); err != nil {
		return Comparison{}, err
	}
	if report.FingerprintVersion != file.FingerprintVersion {
		return Comparison{}, fmt.Errorf("fingerprint version mismatch: report=%s baseline=%s", report.FingerprintVersion, file.FingerprintVersion)
	}
	baselineEntries := make(map[string]Entry, len(file.Entries))
	for _, entry := range file.Entries {
		baselineEntries[entry.Fingerprint] = entry
	}
	current := make(map[string]struct{}, len(report.Findings))
	comparison := Comparison{}
	report.Summary.New = 0
	report.Summary.Existing = 0
	report.Summary.Resolved = 0
	for index := range report.Findings {
		item := &report.Findings[index]
		current[item.Fingerprint] = struct{}{}
		if _, ok := baselineEntries[item.Fingerprint]; ok {
			item.BaselineState = finding.BaselineExisting
			report.Summary.Existing++
			comparison.Existing = append(comparison.Existing, *item)
		} else {
			item.BaselineState = finding.BaselineNew
			report.Summary.New++
			comparison.New = append(comparison.New, *item)
		}
	}
	for _, entry := range file.Entries {
		if _, ok := current[entry.Fingerprint]; !ok {
			comparison.Resolved = append(comparison.Resolved, entry)
		}
	}
	sortEntries(comparison.Resolved)
	report.Summary.Resolved = len(comparison.Resolved)
	return comparison, nil
}

func sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].File != entries[j].File {
			return entries[i].File < entries[j].File
		}
		if entries[i].RuleID != entries[j].RuleID {
			return entries[i].RuleID < entries[j].RuleID
		}
		return entries[i].Fingerprint < entries[j].Fingerprint
	})
}
