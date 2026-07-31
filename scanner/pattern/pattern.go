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
	rules []rules.Compiled
}

func New(compiled []rules.Compiled) *Scanner {
	return &Scanner{rules: compiled}
}

func (s *Scanner) ID() string { return "pattern" }

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean}
	for _, source := range request.Sources {
		if err := ctx.Err(); err != nil {
			result.State, result.Message = finding.ScannerFailed, err.Error()
			break
		}
		findings, err := s.scanSource(ctx, source, request.Root)
		if err != nil {
			result.Message = appendMessage(result.Message, err.Error())
			continue
		}
		result.Findings = append(result.Findings, findings...)
	}
	if len(result.Findings) > 0 && result.State != finding.ScannerFailed {
		result.State = finding.ScannerFindings
	}
	result.Duration = time.Since(started)
	return result
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
	var findings []finding.Finding
	lineNumber := 0
	lineScanner := bufio.NewScanner(file)
	lineScanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for lineScanner.Scan() {
		lineNumber++
		line := lineScanner.Text()
		for _, rule := range s.rules {
			if !rules.MatchesExtension(rule, extension) || !rule.Regex.MatchString(line) {
				continue
			}
			findings = append(findings, finding.Finding{
				RuleID: rule.ID, Tool: s.ID(), Category: rule.Category,
				Severity: rule.Severity, Description: rule.Description,
				Recommendation: rule.Recommendation, Snippet: redact(rule, line),
				Location: finding.Location{File: relative, Line: lineNumber},
			})
		}
	}
	return findings, lineScanner.Err()
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
