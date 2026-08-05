package frontend

import (
	"testing"
)

func TestExtractImportGraphEdges(t *testing.T) {
	source := []byte(`
		import React from 'react';
		import Card from './Card';
		export { Header } from '@/components/Header';
		const db = require('../db');
		const Lazy = () => import('./LazyComponent');
	`)

	tokens, err := Tokenize(source)
	if err != nil {
		t.Fatalf("unexpected tokenize error: %v", err)
	}

	edges := ExtractImportEdges("src/app.js", tokens)
	if len(edges) != 4 {
		t.Fatalf("expected 4 local import edges, got %d: %+v", len(edges), edges)
	}

	expected := []struct {
		spec string
		kind ImportKind
	}{
		{"./Card", KindStaticImport},
		{"@/components/Header", KindExportFrom},
		{"../db", KindRequire},
		{"./LazyComponent", KindDynamicImport},
	}

	for i, exp := range expected {
		if edges[i].ToSpecifier != exp.spec {
			t.Errorf("edge %d: expected specifier %s, got %s", i, exp.spec, edges[i].ToSpecifier)
		}
		if edges[i].Kind != exp.kind {
			t.Errorf("edge %d: expected kind %s, got %s", i, exp.kind, edges[i].Kind)
		}
		if edges[i].FromFile != "src/app.js" {
			t.Errorf("edge %d: expected fromFile src/app.js, got %s", i, edges[i].FromFile)
		}
	}
}
