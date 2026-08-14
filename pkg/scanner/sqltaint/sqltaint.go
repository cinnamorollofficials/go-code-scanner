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

		fileFindings := analyzeFile(fset, src.Path, fileAST)
		findings = append(findings, fileFindings...)
	}

	return scanner.Result{
		Findings: findings,
		State:    finding.ScannerFindings,
		Duration: time.Since(start),
	}
}

func analyzeFile(fset *token.FileSet, relPath string, node *ast.File) []finding.Finding {
	var findings []finding.Finding

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// 1. Build intraprocedural variable assignment map
		vMap := make(map[string]ast.Expr)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if assign, ok := n.(*ast.AssignStmt); ok {
				for i, lhs := range assign.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && i < len(assign.Rhs) {
						vMap[id.Name] = assign.Rhs[i]
					}
				}
			}
			return true
		})

		// 2. Inspect AST nodes within the function
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			funName, selectorName := extractCallNames(call)
			if funName == "" && selectorName == "" {
				return true
			}

			// A. Detect Prepared Statement Sinks (SQLI-012: Tainted Prepared Query Template)
			if isPrepareMethod(selectorName) {
				if len(call.Args) > 0 {
					arg0 := call.Args[0]
					if tpl := reconstructSQLTemplate(fset, relPath, arg0, vMap, 0); tpl != nil && tpl.HasUntrustedHole() {
						pos := fset.Position(call.Pos())
						f := createTaintedPrepareFinding(relPath, pos.Line, selectorName, tpl)
						findings = append(findings, f)
						return true
					}
				}
			}

			// B. Detect ORM escape hatches (SQLI-004) e.g. db.Where(), db.Raw(), db.Order(), db.Having()
			if isORMSinkMethod(selectorName) {
				if len(call.Args) > 0 {
					arg0 := call.Args[0]
					if tpl := reconstructSQLTemplate(fset, relPath, arg0, vMap, 0); tpl != nil && tpl.HasUntrustedHole() {
						pos := fset.Position(call.Pos())
						if hasListExpansionHole(tpl) {
							f := createListExpansionFinding(relPath, pos.Line, selectorName, tpl)
							findings = append(findings, f)
						} else {
							f := createORMIFinding(relPath, pos.Line, selectorName, tpl)
							findings = append(findings, f)
						}
						return true
					}
				}
			}

			// C. Detect Standard DB Sinks (SQLI-001, SQLI-002, SQLI-008, SQLI-011)
			if isDBSinkMethod(selectorName) {
				if len(call.Args) > 0 {
					arg0 := call.Args[0]

					// Check placeholder mismatch (SQLI-008) on constant query strings
					resolvedArg := resolveExpr(arg0, vMap)
					if strLit, ok := resolvedArg.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
						expectedPlaceholders := countPlaceholders(strLit.Value)
						actualArgs := len(call.Args) - 1
						if expectedPlaceholders > 0 && expectedPlaceholders != actualArgs {
							pos := fset.Position(call.Pos())
							f := createBindMismatchFinding(relPath, pos.Line, selectorName, expectedPlaceholders, actualArgs)
							findings = append(findings, f)
						}
					}

					// Check SQL injection (SQLI-001, SQLI-002, or SQLI-011)
					if tpl := reconstructSQLTemplate(fset, relPath, arg0, vMap, 0); tpl != nil && tpl.HasUntrustedHole() {
						pos := fset.Position(call.Pos())
						if hasListExpansionHole(tpl) {
							f := createListExpansionFinding(relPath, pos.Line, selectorName, tpl)
							findings = append(findings, f)
						} else if hasIdentifierHole(tpl) {
							f := createIdentifierSQLIFinding(relPath, pos.Line, selectorName, tpl)
							findings = append(findings, f)
						} else {
							f := createSQLIFinding(relPath, pos.Line, selectorName, tpl)
							findings = append(findings, f)
						}
					}
				}
			}

			// D. Detect SQLSAFE-001: Unbounded Update/Delete without WHERE
			if isUnboundedDeleteOrUpdate(selectorName, call, vMap) {
				pos := fset.Position(call.Pos())
				f := createUnboundedQueryFinding(relPath, pos.Line, selectorName)
				findings = append(findings, f)
			}

			return true
		})
	}

	return findings
}

func resolveExpr(expr ast.Expr, vMap map[string]ast.Expr) ast.Expr {
	if id, ok := expr.(*ast.Ident); ok {
		if target, found := vMap[id.Name]; found {
			return target
		}
	}
	return expr
}

func isPrepareMethod(name string) bool {
	return name == "Prepare" || name == "PrepareContext"
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

func isUnboundedDeleteOrUpdate(name string, call *ast.CallExpr, vMap map[string]ast.Expr) bool {
	if name == "Delete" || name == "Update" || name == "Exec" || name == "ExecContext" {
		if len(call.Args) > 0 {
			resolved := resolveExpr(call.Args[0], vMap)
			if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
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

func reconstructSQLTemplate(fset *token.FileSet, relPath string, expr ast.Expr, vMap map[string]ast.Expr, depth int) *SQLTemplate {
	if depth > 8 {
		return nil
	}

	tpl := &SQLTemplate{
		Kind: KindRawConcatenation,
	}

	pos := fset.Position(expr.Pos())
	tpl.Location = finding.Location{File: relPath, Line: pos.Line}

	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			leftTpl := reconstructSQLTemplate(fset, relPath, e.X, vMap, depth+1)
			rightTpl := reconstructSQLTemplate(fset, relPath, e.Y, vMap, depth+1)
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
		fun, sel := extractCallNames(e)

		// Check strings.Join(slice, ",") inside query
		if (fun == "strings" && sel == "Join") || sel == "Join" {
			argPos := fset.Position(e.Pos())
			hole := &Hole{
				Context:    HoleContextListExpansion,
				Trust:      TrustUntrusted,
				Expression: "strings.Join",
				SourceStep: &finding.DataflowStep{
					Type:        finding.StepSource,
					Location:    finding.Location{File: relPath, Line: argPos.Line},
					Label:       "strings.Join(...) expansion",
					Explanation: "Slice joined as comma-separated values into dynamic SQL IN clause",
				},
			}
			tpl.Segments = []TemplateSegment{{IsConst: false, Hole: hole}}
			tpl.RawText = "strings.Join(...)"
			return tpl
		}

		// Check fmt.Sprintf("SELECT ... %s", var)
		if fun == "fmt" && sel == "Sprintf" {
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
					isListExp := strings.Contains(upperFmt, "IN (%S)") || strings.Contains(upperFmt, "IN(%S)")

					for _, arg := range e.Args[1:] {
						argPos := fset.Position(arg.Pos())
						ctxKind := HoleContextValue
						if isListExp {
							ctxKind = HoleContextListExpansion
						} else if isIdentifier {
							ctxKind = HoleContextIdentifier
						}

						sourceLabel, sourceExpl := detectSourceOrigin(arg, "Dynamic argument formatted into SQL query template")

						hole := &Hole{
							Context:    ctxKind,
							Trust:      TrustUntrusted,
							Expression: fmt.Sprintf("%v", arg),
							SourceStep: &finding.DataflowStep{
								Type:        finding.StepSource,
								Location:    finding.Location{File: relPath, Line: argPos.Line},
								Label:       sourceLabel,
								Explanation: sourceExpl,
							},
						}
						tpl.Segments = append(tpl.Segments, TemplateSegment{IsConst: false, Hole: hole})
					}
					return tpl
				}
			}
		}
	case *ast.Ident:
		// Check if local variable resolves to an RHS expression
		if rhs, found := vMap[e.Name]; found {
			if subTpl := reconstructSQLTemplate(fset, relPath, rhs, vMap, depth+1); subTpl != nil {
				return subTpl
			}
		}

		// Fallback: variable passed as parameter
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

func detectSourceOrigin(expr ast.Expr, fallbackExplanation string) (label string, explanation string) {
	if call, ok := expr.(*ast.CallExpr); ok {
		fun, sel := extractCallNames(call)
		if sel == "Get" || sel == "Query" || sel == "FormValue" || sel == "Param" || sel == "URLParam" {
			return fmt.Sprintf("HTTP Parameter (%s.%s)", fun, sel), "Untrusted HTTP request parameter flowing into SQL sink"
		}
		if (fun == "strings" && sel == "Join") || sel == "Join" {
			return "strings.Join", "Slice elements joined as raw SQL fragment"
		}
	}
	return fmt.Sprintf("%v", expr), fallbackExplanation
}

func hasIdentifierHole(tpl *SQLTemplate) bool {
	for _, seg := range tpl.Segments {
		if seg.Hole != nil && seg.Hole.Context == HoleContextIdentifier {
			return true
		}
	}
	return false
}

func hasListExpansionHole(tpl *SQLTemplate) bool {
	for _, seg := range tpl.Segments {
		if seg.Hole != nil && seg.Hole.Context == HoleContextListExpansion {
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

func createListExpansionFinding(relPath string, line int, method string, tpl *SQLTemplate) finding.Finding {
	dataflow := buildDataflow(tpl, relPath, line, method)
	return finding.Finding{
		ID:             fmt.Sprintf("SQLI-011-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLI-011",
		Tool:           "sqltaint",
		Domain:         finding.Security,
		Category:       "list-expansion",
		Severity:       finding.High,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityLikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("Unsafe list or IN clause expansion using strings.Join or manual string interpolation at %s()", method),
		Recommendation: "Use sqlx.In or generate parameterized bind variable lists (?, ?, ...) for slice queries",
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-011",
		Location:       finding.Location{File: relPath, Line: line},
		Dataflow:       dataflow,
	}
}

func createTaintedPrepareFinding(relPath string, line int, method string, tpl *SQLTemplate) finding.Finding {
	dataflow := buildDataflow(tpl, relPath, line, method)
	return finding.Finding{
		ID:             fmt.Sprintf("SQLI-012-%s-%d", filepath.Base(relPath), line),
		RuleID:         "SQLI-012",
		Tool:           "sqltaint",
		Domain:         finding.Security,
		Category:       "prepared-statement",
		Severity:       finding.High,
		Confidence:     finding.ConfidenceHigh,
		Exploitability: finding.ExploitabilityLikely,
		FindingState:   finding.FindingConfirmed,
		Description:    fmt.Sprintf("Tainted SQL query template passed into statement preparation method %s()", method),
		Recommendation: "Keep the SQL query string passed to db.Prepare strictly constant and bind dynamic values via stmt.Query / stmt.Exec",
		Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-012",
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
