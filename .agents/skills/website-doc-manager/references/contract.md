# Integration & Architectural Contract - `go-code-scanner` Website

## Overview

The `go-code-scanner` website is built using **VitePress**. It serves as the canonical technical documentation for the `security-review` CLI static security scanner tool.

## Key Tool & CLI Specifications

- **CLI Binary**: `security-review` (built from `cmd/security-review/main.go`).
- **Main Subcommands & Flags**:
  - `security-review scan [--root <path>] [--config <path>] [--format json|table|sarif] [--explain <rule>]`
  - `security-review hook install pre-commit`
  - `security-review hook uninstall`
  - `security-review version`
  - `security-review rules`

## Documentation Layout (`website/docs/`)

| Path | Description |
| --- | --- |
| `index.md` | Homepage & landing portal |
| `getting-started/` | Installation, Quickstart, First Scan |
| `features/` | Scan Execution & Policy, Developer Workflow, Reports & Findings |
| `guides/` | Advanced Configuration, CI/CD Integration, Troubleshooting |
| `reference/` | CLI Reference, Config Reference, Rule Catalog (`rules.md`) |
| `author-guide.md` | Style guide for writing docs |
| `security.md` | Security disclosure & vulnerability policy |
| `contributing.md` | Developer contribution guidelines |
| `changelog.md` | Version release history |

## Rule Catalog & Categories

The rule catalog (`website/docs/reference/rules.md`) contains 38 rules grouped into 5 domains:

1. **Security Rules** (🔒 17 rules): Hardcoded credentials, SQL injection, insecure crypto, etc.
2. **Hardening Rules** (🛡️ 6 rules): File permission flags, TLS minimum versions, strict headers.
3. **Reliability Rules** (⚡ 6 rules): Unhandled errors, goroutine leaks, context timeouts.
4. **Quality Rules** (🧹 5 rules): Dead code, anti-patterns, complex code structures.
5. **Governance Rules** (📜 4 rules): License headers, merge conflict markers, dependency restrictions.

## Documentation Generator Contract

- Source: `cmd/gen-rule-catalog/main.go`
- Target: `website/docs/reference/rules.md`
- Rule definitions: `pkg/rules/`
- Command: `go run ./cmd/gen-rule-catalog`
