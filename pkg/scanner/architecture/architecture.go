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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type Layer struct {
	Name  string
	Paths []string
}

type Boundary struct {
	From string
	To   string
}

type Options struct {
	DetectCycles bool
}

type importEdge struct {
	from       string
	to         string
	file       string
	line       int
	importPath string
}

type Scanner struct {
	layers       []Layer
	forbidden    map[string]struct{}
	detectCycles bool
}

func New(layers []Layer, boundaries []Boundary, configured ...Options) *Scanner {
	var options Options
	if len(configured) > 0 {
		options = configured[0]
	}
	forbidden := make(map[string]struct{}, len(boundaries))
	for _, boundary := range boundaries {
		forbidden[boundary.From+"\x00"+boundary.To] = struct{}{}
	}
	return &Scanner{layers: append([]Layer(nil), layers...), forbidden: forbidden, detectCycles: options.DetectCycles}
}

func (s *Scanner) ID() string { return "architecture" }

func (s *Scanner) Describe() scanner.Descriptor {
	return scanner.Descriptor{Domain: finding.Governance, Capabilities: []string{"go-import-graph", "layer-boundaries", "import-cycles"}, SupportedModes: []string{"full", "changed", "staged"}}
}

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean}
	module := modulePath(ctx, request.RepositoryFiles)
	packages := make(map[string]struct{})
	var edges []importEdge
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
		if from == "" && !s.detectCycles {
			continue
		}
		packagePath := pathpkg.Dir(relative)
		packages[packagePath] = struct{}{}
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
			targetPackage := pathpkg.Clean(strings.TrimPrefix(importPath, module+"/"))
			line := fileSet.Position(imported.Pos()).Line
			if s.detectCycles {
				edges = append(edges, importEdge{from: packagePath, to: targetPackage, file: relative, line: line, importPath: importPath})
			}
			targetPath := targetPackage + "/placeholder.go"
			to := s.layerFor(targetPath)
			if _, denied := s.forbidden[from+"\x00"+to]; !denied {
				continue
			}
			result.Findings = append(result.Findings, finding.Finding{
				RuleID: "architecture/forbidden-dependency", Tool: s.ID(), Domain: finding.Governance,
				Category: "architecture_boundary", Severity: finding.High,
				Description:    fmt.Sprintf("Layer %s must not depend on layer %s", from, to),
				Recommendation: "Move the dependency behind an allowed interface or update the reviewed architecture policy",
				Location:       finding.Location{File: relative, Line: line},
				Metadata:       map[string]string{"from_layer": from, "to_layer": to, "import": importPath},
			})
		}
	}
	if s.detectCycles {
		result.Findings = append(result.Findings, s.cycleFindings(edges, packages)...)
	}
	if result.State == finding.ScannerClean && len(result.Findings) > 0 {
		result.State = finding.ScannerFindings
	}
	result.Duration = time.Since(started)
	return result
}

func (s *Scanner) cycleFindings(edges []importEdge, packages map[string]struct{}) []finding.Finding {
	graph := make(map[string][]importEdge)
	for _, edge := range edges {
		if _, local := packages[edge.to]; local {
			graph[edge.from] = append(graph[edge.from], edge)
		}
	}
	for node := range graph {
		sort.Slice(graph[node], func(i, j int) bool {
			if graph[node][i].to != graph[node][j].to {
				return graph[node][i].to < graph[node][j].to
			}
			return graph[node][i].file < graph[node][j].file
		})
	}
	nodes := make([]string, 0, len(packages))
	for node := range packages {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	state := make(map[string]uint8)
	stack := make([]string, 0, len(nodes))
	stackIndex := make(map[string]int)
	seen := make(map[string]struct{})
	var findings []finding.Finding
	var visit func(string)
	visit = func(node string) {
		state[node] = 1
		stackIndex[node] = len(stack)
		stack = append(stack, node)
		for _, edge := range graph[node] {
			switch state[edge.to] {
			case 0:
				visit(edge.to)
			case 1:
				start := stackIndex[edge.to]
				cycle := append([]string(nil), stack[start:]...)
				identity := append([]string(nil), cycle...)
				sort.Strings(identity)
				key := strings.Join(identity, "\x00")
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				path := append(cycle, edge.to)
				findings = append(findings, finding.Finding{
					RuleID: "architecture/import-cycle", Tool: s.ID(), Domain: finding.Governance,
					Category: "architecture_cycle", Severity: finding.High,
					Description:    "Go package import cycle detected: " + strings.Join(path, " -> "),
					Recommendation: "Break the cycle by extracting shared contracts or reversing the dependency through an interface",
					Location:       finding.Location{File: edge.file, Line: edge.line},
					Metadata:       map[string]string{"dependency_path": strings.Join(path, " -> "), "import": edge.importPath},
				})
			}
		}
		stack = stack[:len(stack)-1]
		delete(stackIndex, node)
		state[node] = 2
	}
	for _, node := range nodes {
		if state[node] == 0 {
			visit(node)
		}
	}
	return findings
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
