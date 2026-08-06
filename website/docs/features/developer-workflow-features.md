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
security-review hooks install

# Inspect hook status and core.hooksPath configuration
security-review hooks status

# Uninstall hook cleanly
security-review hooks uninstall
```

::: tip Managed-File Safety Guarantee
Hook files created by `security-review` include a unique signature block. The scanner will never overwrite or modify custom pre-existing git hooks unless explicitly forced with `--overwrite`.
:::

---

## Local Cache Management

To accelerate repeated scans on large codebases, `security-review` caches package ASTs and intermediate analysis results in `$HOME/.cache/security-review` (or `%LOCALAPPDATA%\security-review`).

```sh
# Inspect cache directory size and cached entry count
security-review cache status

# Purge cached ASTs and scan artifacts
security-review cache clear
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
security-review explain secret/gcp-api-key
```

### Configuration Validation

Verify syntax and constraint rules for custom `security-review.json` files:

```sh
security-review config validate .security-review.json
```
