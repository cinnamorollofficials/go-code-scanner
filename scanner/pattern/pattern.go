package pattern

import (
	"bufio"
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type Scanner struct {
	genericRules  []rules.Compiled
	rulesBySuffix map[string][]rules.Compiled
	workers       int
}

func New(compiled []rules.Compiled, workers int) *Scanner {
	s := &Scanner{rulesBySuffix: make(map[string][]rules.Compiled), workers: max(workers, 1)}
	for _, rule := range compiled {
		if len(rule.Extensions) == 0 {
			s.genericRules = append(s.genericRules, rule)
			continue
		}
		for _, suffix := range rule.Extensions {
			suffix = strings.ToLower(suffix)
			s.rulesBySuffix[suffix] = append(s.rulesBySuffix[suffix], rule)
		}
	}
	return s
}

func (s *Scanner) ID() string { return "pattern" }

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean}
	if len(request.Sources) == 0 {
		result.Duration = time.Since(started)
		return result
	}
	jobs := make(chan scanner.Source)
	outcomes := make(chan outcome, len(request.Sources))
	workerCount := min(s.workers, len(request.Sources))
	for range workerCount {
		go func() {
			for source := range jobs {
				items, err := s.scanSource(ctx, source, request.Root)
				outcomes <- outcome{findings: items, err: err}
			}
		}()
	}
	go feedSources(ctx, request.Sources, jobs)
	for completed := 0; completed < len(request.Sources); completed++ {
		select {
		case item := <-outcomes:
			result.Findings = append(result.Findings, item.findings...)
			if item.err != nil {
				result.Message = appendMessage(result.Message, item.err.Error())
			}
		case <-ctx.Done():
			result.State, result.Message = finding.ScannerFailed, ctx.Err().Error()
			result.Duration = time.Since(started)
			return result
		}
	}
	if result.Message != "" {
		result.State = finding.ScannerPartial
	} else if len(result.Findings) > 0 {
		result.State = finding.ScannerFindings
	}
	result.Duration = time.Since(started)
	return result
}

type outcome struct {
	findings []finding.Finding
	err      error
}

func feedSources(ctx context.Context, sources []scanner.Source, jobs chan<- scanner.Source) {
	defer close(jobs)
	for _, source := range sources {
		select {
		case jobs <- source:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scanner) scanSource(ctx context.Context, source scanner.Source, root string) ([]finding.Finding, error) {
	file, err := source.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	relative, err := filepath.Rel(root, source.Path)
	if err != nil {
		relative = source.Path
	}
	relative = filepath.ToSlash(relative)
	extension := strings.ToLower(filepath.Ext(source.Path))
	applicableRules := s.rulesFor(extension)
	var findings []finding.Finding
	lineNumber := 0
	lineScanner := bufio.NewScanner(file)
	lineScanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for lineScanner.Scan() {
		lineNumber++
		line := lineScanner.Text()
		for _, rule := range applicableRules {
			if !rule.Regex.MatchString(line) {
				continue
			}
			findings = append(findings, finding.Finding{
				RuleID: rule.ID, Tool: s.ID(), Domain: rule.Domain, Category: rule.Category,
				Severity: rule.Severity, Description: rule.Description,
				Recommendation: rule.Recommendation, Snippet: redact(rule, line),
				Location: finding.Location{File: relative, Line: lineNumber},
			})
		}
	}
	return findings, lineScanner.Err()
}

func (s *Scanner) rulesFor(extension string) []rules.Compiled {
	result := make([]rules.Compiled, 0, len(s.genericRules)+len(s.rulesBySuffix[extension]))
	result = append(result, s.genericRules...)
	result = append(result, s.rulesBySuffix[extension]...)
	return result
}

func redact(rule rules.Compiled, value string) string {
	if rule.Category == "secret_leak" {
		return "[REDACTED: " + rule.ID + "]"
	}
	value = strings.TrimSpace(value)
	if len(value) > 200 {
		value = value[:200]
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "api_key") || strings.Contains(lower, "apikey") || strings.Contains(lower, "authorization") {
		return "[REDACTED: potentially sensitive source line]"
	}
	return value
}

func appendMessage(current, addition string) string {
	if current == "" {
		return addition
	}
	return current + "; " + addition
}
