package frontend

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type CycleChecker struct {
	cfg      config.Config
	resolver *Resolver
}

func NewCycleChecker(cfg config.Config) *CycleChecker {
	return &CycleChecker{
		cfg:      cfg,
		resolver: NewResolver(cfg),
	}
}

func (c *CycleChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
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

	edges := ExtractImportEdges(cleanRelPath, tokens)
	var findings []finding.Finding

	for _, edge := range edges {
		resolved, ok := c.resolver.Resolve(cleanRelPath, edge.ToSpecifier)
		if !ok || resolved == cleanRelPath {
			continue
		}

		path := []string{cleanRelPath, resolved}
		if cyclePath := c.detectCycleDFS(resolved, cleanRelPath, path, 8); len(cyclePath) > 0 {
			if isCanonicalCycleHead(cleanRelPath, cyclePath) {
				rule, _ := LookupRule("frontend/import-cycle")
				cycleStr := strings.Join(cyclePath, " -> ")
				findings = append(findings, finding.Finding{
					RuleID:         rule.ID,
					Domain:         rule.Domain,
					Category:       rule.Category,
					Severity:       rule.Severity,
					Description:    "Circular import dependency: " + cycleStr,
					Recommendation: rule.Recommendation,
					Documentation:  rule.Documentation,
					Location:       finding.Location{File: cleanRelPath, Line: edge.Line},
					Metadata:       map[string]string{"cycle": cycleStr},
					Tags:           rule.Tags,
				})
			}
		}
	}

	return findings, nil
}

func (c *CycleChecker) detectCycleDFS(current, target string, path []string, maxDepth int) []string {
	if maxDepth <= 0 {
		return nil
	}

	fullPath := filepath.Join(c.cfg.Root, current)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil
	}

	tokens, err := Tokenize(data)
	if err != nil {
		return nil
	}

	edges := ExtractImportEdges(current, tokens)
	for _, edge := range edges {
		resolved, ok := c.resolver.Resolve(current, edge.ToSpecifier)
		if !ok {
			continue
		}

		if resolved == target {
			return append(path, target)
		}

		if containsString(path, resolved) {
			continue
		}

		if res := c.detectCycleDFS(resolved, target, append(path, resolved), maxDepth-1); len(res) > 0 {
			return res
		}
	}

	return nil
}

func isCanonicalCycleHead(head string, cyclePath []string) bool {
	nodes := cyclePath[:len(cyclePath)-1]
	for _, node := range nodes {
		if node < head {
			return false
		}
	}
	return true
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
