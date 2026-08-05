package frontend

import (
	"context"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

// remoteResourcePatterns matches src= or href= attributes with http/https URLs in HTML/template files.
var (
	reSrcHref      = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["']?(https?://[^\s"'>]+)["']?`)
	reIntegrity    = regexp.MustCompile(`(?i)\bintegrity\s*=`)
	reCrossOrigin  = regexp.MustCompile(`(?i)\bcrossorigin\s*=`)
	reScriptTag    = regexp.MustCompile(`(?i)<script\b[^>]*>`)
	reLinkTag      = regexp.MustCompile(`(?i)<link\b[^>]*>`)
	reVersioned    = regexp.MustCompile(`@[\d]+\.[\d]|/[\d]+\.[\d]|\?v=[\d]|/v[\d]+/`)
	reDynamicImport = regexp.MustCompile(`(?i)import\s*\(\s*["']?(https?://[^\s"'>]+)["']?\s*\)`)
)

type RemoteResourceChecker struct {
	cfg config.Config
}

func NewRemoteResourceChecker(cfg config.Config) *RemoteResourceChecker {
	return &RemoteResourceChecker{cfg: cfg}
}

func (c *RemoteResourceChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
	ext := strings.ToLower(filepath.Ext(src.Path))
	if ext != ".html" && ext != ".htm" && ext != ".vue" && ext != ".svelte" {
		return nil, nil
	}

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

	relPath := src.Path
	if c.cfg.Root != "" {
		if rel, err := filepath.Rel(c.cfg.Root, src.Path); err == nil {
			relPath = rel
		}
	}

	text := string(content)
	var findings []finding.Finding

	// Check <script src="..."> and <link href="..."> tags
	for _, lineText := range strings.Split(text, "\n") {
		lineNum := 1
		for i, line := range strings.Split(text, "\n") {
			if line == lineText {
				lineNum = i + 1
				break
			}
		}

		isScriptLine := reScriptTag.MatchString(lineText)
		isLinkLine := reLinkTag.MatchString(lineText)
		if !isScriptLine && !isLinkLine {
			continue
		}

		matches := reSrcHref.FindAllStringSubmatch(lineText, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			url := m[1]
			urlLower := strings.ToLower(url)

			// insecure HTTP
			if strings.HasPrefix(urlLower, "http://") {
				rule, _ := LookupRule("frontend/insecure-resource-url")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Resource loaded over insecure HTTP: " + url,
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: relPath, Line: lineNum},
					Metadata:       map[string]string{"url": url},
					Tags:           rule.Tags,
				})
				continue
			}

			// SRI check for cross-origin script/link (HTTPS only)
			// SRI is only applicable where browser contract supports it (script and link[rel=stylesheet])
			if isScriptLine || (isLinkLine && strings.Contains(strings.ToLower(lineText), "stylesheet")) {
				hasIntegrity := reIntegrity.MatchString(lineText)
				hasCrossOrigin := reCrossOrigin.MatchString(lineText)
				if !hasIntegrity || !hasCrossOrigin {
					rule, _ := LookupRule("frontend/missing-subresource-integrity")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Cross-origin resource without SRI: " + url,
						Recommendation: rule.Recommendation,
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: relPath, Line: lineNum},
						Metadata:       map[string]string{"url": url},
						Tags:           rule.Tags,
					})
				}
			}

			// version check (independent of SRI)
			if !reVersioned.MatchString(url) {
				rule, _ := LookupRule("frontend/unversioned-remote-resource")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Remote resource without version pin: " + url,
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: relPath, Line: lineNum},
					Metadata:       map[string]string{"url": url},
					Tags:           rule.Tags,
				})
			}
		}
	}

	// Dynamic remote imports
	dynMatches := reDynamicImport.FindAllStringSubmatch(text, -1)
	for _, m := range dynMatches {
		if len(m) < 2 {
			continue
		}
		url := m[1]
		rule, _ := LookupRule("frontend/missing-subresource-integrity")
		findings = append(findings, finding.Finding{
			RuleID:         rule.ID,
			Domain:         rule.Domain,
			Category:       rule.Category,
			Severity:       rule.Severity,
			Description:    "Dynamic remote import cannot be integrity-checked: " + url,
			Recommendation: "Avoid dynamic remote imports; bundle dependencies locally",
			Documentation:  rule.Documentation,
			Location:       finding.Location{File: relPath, Line: 1},
			Metadata:       map[string]string{"url": url},
			Tags:           rule.Tags,
		})
	}

	return findings, nil
}
