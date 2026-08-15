---
title: Pre-Commit Hooks Guide
description: Set up automated local git hooks with security-review for fast, isolated staged scanning.
---

# Pre-Commit Hooks Guide

Catch security vulnerabilities and credential leaks before code leaves your local workstation using `security-review`'s native Git hook automation.

## Quick Installation

Run the hook installer from inside your Git repository:

```sh
# Install pre-commit hook into .git/hooks/pre-commit
security-review hook install pre-commit
```

## How It Works

When installed, committing staged files (`git commit`) triggers:

```sh
security-review scan --staged --profile fast
```

### Key Safety Guarantees

1. **Git Index Isolation**: Scans strictly read blob contents staged in the Git index (`git add`). Unstaged modifications in your working directory are ignored, preventing false positives and false negatives.
2. **Sub-Second Performance**: The `--profile fast` flag optimizes scanner execution to complete in `< 1s` so developer workflow is never impeded.
3. **Deterministic Blocking**: If staged files contain findings meeting policy thresholds, the commit is aborted with exit code `1`.

---

## Verifying Hook Status

Inspect the current hook installation and repository state:

```sh
security-review hook status pre-commit
```

Example output:
```text
pre-commit: installed
```

---

## Running Hooks Manually

You can trigger hook validation manually against staged changes without creating a commit:

```sh
security-review hook run pre-commit
```

---

## Uninstalling the Hook

To remove the managed Git hook cleanly:

```sh
security-review hook uninstall pre-commit
```
