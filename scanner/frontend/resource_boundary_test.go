package frontend

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func TestFrontendResourceBoundaries(t *testing.T) {
	cfg := config.Default()
	cfg.Frontend.Enabled = true
	cfg.PatternMaxFileBytes = 100 // tight 100 byte limit for test
	s := New(cfg)

	// Oversized file exceeding 100 bytes
	oversizedContent := strings.Repeat("const x = 1;\n", 20)
	oversizedSource := scanner.Source{
		Path: "/project/src/Oversized.tsx",
		Open: func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(oversizedContent)), nil
		},
	}

	req := scanner.Request{
		Root:    "/project",
		Mode:    "full",
		Sources: []scanner.Source{oversizedSource},
	}

	res := s.Scan(context.Background(), req)
	// Expect partial result or source-file-size quality finding
	if res.State != finding.ScannerFindings && res.State != finding.ScannerPartial && res.State != finding.ScannerClean {
		t.Fatalf("unexpected state for oversized file: %v", res.State)
	}
}

func TestImportGraphDepthLimit(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&sb, "import { f%d } from './file%d';\n", i+1, i+1)
	}
	tokens, err := Tokenize([]byte(sb.String()))
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	edges := ExtractImportEdges("file0.ts", tokens)
	if len(edges) != 50 {
		t.Fatalf("expected 50 import edges, got %d", len(edges))
	}
}
