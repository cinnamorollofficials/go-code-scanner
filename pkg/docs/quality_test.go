package docs_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDocumentationQuality(t *testing.T) {
	docsDir := filepath.Join("..", "..", "website", "docs")
	var mdFiles []string

	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk docs directory: %v", err)
	}

	if len(mdFiles) == 0 {
		t.Fatalf("expected to find markdown files in website/docs, found 0")
	}

	forbiddenPathRegex := regexp.MustCompile(`[C-Z]:\\Users\\[a-zA-Z0-9_\-\.]+`)

	for _, file := range mdFiles {
		relPath, _ := filepath.Rel(docsDir, file)
		t.Run(relPath, func(t *testing.T) {
			contentBytes, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", relPath, err)
			}
			content := string(contentBytes)

			// 1. Check code fences balance
			fenceCount := strings.Count(content, "```")
			if fenceCount%2 != 0 {
				t.Errorf("unmatched code fences (``` count = %d) in %s", fenceCount, relPath)
			}

			// 2. Check forbidden local user paths
			if forbiddenPathRegex.MatchString(content) {
				t.Errorf("file %s contains leaked local absolute user path", relPath)
			}
		})
	}
}
