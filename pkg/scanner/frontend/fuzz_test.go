package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func FuzzFrontendParsersNeverPanic(f *testing.F) {
	seeds := [][]byte{
		[]byte(`import React from 'react'; const x = <div dangerouslySetInnerHTML={{__html: input}} />;`),
		[]byte(`<template><div v-html="raw"></template>`),
		[]byte(`<div>{@html svelteMarkup}</div>`),
		[]byte(`{"dependencies": {"react": "18.2.0", "bad": "latest"}}`),
		[]byte(`import { serverOnly } from 'server-only';`),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// 1. Lexer fuzzing
		tokens, _ := Tokenize(data)

		// 2. Import extraction fuzzing
		_ = ExtractImportEdges("Fuzz.tsx", tokens)

		// 3. Scanner fuzzing
		cfg := config.Default()
		cfg.Frontend.Enabled = true
		s := New(cfg)

		src := mockSource("/project/src/Fuzz.tsx", string(data))
		req := scanner.Request{
			Root:    "/project",
			Mode:    "full",
			Sources: []scanner.Source{src},
		}
		_ = s.Scan(context.Background(), req)
	})
}
