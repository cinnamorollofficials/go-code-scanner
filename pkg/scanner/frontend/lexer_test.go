package frontend

import (
	"testing"
)

func TestTokenizeJSCodeCommentsAndStrings(t *testing.T) {
	src := []byte(`
		// single line comment
		const name = "Alice";
		/* multi-line
		   comment */
		let message = 'hello \'world\'';
		const template = ` + "`" + `value ${1 + 2}` + "`" + `;
		<button onClick={handleClick} className="btn" />
	`)

	tokens, err := Tokenize(src)
	if err != nil {
		t.Fatalf("unexpected error tokenizing JS: %v", err)
	}
	if len(tokens) == 0 {
		t.Fatal("expected non-empty tokens")
	}

	foundComment := false
	foundString := false
	foundTemplate := false
	foundJSXAttr := false

	for _, tok := range tokens {
		switch tok.Type {
		case TokenComment:
			foundComment = true
		case TokenString:
			foundString = true
		case TokenTemplate:
			foundTemplate = true
		case TokenJSXAttribute:
			foundJSXAttr = true
		}
	}

	if !foundComment || !foundString || !foundTemplate || !foundJSXAttr {
		t.Fatalf("missing token types: comment=%v string=%v template=%v jsxAttr=%v",
			foundComment, foundString, foundTemplate, foundJSXAttr)
	}
}

func TestTokenizeTemplateRegions(t *testing.T) {
	src := []byte(`
		<template>
			<div>{{ msg }}</div>
		</template>
		<script lang="ts">
			const count = 0;
		</script>
	`)

	tokens, err := Tokenize(src)
	if err != nil {
		t.Fatalf("unexpected error tokenizing template: %v", err)
	}

	foundScript := false
	foundTemplate := false

	for _, tok := range tokens {
		if tok.Type == TokenScriptRegion {
			foundScript = true
		}
		if tok.Type == TokenTemplateRegion {
			foundTemplate = true
		}
	}

	if !foundScript || !foundTemplate {
		t.Fatalf("missing regions: script=%v template=%v", foundScript, foundTemplate)
	}
}
