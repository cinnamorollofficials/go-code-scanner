package frontend

import (
	"bytes"
	"fmt"
	"strings"
)

type TokenType string

const (
	TokenCode           TokenType = "code"
	TokenComment        TokenType = "comment"
	TokenString         TokenType = "string"
	TokenTemplate       TokenType = "template"
	TokenJSXAttribute   TokenType = "jsx_attr"
	TokenScriptRegion   TokenType = "script"
	TokenTemplateRegion TokenType = "template_region"
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
	Start  int
	End    int
}

// Tokenize parses input JavaScript/TypeScript/HTML/Vue/Svelte source text into bounded lexical tokens.
// It is panic-safe and returns deterministic partial results if malformed or oversized.
func Tokenize(source []byte) (tokens []Token, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("lexer panic: %v", r)
		}
	}()

	const maxSourceBytes = 5 * 1024 * 1024
	if len(source) > maxSourceBytes {
		source = source[:maxSourceBytes]
	}

	if bytes.Contains(source, []byte("<script")) || bytes.Contains(source, []byte("<template")) {
		regions := extractTemplateRegions(source)
		if len(regions) > 0 {
			lastEnd := 0
			for _, r := range regions {
				if r.Start > lastEnd {
					outsideTokens := tokenizeJS(source[lastEnd:r.Start])
					lineOffset := bytesLineOffset(source, lastEnd)
					for idx := range outsideTokens {
						outsideTokens[idx].Line += lineOffset
						outsideTokens[idx].Start += lastEnd
						outsideTokens[idx].End += lastEnd
					}
					tokens = append(tokens, outsideTokens...)
				}
				tokens = append(tokens, r)
				innerTokens := tokenizeJS([]byte(r.Value))
				lineOffset := bytesLineOffset(source, r.Start)
				for idx := range innerTokens {
					innerTokens[idx].Line += lineOffset
					innerTokens[idx].Start += r.Start
					innerTokens[idx].End += r.Start
				}
				tokens = append(tokens, innerTokens...)
				lastEnd = r.End
			}
			if lastEnd < len(source) {
				outsideTokens := tokenizeJS(source[lastEnd:])
				lineOffset := bytesLineOffset(source, lastEnd)
				for idx := range outsideTokens {
					outsideTokens[idx].Line += lineOffset
					outsideTokens[idx].Start += lastEnd
					outsideTokens[idx].End += lastEnd
				}
				tokens = append(tokens, outsideTokens...)
			}
			if len(tokens) > 0 {
				return tokens, nil
			}
		}
	}

	return tokenizeJS(source), nil
}

func bytesLineOffset(src []byte, offset int) int {
	line := 0
	for i := 0; i < offset && i < len(src); i++ {
		if src[i] == '\n' {
			line++
		}
	}
	return line
}

func tokenizeJS(src []byte) []Token {
	var tokens []Token
	n := len(src)
	line := 1
	col := 1
	i := 0

	for i < n {
		ch := src[i]

		if ch == '\n' {
			line++
			col = 1
			i++
			continue
		}

		if i+1 < n && ch == '/' && src[i+1] == '/' {
			start := i
			startLine, startCol := line, col
			i += 2
			col += 2
			for i < n && src[i] != '\n' {
				i++
				col++
			}
			tokens = append(tokens, Token{
				Type: TokenComment, Value: string(src[start:i]),
				Line: startLine, Column: startCol, Start: start, End: i,
			})
			continue
		}

		if i+1 < n && ch == '/' && src[i+1] == '*' {
			start := i
			startLine, startCol := line, col
			i += 2
			col += 2
			for i < n {
				if src[i] == '\n' {
					line++
					col = 1
					i++
					continue
				}
				if i+1 < n && src[i] == '*' && src[i+1] == '/' {
					i += 2
					col += 2
					break
				}
				i++
				col++
			}
			tokens = append(tokens, Token{
				Type: TokenComment, Value: string(src[start:i]),
				Line: startLine, Column: startCol, Start: start, End: i,
			})
			continue
		}

		if ch == '\'' || ch == '"' {
			quote := ch
			start := i
			startLine, startCol := line, col
			i++
			col++
			escaped := false
			for i < n {
				cur := src[i]
				if cur == '\n' {
					line++
					col = 1
					i++
					break
				}
				if escaped {
					escaped = false
				} else if cur == '\\' {
					escaped = true
				} else if cur == quote {
					i++
					col++
					break
				}
				i++
				col++
			}
			tokens = append(tokens, Token{
				Type: TokenString, Value: string(src[start:i]),
				Line: startLine, Column: startCol, Start: start, End: i,
			})
			continue
		}

		if ch == '`' {
			start := i
			startLine, startCol := line, col
			i++
			col++
			escaped := false
			depth := 0
			for i < n {
				cur := src[i]
				if cur == '\n' {
					line++
					col = 1
				}
				if escaped {
					escaped = false
				} else if cur == '\\' {
					escaped = true
				} else if cur == '$' && i+1 < n && src[i+1] == '{' {
					depth++
					i += 2
					col += 2
					continue
				} else if cur == '}' && depth > 0 {
					depth--
					i++
					col++
					continue
				} else if cur == '`' && depth == 0 {
					i++
					col++
					break
				}
				i++
				col++
			}
			tokens = append(tokens, Token{
				Type: TokenTemplate, Value: string(src[start:i]),
				Line: startLine, Column: startCol, Start: start, End: i,
			})
			continue
		}

		if isAlpha(ch) {
			start := i
			startLine, startCol := line, col
			for i < n && (isAlphaNum(src[i]) || src[i] == '-' || src[i] == '_') {
				i++
				col++
			}
			word := string(src[start:i])
			j := i
			for j < n && (src[j] == ' ' || src[j] == '\t') {
				j++
			}
			if j < n && src[j] == '=' {
				tokens = append(tokens, Token{
					Type: TokenJSXAttribute, Value: word,
					Line: startLine, Column: startCol, Start: start, End: i,
				})
				continue
			}
			tokens = append(tokens, Token{
				Type: TokenCode, Value: word,
				Line: startLine, Column: startCol, Start: start, End: i,
			})
			continue
		}

		start := i
		startLine, startCol := line, col
		i++
		col++
		tokens = append(tokens, Token{
			Type: TokenCode, Value: string(src[start:i]),
			Line: startLine, Column: startCol, Start: start, End: i,
		})
	}

	return tokens
}

func extractTemplateRegions(src []byte) []Token {
	var tokens []Token
	s := string(src)

	lower := strings.ToLower(s)
	idx := 0
	for idx < len(lower) {
		scriptStart := strings.Index(lower[idx:], "<script")
		if scriptStart == -1 {
			break
		}
		scriptStart += idx
		tagClose := strings.IndexByte(lower[scriptStart:], '>')
		if tagClose == -1 {
			break
		}
		contentStart := scriptStart + tagClose + 1
		if contentStart > len(s) {
			break
		}
		scriptEnd := strings.Index(lower[contentStart:], "</script>")
		if scriptEnd == -1 {
			contentEnd := len(s)
			if contentStart <= contentEnd {
				tokens = append(tokens, Token{
					Type:  TokenScriptRegion,
					Value: s[contentStart:contentEnd],
					Start: contentStart,
					End:   contentEnd,
				})
			}
			break
		}
		contentEnd := contentStart + scriptEnd
		if contentStart <= contentEnd && contentEnd <= len(s) {
			tokens = append(tokens, Token{
				Type:  TokenScriptRegion,
				Value: s[contentStart:contentEnd],
				Start: contentStart,
				End:   contentEnd,
			})
		}
		idx = contentEnd + len("</script>")
	}

	idx = 0
	for idx < len(lower) {
		tmplStart := strings.Index(lower[idx:], "<template")
		if tmplStart == -1 {
			break
		}
		tmplStart += idx
		tagClose := strings.IndexByte(lower[tmplStart:], '>')
		if tagClose == -1 {
			break
		}
		contentStart := tmplStart + tagClose + 1
		if contentStart > len(s) {
			break
		}
		tmplEnd := strings.Index(lower[contentStart:], "</template>")
		if tmplEnd == -1 {
			contentEnd := len(s)
			if contentStart <= contentEnd {
				tokens = append(tokens, Token{
					Type:  TokenTemplateRegion,
					Value: s[contentStart:contentEnd],
					Start: contentStart,
					End:   contentEnd,
				})
			}
			break
		}
		contentEnd := contentStart + tmplEnd
		if contentStart <= contentEnd && contentEnd <= len(s) {
			tokens = append(tokens, Token{
				Type:  TokenTemplateRegion,
				Value: s[contentStart:contentEnd],
				Start: contentStart,
				End:   contentEnd,
			})
		}
		idx = contentEnd + len("</template>")
	}

	return tokens
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '$' || b == '_'
}

func isAlphaNum(b byte) bool {
	return isAlpha(b) || (b >= '0' && b <= '9')
}
