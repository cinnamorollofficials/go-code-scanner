---
title: Developer Workflow Features
description: Git hooks integration, local cache management, dry-run fixes, rule explanations, and utility commands.
---

# Developer Workflow Features

Explore developer workflow tools provided by `security-review`, including git hook lifecycle automation, cache management, safe auto-fixing, rule explanation, and config validation.

## Git Hooks Integration

`security-review` includes native git hook management to automate pre-commit scans across your development team.

```sh
# Install pre-commit hook in active repository
security-review hook install pre-commit

# Inspect hook status and active installation state
security-review hook status pre-commit

# Uninstall pre-commit hook cleanly
security-review hook uninstall pre-commit
```

::: tip Managed-File Safety Guarantee
Hook files created by `security-review` include a unique signature block. The scanner will never overwrite or modify custom pre-existing git hooks.
:::

---

## Local Cache Management

To accelerate repeated scans on large codebases, `security-review` caches package ASTs and intermediate analysis results in `.go-code-scanner-cache` (configurable via `--dir`).

```sh
# Inspect cache directory size and cached entry count
security-review cache stats

# Purge cached ASTs and scan artifacts
security-review cache clean
```

---

## Auto-Fixing & Dry Runs

For rules with deterministic remediation logic, `security-review` can automatically apply fixes:

```sh
# Preview auto-fix changes without modifying workspace files
security-review scan --fix --dry-run

# Apply deterministic fixes directly to source files
security-review scan --fix
```

---

## Utility Commands

### Rule Explanation

Get detailed remediation guidance and examples for any rule ID:

```sh
security-review scan --explain hardcoded-credential
```

### Configuration Validation

Verify syntax and constraint rules for custom `security-review.json` files:

```sh
security-review config validate .security-review.json
```
