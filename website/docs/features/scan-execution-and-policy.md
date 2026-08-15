---
title: Scan Execution & Policy
description: Deep dive into discovery modes, scope isolation, performance profiles, and security policy thresholds.
---

# Scan Execution & Policy

Learn how `security-review` discovers files, isolates index snapshots, applies performance profiles, and enforces severity thresholds.

## Discovery Modes

| Mode | Flag | Description | Typical Use Case |
| :--- | :--- | :--- | :--- |
| **Full** | *(default)* | Scans all recognized files across the entire workspace directory when neither `--staged` nor `--changed` is supplied. | Nightly CI, baseline generation, release audits. |
| **Changed** | `--changed` | Scans files modified relative to git target branch or uncommitted workspace diffs. | Pull Request (PR) validation builds. |
| **Staged** | `--staged` | Scans only files staged in the git index (`git add`). Operates in strict index isolation. | Pre-commit git hooks. |

::: important Git Index Isolation Guarantee
When running with `--staged`, `security-review` materializes a temporary snapshot directly from `git index` objects rather than reading working tree files. This guarantees that unstaged edits in your working directory will never cause false positives or false negatives in pre-commit checks.
:::

---

## Scan Scope

Scope filters target specific architecture components:

- **`all`** *(default)*: Analyzes Go backend packages, configuration files, supply chain dependencies, and frontend assets.
- **`server`**: Limits execution to Go source files (`*.go`), Go module files (`go.mod`, `go.sum`), and backend configurations.
- **`client`**: Limits execution to frontend applications (React, Vue, Svelte, Next.js, Nuxt) and client static assets.

```sh
# Scan only Go backend components
security-review scan --scope server

# Scan only frontend application assets
security-review scan --scope client
```

---

## Performance Profiles

Profiles tune active rule sets, AST depth, and scanner timeout limits:

| Profile | Active Rule Capabilities | Speed Target | Use Case |
| :--- | :--- | :--- | :--- |
| **`fast`** | High-confidence secret detection and fast rules | `< 1s` | Pre-commit hooks |
| **`standard`** | Secret + SAST + Supply Chain + Governance rules | `< 5s` | PR validation builds |
| **`full`** | All rules + Deep AST & SQL taint analysis + Architecture checks | Thorough | Release certification |
| **`frontend`** | Client rules + Framework sanitizers + Import cycle detection | Fast | Frontend audits |

---

## Policy Thresholds & Failure Controls

### `--fail-on <SEVERITY>`

Defines the minimum severity required to trigger exit code `1` when `--ci` is enabled:

```sh
# Fails build only if Critical or High severity findings are found
security-review scan --ci --fail-on HIGH
```

Severity hierarchy: `CRITICAL` > `HIGH` > `MEDIUM` > `LOW`.

Note: `--fail-on` configures the severity threshold, but process exit code `1` still requires `--ci`.

### `--new-only`

Ignores existing baseline findings and fails execution only if **new** findings are introduced by the current commit or PR diff:

```sh
security-review scan --ci --baseline .security-baseline.json --new-only
```
