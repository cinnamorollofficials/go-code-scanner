package frontend

import (
	"testing"
)

func FuzzLexer(f *testing.F) {
	seeds := []string{
		`const x = "hello \"world\"";`,
		"const y = `template ${nested + `inside ${1}`} end`;",
		"// single line comment\nlet a = 1; /* multi\nline */",
		`<script lang="ts">\nlet count = 0;\n</script>`,
		`<template><div>{{ count }}</div></template>`,
		`<div dangerouslySetInnerHTML={{ __html: unescaped }} />`,
		`const truncated = "unclosed string`,
		"const truncatedTmpl = `unclosed template ${x",
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		tokens, err := Tokenize(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = tokens
	})
}
