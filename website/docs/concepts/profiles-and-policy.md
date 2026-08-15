---
title: Profiles and Policy
description: "For policy owners: understand performance profiles, six canonical domains, severity thresholds, and commit gating."
---

# Profiles and Policy

Understand how `security-review` tunes scanner execution speed using performance profiles and enforces deterministic security gates using policy thresholds.

---

## Performance Profiles

Profiles configure active scanner capabilities and AST recursion limits to match specific workflow time budgets:

| Profile | Active capabilities | Performance expectation | Typical use |
| :--- | :--- | :---: | :--- |
| **`fast`** | High-confidence secret detection and fast rules | One-second benchmark budget for the staged fixture | Pre-commit Git hooks |
| **`standard`** | Secret, SAST, supply-chain, and governance rules | Repository-dependent | Pull-request validation |
| **`full`** | All rules, deeper AST and SQL taint analysis, and architecture checks | Repository-dependent | Nightly builds and release audits |
| **`frontend`** | Client rules, framework sanitizers, and import cycle detection | Repository-dependent | Frontend-focused audits |

Select a profile via the `--profile` CLI flag:

```sh
security-review scan --profile fast
```

---

## The 6 Canonical Policy Domains

Every security rule belongs to one of six canonical domains:

1. **`security`**: Hardcoded secrets, SQL injection, command execution, and authentication bypasses.
2. **`hardening`**: Defensive configurations, TLS settings, and security headers.
3. **`reliability`**: Goroutine leaks, unhandled errors, and context cancellations.
4. **`quality`**: Dead code, anti-patterns, and bad practices.
5. **`supply_chain`**: Dependency vulnerabilities and untrusted packages.
6. **`governance`**: License requirements, merge conflict markers, and architectural constraints.

---

## Severity Levels

`security-review` defines four strict severity levels:

```text
CRITICAL > HIGH > MEDIUM > LOW
```

---

## Policy Evaluation and Exit Codes

`security-review` operates as a strict commit and CI gate:

- **Local Interactive Scans**: By default, `security-review scan` prints findings to stdout, writes the report, and returns **exit code `0`** to support local exploration without breaking shell loops.
- **CI Enforcement**: Passing the **`--ci`** flag returns **exit code `1`** whenever active findings meet or exceed the `--fail-on` threshold.

```sh
# Fails CI if ANY active findings exist
security-review scan --ci

# Fails CI only if High or Critical findings exist
security-review scan --ci --fail-on HIGH
```
