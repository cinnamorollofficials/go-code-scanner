package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
)

type DomainInfo struct {
	Domain      finding.Domain
	Title       string
	Description string
}

func detectLanguage(exts []string) string {
	if len(exts) == 0 {
		return "go"
	}
	for _, ext := range exts {
		switch ext {
		case ".go":
			return "go"
		case ".ts", ".tsx", ".js", ".jsx":
			return "ts"
		case ".json":
			return "json"
		case ".yaml", ".yml":
			return "yaml"
		}
	}
	return "text"
}

func main() {
	allRules := rules.Default()

	domains := []DomainInfo{
		{
			Domain:      finding.Security,
			Title:       "🔒 Security Rules",
			Description: "Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws.",
		},
		{
			Domain:      finding.Hardening,
			Title:       "🛡️ Hardening Rules",
			Description: "Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings.",
		},
		{
			Domain:      finding.Reliability,
			Title:       "⚡ Reliability Rules",
			Description: "Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes.",
		},
		{
			Domain:      finding.Quality,
			Title:       "🧹 Quality Rules",
			Description: "Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements.",
		},
		{
			Domain:      finding.Governance,
			Title:       "📜 Governance Rules",
			Description: "Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints.",
		},
	}

	// Group rules by domain
	grouped := make(map[finding.Domain][]rules.Rule)
	for _, r := range allRules {
		grouped[r.Domain] = append(grouped[r.Domain], r)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("title: Rule Catalog\n")
	buf.WriteString("description: Complete catalog of default built-in security, secret, governance, and quality rules with Do's and Don'ts code examples.\n")
	buf.WriteString("---\n\n")

	buf.WriteString("# Built-In Rule Catalog\n\n")
	buf.WriteString("Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog includes detailed guidance, recommendations, and **Do's and Don'ts** code examples for each rule.\n\n")

	// Domain Overview Table
	buf.WriteString("## Domain Overview\n\n")
	buf.WriteString("| Domain | Icon | Total Rules | Scope & Focus |\n")
	buf.WriteString("| :--- | :---: | :---: | :--- |\n")
	for _, info := range domains {
		count := len(grouped[info.Domain])
		icon := strings.Fields(info.Title)[0]
		titleWithoutIcon := strings.TrimSpace(strings.TrimPrefix(info.Title, icon))
		buf.WriteString(fmt.Sprintf("| **%s** | %s | %d | %s |\n", titleWithoutIcon, icon, count, info.Description))
	}
	buf.WriteString("\n---\n\n")

	// Render each Domain Section
	for _, info := range domains {
		domainRules := grouped[info.Domain]
		if len(domainRules) == 0 {
			continue
		}

		buf.WriteString(fmt.Sprintf("## %s\n\n", info.Title))
		buf.WriteString(fmt.Sprintf("%s\n\n", info.Description))

		// Table for this domain
		buf.WriteString("| Rule ID | Severity | Category | Description |\n")
		buf.WriteString("| :--- | :--- | :--- | :--- |\n")

		for _, r := range domainRules {
			desc := strings.ReplaceAll(r.Description, "\n", " ")
			buf.WriteString(fmt.Sprintf("| [`%s`](#%s) | `%s` | `%s` | %s |\n", r.ID, r.ID, r.Severity, r.Category, desc))
		}

		buf.WriteString("\n### Details & Guidance\n\n")
		for _, r := range domainRules {
			buf.WriteString(fmt.Sprintf("#### `%s`\n\n", r.ID))
			buf.WriteString(fmt.Sprintf("- **Domain**: `%s`\n", r.Domain))
			buf.WriteString(fmt.Sprintf("- **Severity**: `%s`\n", r.Severity))
			buf.WriteString(fmt.Sprintf("- **Category**: `%s`\n\n", r.Category))
			buf.WriteString(fmt.Sprintf("**Description**: %s\n\n", r.Description))
			if r.Recommendation != "" {
				buf.WriteString(fmt.Sprintf("**Recommendation**: %s\n\n", r.Recommendation))
			}

			lang := detectLanguage(r.Extensions)

			if r.UnsafeExample != "" || r.SafeExample != "" {
				buf.WriteString("##### ❌ Don't (Unsafe)\n\n")
				buf.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", lang, strings.TrimSpace(r.UnsafeExample)))

				buf.WriteString("##### ✅ Do (Recommended)\n\n")
				buf.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", lang, strings.TrimSpace(r.SafeExample)))
			}

			buf.WriteString("---\n\n")
		}
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

	fmt.Printf("Successfully generated grouped rule catalog with code examples at %s (%d rules across %d domains)\n", targetPath, len(allRules), len(domains))
}
