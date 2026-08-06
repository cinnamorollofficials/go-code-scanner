---
layout: home

hero:
  name: "Go Code Scanner"
  text: "Policy-driven, offline-first security analysis CLI"
  tagline: Enterprise-grade security gate, secret detection, dependency auditing, and SAST for Go codebases.
  actions:
    - theme: brand
      text: Quick Start
      link: #quick-start-paths
    - theme: alt
      text: View Documentation
      link: /getting-started/
    - theme: alt
      text: View on GitHub
      link: https://github.com/cinnamorollofficials/go-code-scanner

features:
  - title: 🔒 Offline-First & Private
    details: Runs 100% locally on your machine or build runner. No code, AST, or repository metadata is transmitted to remote servers.
  - title: 🛡️ Policy-Driven Commit Gate
    details: Enforces security compliance using deterministic exit codes (0 for pass, 1 for policy violations), baselines, and suppressions.
  - title: 🌐 6 Finding Domains
    details: Comprehensive coverage across Secrets, SAST, Vulnerabilities, Governance, Architecture, and Frontend security threats.
  - title: 📊 Multi-Format Reporting
    details: Outputs rich terminal colors, machine-readable JSON, SARIF for GitHub Security Code Scanning, and JUnit XML for CI test reporting.
---

## Overview

**`security-review` (Go Code Scanner)** is a high-performance, single-binary CLI tool designed to prevent security flaws, secret leaks, and governance regressions from entering production.

::: tip Policy-Driven Gate
Unlike conventional linters that spam developers with informational alerts, `security-review` operates as a strict **commit and release gate**. It returns exit code `1` only when policy-enforcing thresholds (such as `--ci` or `--fail-on HIGH`) are violated.
:::

---

## 6 Security Finding Domains

1. **Secrets Exposure**: Detects hardcoded API keys, private keys, JWT tokens, AWS credentials, and high-entropy strings.
2. **Static Application Security Testing (SAST)**: Identifies native Go vulnerabilities including SQL injection, command injection, path traversal, and unsafe pointer usage.
3. **Supply Chain & Vulnerabilities**: Audits direct and transitive Go module dependencies against known CVE databases.
4. **Governance & Compliance**: Enforces module naming conventions, required license files, security policy headers, and release manifests.
5. **Architecture & Design**: Detects circular package imports, package boundary violations, and resource boundary exceedances.
6. **Frontend & Ecosystem Scanning**: Analyzes React, Vue, Svelte, Next.js, and Nuxt assets for unsanitized HTML injection, secret leaks, and import cycles.

---

## Quick-Start Paths {#quick-start-paths}

Choose one of three copyable quick-start paths below to start scanning immediately:

### Path 1: Local Scan

Run a full security scan on your active working directory:

```sh
# Install binary from release or source
go install github.com/cinnamorollofficials/go-code-scanner/cmd/security-review@latest

# Run local scan
security-review scan
```

### Path 2: Staged Pre-Commit Hook Scan

Scan only git-staged changes before committing to maintain sub-second hook speed:

```sh
# Fast scan restricted to staged git diff
security-review scan --mode staged --profile fast
```

### Path 3: CI SARIF Scan (GitHub Actions)

Generate a SARIF report and fail the build if high-severity findings exist:

```sh
# Generate SARIF report for GitHub Code Scanning upload
security-review scan --ci --fail-on HIGH --format sarif --output results.sarif
```
