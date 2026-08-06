package sqltaint

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type Scanner struct {
	id string
}

func New() *Scanner {
	return &Scanner{id: "sqltaint"}
}

func (s *Scanner) ID() string {
	return s.id
}

func (s *Scanner) Describe() scanner.Descriptor {
	return scanner.Descriptor{
		Domain:          finding.Security,
		Version:         "1.0.0",
		Capabilities:    []string{"ast-analysis", "sql-taint", "dataflow-tracing"},
		SupportedModes:  []string{"full", "changed", "staged"},
		RequiresNetwork: false,
	}
}

func (s *Scanner) Scan(ctx context.Context, req scanner.Request) scanner.Result {
	start := time.Now()
	var findings []finding.Finding

	sources := req.Sources
	if len(sources) == 0 {
		sources = req.Files
	}

	for _, src := range sources {
		if ctx.Err() != nil {
			return scanner.Result{
				State:    finding.ScannerFailed,
				Duration: time.Since(start),
				Failure:  scanner.FailureCanceled,
			}
		}

		if !strings.HasSuffix(src.Path, ".go") || strings.HasSuffix(src.Path, "_test.go") {
			continue
		}

		r, err := src.Open(ctx)
		if err != nil {
			continue
		}
		code, err := io.ReadAll(r)
		_ = r.Close()
		if err != nil {
			continue
		}

		fset := token.NewFileSet()
		fileAST, err := parser.ParseFile(fset, src.Path, code, parser.ParseComments)
		if err != nil {
			continue
		}

		fileFindings := analyzeAST(fset, src.Path, fileAST)
		findings = append(findings, fileFindings...)
	}

	return scanner.Result{
		Findings: findings,
		State:    finding.ScannerFindings,
		Duration: time.Since(start),
	}
}

func analyzeAST(fset *token.FileSet, relPath string, node *ast.File) []finding.Finding {
	var findings []finding.Finding

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		funName, selectorName := extractCallNames(call)
		if funName == "" && selectorName == "" {
			return true
		}

		// Detect DB sinks: Query, Exec, QueryRow, Raw, Where
		if isDBSinkMethod(selectorName) {
			if len(call.Args) > 0 {
				arg0 := call.Args[0]
				if tpl := reconstructSQLTemplate(fset, relPath, arg0); tpl != nil && tpl.HasUntrustedHole() {
					pos := fset.Position(call.Pos())
					f := createSQLIFinding(relPath, pos.Line, selectorName, tpl)
					findings = append(findings, f)
				}
			}
		}

		// Detect SQLSAFE-001: Unbounded Update/Delete without WHERE
		if isUnboundedDeleteOrUpdate(selectorName, call) {
			pos := fset.Position(call.Pos())
			f := createUnboundedQueryFinding(relPath, pos.Line, selectorName)
			findings = append(findings, f)
		}

		return true
	})

	return findings
}

func isDBSinkMethod(name string) bool {
	switch name {
	case "Query", "QueryRow", "Exec", "ExecContext", "QueryContext", "QueryRowContext", "Raw", "ExecRaw":
		return true
	default:
		return false
	}
}

func isUnboundedDeleteOrUpdate(name string, call *ast.CallExpr) bool {
	if name == "Delete" || name == "Update" || name == "Exec" {
		if len(call.Args) > 0 {
			if strLit, ok := call.Args[0].(*ast.BasicLit); ok && strLit.Kind == token.STRING {
				upper := strings.ToUpper(strLit.Value)
				if (strings.Contains(upper, "DELETE FROM") || strings.Contains(upper, "UPDATE ")) && !strings.Contains(upper, "WHERE") {
					return true
				}
			}
		}
	}
	return false
}

func extractCallNames(call *ast.CallExpr) (funName string, selectorName string) {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name, ""
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok {
			return id.Name, fn.Sel.Name
		}
		return "", fn.Sel.Name
	}
	return "", ""
}

func reconstructSQLTemplate(fset *token.FileSet, relPath string, expr ast.Expr) *SQLTemplate {
	tpl := &SQLTemplate{
		Kind: KindRawConcatenation,
	}

	pos := fset.Position(expr.Pos())
	tpl.Location = finding.Location{File: relPath, Line: pos.Line}

	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			leftTpl := reconstructSQLTemplate(fset, relPath, e.X)
			rightTpl := reconstructSQLTemplate(fset, relPath, e.Y)
			if leftTpl != nil && rightTpl != nil {
				tpl.Segments = append(leftTpl.Segments, rightTpl.Segments...)
				tpl.RawText = leftTpl.RawText + " + " + rightTpl.RawText
				return tpl
			}
		}
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			tpl.Segments = []TemplateSegment{
				{IsConst: true, Text: e.Value},
			}
			tpl.RawText = e.Value
			return tpl
		}
	case *ast.CallExpr:
		// fmt.Sprintf("SELECT ... %s", var)
		if fun, sel := extractCallNames(e); fun == "fmt" && sel == "Sprintf" {
			if len(e.Args) > 0 {
				if fmtStr, ok := e.Args[0].(*ast.BasicLit); ok && fmtStr.Kind == token.STRING {
					tpl.RawText = fmtStr.Value
					tpl.Segments = append(tpl.Segments, TemplateSegment{IsConst: true, Text: fmtStr.Value})
					for _, arg := range e.Args[1:] {
						argPos := fset.Position(arg.Pos())
						hole := &Hole{
							Context:    HoleContextValue,
							Trust:      TrustUntrusted,
							Expression: fmt.Sprintf("%v", arg),
							SourceStep: &finding.DataflowStep{
								Type:        finding.StepSource,
								Location:    finding.Location{File: relPath, Line: argPos.Line},
								Label:       "Format argument",
								Explanation: "Dynamic variable formatted into SQL query template",
							},
						}
						tpl.Segments = append(tpl.Segments, TemplateSegment{IsConst: false, Hole: hole})
					}
					return tpl
				}
			}
		}
	case *ast.Ident:
		// Variable expression passed into query
		tpl.RawText = e.Name
		hole := &Hole{
			Context:    HoleContextValue,
			Trust:      TrustUntrusted,
			Expression: e.Name,
			SourceStep: &finding.DataflowStep{
				Type:        finding.StepSource,
				Location:    finding.Location{File: relPath, Line: pos.Line},
				Label:       e.Name,
				Explanation: "Dynamic string variable concatenated into SQL query execution",
			},
		}
		tpl.Segments = []TemplateSegment{{IsConst: false, Hole: hole}}
		return tpl
	}

	return nil
}

func createSQLIFinding(relPath string, line int, method string, tpl *SQLTemplate) finding.Finding {
	dataflow := []finding.DataflowStep{}
	for _, seg := range tpl.Segments {
		if seg.Hole != nil && seg.Hole.SourceStep != nil {
			dataflow = append(dataflow, *seg.Hole.SourceStep)
		}
	}
	dataflow = append(dataflow, finding.DataflowStep{
		Type:        finding.StepSink,
		Location:    finding.Location{File: relPath, Line: line},
		Label:       method + "(query)",
		Explanation: "Dynamic SQL template executed at database driver sink",
	})

	return finding.Finding{
		ID:             fmt.Sprintf("SQLI-001-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLI-001",
		Tool:           "sqltaint",
		Domain:         finding.Security,
		Category:       "sql-injection",
		Severity:       finding.High,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityLikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("Untrusted value concatenated or formatted into executable SQL at %s()", method),
		Recommendation: "Use parameterized queries ($1, ?, :name) instead of string concatenation or fmt.Sprintf",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/rules/SQLI-001.md",
		Location:       finding.Location{File: relPath, Line: line},
		Dataflow:       dataflow,
	}
}

func createUnboundedQueryFinding(relPath string, line int, method string) finding.Finding {
	return finding.Finding{
		ID:             fmt.Sprintf("SQLSAFE-001-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLSAFE-001",
		Tool:           "sqltaint",
		Domain:         finding.Reliability,
		Category:       "destructive-query",
		Severity:       finding.High,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityLikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("Unbounded UPDATE or DELETE query without a WHERE clause at %s()", method),
		Recommendation: "Always specify a WHERE clause or explicit target filter to prevent accidental table-wide mutation",
		Location:       finding.Location{File: relPath, Line: line},
	}
}
