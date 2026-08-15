---
title: Scanner and Adapter Compatibility
description: "For configuration authors: compare built-in scanners and external adapters by domain, network use, and parser format."
---

# Scanner and Adapter Compatibility

`security-review` includes native AST scanners and supports external tool adapters. Below is the compatibility and requirement matrix.

| Adapter ID | Executable | Domain | Offline compatible | Parser format | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `gofmt` | `gofmt` | `quality` | Yes | Paths list | Checks Go source code formatting against gofmt standards. |
| `go-vet` | `go` | `reliability` | Yes | Compiler text | Examines Go source code and reports suspicious constructs. |
| `go-test` | `go` | `reliability` | Yes | Test output | Executes Go test suites across workspace packages. |
| `govulncheck` | `govulncheck` | `supply_chain` | No (network required) | JSON stream | Official Go vulnerability scanner for known module CVEs. |
| `gosec` | `gosec` | `security` | Yes | JSON report | Inspects Go AST for security flaws and unsafe practices. |
| `gitleaks` | `gitleaks` | `security` | Yes | JSON array | High-performance secret and credential detector. |
| `trivy` | `trivy` | `supply_chain` | No (network required) | JSON vulnerability schema | Vulnerability scanner for containers and dependencies. |
| `osv-scanner` | `osv-scanner` | `supply_chain` | No (network required) | OSV JSON | Vulnerability scanner using Open Source Vulnerabilities database. |
| `semgrep` | `semgrep` | `security` | Yes | JSON output | Multi-language lightweight static analysis engine. |
| `sqltaint` | `built-in` | `security` | Yes | AST / Dataflow | Native Go AST and intraprocedural SQL taint analysis engine. |
| `eslint` | `eslint` | `quality` | Yes | JSON format | Pluggable linting utility for JavaScript and TypeScript. |
| `tsc` | `tsc` | `quality` | Yes | Compiler text | TypeScript language compiler type checker. |
| `biome` | `biome` | `quality` | Yes | JSON report | Fast linter and formatter for JavaScript/TypeScript. |
