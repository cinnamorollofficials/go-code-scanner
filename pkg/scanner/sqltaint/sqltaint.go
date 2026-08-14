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
		Version:         "2.0.0",
		Capabilities:    []string{"ast-analysis", "sql-taint", "dataflow-tracing", "interprocedural", "router-models"},
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

	pkgCtx := &PackageAnalysisContext{
		Files:     make(map[string]*ast.File),
		FileSets:  make(map[string]*token.FileSet),
		Functions: make(map[string]*FunctionSummary),
		CallEdges: make([]CallGraphEdge, 0),
	}

	// Pass 1: Parse all Go files into ASTs and register functions
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

		pkgCtx.Files[src.Path] = fileAST
		pkgCtx.FileSets[src.Path] = fset

		// Register functions and methods
		for _, decl := range fileAST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}

			summary := extractFunctionSummary(fset, src.Path, fn)
			key := functionKey(summary.ReceiverType, summary.Name)
			pkgCtx.Functions[key] = summary
		}
	}

	// Pass 2: Build Call Graph Edges across files
	for relPath, fileAST := range pkgCtx.Files {
		fset := pkgCtx.FileSets[relPath]
		for _, decl := range fileAST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			callerName := fn.Name.Name
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				funName, selName := extractCallNames(call)
				callee := selName
				if callee == "" {
					callee = funName
				}
				if callee != "" {
					pos := fset.Position(call.Pos())
					pkgCtx.CallEdges = append(pkgCtx.CallEdges, CallGraphEdge{
						CallerFile: relPath,
						CallerFunc: callerName,
						CalleeFunc: callee,
						CallExpr:   call,
						Line:       pos.Line,
					})
				}
				return true
			})
		}
	}

	// Pass 3: Intraprocedural and Interprocedural Security Analysis
	for relPath, fileAST := range pkgCtx.Files {
		fset := pkgCtx.FileSets[relPath]
		fileFindings := analyzeFileWithContext(fset, relPath, fileAST, pkgCtx)
		findings = append(findings, fileFindings...)
	}

	return scanner.Result{
		Findings: findings,
		State:    finding.ScannerFindings,
		Duration: time.Since(start),
	}
}

func functionKey(receiverType, name string) string {
	if receiverType != "" {
		return receiverType + "." + name
	}
	return name
}

func extractFunctionSummary(fset *token.FileSet, file string, fn *ast.FuncDecl) *FunctionSummary {
	var recvType string
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvType = exprToString(fn.Recv.List[0].Type)
		recvType = strings.TrimPrefix(recvType, "*")
	}

	params := make([]string, 0)
	isHTTPHandler := false
	routerType := ""

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			fieldType := exprToString(field.Type)
			if strings.Contains(fieldType, "gin.Context") {
				isHTTPHandler = true
				routerType = "gin"
			} else if strings.Contains(fieldType, "echo.Context") {
				isHTTPHandler = true
				routerType = "echo"
			} else if strings.Contains(fieldType, "fiber.Ctx") {
				isHTTPHandler = true
				routerType = "fiber"
			} else if strings.Contains(fieldType, "http.ResponseWriter") || strings.Contains(fieldType, "*http.Request") || strings.Contains(fieldType, "http.Request") {
				isHTTPHandler = true
				routerType = "net/http"
			}

			for _, name := range field.Names {
				params = append(params, name.Name)
			}
		}
	}

	pos := fset.Position(fn.Pos())
	return &FunctionSummary{
		Name:          fn.Name.Name,
		ReceiverType:  recvType,
		File:          file,
		Line:          pos.Line,
		Params:        params,
		Decl:          fn,
		Fset:          fset,
		IsHTTPHandler: isHTTPHandler,
		RouterType:    routerType,
	}
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	default:
		return fmt.Sprintf("%v", expr)
	}
}

func analyzeFileWithContext(fset *token.FileSet, relPath string, node *ast.File, pkgCtx *PackageAnalysisContext) []finding.Finding {
	var findings []finding.Finding

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		funcName := fn.Name.Name
		fnSummary := extractFunctionSummary(fset, relPath, fn)

		// 1. Build variable assignment map & identify tainted variables from HTTP router sources
		vMap := make(map[string]ast.Expr)
		taintedVars := make(map[string]*finding.DataflowStep)

		// If function is an HTTP handler or takes request params, mark params as tainted
		if fnSummary.IsHTTPHandler {
			for _, p := range fnSummary.Params {
				if p != "w" && p != "rw" && p != "res" {
					taintedVars[p] = &finding.DataflowStep{
						Type:        finding.StepSource,
						Location:    finding.Location{File: relPath, Line: fnSummary.Line},
						Label:       fmt.Sprintf("%s handler parameter: %s", fnSummary.RouterType, p),
						Explanation: "HTTP request entrypoint parameter",
					}
				}
			}
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if assign, ok := n.(*ast.AssignStmt); ok {
				for i, lhs := range assign.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && i < len(assign.Rhs) {
						rhs := assign.Rhs[i]
						vMap[id.Name] = rhs

						// Check if RHS is a router input extraction
						if step := extractRouterSourceStep(fset, relPath, rhs); step != nil {
							taintedVars[id.Name] = step
						}
					}
				}
			}
			return true
		})

		// 2. Interprocedural Call Graph Taint Propagation:
		// Check if caller invokes helper functions passing tainted variables
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			funName, selName := extractCallNames(call)
			targetName := selName
			if targetName == "" {
				targetName = funName
			}

			// Check if target function exists in package
			for _, targetSummary := range pkgCtx.Functions {
				if targetSummary.Name == targetName && targetSummary.Decl != nil && targetSummary.Decl.Body != nil {
					// Check if any argument is tainted
					for argIdx, arg := range call.Args {
						var sourceStep *finding.DataflowStep
						if id, ok := arg.(*ast.Ident); ok {
							if step, ok := taintedVars[id.Name]; ok {
								sourceStep = step
							}
						} else if step := extractRouterSourceStep(fset, relPath, arg); step != nil {
							sourceStep = step
						}

						if sourceStep != nil && argIdx < len(targetSummary.Params) {
							paramName := targetSummary.Params[argIdx]
							pos := fset.Position(call.Pos())

							// Propagate into callee body
							calleeFindings := traceInterproceduralSink(
								targetSummary,
								paramName,
								sourceStep,
								relPath,
								pos.Line,
								funcName,
								targetName,
							)
							findings = append(findings, calleeFindings...)
						}
					}
				}
			}
			return true
		})

		// 3. Inspect AST nodes for Intraprocedural Rules
		hasTxParam := false
		if fn.Type.Params != nil {
			for _, field := range fn.Type.Params.List {
				fieldType := exprToString(field.Type)
				if strings.Contains(fieldType, "sql.Tx") || (strings.Contains(fieldType, "gorm.DB") && strings.Contains(strings.ToLower(exprToString(field.Type)), "tx")) {
					hasTxParam = true
				}
			}
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			funName, selectorName := extractCallNames(call)
			if funName == "" && selectorName == "" {
				return true
			}

			// A. Detect Prepared Statement Sinks (SQLI-012)
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

			// D. SQLSAFE-001: Unbounded Update/Delete without WHERE
			if isUnboundedDeleteOrUpdate(selectorName, call, vMap) {
				pos := fset.Position(call.Pos())
				f := createUnboundedQueryFinding(relPath, pos.Line, selectorName)
				findings = append(findings, f)
			}

			// E. SQLAUTH-001: Missing Tenant Constraint in Multi-Tenant Query
			if f, ok := detectSQLAUTH001(fset, relPath, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// F. SQLAUTH-002: Insecure Direct Object Reference (IDOR) on Single Entity Lookup
			if f, ok := detectSQLAUTH002(fset, relPath, selectorName, call, vMap, taintedVars); ok {
				findings = append(findings, f)
			}

			// G. SQLAUTH-003: Auth Filter Dropped in Raw Query
			if f, ok := detectSQLAUTH003(fset, relPath, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// H. SQLAUTH-004: Row-Level Security Assumption Mismatch (Superuser / Bypass)
			if f, ok := detectSQLAUTH004(fset, relPath, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// I. SQLSAFE-003: Non-Atomic Read-Modify-Write on Balance/Inventory
			if f, ok := detectSQLSAFE003(fset, relPath, fn.Body, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// J. SQLSAFE-004: Transaction Boundary Loss (goroutine / unmanaged DB call)
			if f, ok := detectSQLSAFE004(fset, relPath, hasTxParam, selectorName, call); ok {
				findings = append(findings, f)
			}

			// K. SQLSAFE-005: Incorrect AND/OR Precedence in Query Builder / Raw SQL
			if f, ok := detectSQLSAFE005(fset, relPath, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// L. SQLSAFE-006: Soft-Delete Bypass in Raw Queries
			if f, ok := detectSQLSAFE006(fset, relPath, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// M. DBPERF-001: Unbounded Result Set
			if f, ok := detectDBPERF001(fset, relPath, selectorName, call, vMap); ok {
				findings = append(findings, f)
			}

			// N. DBSEC-003: Database Error Exposed to Untrusted Client
			if f, ok := detectDBSEC003(fset, relPath, funName, selectorName, call); ok {
				findings = append(findings, f)
			}

			return true
		})

		// 4. Detect Loop Queries (DBPERF-002: N+1 Query in Loops)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch loop := n.(type) {
			case *ast.RangeStmt:
				loopFindings := findQueriesInLoop(fset, relPath, loop.Body)
				findings = append(findings, loopFindings...)
			case *ast.ForStmt:
				loopFindings := findQueriesInLoop(fset, relPath, loop.Body)
				findings = append(findings, loopFindings...)
			}
			return true
		})
	}

	return findings
}

func extractRouterSourceStep(fset *token.FileSet, relPath string, expr ast.Expr) *finding.DataflowStep {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}

	funName, selName := extractCallNames(call)
	pos := fset.Position(call.Pos())

	// Gin / Echo / Fiber router getters
	switch selName {
	case "Param", "Query", "QueryParam", "PostForm", "FormValue", "Params", "Body", "GetString", "GetHeader", "Header":
		return &finding.DataflowStep{
			Type:        finding.StepSource,
			Location:    finding.Location{File: relPath, Line: pos.Line},
			Label:       fmt.Sprintf("HTTP parameter extraction: %s.%s()", funName, selName),
			Explanation: "Untrusted HTTP request parameter read from route/query/body",
		}
	}

	// Chi router: chi.URLParam(r, "id")
	if funName == "chi" && selName == "URLParam" {
		return &finding.DataflowStep{
			Type:        finding.StepSource,
			Location:    finding.Location{File: relPath, Line: pos.Line},
			Label:       "chi.URLParam(r, ...)",
			Explanation: "Untrusted HTTP URL route parameter from chi router",
		}
	}

	// Gorilla Mux: mux.Vars(r)["id"]
	if funName == "mux" && selName == "Vars" {
		return &finding.DataflowStep{
			Type:        finding.StepSource,
			Location:    finding.Location{File: relPath, Line: pos.Line},
			Label:       "mux.Vars(r)",
			Explanation: "Untrusted HTTP URL route parameter from gorilla/mux",
		}
	}

	// net/http: r.URL.Query().Get(...)
	if selName == "Get" {
		return &finding.DataflowStep{
			Type:        finding.StepSource,
			Location:    finding.Location{File: relPath, Line: pos.Line},
			Label:       "r.URL.Query().Get(...)",
			Explanation: "Untrusted query parameter from net/http Request",
		}
	}

	return nil
}

func traceInterproceduralSink(
	callee *FunctionSummary,
	paramName string,
	sourceStep *finding.DataflowStep,
	callerFile string,
	callerLine int,
	callerName string,
	calleeName string,
) []finding.Finding {
	var findings []finding.Finding
	if callee.Decl == nil || callee.Decl.Body == nil {
		return findings
	}

	fset := callee.Fset
	if fset == nil {
		fset = token.NewFileSet()
	}

	// Build callee vMap
	vMap := make(map[string]ast.Expr)
	ast.Inspect(callee.Decl.Body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for i, lhs := range assign.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && i < len(assign.Rhs) {
					vMap[id.Name] = assign.Rhs[i]
				}
			}
		}
		return true
	})

	// Inspect callee body for query sinks using paramName
	ast.Inspect(callee.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		_, selectorName := extractCallNames(call)
		if isDBSinkMethod(selectorName) || isORMSinkMethod(selectorName) {
			if len(call.Args) > 0 {
				arg0 := call.Args[0]
				tpl := reconstructSQLTemplate(fset, callee.File, arg0, vMap, 0)
				if tpl != nil {
					// Check if tpl uses paramName
					usesParam := false
					for _, seg := range tpl.Segments {
						if seg.Hole != nil && (seg.Hole.Expression == paramName || strings.Contains(tpl.RawText, paramName)) {
							usesParam = true
							break
						}
					}

					if usesParam {
						pos := fset.Position(call.Pos())
						steps := []finding.DataflowStep{
							*sourceStep,
							{
								Type:        finding.StepPropagator,
								Location:    finding.Location{File: callerFile, Line: callerLine},
								Label:       fmt.Sprintf("%s() -> %s(%s)", callerName, calleeName, paramName),
								Explanation: "Tainted HTTP parameter passed across function boundary",
							},
							{
								Type:        finding.StepPropagator,
								Location:    finding.Location{File: callee.File, Line: callee.Line},
								Label:       fmt.Sprintf("func %s(%s ...)", calleeName, paramName),
								Explanation: "Callee receives tainted parameter",
							},
							{
								Type:        finding.StepSink,
								Location:    finding.Location{File: callee.File, Line: pos.Line},
								Label:       selectorName + "(query)",
								Explanation: "Dynamic query executed with interprocedural untrusted data",
							},
						}

						findings = append(findings, finding.Finding{
							ID:             fmt.Sprintf("SQLI-001-%s-%d", filepath.Base(callee.File), pos.Line),
							RuleID:         "SQLI-001",
							Tool:           "sqltaint",
							Domain:         finding.Security,
							Category:       "sql-injection",
							Severity:       finding.High,
							Confidence:     finding.ConfidenceHigh,
							Exploitability: finding.ExploitabilityLikely,
							FindingState:   finding.FindingConfirmed,
							Description:    fmt.Sprintf("Untrusted value from %s() propagated across calls reaches %s() in %s", callerName, selectorName, calleeName),
							Recommendation: "Use parameterized queries ($1, ?, :name) instead of concatenating interprocedural arguments",
							Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqli-001",
							Location:       finding.Location{File: callee.File, Line: pos.Line},
							Dataflow:       steps,
						})
					}
				}
			}
		}
		return true
	})

	return findings
}

// Multi-tenant tables pattern
var multiTenantTableRegex = regexp.MustCompile(`(?i)\b(accounts|tenants|organizations|workspaces|orders|projects|teams|members|invoices|customers|subscriptions|departments)\b`)

func detectSQLAUTH001(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if !isDBSinkMethod(method) && !isORMSinkMethod(method) {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		if multiTenantTableRegex.MatchString(strLit.Value) {
			hasTenantFilter := strings.Contains(upper, "TENANT_ID") ||
				strings.Contains(upper, "ORG_ID") ||
				strings.Contains(upper, "ACCOUNT_ID") ||
				strings.Contains(upper, "WORKSPACE_ID") ||
				strings.Contains(upper, "COMPANY_ID")

			if !hasTenantFilter && (strings.Contains(upper, "SELECT") || strings.Contains(upper, "UPDATE") || strings.Contains(upper, "DELETE")) {
				pos := fset.Position(call.Pos())
				return finding.Finding{
					ID:             fmt.Sprintf("SQLAUTH-001-%s-%d", filepath.Base(relPath), pos.Line),
					RuleID:         "SQLAUTH-001",
					Tool:           "sqltaint",
					Domain:         finding.Security,
					Category:       "multi-tenant-isolation",
					Severity:       finding.High,
					Confidence:     finding.ConfidenceHigh,
					Exploitability: finding.ExploitabilityLikely,
					FindingState:   finding.FindingConfirmed,
					Description:    fmt.Sprintf("Multi-tenant entity queried at %s() without tenant_id/org_id scoping constraint", method),
					Recommendation: "Enforce explicit tenant_id or organization_id filtering on all multi-tenant queries to prevent cross-tenant data access",
					Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlauth-001",
					Location:       finding.Location{File: relPath, Line: pos.Line},
				}, true
			}
		}
	}
	return finding.Finding{}, false
}

var sensitiveResourceRegex = regexp.MustCompile(`(?i)\b(orders|documents|invoices|transfers|payments|cards|records|profiles|files)\b`)

func detectSQLAUTH002(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr, taintedVars map[string]*finding.DataflowStep) (finding.Finding, bool) {
	if !isDBSinkMethod(method) && !isORMSinkMethod(method) {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		if sensitiveResourceRegex.MatchString(strLit.Value) && (strings.Contains(upper, "WHERE ID =") || strings.Contains(upper, "WHERE ID=")) {
			hasOwnerScope := strings.Contains(upper, "USER_ID") || strings.Contains(upper, "OWNER_ID") || strings.Contains(upper, "ACCOUNT_ID")
			if !hasOwnerScope {
				pos := fset.Position(call.Pos())
				return finding.Finding{
					ID:             fmt.Sprintf("SQLAUTH-002-%s-%d", filepath.Base(relPath), pos.Line),
					RuleID:         "SQLAUTH-002",
					Tool:           "sqltaint",
					Domain:         finding.Security,
					Category:       "authorization-idor",
					Severity:       finding.High,
					Confidence:     finding.ConfidenceHigh,
					Exploitability: finding.ExploitabilityLikely,
					FindingState:   finding.FindingConfirmed,
					Description:    fmt.Sprintf("Sensitive resource queried solely by object ID at %s() without user ownership scoping (IDOR risk)", method),
					Recommendation: "Scope entity lookups by both the object ID and authenticated user/account ID (e.g. WHERE id = $1 AND user_id = $2)",
					Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlauth-002",
					Location:       finding.Location{File: relPath, Line: pos.Line},
				}, true
			}
		}
	}
	return finding.Finding{}, false
}

func detectSQLAUTH003(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if method != "Raw" && method != "Exec" && method != "Query" {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		// Detect raw query bypassing auth filters on protected tables
		if (strings.Contains(upper, "SELECT * FROM") || strings.Contains(upper, "DELETE FROM")) &&
			(strings.Contains(upper, "USERS") || strings.Contains(upper, "ROLES") || strings.Contains(upper, "PERMISSIONS") || strings.Contains(upper, "AUDIT_LOGS")) &&
			!strings.Contains(upper, "WHERE") {
			pos := fset.Position(call.Pos())
			return finding.Finding{
				ID:             fmt.Sprintf("SQLAUTH-003-%s-%d", filepath.Base(relPath), pos.Line),
				RuleID:         "SQLAUTH-003",
				Tool:           "sqltaint",
				Domain:         finding.Security,
				Category:       "raw-query-bypass",
				Severity:       finding.High,
				Confidence:     finding.ConfidenceHigh,
				Exploitability: finding.ExploitabilityLikely,
				FindingState:   finding.FindingConfirmed,
				Description:    fmt.Sprintf("Raw query at %s() bypasses standard ORM authorization scopes and permission filters", method),
				Recommendation: "Ensure raw queries replicate all security barriers, role restrictions, and tenant scopes provided by ORM repositories",
				Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlauth-003",
				Location:       finding.Location{File: relPath, Line: pos.Line},
			}, true
		}
	}
	return finding.Finding{}, false
}

func detectSQLAUTH004(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if !isDBSinkMethod(method) {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		if strings.Contains(upper, "SET ROLE POSTGRES") || strings.Contains(upper, "BYPASSRLS") || strings.Contains(upper, "SET ROLE ROOT") || strings.Contains(upper, "SET ROLE SUPERUSER") {
			pos := fset.Position(call.Pos())
			return finding.Finding{
				ID:             fmt.Sprintf("SQLAUTH-004-%s-%d", filepath.Base(relPath), pos.Line),
				RuleID:         "SQLAUTH-004",
				Tool:           "sqltaint",
				Domain:         finding.Security,
				Category:       "rls-misconfiguration",
				Severity:       finding.High,
				Confidence:     finding.ConfidenceHigh,
				Exploitability: finding.ExploitabilityLikely,
				FindingState:   finding.FindingConfirmed,
				Description:    fmt.Sprintf("Database query at %s() assumes Row-Level Security but explicitly switches to superuser or bypass role", method),
				Recommendation: "Connect and execute application queries using least-privilege non-superuser roles to enforce database Row-Level Security",
				Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlauth-004",
				Location:       finding.Location{File: relPath, Line: pos.Line},
			}, true
		}
	}
	return finding.Finding{}, false
}

var financialFieldRegex = regexp.MustCompile(`(?i)\b(balance|stock|quota|points|credits|quantity|inventory|coins|wallet)\b`)

func detectSQLSAFE003(fset *token.FileSet, relPath string, body *ast.BlockStmt, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if !isDBSinkMethod(method) {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		if strings.Contains(upper, "UPDATE") && financialFieldRegex.MatchString(strLit.Value) {
			// Check if function performs a SELECT without FOR UPDATE
			hasSelectWithoutLock := false
			ast.Inspect(body, func(n ast.Node) bool {
				if innerCall, ok := n.(*ast.CallExpr); ok {
					_, innerMethod := extractCallNames(innerCall)
					if (innerMethod == "Query" || innerMethod == "QueryRow" || innerMethod == "Get") && len(innerCall.Args) > 0 {
						resInner := resolveExpr(innerCall.Args[0], vMap)
						if innerLit, ok := resInner.(*ast.BasicLit); ok && innerLit.Kind == token.STRING {
							innerUpper := strings.ToUpper(innerLit.Value)
							if strings.Contains(innerUpper, "SELECT") && !strings.Contains(innerUpper, "FOR UPDATE") && financialFieldRegex.MatchString(innerLit.Value) {
								hasSelectWithoutLock = true
							}
						}
					}
				}
				return true
			})

			if hasSelectWithoutLock {
				pos := fset.Position(call.Pos())
				return finding.Finding{
					ID:             fmt.Sprintf("SQLSAFE-003-%s-%d", filepath.Base(relPath), pos.Line),
					RuleID:         "SQLSAFE-003",
					Tool:           "sqltaint",
					Domain:         finding.Reliability,
					Category:       "concurrency-hazard",
					Severity:       finding.High,
					Confidence:     finding.ConfidenceHigh,
					Exploitability: finding.ExploitabilityLikely,
					FindingState:   finding.FindingConfirmed,
					Description:    fmt.Sprintf("Non-atomic read-modify-write pattern detected on balance/inventory field at %s() without row locking", method),
					Recommendation: "Use SELECT ... FOR UPDATE within a transaction or perform atomic SQL mutations (e.g. SET balance = balance - $1 WHERE balance >= $1)",
					Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlsafe-003",
					Location:       finding.Location{File: relPath, Line: pos.Line},
				}, true
			}
		}
	}
	return finding.Finding{}, false
}

func detectSQLSAFE004(fset *token.FileSet, relPath string, hasTxParam bool, method string, call *ast.CallExpr) (finding.Finding, bool) {
	if !hasTxParam {
		return finding.Finding{}, false
	}

	funName, _ := extractCallNames(call)
	// If method is executed on global `db` instead of `tx` within a transactional function
	if funName == "db" && isDBSinkMethod(method) {
		pos := fset.Position(call.Pos())
		return finding.Finding{
			ID:             fmt.Sprintf("SQLSAFE-004-%s-%d", filepath.Base(relPath), pos.Line),
			RuleID:         "SQLSAFE-004",
			Tool:           "sqltaint",
			Domain:         finding.Reliability,
			Category:       "transaction-loss",
			Severity:       finding.High,
			Confidence:     finding.ConfidenceHigh,
			Exploitability: finding.ExploitabilityLikely,
			FindingState:   finding.FindingConfirmed,
			Description:    fmt.Sprintf("Database operation db.%s() executes on global connection pool escaping active transaction boundary", method),
			Recommendation: "Execute queries using the active transaction handle (tx.%s) to guarantee atomic rollback on error",
			Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlsafe-004",
			Location:       finding.Location{File: relPath, Line: pos.Line},
		}, true
	}
	return finding.Finding{}, false
}

var unparenthesizedAndOrRegex = regexp.MustCompile(`(?i)\bWHERE\s+[^()]+AND\s+[^()]+OR\s+[^()]+;?$`)

func detectSQLSAFE005(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if !isDBSinkMethod(method) && !isORMSinkMethod(method) {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		if unparenthesizedAndOrRegex.MatchString(strLit.Value) {
			pos := fset.Position(call.Pos())
			return finding.Finding{
				ID:             fmt.Sprintf("SQLSAFE-005-%s-%d", filepath.Base(relPath), pos.Line),
				RuleID:         "SQLSAFE-005",
				Tool:           "sqltaint",
				Domain:         finding.Reliability,
				Category:       "logic-operator-precedence",
				Severity:       finding.High,
				Confidence:     finding.ConfidenceHigh,
				Exploitability: finding.ExploitabilityLikely,
				FindingState:   finding.FindingConfirmed,
				Description:    fmt.Sprintf("Query at %s() contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence", method),
				Recommendation: "Explicitly group logical expressions with parentheses to avoid inadvertent filter bypass or tenant leakage",
				Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlsafe-005",
				Location:       finding.Location{File: relPath, Line: pos.Line},
			}, true
		}
	}
	return finding.Finding{}, false
}

var softDeleteTableRegex = regexp.MustCompile(`(?i)\b(users|accounts|posts|articles|products|documents|orders)\b`)

func detectSQLSAFE006(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if method != "Query" && method != "QueryRow" && method != "Raw" {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		if strings.Contains(upper, "SELECT") && softDeleteTableRegex.MatchString(strLit.Value) {
			if !strings.Contains(upper, "DELETED_AT") {
				pos := fset.Position(call.Pos())
				return finding.Finding{
					ID:             fmt.Sprintf("SQLSAFE-006-%s-%d", filepath.Base(relPath), pos.Line),
					RuleID:         "SQLSAFE-006",
					Tool:           "sqltaint",
					Domain:         finding.Reliability,
					Category:       "soft-delete-bypass",
					Severity:       finding.Medium,
					Confidence:     finding.ConfidenceHigh,
					Exploitability: finding.ExploitabilityLikely,
					FindingState:   finding.FindingConfirmed,
					Description:    fmt.Sprintf("Raw query at %s() omits deleted_at IS NULL condition on soft-deletable entity table", method),
					Recommendation: "Include 'deleted_at IS NULL' in WHERE clauses when querying tables that use soft deletion",
					Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#sqlsafe-006",
					Location:       finding.Location{File: relPath, Line: pos.Line},
				}, true
			}
		}
	}
	return finding.Finding{}, false
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
		if sel == "Get" || sel == "Query" || sel == "QueryParam" || sel == "FormValue" || sel == "Param" || sel == "Params" || sel == "URLParam" {
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

var unboundedPublicTableRegex = regexp.MustCompile(`(?i)\b(users|orders|accounts|logs|events|transactions|messages)\b`)

func detectDBPERF001(fset *token.FileSet, relPath, method string, call *ast.CallExpr, vMap map[string]ast.Expr) (finding.Finding, bool) {
	if method != "Query" && method != "Raw" && method != "Select" {
		return finding.Finding{}, false
	}
	if len(call.Args) == 0 {
		return finding.Finding{}, false
	}

	resolved := resolveExpr(call.Args[0], vMap)
	if strLit, ok := resolved.(*ast.BasicLit); ok && strLit.Kind == token.STRING {
		upper := strings.ToUpper(strLit.Value)
		if strings.Contains(upper, "SELECT") && unboundedPublicTableRegex.MatchString(strLit.Value) {
			if !strings.Contains(upper, "LIMIT") && !strings.Contains(upper, "WHERE ID =") && !strings.Contains(upper, "WHERE ID=") {
				pos := fset.Position(call.Pos())
				return finding.Finding{
					ID:             fmt.Sprintf("DBPERF-001-%s-%d", filepath.Base(relPath), pos.Line),
					RuleID:         "DBPERF-001",
					Tool:           "sqltaint",
					Domain:         finding.Reliability,
					Category:       "query-performance",
					Severity:       finding.Medium,
					Confidence:     finding.ConfidenceHigh,
					Exploitability: finding.ExploitabilityLikely,
					FindingState:   finding.FindingConfirmed,
					Description:    fmt.Sprintf("Public dataset queried at %s() without an explicit LIMIT or pagination boundary", method),
					Recommendation: "Always enforce LIMIT and pagination to prevent unbounded result sets and excessive memory consumption",
					Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#dbperf-001",
					Location:       finding.Location{File: relPath, Line: pos.Line},
				}, true
			}
		}
	}
	return finding.Finding{}, false
}

func findQueriesInLoop(fset *token.FileSet, relPath string, loopBody *ast.BlockStmt) []finding.Finding {
	var findings []finding.Finding
	if loopBody == nil {
		return findings
	}

	ast.Inspect(loopBody, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		_, selectorName := extractCallNames(call)
		if isDBSinkMethod(selectorName) || isORMSinkMethod(selectorName) {
			pos := fset.Position(call.Pos())
			findings = append(findings, finding.Finding{
				ID:             fmt.Sprintf("DBPERF-002-%s-%d", filepath.Base(relPath), pos.Line),
				RuleID:         "DBPERF-002",
				Tool:           "sqltaint",
				Domain:         finding.Reliability,
				Category:       "n-plus-one",
				Severity:       finding.High,
				Confidence:     finding.ConfidenceHigh,
				Exploitability: finding.ExploitabilityLikely,
				FindingState:   finding.FindingConfirmed,
				Description:    fmt.Sprintf("Database query %s() executed inside loop (N+1 query pattern)", selectorName),
				Recommendation: "Batch queries using WHERE id IN (...) or JOINs to fetch data in a single roundtrip rather than in a loop",
				Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#dbperf-002",
				Location:       finding.Location{File: relPath, Line: pos.Line},
			})
		}
		return true
	})

	return findings
}

func detectDBSEC003(fset *token.FileSet, relPath, funName, selName string, call *ast.CallExpr) (finding.Finding, bool) {
	// Check c.JSON / c.String / http.Error / w.Write passing err.Error()
	isHTTPResp := selName == "JSON" || selName == "String" || selName == "Error" || (funName == "w" && selName == "Write")
	if !isHTTPResp {
		return finding.Finding{}, false
	}

	hasErrError := false
	for _, arg := range call.Args {
		ast.Inspect(arg, func(n ast.Node) bool {
			if innerCall, ok := n.(*ast.CallExpr); ok {
				_, innerSel := extractCallNames(innerCall)
				if innerSel == "Error" {
					hasErrError = true
				}
			}
			return true
		})
	}

	if hasErrError {
		pos := fset.Position(call.Pos())
		return finding.Finding{
			ID:             fmt.Sprintf("DBSEC-003-%s-%d", filepath.Base(relPath), pos.Line),
			RuleID:         "DBSEC-003",
			Tool:           "sqltaint",
			Domain:         finding.Security,
			Category:       "information_exposure",
			Severity:       finding.High,
			Confidence:     finding.ConfidenceHigh,
			Exploitability: finding.ExploitabilityLikely,
			FindingState:   finding.FindingConfirmed,
			Description:    fmt.Sprintf("Internal database/driver error passed directly to HTTP client response at %s()", selName),
			Recommendation: "Log the raw error securely on the server and return a sanitized, generic error message to the client",
			Documentation:  "https://cinnamorollofficials.github.io/go-code-scanner/reference/rules#dbsec-003",
			Location:       finding.Location{File: relPath, Line: pos.Line},
		}, true
	}

	return finding.Finding{}, false
}

