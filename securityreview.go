package securityreview

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/discovery"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
	"github.com/cinnamorollofficials/go-code-scanner/scanner/adapters"
	architecturescanner "github.com/cinnamorollofficials/go-code-scanner/scanner/architecture"
	commandscanner "github.com/cinnamorollofficials/go-code-scanner/scanner/command"
	patternscanner "github.com/cinnamorollofficials/go-code-scanner/scanner/pattern"
	"github.com/cinnamorollofficials/go-code-scanner/suppression"
)

const SchemaVersion = "1.0"
const FingerprintVersion = "3"

type Reviewer interface {
	Run(context.Context) (*finding.Report, error)
}

type reviewer struct {
	config      config.Config
	scanners    []registeredScanner
	now         func() time.Time
	toolVersion string
	configHash  string
	ruleSetHash string
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

func WithToolVersion(version string) Option {
	return func(r *reviewer) error {
		r.toolVersion = strings.TrimSpace(version)
		return nil
	}
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
	configHash, err := hashJSON(cfg)
	if err != nil {
		return nil, fmt.Errorf("hash configuration: %w", err)
	}
	ruleValues := make([]rules.Rule, len(compiled))
	for index := range compiled {
		ruleValues[index] = compiled[index].Rule
	}
	ruleSetHash, err := hashJSON(ruleValues)
	if err != nil {
		return nil, fmt.Errorf("hash rule set: %w", err)
	}
	headerPolicies := make([]patternscanner.HeaderPolicy, len(cfg.Governance.RequiredHeaders))
	for index, header := range cfg.Governance.RequiredHeaders {
		headerPolicies[index] = patternscanner.HeaderPolicy{
			ID: header.ID, Paths: header.Paths, Pattern: header.Pattern, MaxLines: header.MaxLines,
			Severity: header.Severity, Description: header.Description, Recommendation: header.Recommendation,
		}
	}
	ownershipPolicies := make([]patternscanner.OwnershipPolicy, len(cfg.Governance.OwnershipRules))
	for index, rule := range cfg.Governance.OwnershipRules {
		ownershipPolicies[index] = patternscanner.OwnershipPolicy{Path: rule.Path, Owners: append([]string(nil), rule.Owners...), Severity: rule.Severity}
	}
	r := &reviewer{
		config: cfg,
		scanners: []registeredScanner{{scanner: patternscanner.New(compiled, cfg.Workers, patternscanner.Limits{
			MaxFileBytes: cfg.PatternMaxFileBytes, MaxLineBytes: cfg.PatternMaxLineBytes,
			QualityMaxFileBytes: cfg.QualityMaxFileBytes, QualityMaxLineLength: cfg.QualityMaxLineLength,
			DependencyAllowlist: cfg.SupplyChain.DependencyAllowlist, DependencyDenylist: cfg.SupplyChain.DependencyDenylist,
			LicenseAllowlist: cfg.SupplyChain.LicenseAllowlist, LicenseDenylist: cfg.SupplyChain.LicenseDenylist,
			RequiredFiles:   cfg.Governance.RequiredFiles,
			RequiredHeaders: headerPolicies,
			OwnershipFile:   cfg.Governance.OwnershipFile,
			OwnershipRules:  ownershipPolicies,
		}), required: true}},
		now:         time.Now,
		configHash:  configHash,
		ruleSetHash: ruleSetHash,
	}
	if len(cfg.Architecture.Layers) > 0 || cfg.Architecture.DetectCycles {
		layers := make([]architecturescanner.Layer, len(cfg.Architecture.Layers))
		for index, layer := range cfg.Architecture.Layers {
			layers[index] = architecturescanner.Layer{Name: layer.Name, Paths: append([]string(nil), layer.Paths...)}
		}
		boundaries := make([]architecturescanner.Boundary, len(cfg.Architecture.ForbiddenDependencies))
		for index, boundary := range cfg.Architecture.ForbiddenDependencies {
			boundaries[index] = architecturescanner.Boundary{From: boundary.From, To: boundary.To}
		}
		r.scanners = append(r.scanners, registeredScanner{scanner: architecturescanner.New(layers, boundaries, architecturescanner.Options{
			DetectCycles: cfg.Architecture.DetectCycles,
		}), required: true})
	}
	configuredIDs := make([]string, 0, len(cfg.Scanners))
	for id, configured := range cfg.Scanners {
		if configured.Type == "command" || configured.Type == "adapter" {
			configuredIDs = append(configuredIDs, id)
		}
	}
	sort.Strings(configuredIDs)
	for _, id := range configuredIDs {
		configured := cfg.Scanners[id]
		var source scanner.Scanner
		var err error
		if configured.Type == "adapter" {
			source, err = adapters.New(id, configured.Adapter, configured.AdapterOptions())
		} else {
			source, err = commandscanner.New(configured.CommandSpec(id))
		}
		if err != nil {
			return nil, err
		}
		r.scanners = append(r.scanners, registeredScanner{scanner: source, required: configured.Required})
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
	files, err := discovery.Files(ctx, r.config)
	if err != nil {
		return nil, err
	}
	repositoryFiles, err := discovery.RepositoryFiles(ctx, r.config)
	if err != nil {
		return nil, err
	}
	request := scanner.Request{Root: r.config.Root, Mode: string(r.config.Mode), Sources: sources, Files: files, RepositoryFiles: repositoryFiles}
	var all []finding.Finding
	statuses := make([]finding.ScannerStatus, 0, len(r.scanners))
	var operationalErrors []error
	var warnings []string
	for _, outcome := range r.runScanners(ctx, request) {
		all = append(all, outcome.findings...)
		statuses = append(statuses, outcome.status)
		if outcome.failure == nil {
			continue
		}
		if outcome.status.Required {
			operationalErrors = append(operationalErrors, outcome.failure)
		} else {
			warnings = append(warnings, outcome.failure.Error())
		}
	}
	all = normalize(all)
	suppressions, err := suppression.Load(resolvePath(r.config.Root, r.config.SuppressionFile))
	if err != nil {
		return nil, err
	}
	active, suppressed, stale := suppression.Apply(all, suppressions, started)
	report := &finding.Report{
		SchemaVersion: SchemaVersion, FingerprintVersion: FingerprintVersion,
		ToolVersion: r.toolVersion, ConfigHash: r.configHash, RuleSetHash: r.ruleSetHash,
		Timestamp: started, Duration: time.Since(started),
		ScanMode: string(r.config.Mode), Project: r.config.Project,
		Findings: active, SuppressionsApplied: suppressed,
		StaleSuppressionFiles: stale, Scanners: statuses, Warnings: warnings,
	}
	report.Summary = summarize(active, suppressed, stale)
	return report, errors.Join(operationalErrors...)
}

func hashJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return fmt.Sprintf("%x", digest[:8]), nil
}

type scannerOutcome struct {
	findings []finding.Finding
	status   finding.ScannerStatus
	failure  error
}

func (r *reviewer) runScanners(ctx context.Context, request scanner.Request) []scannerOutcome {
	outcomes := make([]scannerOutcome, len(r.scanners))
	if len(r.scanners) == 0 {
		return outcomes
	}

	jobs := make(chan int)
	workerCount := min(r.config.Workers, len(r.scanners))
	done := make(chan struct{}, workerCount)
	for range workerCount {
		go func() {
			defer func() { done <- struct{}{} }()
			for index := range jobs {
				outcomes[index] = r.runScanner(ctx, r.scanners[index], request)
			}
		}()
	}
	for index := range r.scanners {
		jobs <- index
	}
	close(jobs)
	for range workerCount {
		<-done
	}
	return outcomes
}

func (r *reviewer) runScanner(ctx context.Context, registered registeredScanner, request scanner.Request) scannerOutcome {
	source := registered.scanner
	required := registered.required
	descriptor := describeScanner(source)
	configured, hasConfig := r.config.Scanners[source.ID()]
	if hasConfig {
		required = configured.Required
	}
	if r.config.SelectedProfile != "" && !profileContains(r.config.Profiles[r.config.SelectedProfile], source.ID()) {
		return scannerOutcome{status: finding.ScannerStatus{
			ID: source.ID(), State: finding.ScannerSkipped, Required: required,
			Message: fmt.Sprintf("not included in profile %s", r.config.SelectedProfile),
			Domain:  descriptor.Domain, Capabilities: descriptor.Capabilities, SupportedModes: descriptor.SupportedModes,
		}}
	}
	if len(descriptor.SupportedModes) > 0 && !profileContains(descriptor.SupportedModes, request.Mode) {
		return scannerOutcome{status: finding.ScannerStatus{
			ID: source.ID(), State: finding.ScannerSkipped, Required: required,
			Message: fmt.Sprintf("scan mode %s is not supported", request.Mode),
			Domain:  descriptor.Domain, Capabilities: descriptor.Capabilities, SupportedModes: descriptor.SupportedModes,
		}}
	}
	if hasConfig && !configured.Enabled {
		return scannerOutcome{status: finding.ScannerStatus{
			ID: source.ID(), State: finding.ScannerSkipped, Required: required,
			Message: "disabled by configuration",
			Domain:  descriptor.Domain, Capabilities: descriptor.Capabilities, SupportedModes: descriptor.SupportedModes,
		}}
	}

	timeout, _ := configured.TimeoutDuration()
	result := executeScanner(ctx, source, request, timeout)
	if result.Version == "" {
		result.Version = descriptor.Version
	}
	if result.State == finding.ScannerPartial && result.Failure == "" {
		result.Failure = scanner.FailurePartial
	}
	if result.State == finding.ScannerFailed && result.Failure == "" {
		result.Failure = scanner.FailureExecution
	}
	outcome := scannerOutcome{
		findings: result.Findings,
		status: finding.ScannerStatus{
			ID: source.ID(), State: result.State, Duration: result.Duration,
			Message: result.Message, Version: result.Version, Required: required,
			Domain: descriptor.Domain, Capabilities: descriptor.Capabilities,
			SupportedModes: descriptor.SupportedModes, FailureKind: string(result.Failure),
		},
	}
	if result.State == finding.ScannerFailed || result.State == finding.ScannerPartial {
		outcome.failure = fmt.Errorf("scanner %s failed: %s", source.ID(), result.Message)
	}
	return outcome
}

func describeScanner(source scanner.Scanner) scanner.Descriptor {
	if described, ok := source.(scanner.Described); ok {
		descriptor := described.Describe()
		descriptor.Capabilities = append([]string(nil), descriptor.Capabilities...)
		descriptor.SupportedModes = append([]string(nil), descriptor.SupportedModes...)
		return descriptor
	}
	return scanner.Descriptor{Domain: finding.Security}
}

func profileContains(scanners []string, id string) bool {
	for _, candidate := range scanners {
		if candidate == id {
			return true
		}
	}
	return false
}

func executeScanner(ctx context.Context, source scanner.Scanner, request scanner.Request, timeout time.Duration) (result scanner.Result) {
	if timeout <= 0 {
		return scanSafely(ctx, source, request)
	}

	started := time.Now()
	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	results := make(chan scanner.Result, 1)
	go func() {
		results <- scanSafely(scanCtx, source, request)
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
			State: finding.ScannerFailed, Message: message, Duration: time.Since(started), Failure: scanner.FailureTimeout,
		}
	}
}

func scanSafely(ctx context.Context, source scanner.Scanner, request scanner.Request) (result scanner.Result) {
	started := time.Now()
	defer func() {
		if recovered := recover(); recovered != nil {
			result = scanner.Result{
				State:    finding.ScannerFailed,
				Message:  fmt.Sprintf("panic: %v", recovered),
				Duration: time.Since(started),
				Failure:  scanner.FailurePanic,
			}
		}
	}()
	return source.Scan(ctx, request)
}

func normalize(input []finding.Finding) []finding.Finding {
	input = append([]finding.Finding(nil), input...)
	sortFindings(input)
	seen := make(map[string]struct{}, len(input))
	occurrences := make(map[string]int, len(input))
	output := make([]finding.Finding, 0, len(input))
	for _, item := range input {
		// Findings produced by scanners built against the original API predate
		// domains. The project was security-only at that point, so security is
		// the compatible classification for an omitted value.
		if item.Domain == "" {
			item.Domain = finding.Security
		}
		path := filepath.ToSlash(item.Location.File)
		symbol := normalizeFingerprintText(item.Metadata["symbol"])
		deduplicationKey := fmt.Sprintf("%s\x00%s\x00%d\x00%s\x00%s", item.RuleID, path, item.Location.Line, item.Description, symbol)
		if _, ok := seen[deduplicationKey]; ok {
			continue
		}
		seen[deduplicationKey] = struct{}{}
		identity := strings.Join([]string{
			FingerprintVersion, string(item.Domain), item.RuleID, path,
			normalizeFingerprintText(item.Description), normalizeFingerprintText(item.Snippet), symbol,
		}, "\x00")
		occurrences[identity]++
		fingerprintInput := fmt.Sprintf("%s\x00%d", identity, occurrences[identity])
		fingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(fingerprintInput)))
		item.Fingerprint = fingerprint[:16]
		output = append(output, item)
	}
	sortFindings(output)
	for index := range output {
		output[index].ID = fmt.Sprintf("F-%04d", index+1)
	}
	return output
}

func sortFindings(items []finding.Finding) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Severity.Rank() != items[j].Severity.Rank() {
			return items[i].Severity.Rank() < items[j].Severity.Rank()
		}
		if items[i].Location.File != items[j].Location.File {
			return items[i].Location.File < items[j].Location.File
		}
		if items[i].Location.Line != items[j].Location.Line {
			return items[i].Location.Line < items[j].Location.Line
		}
		if items[i].RuleID != items[j].RuleID {
			return items[i].RuleID < items[j].RuleID
		}
		if items[i].Description != items[j].Description {
			return items[i].Description < items[j].Description
		}
		return items[i].Tool < items[j].Tool
	})
}

func normalizeFingerprintText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func summarize(active, suppressed []finding.Finding, stale []string) finding.Summary {
	summary := finding.Summary{
		Total: len(active), Suppressed: len(suppressed), StaleSuppressions: len(stale),
		ByDomain: make(map[finding.Domain]int),
	}
	for _, item := range active {
		summary.ByDomain[item.Domain]++
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
