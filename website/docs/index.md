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
    details: Enforces security compliance using deterministic exit codes (0 for pass, 1 for policy violations under CI), baselines, and suppressions.
  - title: 🌐 6 Canonical Policy Domains
    details: Comprehensive coverage across Quality, Reliability, Hardening, Security, Supply Chain, and Governance domains.
  - title: 📊 Multi-Format Reporting
    details: Outputs rich terminal summary, JSON (default artifact), SARIF for GitHub Security Code Scanning, and JUnit XML for CI reporting.
---

## Overview

**`security-review` (Go Code Scanner)** is a high-performance, single-binary CLI tool designed to prevent security flaws, secret leaks, and governance regressions from entering production.

::: tip Policy-Driven Gate
Unlike conventional linters that report purely informational warnings, `security-review` operates as a strict **commit and release gate**. It returns exit code `1` when policy thresholds are violated and the **`--ci`** flag is supplied.
:::

---

## 6 Canonical Policy Domains

1. **Security (`security`)**: Hardcoded API keys, private keys, database credentials, SQL injection, command injection, and tainted dataflows.
2. **Hardening (`hardening`)**: Insecure file permissions, weak TLS configurations, and missing security headers.
3. **Reliability (`reliability`)**: Unhandled error returns, goroutine leaks, and missing context cancellations.
4. **Quality (`quality`)**: Dead code, anti-patterns, empty handlers, and dangerous type conversions.
5. **Supply Chain (`supply_chain`)**: Vulnerable dependencies and untrusted third-party packages.
6. **Governance (`governance`)**: License compliance, unresolved merge conflict markers, and architectural constraints.

---

## Quick-Start Paths {#quick-start-paths}

Choose one of three copyable quick-start paths below to start scanning immediately:

### Path 1: Local Scan

Run a full security scan on your active working directory:

```sh
# Install binary via Go toolchain
go install github.com/cinnamorollofficials/go-code-scanner/cmd/security-review@latest

# Run local scan (writes security_findings.json and displays terminal summary)
security-review scan
```

### Path 2: Staged Pre-Commit Hook Scan

Scan only git-staged changes before committing to maintain sub-second hook speed:

```sh
# Fast scan restricted to staged git diff
security-review scan --staged --profile fast
```

### Path 3: CI SARIF Scan (GitHub Actions)

Generate a SARIF report and fail the build if high-severity findings exist:

```sh
# Generate SARIF report and enforce failure on High/Critical findings
security-review scan --ci --fail-on HIGH --format sarif --output results.sarif
```
