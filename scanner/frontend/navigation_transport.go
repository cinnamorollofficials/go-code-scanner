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

type NavigationTransportChecker struct {
	cfg config.Config
}

func NewNavigationTransportChecker(cfg config.Config) *NavigationTransportChecker {
	return &NavigationTransportChecker{cfg: cfg}
}

var (
	httpPrefixRegex     = regexp.MustCompile(`(?i)http://([a-zA-Z0-9\.\-_:]+)`)
	sensitiveQueryRegex = regexp.MustCompile(`(?i)[?&](token|secret|password|auth_key|api_key)=`)
)

func isInsecureHttpUrl(val string) bool {
	matches := httpPrefixRegex.FindStringSubmatch(val)
	if len(matches) < 2 {
		return false
	}
	host := strings.ToLower(matches[1])
	if host == "localhost" || strings.HasPrefix(host, "localhost:") || host == "127.0.0.1" || strings.HasPrefix(host, "127.0.0.1:") || host == "[::1]" || strings.HasPrefix(host, "[::1]:") {
		return false
	}
	return true
}

func (c *NavigationTransportChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
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

		if tok.Type == TokenString || tok.Type == TokenTemplate {
			val := tok.Value
			lowerVal := strings.ToLower(val)

			if strings.Contains(lowerVal, "javascript:") {
				rule, _ := LookupRule("frontend/unsafe-navigation")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "javascript: pseudo-protocol URL used in client source",
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
					Tags:           rule.Tags,
				})
			}

			if isInsecureHttpUrl(val) {
				rule, _ := LookupRule("frontend/unsafe-transport")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Insecure non-localhost HTTP URL used in client code",
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
					Tags:           rule.Tags,
				})
			}

			if sensitiveQueryRegex.MatchString(val) {
				rule, _ := LookupRule("frontend/sensitive-query-param")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Sensitive credential placed in URL query parameter string",
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
					Tags:           rule.Tags,
				})
			}
			continue
		}

		if tok.Type != TokenCode && tok.Type != TokenJSXAttribute {
			continue
		}

		val := tok.Value

		if val == "href" || val == "assign" || val == "replace" {
			if i > 1 && tokens[i-1].Value == "." && (tokens[i-2].Value == "location" || tokens[i-2].Value == "window") {
				var rightHand []Token
				if i+1 < n && tokens[i+1].Value == "(" {
					rightHand = getArgTokens(tokens, i+1)
				} else {
					rightHand = getRightHandTokens(tokens, i+1)
				}
				if len(rightHand) > 0 && !isStaticLiteralExpr(rightHand) {
					rule, _ := LookupRule("frontend/unsafe-navigation")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Untrusted variable assigned to browser navigation sink",
						Recommendation: rule.Recommendation,
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
						Tags:           rule.Tags,
					})
				}
			}
			continue
		}

		if val == "open" && (i == 0 || (i > 1 && tokens[i-1].Value == "." && (tokens[i-2].Value == "window" || tokens[i-2].Value == "globalThis"))) {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+1)
				hasBlankTarget := false
				hasNoOpener := false
				for _, a := range args {
					lowerA := strings.ToLower(a.Value)
					if strings.Contains(lowerA, "_blank") {
						hasBlankTarget = true
					}
					if strings.Contains(lowerA, "noopener") || strings.Contains(lowerA, "noreferrer") {
						hasNoOpener = true
					}
				}
				if hasBlankTarget && !hasNoOpener {
					rule, _ := LookupRule("frontend/unsafe-navigation")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "window.open with _blank target missing noopener/noreferrer features",
						Recommendation: "Pass 'noopener,noreferrer' as window.open features when using target _blank",
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
