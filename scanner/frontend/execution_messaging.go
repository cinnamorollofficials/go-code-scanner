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

type ExecutionMessagingChecker struct {
	cfg config.Config
}

func NewExecutionMessagingChecker(cfg config.Config) *ExecutionMessagingChecker {
	return &ExecutionMessagingChecker{cfg: cfg}
}

func (c *ExecutionMessagingChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
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
		if tok.Type != TokenCode {
			continue
		}

		val := tok.Value

		// 1. eval(...)
		if val == "eval" {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+2)
				if len(args) > 0 && !isStaticLiteralExpr(args) {
					rule, _ := LookupRule("frontend/unsafe-execution")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Unsafe variable passed to eval execution sink",
						Recommendation: rule.Recommendation,
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
						Tags:           rule.Tags,
					})
				}
			}
			continue
		}

		// 2. new Function(...) or Function(...)
		if val == "Function" {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+2)
				if len(args) > 0 && !isStaticLiteralExpr(args) {
					rule, _ := LookupRule("frontend/unsafe-execution")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Dynamic string passed to Function constructor",
						Recommendation: rule.Recommendation,
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
						Tags:           rule.Tags,
					})
				}
			}
			continue
		}

		// 3. String-based timers: setTimeout("code", 100) / setInterval("code", 100)
		if val == "setTimeout" || val == "setInterval" {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+2)
				if len(args) > 0 && (args[0].Type == TokenString || args[0].Type == TokenTemplate) {
					rule, _ := LookupRule("frontend/unsafe-execution")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "String-based code execution in timer",
						Recommendation: "Pass a function closure instead of a string to setTimeout/setInterval",
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
						Tags:           rule.Tags,
					})
				}
			}
			continue
		}

		// 4. Wildcard postMessage: postMessage(data, "*")
		if val == "postMessage" {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+2)
				if hasWildcardOriginArg(args) {
					rule, _ := LookupRule("frontend/unsafe-messaging")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "postMessage sent with wildcard '*' target origin",
						Recommendation: rule.Recommendation,
						Documentation:  rule.Documentation,
						Location:       finding.Location{File: cleanRelPath, Line: tok.Line},
						Tags:           rule.Tags,
					})
				}
			}
			continue
		}

		// 5. addEventListener('message', ...) handler without origin check
		if val == "addEventListener" && i+2 < n && tokens[i+1].Value == "(" {
			args := getArgTokens(tokens, i+2)
			if len(args) > 0 && strings.Contains(args[0].Value, "message") {
				if !hasOriginCheck(tokens, i) {
					rule, _ := LookupRule("frontend/unsafe-messaging")
					findings = append(findings, finding.Finding{
						RuleID:         rule.ID,
						Domain:         rule.Domain,
						Category:       rule.Category,
						Severity:       rule.Severity,
						Description:    "Message event listener missing origin validation",
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

func getArgTokens(tokens []Token, startIdx int) []Token {
	n := len(tokens)
	i := startIdx
	depth := 0
	var args []Token
	for i < n {
		tok := tokens[i]
		if tok.Value == "(" {
			depth++
			if depth == 1 {
				i++
				continue
			}
		} else if tok.Value == ")" {
			depth--
			if depth == 0 {
				break
			}
		}
		if depth > 0 {
			args = append(args, tok)
		}
		i++
	}
	return args
}

func hasWildcardOriginArg(args []Token) bool {
	for _, tok := range args {
		if tok.Type == TokenString && (tok.Value == `"*"` || tok.Value == `'*'`) {
			return true
		}
	}
	return false
}

func hasOriginCheck(tokens []Token, listenerIdx int) bool {
	n := len(tokens)
	end := listenerIdx + 60
	if end > n {
		end = n
	}
	for i := listenerIdx; i < end; i++ {
		val := tokens[i].Value
		if val == "origin" {
			return true
		}
	}
	return false
}
