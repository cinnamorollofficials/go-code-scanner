package frontend

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type TelemetryPrivacyChecker struct {
	cfg config.Config
}

func NewTelemetryPrivacyChecker(cfg config.Config) *TelemetryPrivacyChecker {
	return &TelemetryPrivacyChecker{cfg: cfg}
}

var defaultPIINames = []string{
	"password", "passwd", "secret", "ssn", "creditcard", "credit_card",
	"email", "user_email", "phone", "phone_number", "dob", "birthdate",
	"national_id", "nationalid",
}

func (c *TelemetryPrivacyChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
	if src.Open == nil {
		return nil, nil
	}

	lowerPath := strings.ToLower(src.Path)
	if strings.Contains(lowerPath, ".test.") || strings.Contains(lowerPath, ".spec.") || strings.Contains(lowerPath, "__tests__") {
		return nil, nil
	}

	rc, err := src.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	tokens, err := Tokenize(content)
	if err != nil {
		return nil, nil
	}

	relPath := src.Path
	if c.cfg.Root != "" {
		if rel, err := filepath.Rel(c.cfg.Root, src.Path); err == nil {
			relPath = rel
		}
	}
	cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))

	var findings []finding.Finding

	n := len(tokens)
	for i := 0; i < n; i++ {
		tok := tokens[i]
		if tok.Type != TokenCode && tok.Type != TokenJSXAttribute {
			continue
		}

		val := tok.Value

		isConsoleLog := (val == "log" || val == "info" || val == "warn" || val == "error" || val == "debug") && i > 1 && tokens[i-1].Value == "." && tokens[i-2].Value == "console"
		isTelemetryCall := (val == "track" || val == "logEvent" || val == "captureException" || val == "setContext") && i > 1 && tokens[i-1].Value == "."

		if !isConsoleLog && !isTelemetryCall {
			continue
		}

		args := getArgTokens(tokens, i+1)
		piiField := findPIIFieldName(args, defaultPIINames)
		if piiField != "" {
			rule, _ := LookupRule("frontend/telemetry-privacy-leak")
			findings = append(findings, finding.Finding{
				RuleID:         rule.ID,
				Domain:         rule.Domain,
				Category:       rule.Category,
				Severity:       rule.Severity,
				Description:    "Potential PII field (" + piiField + ") logged to client console or telemetry stream",
				Recommendation: rule.Recommendation,
				Documentation:  rule.Documentation,
				Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
				Tags:           rule.Tags,
			})
		}
	}

	return findings, nil
}

func findPIIFieldName(args []Token, defaultPIIs []string) string {
	for _, tok := range args {
		val := strings.ToLower(tok.Value)
		val = strings.Trim(val, `"'`+"`")
		for _, pii := range defaultPIIs {
			if val == pii || strings.Contains(val, pii) {
				return pii
			}
		}
	}
	return ""
}
