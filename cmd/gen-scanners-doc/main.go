package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type AdapterInfo struct {
	ID              string
	Executable      string
	Domain          string
	RequiresNetwork bool
	ParserFormat    string
	Description     string
}

func main() {
	adaptersList := []AdapterInfo{
		{ID: "gofmt", Executable: "gofmt", Domain: "quality", RequiresNetwork: false, ParserFormat: "Paths list", Description: "Checks Go source code formatting against gofmt standards."},
		{ID: "go-vet", Executable: "go", Domain: "reliability", RequiresNetwork: false, ParserFormat: "Compiler text", Description: "Examines Go source code and reports suspicious constructs."},
		{ID: "go-test", Executable: "go", Domain: "reliability", RequiresNetwork: false, ParserFormat: "Test output", Description: "Executes Go test suites across workspace packages."},
		{ID: "govulncheck", Executable: "govulncheck", Domain: "vulnerabilities", RequiresNetwork: true, ParserFormat: "JSON stream", Description: "Official Go vulnerability scanner for known module CVEs."},
		{ID: "gosec", Executable: "gosec", Domain: "security", RequiresNetwork: false, ParserFormat: "JSON report", Description: "Inspects Go AST for security flaws and unsafe practices."},
		{ID: "gitleaks", Executable: "gitleaks", Domain: "secrets", RequiresNetwork: false, ParserFormat: "JSON array", Description: "High-performance secret and credential detector."},
		{ID: "trivy", Executable: "trivy", Domain: "vulnerabilities", RequiresNetwork: true, ParserFormat: "JSON vulnerability schema", Description: "Comprehensive vulnerability scanner for containers and dependencies."},
		{ID: "osv-scanner", Executable: "osv-scanner", Domain: "vulnerabilities", RequiresNetwork: true, ParserFormat: "OSV JSON", Description: "Vulnerability scanner using Open Source Vulnerabilities database."},
		{ID: "semgrep", Executable: "semgrep", Domain: "security", RequiresNetwork: false, ParserFormat: "JSON output", Description: "Multi-language lightweight static analysis engine."},
		{ID: "sqltaint", Executable: "built-in", Domain: "security", RequiresNetwork: false, ParserFormat: "AST / Dataflow", Description: "Native Go AST and intraprocedural SQL taint analysis engine."},
		{ID: "eslint", Executable: "eslint", Domain: "frontend", RequiresNetwork: false, ParserFormat: "JSON format", Description: "Pluggable linting utility for JavaScript and TypeScript."},
		{ID: "tsc", Executable: "tsc", Domain: "frontend", RequiresNetwork: false, ParserFormat: "Compiler text", Description: "TypeScript language compiler type checker."},
		{ID: "biome", Executable: "biome", Domain: "frontend", RequiresNetwork: false, ParserFormat: "JSON report", Description: "Fast linter and formatter for JavaScript/TypeScript."},
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("title: Scanner & Adapter Compatibility Reference\n")
	buf.WriteString("description: Complete list of built-in scanners, external adapters, network requirements, and parser formats.\n")
	buf.WriteString("---\n\n")

	buf.WriteString("# Scanner & Adapter Compatibility Reference\n\n")
	buf.WriteString("`security-review` includes native AST scanners and supports external tool adapters. Below is the compatibility and requirement matrix.\n\n")

	buf.WriteString("| Adapter ID | Executable | Domain | Offline Compatible | Parser Format | Description |\n")
	buf.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")

	for _, a := range adaptersList {
		offlineStr := "Yes 🔒"
		if a.RequiresNetwork {
			offlineStr = "No 🌐"
		}
		buf.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s | %s | %s |\n", a.ID, a.Executable, a.Domain, offlineStr, a.ParserFormat, a.Description))
	}

	targetPath := filepath.Join("website", "docs", "reference", "scanners.md")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing scanners reference: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated scanner compatibility reference at %s\n", targetPath)
}
