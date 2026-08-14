package sqltaint

import (
	"go/ast"
	"go/token"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

type HoleContext string

const (
	HoleContextValue          HoleContext = "value"
	HoleContextIdentifier     HoleContext = "identifier"
	HoleContextTable          HoleContext = "table"
	HoleContextColumn         HoleContext = "column"
	HoleContextOrderDirection HoleContext = "order-direction"
	HoleContextOperator       HoleContext = "operator"
	HoleContextKeyword        HoleContext = "keyword"
	HoleContextClause         HoleContext = "clause"
	HoleContextListExpansion  HoleContext = "list-expansion"
	HoleContextUnknown        HoleContext = "unknown"
)

type TaintTrust string

const (
	TrustConstant  TaintTrust = "trusted_constant"
	TrustSanitizer TaintTrust = "sanitized"
	TrustUntrusted TaintTrust = "untrusted"
	TrustUnknown   TaintTrust = "unknown"
)

type Hole struct {
	Context    HoleContext           `json:"context"`
	Trust      TaintTrust            `json:"trust"`
	Expression string                `json:"expression,omitempty"`
	SourceStep *finding.DataflowStep `json:"source_step,omitempty"`
}

type TemplateKind string

const (
	KindRawConcatenation  TemplateKind = "raw_concatenation"
	KindPreparedStatement TemplateKind = "prepared_statement"
	KindBoundParam        TemplateKind = "bound_parameterized"
	KindORMBuilder        TemplateKind = "orm_builder"
)

type TemplateSegment struct {
	IsConst bool   `json:"is_const"`
	Text    string `json:"text,omitempty"`
	Hole    *Hole  `json:"hole,omitempty"`
}

type SQLTemplate struct {
	Kind     TemplateKind      `json:"kind"`
	RawText  string            `json:"raw_text"`
	Segments []TemplateSegment `json:"segments"`
	Location finding.Location  `json:"location"`
}

func (t SQLTemplate) HasUntrustedHole() bool {
	for _, seg := range t.Segments {
		if seg.Hole != nil && seg.Hole.Trust == TrustUntrusted {
			return true
		}
	}
	return false
}

// FunctionSummary captures metadata for interprocedural analysis across functions
type FunctionSummary struct {
	Name         string
	ReceiverType string
	File         string
	Line         int
	Params       []string
	Returns      []ast.Expr
	Decl         *ast.FuncDecl
	Fset         *token.FileSet
	IsHTTPHandler bool
	RouterType   string // "gin", "echo", "chi", "fiber", "mux", "net/http"
}

// CallGraphEdge represents a function call from caller to callee
type CallGraphEdge struct {
	CallerFile string
	CallerFunc string
	CalleeFunc string
	CallExpr   *ast.CallExpr
	Line       int
}

// PackageAnalysisContext holds the multi-file ASTs and call graph
type PackageAnalysisContext struct {
	Files      map[string]*ast.File
	FileSets   map[string]*token.FileSet
	Functions  map[string]*FunctionSummary
	CallEdges  []CallGraphEdge
}

