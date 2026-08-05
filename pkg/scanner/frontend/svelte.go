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

type SvelteChecker struct {
	cfg        config.Config
	classifier *Classifier
}

func NewSvelteChecker(cfg config.Config) *SvelteChecker {
	return &SvelteChecker{
		cfg:        cfg,
		classifier: NewClassifier(cfg),
	}
}

func (c *SvelteChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
	if src.Open == nil {
		return nil, nil
	}

	scope := c.classifier.Classify(ctx, src)

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
		val := tok.Value

		if strings.Contains(val, "@html") || (val == "html" && i > 0 && (tokens[i-1].Value == "@" || tokens[i-1].Value == "{@")) {
			exprTokens := getRightHandTokens(tokens, i+1)
			if len(exprTokens) > 0 && !isStaticLiteralExpr(exprTokens) && !isSanitizedExpr(exprTokens, c.cfg.Frontend.RecognizeSanitizers) {
				rule, _ := LookupRule("frontend/svelte-html")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Svelte {@html} tag rendering un-sanitized dynamic expression",
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
					Tags:           rule.Tags,
				})
			}
		}

		if scope == ScopeClient {
			if val == "import" || val == "from" {
				for j := i + 1; j < n && j < i+6; j++ {
					if tokens[j].Type == TokenString {
						mod := strings.Trim(tokens[j].Value, `"'`+"`")
						if strings.Contains(mod, "$env/static/private") || strings.Contains(mod, "$env/dynamic/private") {
							rule, _ := LookupRule("frontend/sveltekit-private-env-in-client")
							findings = append(findings, finding.Finding{
								RuleID:         rule.ID,
								Domain:         rule.Domain,
								Category:       rule.Category,
								Severity:       rule.Severity,
								Description:    "Client component importing SvelteKit private env module " + mod,
								Recommendation: rule.Recommendation,
								Documentation:  rule.Documentation,
								Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
								Tags:           rule.Tags,
							})
							break
						}
						if strings.Contains(mod, ".server") || strings.Contains(mod, "$lib/server") {
							rule, _ := LookupRule("frontend/sveltekit-server-module-in-client")
							findings = append(findings, finding.Finding{
								RuleID:         rule.ID,
								Domain:         rule.Domain,
								Category:       rule.Category,
								Severity:       rule.Severity,
								Description:    "Client component importing SvelteKit server module " + mod,
								Recommendation: rule.Recommendation,
								Documentation:  rule.Documentation,
								Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
								Tags:           rule.Tags,
							})
							break
						}
					}
				}
			}
		}
	}

	return findings, nil
}
