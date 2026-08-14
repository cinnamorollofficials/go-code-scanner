package sqltaint

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"regexp"
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

		// 1. Detect ORM escape hatches (SQLI-004) e.g. db.Where(), db.Raw(), db.Order(), db.Having()
		if isORMSinkMethod(selectorName) {
			if len(call.Args) > 0 {
				arg0 := call.Args[0]
				if tpl := reconstructSQLTemplate(fset, relPath, arg0); tpl != nil && tpl.HasUntrustedHole() {
					pos := fset.Position(call.Pos())
					f := createORMIFinding(relPath, pos.Line, selectorName, tpl)
					findings = append(findings, f)
					return true
				}
			}
		}

		// 2. Detect Standard DB Sinks (SQLI-001, SQLI-002, SQLI-008)
		if isDBSinkMethod(selectorName) {
			if len(call.Args) > 0 {
				arg0 := call.Args[0]

				// Check placeholder mismatch (SQLI-008) on constant query strings
				if strLit, ok := arg0.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
					expectedPlaceholders := countPlaceholders(strLit.Value)
					actualArgs := len(call.Args) - 1
					if expectedPlaceholders > 0 && expectedPlaceholders != actualArgs {
						pos := fset.Position(call.Pos())
						f := createBindMismatchFinding(relPath, pos.Line, selectorName, expectedPlaceholders, actualArgs)
						findings = append(findings, f)
					}
				}

				// Check SQL injection (SQLI-001 or SQLI-002)
				if tpl := reconstructSQLTemplate(fset, relPath, arg0); tpl != nil && tpl.HasUntrustedHole() {
					pos := fset.Position(call.Pos())
					if hasIdentifierHole(tpl) {
						f := createIdentifierSQLIFinding(relPath, pos.Line, selectorName, tpl)
						findings = append(findings, f)
					} else {
						f := createSQLIFinding(relPath, pos.Line, selectorName, tpl)
						findings = append(findings, f)
					}
				}
			}
		}

		// 3. Detect SQLSAFE-001: Unbounded Update/Delete without WHERE
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
	case "Query", "QueryRow", "Exec", "ExecContext", "QueryContext", "QueryRowContext", "Select", "Get", "NamedExec", "NamedQuery":
		return true
	default:
		return false
	}
}

func isORMSinkMethod(name string) bool {
	switch name {
	case "Raw", "Where", "Order", "Having", "Group", "Not", "Or":
		return true
	default:
		return false
	}
}

func isUnboundedDeleteOrUpdate(name string, call *ast.CallExpr) bool {
	if name == "Delete" || name == "Update" || name == "Exec" || name == "ExecContext" {
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
					upperFmt := strings.ToUpper(fmtStr.Value)
					isIdentifier := strings.Contains(upperFmt, "FROM %S") ||
						strings.Contains(upperFmt, "INTO %S") ||
						strings.Contains(upperFmt, "TABLE %S") ||
						strings.Contains(upperFmt, "UPDATE %S") ||
						strings.Contains(upperFmt, "JOIN %S") ||
						strings.Contains(upperFmt, "ORDER BY %S")

					for _, arg := range e.Args[1:] {
						argPos := fset.Position(arg.Pos())
						ctxKind := HoleContextValue
						if isIdentifier {
							ctxKind = HoleContextIdentifier
						}
						hole := &Hole{
							Context:    ctxKind,
							Trust:      TrustUntrusted,
							Expression: fmt.Sprintf("%v", arg),
							SourceStep: &finding.DataflowStep{
								Type:        finding.StepSource,
								Location:    finding.Location{File: relPath, Line: argPos.Line},
								Label:       fmt.Sprintf("%v", arg),
								Explanation: "Dynamic argument formatted into SQL query template",
							},
						}
						tpl.Segments = append(tpl.Segments, TemplateSegment{IsConst: false, Hole: hole})
					}
					return tpl
				}
			}
		}
	case *ast.Ident:
		// Variable expression passed directly into query
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

func hasIdentifierHole(tpl *SQLTemplate) bool {
	for _, seg := range tpl.Segments {
		if seg.Hole != nil && seg.Hole.Context == HoleContextIdentifier {
			return true
		}
	}
	return false
}

var postgresParamRegex = regexp.MustCompile(`\$[0-9]+`)

func countPlaceholders(query string) int {
	qMarks := strings.Count(query, "?")
	if qMarks > 0 {
		return qMarks
	}
	matches := postgresParamRegex.FindAllString(query, -1)
	return len(matches)
}

func createSQLIFinding(relPath string, line int, method string, tpl *SQLTemplate) finding.Finding {
	dataflow := buildDataflow(tpl, relPath, line, method)
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
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-001",
		Location:       finding.Location{File: relPath, Line: line},
		Dataflow:       dataflow,
	}
}

func createIdentifierSQLIFinding(relPath string, line int, method string, tpl *SQLTemplate) finding.Finding {
	dataflow := buildDataflow(tpl, relPath, line, method)
	return finding.Finding{
		ID:             fmt.Sprintf("SQLI-002-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLI-002",
		Tool:           "sqltaint",
		Domain:         finding.Security,
		Category:       "sql-injection",
		Severity:       finding.High,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityLikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("Untrusted table, column, or identifier dynamically interpolated into SQL at %s()", method),
		Recommendation: "Validate SQL identifiers against an explicit allow-list of known safe column/table names before interpolation",
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-002",
		Location:       finding.Location{File: relPath, Line: line},
		Dataflow:       dataflow,
	}
}

func createORMIFinding(relPath string, line int, method string, tpl *SQLTemplate) finding.Finding {
	dataflow := buildDataflow(tpl, relPath, line, method)
	return finding.Finding{
		ID:             fmt.Sprintf("SQLI-004-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLI-004",
		Tool:           "sqltaint",
		Domain:         finding.Security,
		Category:       "orm-escape-hatch",
		Severity:       finding.High,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityLikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("Unsafe raw ORM escape hatch called with dynamic or concatenated string at %s()", method),
		Recommendation: "Pass parameters as separate arguments to ORM clauses (e.g. db.Where(\"name = ?\", val)) rather than dynamic string formatting",
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-004",
		Location:       finding.Location{File: relPath, Line: line},
		Dataflow:       dataflow,
	}
}

func createBindMismatchFinding(relPath string, line int, method string, expected int, actual int) finding.Finding {
	return finding.Finding{
		ID:             fmt.Sprintf("SQLI-008-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLI-008",
		Tool:           "sqltaint",
		Domain:         finding.Security,
		Category:       "bind-mismatch",
		Severity:       finding.Medium,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityUnlikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("SQL placeholder count mismatch at %s(): query specifies %d placeholders but %d parameters were passed", method, expected, actual),
		Recommendation: "Ensure the number of bind placeholders ($1, ?) matches the count of passed query arguments exactly",
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-008",
		Location:       finding.Location{File: relPath, Line: line},
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
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlsafe-001",
		Location:       finding.Location{File: relPath, Line: line},
	}
}

func buildDataflow(tpl *SQLTemplate, relPath string, line int, method string) []finding.DataflowStep {
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
		Explanation: "Dynamic SQL template executed at database driver/ORM sink",
	})
	return dataflow
}
