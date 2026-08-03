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

type ReactNextChecker struct {
	cfg        config.Config
	classifier *Classifier
}

func NewReactNextChecker(cfg config.Config) *ReactNextChecker {
	return &ReactNextChecker{
		cfg:        cfg,
		classifier: NewClassifier(cfg),
	}
}

var nodeServerOnlyModules = map[string]struct{}{
	"fs": {}, "crypto": {}, "path": {}, "child_process": {},
	"net": {}, "dns": {}, "tls": {}, "server-only": {},
}

func (c *ReactNextChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
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

		if val == "dangerouslySetInnerHTML" {
			exprTokens := getRightHandTokens(tokens, i+1)
			if len(exprTokens) > 0 && !isStaticLiteralExpr(exprTokens) && !isSanitizedExpr(exprTokens, c.cfg.Frontend.RecognizeSanitizers) {
				rule, _ := LookupRule("frontend/react-dangerously-set-inner-html")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "React component using dangerouslySetInnerHTML with un-sanitized dynamic expression",
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
						if _, ok := nodeServerOnlyModules[mod]; ok {
							rule, _ := LookupRule("frontend/next-server-module-in-client")
							findings = append(findings, finding.Finding{
								RuleID:         rule.ID,
								Domain:         rule.Domain,
								Category:       rule.Category,
								Severity:       rule.Severity,
								Description:    "Client module importing Node/server-only module " + mod,
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

			if val == "process" && i+4 < n && tokens[i+1].Value == "." && tokens[i+2].Value == "env" && tokens[i+3].Value == "." {
				envVar := tokens[i+4].Value
				if envVar != "" && !strings.HasPrefix(envVar, "NEXT_PUBLIC_") && !strings.HasPrefix(envVar, "NODE_ENV") && !strings.HasPrefix(envVar, "PUBLIC_") {
					rule, _ := LookupRule("frontend/next-private-env-in-client")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Client module attempting to read private server environment variable process.env." + envVar,
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
