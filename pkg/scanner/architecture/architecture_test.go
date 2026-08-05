package architecture

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestScannerReportsForbiddenGoDependency(t *testing.T) {
	source := New([]Layer{
		{Name: "domain", Paths: []string{"internal/domain/*.go"}},
		{Name: "infra", Paths: []string{"internal/infra/*.go"}},
	}, []Boundary{{From: "domain", To: "infra"}})
	files := []scanner.Source{
		memorySource("/repo/go.mod", "module example.test/app\n"),
		memorySource("/repo/internal/domain/service.go", "package domain\n\nimport \"example.test/app/internal/infra\"\n"),
		memorySource("/repo/internal/infra/store.go", "package infra\n"),
	}
	result := source.Scan(context.Background(), scanner.Request{Root: "/repo", Mode: "staged", RepositoryFiles: files})
	if result.State != finding.ScannerFindings || len(result.Findings) != 1 {
		t.Fatalf("unexpected architecture result: %+v", result)
	}
	item := result.Findings[0]
	if item.Location.File != "internal/domain/service.go" || item.Location.Line != 3 || item.Metadata["from_layer"] != "domain" || item.Metadata["to_layer"] != "infra" {
		t.Fatalf("unexpected architecture finding: %+v", item)
	}
}

func TestScannerAllowsConfiguredDirectionAndExternalImports(t *testing.T) {
	source := New([]Layer{
		{Name: "domain", Paths: []string{"internal/domain/*.go"}},
		{Name: "infra", Paths: []string{"internal/infra/*.go"}},
	}, []Boundary{{From: "domain", To: "infra"}})
	files := []scanner.Source{
		memorySource("/repo/go.mod", "module example.test/app\n"),
		memorySource("/repo/internal/infra/store.go", "package infra\nimport (\n \"example.test/app/internal/domain\"\n \"net/http\"\n)\n"),
	}
	result := source.Scan(context.Background(), scanner.Request{Root: "/repo", Mode: "full", RepositoryFiles: files})
	if result.State != finding.ScannerClean || len(result.Findings) != 0 {
		t.Fatalf("allowed imports produced findings: %+v", result)
	}
}

func TestScannerReportsDeterministicGoImportCycle(t *testing.T) {
	source := New(nil, nil, Options{DetectCycles: true})
	files := []scanner.Source{
		memorySource("/repo/go.mod", "module example.test/app\n"),
		memorySource("/repo/internal/a/a.go", "package a\nimport \"example.test/app/internal/b\"\n"),
		memorySource("/repo/internal/b/b.go", "package b\nimport \"example.test/app/internal/c\"\n"),
		memorySource("/repo/internal/c/c.go", "package c\nimport \"example.test/app/internal/a\"\n"),
	}
	request := scanner.Request{Root: "/repo", Mode: "full", RepositoryFiles: files}
	first := source.Scan(context.Background(), request)
	second := source.Scan(context.Background(), request)
	if first.State != finding.ScannerFindings || len(first.Findings) != 1 {
		t.Fatalf("unexpected cycle result: %+v", first)
	}
	if second.State != finding.ScannerFindings || len(second.Findings) != 1 {
		t.Fatalf("unexpected repeated cycle result: %+v", second)
	}
	wantPath := "internal/a -> internal/b -> internal/c -> internal/a"
	if first.Findings[0].RuleID != "architecture/import-cycle" || first.Findings[0].Metadata["dependency_path"] != wantPath {
		t.Fatalf("unexpected cycle finding: %+v", first.Findings[0])
	}
	if second.Findings[0].Metadata["dependency_path"] != first.Findings[0].Metadata["dependency_path"] || second.Findings[0].Location != first.Findings[0].Location {
		t.Fatalf("cycle finding is not deterministic: first=%+v second=%+v", first.Findings[0], second.Findings[0])
	}
}

func memorySource(path, content string) scanner.Source {
	return scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(content)), nil
	}}
}
