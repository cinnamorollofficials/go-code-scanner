package frontend

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type BoundaryChecker struct {
	cfg        config.Config
	classifier *Classifier
	resolver   *Resolver
}

func NewBoundaryChecker(cfg config.Config) *BoundaryChecker {
	return &BoundaryChecker{
		cfg:        cfg,
		classifier: NewClassifier(cfg),
		resolver:   NewResolver(cfg),
	}
}

func (c *BoundaryChecker) Check(ctx context.Context, src scanner.Source) ([]finding.Finding, error) {
	if src.Open == nil {
		return nil, nil
	}

	scope := c.classifier.Classify(ctx, src)
	if scope != ScopeClient {
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
		if !ok {
			continue
		}

		targetScope := c.classifier.classifyPathOnly(resolved)
		if targetScope == ScopeServer {
			rule, _ := LookupRule("frontend/client-server-boundary-violation")
			findings = append(findings, finding.Finding{
				RuleID:         rule.ID,
				Domain:         rule.Domain,
				Category:       rule.Category,
				Severity:       rule.Severity,
				Description:    "Client code imports server-only module or server root: " + resolved,
				Recommendation: rule.Recommendation,
				Documentation:  rule.Documentation,
				Location:       finding.Location{File: cleanRelPath, Line: edge.Line},
				Metadata:       map[string]string{"dependency_target": resolved},
				Tags:           rule.Tags,
			})
		}
	}

	return findings, nil
}

func (c *Classifier) classifyPathOnly(cleanPath string) Scope {
	if scope := c.classifyByConfiguredRoots(cleanPath); scope != ScopeUnknown {
		return scope
	}
	if scope := c.classifyByFileConventions(cleanPath); scope != ScopeUnknown {
		return scope
	}
	if fullPath := filepath.Join(c.cfg.Root, cleanPath); fileExists(fullPath) {
		if data, err := os.ReadFile(fullPath); err == nil {
			if strings.Contains(string(data), `"use server"`) || strings.Contains(string(data), `'use server'`) {
				return ScopeServer
			}
		}
	}
	return ScopeUnknown
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}
