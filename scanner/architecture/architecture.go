package architecture

import (
	"bufio"
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type Layer struct {
	Name  string
	Paths []string
}

type Boundary struct {
	From string
	To   string
}

type Scanner struct {
	layers    []Layer
	forbidden map[string]struct{}
}

func New(layers []Layer, boundaries []Boundary) *Scanner {
	forbidden := make(map[string]struct{}, len(boundaries))
	for _, boundary := range boundaries {
		forbidden[boundary.From+"\x00"+boundary.To] = struct{}{}
	}
	return &Scanner{layers: append([]Layer(nil), layers...), forbidden: forbidden}
}

func (s *Scanner) ID() string { return "architecture" }

func (s *Scanner) Describe() scanner.Descriptor {
	return scanner.Descriptor{Domain: finding.Governance, Capabilities: []string{"go-import-graph", "layer-boundaries"}, SupportedModes: []string{"full", "changed", "staged"}}
}

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean}
	module := modulePath(ctx, request.RepositoryFiles)
	for _, source := range request.RepositoryFiles {
		if err := ctx.Err(); err != nil {
			result.State, result.Failure, result.Message = finding.ScannerFailed, scanner.FailureCanceled, err.Error()
			result.Duration = time.Since(started)
			return result
		}
		if strings.ToLower(filepath.Ext(source.Path)) != ".go" {
			continue
		}
		relative, err := filepath.Rel(request.Root, source.Path)
		if err != nil {
			continue
		}
		relative = filepath.ToSlash(relative)
		from := s.layerFor(relative)
		if from == "" {
			continue
		}
		reader, err := source.Open(ctx)
		if err != nil {
			result.State, result.Failure = finding.ScannerPartial, scanner.FailurePartial
			result.Message = appendArchitectureMessage(result.Message, fmt.Sprintf("read %s: %v", relative, err))
			continue
		}
		fileSet := token.NewFileSet()
		parsed, parseErr := parser.ParseFile(fileSet, relative, reader, parser.ImportsOnly)
		_ = reader.Close()
		if parseErr != nil {
			result.State, result.Failure = finding.ScannerPartial, scanner.FailurePartial
			result.Message = appendArchitectureMessage(result.Message, fmt.Sprintf("parse %s: %v", relative, parseErr))
			continue
		}
		for _, imported := range parsed.Imports {
			importPath, unquoteErr := strconv.Unquote(imported.Path.Value)
			if unquoteErr != nil || module == "" || !strings.HasPrefix(importPath, module+"/") {
				continue
			}
			targetPath := strings.TrimPrefix(importPath, module+"/") + "/placeholder.go"
			to := s.layerFor(targetPath)
			if _, denied := s.forbidden[from+"\x00"+to]; !denied {
				continue
			}
			result.Findings = append(result.Findings, finding.Finding{
				RuleID: "architecture/forbidden-dependency", Tool: s.ID(), Domain: finding.Governance,
				Category: "architecture_boundary", Severity: finding.High,
				Description:    fmt.Sprintf("Layer %s must not depend on layer %s", from, to),
				Recommendation: "Move the dependency behind an allowed interface or update the reviewed architecture policy",
				Location:       finding.Location{File: relative, Line: fileSet.Position(imported.Pos()).Line},
				Metadata:       map[string]string{"from_layer": from, "to_layer": to, "import": importPath},
			})
		}
	}
	if result.State == finding.ScannerClean && len(result.Findings) > 0 {
		result.State = finding.ScannerFindings
	}
	result.Duration = time.Since(started)
	return result
}

func (s *Scanner) layerFor(path string) string {
	path = filepath.ToSlash(path)
	for _, layer := range s.layers {
		for _, pattern := range layer.Paths {
			if matched, _ := pathpkg.Match(pattern, path); matched {
				return layer.Name
			}
		}
	}
	return ""
}

func modulePath(ctx context.Context, files []scanner.Source) string {
	for _, source := range files {
		if filepath.Base(source.Path) != "go.mod" {
			continue
		}
		reader, err := source.Open(ctx)
		if err != nil {
			return ""
		}
		scan := bufio.NewScanner(io.LimitReader(reader, 64*1024))
		for scan.Scan() {
			fields := strings.Fields(scan.Text())
			if len(fields) == 2 && fields[0] == "module" {
				_ = reader.Close()
				return fields[1]
			}
		}
		_ = reader.Close()
	}
	return ""
}

func appendArchitectureMessage(current, addition string) string {
	if current == "" {
		return addition
	}
	return current + "; " + addition
}
