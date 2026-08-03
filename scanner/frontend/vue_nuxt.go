package frontend

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type VueNuxtChecker struct {
	cfg        config.Config
	classifier *Classifier
}

func NewVueNuxtChecker(cfg config.Config) *VueNuxtChecker {
	return &VueNuxtChecker{
		cfg:        cfg,
		classifier: NewClassifier(cfg),
	}
}

func (c *VueNuxtChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
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

		if strings.HasPrefix(val, "v-html") {
			exprTokens := getRightHandTokens(tokens, i+1)
			if len(exprTokens) > 0 && !isStaticLiteralExpr(exprTokens) && !isSanitizedExpr(exprTokens, c.cfg.Frontend.RecognizeSanitizers) {
				rule, _ := LookupRule("frontend/vue-v-html")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Vue v-html directive bound to un-sanitized dynamic expression",
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
						if strings.Contains(mod, "/server/") || strings.HasPrefix(mod, "#server") || strings.HasPrefix(mod, "~/server/") || strings.HasPrefix(mod, "@/server/") {
							rule, _ := LookupRule("frontend/nuxt-server-import-in-client")
							findings = append(findings, finding.Finding{
								RuleID:         rule.ID,
								Domain:         rule.Domain,
								Category:       rule.Category,
								Severity:       rule.Severity,
								Description:    "Client code importing Nuxt server module " + mod,
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

			if (val == "useRuntimeConfig" || val == "runtimeConfig" || val == "config") && i+2 < n && tokens[i+1].Value == "." {
				field := tokens[i+2].Value
				if field == "(" && i+4 < n && tokens[i+3].Value == "." {
					field = tokens[i+4].Value
				}
				if field != "" && field != "(" && field != ")" && field != "public" && field != "app" && field != "global" && isSensitiveConfigField(field) {
					rule, _ := LookupRule("frontend/nuxt-private-runtime-config")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Client component reading private Nuxt runtimeConfig field " + field,
						Recommendation: rule.Recommendation,
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

func isSensitiveConfigField(field string) bool {
	lower := strings.ToLower(field)
	return strings.Contains(lower, "secret") || strings.Contains(lower, "private") || strings.Contains(lower, "password") || strings.Contains(lower, "key") || strings.Contains(lower, "token")
}
