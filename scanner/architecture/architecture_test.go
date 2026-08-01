package architecture

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
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

func memorySource(path, content string) scanner.Source {
	return scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(content)), nil
	}}
}
