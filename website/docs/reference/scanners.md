---
title: Scanner & Adapter Compatibility Reference
description: Complete list of built-in scanners, external adapters, network requirements, and parser formats.
---

# Scanner & Adapter Compatibility Reference

`security-review` includes native AST scanners and supports external tool adapters. Below is the compatibility and requirement matrix.

| Adapter ID | Executable | Domain | Offline Compatible | Parser Format | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `gofmt` | `gofmt` | `quality` | Yes 🔒 | Paths list | Checks Go source code formatting against gofmt standards. |
| `go-vet` | `go` | `reliability` | Yes 🔒 | Compiler text | Examines Go source code and reports suspicious constructs. |
| `go-test` | `go` | `reliability` | Yes 🔒 | Test output | Executes Go test suites across workspace packages. |
| `govulncheck` | `govulncheck` | `vulnerabilities` | No 🌐 | JSON stream | Official Go vulnerability scanner for known module CVEs. |
| `gosec` | `gosec` | `security` | Yes 🔒 | JSON report | Inspects Go AST for security flaws and unsafe practices. |
| `gitleaks` | `gitleaks` | `secrets` | Yes 🔒 | JSON array | High-performance secret and credential detector. |
| `trivy` | `trivy` | `vulnerabilities` | No 🌐 | JSON vulnerability schema | Comprehensive vulnerability scanner for containers and dependencies. |
| `osv-scanner` | `osv-scanner` | `vulnerabilities` | No 🌐 | OSV JSON | Vulnerability scanner using Open Source Vulnerabilities database. |
| `semgrep` | `semgrep` | `security` | Yes 🔒 | JSON output | Multi-language lightweight static analysis engine. |
| `eslint` | `eslint` | `frontend` | Yes 🔒 | JSON format | Pluggable linting utility for JavaScript and TypeScript. |
| `tsc` | `tsc` | `frontend` | Yes 🔒 | Compiler text | TypeScript language compiler type checker. |
| `biome` | `biome` | `frontend` | Yes 🔒 | JSON report | Fast linter and formatter for JavaScript/TypeScript. |
