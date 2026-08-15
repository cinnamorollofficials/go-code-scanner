---
layout: home
title: Go Code Scanner Documentation
description: Install, configure, and operate the policy-driven Go Code Scanner CLI.

hero:
  name: "Go Code Scanner"
  text: "Policy-driven, offline-first security analysis CLI"
  tagline: Single-binary commit gate, secret detection, dependency auditing, and AST/SQL taint analysis for Go codebases.
  actions:
    - theme: brand
      text: Quick Start
      link: #quick-start-paths
    - theme: alt
      text: Choose Your Goal
      link: #choose-your-goal
    - theme: alt
      text: View on GitHub
      link: https://github.com/cinnamorollofficials/go-code-scanner

features:
  - title: 🔒 Offline-First & Private
    details: Runs entirely on your local machine or self-hosted CI runner. No code, ASTs, or metadata are transmitted over the network.
  - title: 🛡️ Policy-Driven Commit Gate
    details: Enforces security compliance using deterministic exit codes (0 for pass, 1 for CI policy violations), baselines, and suppressions.
  - title: 🌐 6 Canonical Policy Domains
    details: Comprehensive coverage across Quality, Reliability, Hardening, Security, Supply Chain, and Governance domains.
  - title: 📊 Multi-Format Reporting
    details: Generates human-readable terminal summaries, JSON (default artifact), SARIF for GitHub Security tab, and JUnit XML for CI test reports.
---

## Choose Your Goal {#choose-your-goal}

<div class="tip custom-block" style="padding-top: 12px">

| Goal | Description | Recommended Guide |
| :--- | :--- | :--- |
| **Run a Local Scan** | Analyze your current workspace in seconds and review findings. | [First Scan Guide](/getting-started/first-scan) |
| **Set Up Pre-Commit Hook** | Block secret leaks and vulnerabilities before code is committed. | [Pre-Commit Hooks](/guides/pre-commit-hooks) |
| **Integrate with CI/CD** | Add automated commit and PR security gates with SARIF upload. | [Five-Minute CI Setup](/getting-started/ci-setup) |
| **Adopt in Existing Codebase** | Roll out without blocking velocity using finding baselines. | [Gradual Adoption Guide](/guides/baselines) |
| **Lookup CLI & Rules** | Browse flags, configuration options, and rule remediation examples. | [CLI Reference](/reference/cli) · [Rule Catalog](/reference/rule-catalog) |

</div>

---

## Quick-Start Paths {#quick-start-paths}

Choose one of three copyable quick-start paths below to start scanning immediately:

### Path 1: Local Scan

Run a full security scan on your active working directory:

```sh
# 1. Install security-review CLI binary via Go toolchain
go install github.com/cinnamorollofficials/go-code-scanner/cmd/security-review@latest

# 2. Run scan (writes security_findings.json and prints terminal summary)
security-review scan
```

::: details Expected Terminal Output
The following output was captured from a minimal Go fixture using the default
scan settings. Only the temporary absolute report path has been shortened.

```text
Code review: security-review (full)
  scanner pattern          clean
Findings: 0 | critical=0 high=0 medium=0 low=0 | suppressed=0 stale=0
Report: /path/to/my-project/security_findings.json
```
:::

::: tip Local vs. CI Exit Behavior
In local interactive execution, `security-review scan` outputs findings and returns **exit code `0`** so it does not interrupt local development. To return exit code `1` on policy violations, pass the **`--ci`** flag.
:::

---

### Path 2: Staged Pre-Commit Hook Scan

Scan only Git-staged changes before committing to maintain sub-second execution speed:

```sh
# Fast scan strictly isolated to staged Git index diff
security-review scan --staged --profile fast
```

Learn how to automate this automatically on every commit in the [Pre-Commit Hooks Guide](/guides/pre-commit-hooks).

---

### Path 3: CI SARIF Scan (GitHub Actions)

Generate a SARIF report and fail the build if High or Critical findings exist:

```sh
# Generate SARIF report and fail CI on High/Critical findings
security-review scan --ci --fail-on HIGH --format sarif --output results.sarif
```

See the complete [Five-Minute CI Setup](/getting-started/ci-setup) and [GitHub Actions Guide](/guides/ci-integrations).

---

## 6 Canonical Policy Domains

`security-review` classifies all rules into 6 distinct policy areas:

1. **Security (`security`)**: Hardcoded API keys, private keys, database credentials, SQL injection, and tainted dataflow sinks.
2. **Hardening (`hardening`)**: Insecure file permissions, weak TLS versions, and missing security headers.
3. **Reliability (`reliability`)**: Unhandled error returns, goroutine leaks, and missing context cancellations.
4. **Quality (`quality`)**: Dead code, anti-patterns, empty handlers, and dangerous type conversions.
5. **Supply Chain (`supply_chain`)**: Vulnerable dependencies and untrusted third-party packages.
6. **Governance (`governance`)**: License compliance, unresolved merge conflict markers, and architectural constraints.
