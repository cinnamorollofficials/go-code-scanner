package cache

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func TestHashSourcesIsDeterministicAndContentSensitive(t *testing.T) {
	source := func(path, content string) scanner.Source {
		return scanner.Source{Path: path, Open: func(context.Context) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(content)), nil }}
	}
	first, err := HashSources(context.Background(), "/repo", []scanner.Source{source("/repo/b.go", "b"), source("/repo/a.go", "a")})
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashSources(context.Background(), "/repo", []scanner.Source{source("/repo/a.go", "a"), source("/repo/b.go", "b")})
	if err != nil {
		t.Fatal(err)
	}
	if first["a.go"] != second["a.go"] || first["b.go"] != second["b.go"] {
		t.Fatalf("source order changed hashes: first=%v second=%v", first, second)
	}
	changed, _ := HashSources(context.Background(), "/repo", []scanner.Source{source("/repo/a.go", "changed")})
	if changed["a.go"] == first["a.go"] {
		t.Fatal("content change did not invalidate hash")
	}
}

func TestHashSourcesRejectsOutsideRoot(t *testing.T) {
	source := scanner.Source{Path: "/outside/a.go", Open: func(context.Context) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("a")), nil }}
	if _, err := HashSources(context.Background(), "/repo", []scanner.Source{source}); err == nil {
		t.Fatal("outside-root cache source accepted")
	}
}
