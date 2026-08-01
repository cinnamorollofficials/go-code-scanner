package securityreview

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/discovery"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
	patternscanner "github.com/cinnamorollofficials/go-code-scanner/scanner/pattern"
	"github.com/cinnamorollofficials/go-code-scanner/suppression"
)

const SchemaVersion = "1.0"

type Reviewer interface {
	Run(context.Context) (*finding.Report, error)
}

type reviewer struct {
	config   config.Config
	scanners []registeredScanner
	now      func() time.Time
}

type registeredScanner struct {
	scanner  scanner.Scanner
	required bool
}

type Option func(*reviewer) error

func WithScanner(value scanner.Scanner) Option {
	return withScanner(value, false)
}

func WithRequiredScanner(value scanner.Scanner) Option {
	return withScanner(value, true)
}

func withScanner(value scanner.Scanner, required bool) Option {
	return func(r *reviewer) error {
		if value == nil {
			return fmt.Errorf("scanner cannot be nil")
		}
		r.scanners = append(r.scanners, registeredScanner{scanner: value, required: required})
		return nil
	}
}

func New(cfg config.Config, options ...Option) (Reviewer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	compiled, err := rules.Load(resolvePaths(cfg.Root, cfg.RuleFiles))
	if err != nil {
		return nil, err
	}
	r := &reviewer{
		config:   cfg,
		scanners: []registeredScanner{{scanner: patternscanner.New(compiled, cfg.Workers), required: true}},
		now:      time.Now,
	}
	for _, option := range options {
		if err := option(r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *reviewer) Run(ctx context.Context) (*finding.Report, error) {
	started := r.now().UTC()
	sources, err := discovery.Sources(ctx, r.config)
	if err != nil {
		return nil, err
	}
	request := scanner.Request{Root: r.config.Root, Mode: string(r.config.Mode), Sources: sources}
	var all []finding.Finding
	statuses := make([]finding.ScannerStatus, 0, len(r.scanners))
	var operationalErrors []error
	var warnings []string
	for _, registered := range r.scanners {
		source := registered.scanner
		required := registered.required
		configured, hasConfig := r.config.Scanners[source.ID()]
		if hasConfig {
			required = configured.Required
		}
		if hasConfig && !configured.Enabled {
			statuses = append(statuses, finding.ScannerStatus{
				ID: source.ID(), State: finding.ScannerSkipped, Required: required,
				Message: "disabled by configuration",
			})
			continue
		}
		timeout, _ := configured.TimeoutDuration()
		result := executeScanner(ctx, source, request, timeout)
		all = append(all, result.Findings...)
		statuses = append(statuses, finding.ScannerStatus{
			ID: source.ID(), State: result.State, Duration: result.Duration,
			Message: result.Message, Version: result.Version, Required: required,
		})
		if result.State == finding.ScannerFailed || result.State == finding.ScannerPartial {
			failure := fmt.Errorf("scanner %s failed: %s", source.ID(), result.Message)
			if required {
				operationalErrors = append(operationalErrors, failure)
			} else {
				warnings = append(warnings, failure.Error())
			}
		}
	}
	all = normalize(all)
	suppressions, err := suppression.Load(resolvePath(r.config.Root, r.config.SuppressionFile))
	if err != nil {
		return nil, err
	}
	active, suppressed, stale := suppression.Apply(all, suppressions, started)
	report := &finding.Report{
		SchemaVersion: SchemaVersion, Timestamp: started, Duration: time.Since(started),
		ScanMode: string(r.config.Mode), Project: r.config.Project,
		Findings: active, SuppressionsApplied: suppressed,
		StaleSuppressionFiles: stale, Scanners: statuses, Warnings: warnings,
	}
	report.Summary = summarize(active, suppressed, stale)
	return report, errors.Join(operationalErrors...)
}

func executeScanner(ctx context.Context, source scanner.Scanner, request scanner.Request, timeout time.Duration) scanner.Result {
	if timeout <= 0 {
		return source.Scan(ctx, request)
	}

	started := time.Now()
	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	results := make(chan scanner.Result, 1)
	go func() {
		results <- source.Scan(scanCtx, request)
	}()

	select {
	case result := <-results:
		return result
	case <-scanCtx.Done():
		message := scanCtx.Err().Error()
		if errors.Is(scanCtx.Err(), context.DeadlineExceeded) {
			message = fmt.Sprintf("timeout after %s", timeout)
		}
		return scanner.Result{
			State: finding.ScannerFailed, Message: message, Duration: time.Since(started),
		}
	}
}

func normalize(input []finding.Finding) []finding.Finding {
	seen := make(map[string]struct{}, len(input))
	output := make([]finding.Finding, 0, len(input))
	for _, item := range input {
		// Findings produced by scanners built against the original API predate
		// domains. The project was security-only at that point, so security is
		// the compatible classification for an omitted value.
		if item.Domain == "" {
			item.Domain = finding.Security
		}
		key := fmt.Sprintf("%s\x00%s\x00%d\x00%s", item.RuleID, filepath.ToSlash(item.Location.File), item.Location.Line, item.Description)
		fingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
		if _, ok := seen[fingerprint]; ok {
			continue
		}
		seen[fingerprint] = struct{}{}
		item.Fingerprint = fingerprint[:16]
		output = append(output, item)
	}
	sort.SliceStable(output, func(i, j int) bool {
		if output[i].Severity.Rank() != output[j].Severity.Rank() {
			return output[i].Severity.Rank() < output[j].Severity.Rank()
		}
		if output[i].Location.File != output[j].Location.File {
			return output[i].Location.File < output[j].Location.File
		}
		return output[i].Location.Line < output[j].Location.Line
	})
	for index := range output {
		output[index].ID = fmt.Sprintf("F-%04d", index+1)
	}
	return output
}

func summarize(active, suppressed []finding.Finding, stale []string) finding.Summary {
	summary := finding.Summary{Total: len(active), Suppressed: len(suppressed), StaleSuppressions: len(stale)}
	for _, item := range active {
		switch item.Severity {
		case finding.Critical:
			summary.Critical++
		case finding.High:
			summary.High++
		case finding.Medium:
			summary.Medium++
		case finding.Low:
			summary.Low++
		}
	}
	return summary
}

func resolvePaths(root string, paths []string) []string {
	resolved := make([]string, len(paths))
	for index, path := range paths {
		resolved[index] = resolvePath(root, path)
	}
	return resolved
}

func resolvePath(root, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
