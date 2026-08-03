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

type DOMInjectionChecker struct {
	cfg config.Config
}

func NewDOMInjectionChecker(cfg config.Config) *DOMInjectionChecker {
	return &DOMInjectionChecker{cfg: cfg}
}

func (c *DOMInjectionChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
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

	sanitizers := c.cfg.Frontend.RecognizeSanitizers
	if len(sanitizers) == 0 {
		sanitizers = []string{"DOMPurify.sanitize", "sanitizeHtml", "sanitize"}
	}

	relPath := src.Path
	if c.cfg.Root != "" {
		if rel, err := filepath.Rel(c.cfg.Root, src.Path); err == nil {
			relPath = rel
		}
	}
	cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))

	var findings []finding.Finding

	for i, tok := range tokens {
		if tok.Type != TokenCode && tok.Type != TokenJSXAttribute {
			continue
		}

		val := tok.Value
		if val == "innerHTML" || val == "outerHTML" || val == "insertAdjacentHTML" || val == "write" || val == "writeln" {
			if (val == "write" || val == "writeln") && i > 0 {
				if tokens[i-1].Value != "." && (i < 2 || tokens[i-2].Value != "document") {
					continue
				}
			}
		} else {
			continue
		}

		exprTokens := getRightHandTokens(tokens, i+1)
		if len(exprTokens) == 0 {
			continue
		}

		if isStaticLiteralExpr(exprTokens) {
			continue
		}

		if isSanitizedExpr(exprTokens, sanitizers) {
			continue
		}

		rule, _ := LookupRule("frontend/dom-injection")
		f := finding.Finding{
			RuleID:         rule.ID,
			Domain:         rule.Domain,
			Category:       rule.Category,
			Severity:       rule.Severity,
			Description:    rule.Description,
			Recommendation: rule.Recommendation,
			Documentation:  rule.Documentation,
			Location: finding.Location{
				File: cleanRelPath,
				Line: tok.Line,
			},
			Tags: rule.Tags,
		}
		findings = append(findings, f)
	}

	return findings, nil
}

func getRightHandTokens(tokens []Token, startIdx int) []Token {
	n := len(tokens)
	i := startIdx
	for i < n && (tokens[i].Value == "=" || tokens[i].Value == "(" || tokens[i].Value == " " || tokens[i].Value == "\t") {
		i++
	}
	var expr []Token
	for i < n {
		tok := tokens[i]
		if tok.Value == ";" || tok.Value == "\n" {
			break
		}
		expr = append(expr, tok)
		i++
	}
	return expr
}

func isStaticLiteralExpr(tokens []Token) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, tok := range tokens {
		if tok.Type != TokenString && tok.Value != "+" && tok.Value != " " && tok.Value != "\t" {
			return false
		}
	}
	return true
}

func isSanitizedExpr(tokens []Token, sanitizers []string) bool {
	exprStr := ""
	for _, tok := range tokens {
		exprStr += tok.Value
	}
	lowerExpr := strings.ToLower(exprStr)
	if strings.Contains(lowerExpr, "trustedtypes") || strings.Contains(lowerExpr, "createhtml") {
		return true
	}
	for _, san := range sanitizers {
		sanClean := strings.ToLower(strings.TrimSpace(san))
		if sanClean != "" && strings.Contains(lowerExpr, sanClean) {
			return true
		}
	}
	return false
}
