---
title: Scan Modes and Isolation
description: "For developers and policy owners: compare full workspace, changed-file, and isolated staged Git index scans."
---

# Scan Modes and Isolation

`security-review` supports three distinct discovery modes to balance scanning speed with comprehensive coverage.

---

## Discovery Modes

| Mode | Trigger | Description | Typical Use Case |
| :--- | :--- | :--- | :--- |
| **Full** | Default (no flags) | Scans all recognized files across the entire workspace directory. | Nightly CI, baseline generation, release audits. |
| **Changed** | `--changed` | Scans only files modified relative to Git `HEAD` or uncommitted workspace diffs. | Pull Request (PR) validation builds. |
| **Staged** | `--staged` | Scans only files staged in the Git index (`git add`). | Pre-commit Git hooks. |

---

## Git Index Isolation Guarantee

When running with **`--staged`**, `security-review` does not read files directly from the working tree. Instead, it:

1. Queries the Git index for staged blob objects using NUL-delimited commands (`git diff-index --cached -z`).
2. Materializes a clean temporary snapshot directly from indexed Git object blobs.
3. Executes active scanners against the isolated snapshot.
4. Cleans up temporary artifacts immediately upon scan completion.

::: tip Why Isolation Matters
This strict index isolation guarantees that unstaged edits, WIP modifications, or temporary debug logs in your working tree will **never** cause false positives or false negatives during pre-commit checks.
:::

---

## Scan Scope Filtering

In addition to discovery modes, you can restrict analysis by application layer using the `--scope` flag:

- **`all`** *(default)*: Analyzes Go backend, configuration, dependencies, and frontend assets.
- **`server`**: Limits execution to Go source files (`*.go`), module files (`go.mod`, `go.sum`), and backend configurations.
- **`client`**: Limits execution to frontend applications (React, Vue, Svelte, Next.js, Nuxt) and static client assets.

```sh
# Scan only backend Go code
security-review scan --scope server

# Scan only client assets
security-review scan --scope client
```
