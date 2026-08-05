package frontend

import (
	"sort"
	"strings"
)

type ImportKind string

const (
	KindStaticImport  ImportKind = "static_import"
	KindExportFrom    ImportKind = "export_from"
	KindRequire       ImportKind = "require"
	KindDynamicImport ImportKind = "dynamic_import"
)

type ImportEdge struct {
	FromFile    string     `json:"from_file"`
	ToSpecifier string     `json:"to_specifier"`
	Line        int        `json:"line"`
	Kind        ImportKind `json:"kind"`
}

// ExtractImportEdges parses lexical tokens to extract local import dependencies.
func ExtractImportEdges(fromFile string, tokens []Token) []ImportEdge {
	var edges []ImportEdge
	n := len(tokens)

	for i := 0; i < n; i++ {
		tok := tokens[i]
		val := tok.Value

		if val == "import" {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+1)
				if len(args) > 0 && args[0].Type == TokenString {
					spec := strings.Trim(args[0].Value, `"'`+"`")
					if IsLocalSpecifier(spec) {
						edges = append(edges, ImportEdge{
							FromFile:    fromFile,
							ToSpecifier: spec,
							Line:        tok.Line,
							Kind:        KindDynamicImport,
						})
					}
				}
				continue
			}

			for j := i + 1; j < n && j < i+10; j++ {
				if tokens[j].Type == TokenString {
					spec := strings.Trim(tokens[j].Value, `"'`+"`")
					if IsLocalSpecifier(spec) {
						edges = append(edges, ImportEdge{
							FromFile:    fromFile,
							ToSpecifier: spec,
							Line:        tok.Line,
							Kind:        KindStaticImport,
						})
					}
					break
				}
			}
		}

		if val == "export" {
			for j := i + 1; j < n && j < i+15; j++ {
				if tokens[j].Value == "from" {
					for k := j + 1; k < n && k < j+5; k++ {
						if tokens[k].Type == TokenString {
							spec := strings.Trim(tokens[k].Value, `"'`+"`")
							if IsLocalSpecifier(spec) {
								edges = append(edges, ImportEdge{
									FromFile:    fromFile,
									ToSpecifier: spec,
									Line:        tok.Line,
									Kind:        KindExportFrom,
								})
							}
							break
						}
					}
					break
				}
			}
		}

		if val == "require" {
			if i+1 < n && tokens[i+1].Value == "(" {
				args := getArgTokens(tokens, i+1)
				if len(args) > 0 && args[0].Type == TokenString {
					spec := strings.Trim(args[0].Value, `"'`+"`")
					if IsLocalSpecifier(spec) {
						edges = append(edges, ImportEdge{
							FromFile:    fromFile,
							ToSpecifier: spec,
							Line:        tok.Line,
							Kind:        KindRequire,
						})
					}
				}
			}
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Line != edges[j].Line {
			return edges[i].Line < edges[j].Line
		}
		if edges[i].ToSpecifier != edges[j].ToSpecifier {
			return edges[i].ToSpecifier < edges[j].ToSpecifier
		}
		return edges[i].Kind < edges[j].Kind
	})

	return edges
}

// IsLocalSpecifier checks if a module specifier points to a local file/alias rather than an external package.
func IsLocalSpecifier(spec string) bool {
	if spec == "" {
		return false
	}
	return strings.HasPrefix(spec, ".") ||
		strings.HasPrefix(spec, "/") ||
		strings.HasPrefix(spec, "@/") ||
		strings.HasPrefix(spec, "~/") ||
		strings.HasPrefix(spec, "$lib/") ||
		strings.HasPrefix(spec, "#") ||
		strings.Contains(spec, ".server")
}
