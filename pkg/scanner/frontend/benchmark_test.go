package frontend

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func BenchmarkFrontendScan(b *testing.B) {
	cfg := config.Default()
	cfg.Frontend.Enabled = true
	cfg.Frontend.RecognizeSanitizers = []string{"dompurify"}
	s := New(cfg)

	content := `import React from 'react';
import { useState } from 'react';
import DOMPurify from 'dompurify';

export const MyComponent = ({ html }: { html: string }) => {
  const [count, setCount] = useState(0);
  return (
    <div>
      <span>{count}</span>
      <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(html) }} />
    </div>
  );
};`

	sources := make([]scanner.Source, 50)
	for i := range sources {
		path := fmt.Sprintf("/project/src/Component%02d.tsx", i)
		sources[i] = scanner.Source{
			Path: path,
			Open: func(context.Context) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(content)), nil
			},
		}
	}

	req := scanner.Request{
		Root:    "/project",
		Mode:    "full",
		Sources: sources,
	}

	b.ReportAllocs()
	for b.Loop() {
		res := s.Scan(context.Background(), req)
		if res.State != finding.ScannerClean {
			b.Fatalf("unexpected benchmark result: %+v", res)
		}
	}
}

func BenchmarkFrontendImportExtraction(b *testing.B) {
	code := `
import { a, b, c } from './moduleA';
import type { T } from '@/types';
import * as Helper from '../utils/helper';
require('dotenv').config();
export { foo } from './foo';
`
	tokens, _ := Tokenize([]byte(code))
	b.ReportAllocs()
	for b.Loop() {
		edges := ExtractImportEdges("App.tsx", tokens)
		if len(edges) == 0 {
			b.Fatal("expected import edges")
		}
	}
}
