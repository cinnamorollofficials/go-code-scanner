package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
)

func main() {
	allRules := rules.Default()

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("title: Rule Catalog\n")
	buf.WriteString("description: Complete catalog of default built-in security, secret, governance, and quality rules.\n")
	buf.WriteString("---\n\n")

	buf.WriteString("# Built-In Rule Catalog\n\n")
	buf.WriteString("Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog is automatically generated from Go rule registries.\n\n")

	buf.WriteString("| Rule ID | Domain | Severity | Category | Description |\n")
	buf.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")

	for _, r := range allRules {
		desc := strings.ReplaceAll(r.Description, "\n", " ")
		buf.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | %s |\n", r.ID, r.Domain, r.Severity, r.Category, desc))
	}

	buf.WriteString("\n\n## Rule Details & Guidance\n\n")
	for _, r := range allRules {
		buf.WriteString(fmt.Sprintf("### `%s`\n\n", r.ID))
		buf.WriteString(fmt.Sprintf("- **Domain**: `%s`\n", r.Domain))
		buf.WriteString(fmt.Sprintf("- **Severity**: `%s`\n", r.Severity))
		buf.WriteString(fmt.Sprintf("- **Category**: `%s`\n\n", r.Category))
		buf.WriteString(fmt.Sprintf("**Description**: %s\n\n", r.Description))
		buf.WriteString(fmt.Sprintf("**Recommendation**: %s\n\n---\n\n", r.Recommendation))
	}

	targetPath := filepath.Join("website", "docs", "reference", "rules.md")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing rule catalog: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated rule catalog at %s (%d rules)\n", targetPath, len(allRules))
}
