package frontend

import (
	"context"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type SecretExposureChecker struct {
	cfg config.Config
}

func NewSecretExposureChecker(cfg config.Config) *SecretExposureChecker {
	return &SecretExposureChecker{cfg: cfg}
}

var (
	secretEnvNameRegex = regexp.MustCompile(`(?i)(NEXT_PUBLIC_|VITE_|REACT_APP_)[A-Z0-9_]*(SECRET|PRIVATE_KEY|PASSWORD|TOKEN|AUTH_KEY)[A-Z0-9_]*`)
	privateKeyRegex    = regexp.MustCompile(`-----BEGIN (RSA|EC|OPENSSH|DSA|PRIVATE) KEY-----`)
)

func (c *SecretExposureChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
	if src.Open == nil {
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

	if loc := privateKeyRegex.FindIndex(content); loc != nil {
		line := bytesLineNumber(content, loc[0])
		rule, _ := LookupRule("frontend/client-credential-exposure")
		findings = append(findings, finding.Finding{
			RuleID:         rule.ID,
			Domain:         rule.Domain,
			Category:       rule.Category,
			Severity:       rule.Severity,
			Description:    "Embedded private key detected in client source",
			Recommendation: "Move private keys out of client bundle into secure server-side storage",
			Documentation:  rule.Documentation,
			Location:       finding.Location{File: cleanRelPath, Line: line},
			Tags:           rule.Tags,
		})
	}

	n := len(tokens)
	for i := 0; i < n; i++ {
		tok := tokens[i]
		if tok.Type != TokenCode && tok.Type != TokenString {
			continue
		}

		val := tok.Value

		if secretEnvNameRegex.MatchString(val) {
			rule, _ := LookupRule("frontend/client-credential-exposure")
			findings = append(findings, finding.Finding{
				RuleID:         rule.ID,
				Domain:         rule.Domain,
				Category:       rule.Category,
				Severity:       rule.Severity,
				Description:    "Secret or private credential exposed via public client environment variable",
				Recommendation: "Do not expose secret keys or private tokens through public environment variable prefixes",
				Documentation:  rule.Documentation,
				Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
				Tags:           rule.Tags,
			})
			continue
		}

		if (val == "localStorage" || val == "sessionStorage" || val == "indexedDB") && i+2 < n && tokens[i+1].Value == "." {
			method := tokens[i+2].Value
			if method == "setItem" || method == "put" || method == "add" {
				args := getArgTokens(tokens, i+3)
				if len(args) > 0 && isSensitiveStorageKey(args[0].Value) {
					rule, _ := LookupRule("frontend/client-credential-exposure")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Storing sensitive token or credential in unencrypted browser storage",
						Recommendation: "Avoid storing sensitive tokens or credentials in localStorage/sessionStorage; use secure HTTP-only cookies",
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
						Tags:           rule.Tags,
					})
				}
			}
		}
	}

	return findings, nil
}

func isSensitiveStorageKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "auth_key") || strings.Contains(lower, "api_key")
}

func bytesLineNumber(content []byte, offset int) int {
	line := 1
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}
