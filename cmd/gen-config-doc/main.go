package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
)

func main() {
	meta, err := config.GenerateMetadata()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating metadata: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("title: Complete Field Reference\n")
	buf.WriteString("description: Automatically generated configuration reference tables from Go struct definitions.\n")
	buf.WriteString("---\n\n")

	buf.WriteString("# Complete Configuration Field Reference\n\n")
	buf.WriteString("This page is automatically generated from Go struct definitions in `pkg/config`. Do not edit manually.\n\n")

	buf.WriteString("| Field Path | Type | Default | Required | Description |\n")
	buf.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")

	for _, f := range meta.Fields {
		defStr := fmt.Sprintf("%v", f.Default)
		if strings.HasPrefix(f.Type, "[]") || strings.HasPrefix(f.Type, "map") {
			if defBytes, err := fmt.Printf(""); err == nil {
				_ = defBytes
			}
		}
		reqStr := "No"
		if f.Required {
			reqStr = "Yes"
		}
		buf.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s | %s |\n", f.Path, f.Type, defStr, reqStr, f.Description))
	}

	targetPath := filepath.Join("website", "docs", "reference", "config", "generated-reference.md")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing generated reference: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated configuration reference at %s\n", targetPath)
}
