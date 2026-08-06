---
title: CLI Command Reference
description: Complete usage reference for all security-review CLI commands, flags, defaults, and exit codes.
---

# CLI Command Reference

`security-review` provides command-line tools for security scanning, git hooks, baseline management, suppressions, and cache maintenance.

```sh
security-review [command] [flags]
```

## Global Command Summary

| Command | Description |
| :--- | :--- |
| **[`scan`](#scan)** | Runs security scans against Go, configuration, and frontend source files. |
| **[`config`](#config)** | Validates `security-review.json` syntax and constraint rules. |
| **[`hook`](#hook)** | Installs, inspects, runs, and uninstalls git hooks (`pre-commit`, `commit-msg`, `pre-push`). |
| **[`baseline`](#baseline)** | Manages finding baseline snapshots (`create`, `check`). |
| **[`suppress`](#suppress)** | Manages rule suppression rules (`add`, `list`, `check`). |
| **[`cache`](#cache)** | Inspects AST cache status and clears cached scan artifacts (`status`, `clear`). |
| **[`release`](#release)** | Verifies release builds and produces release notes. |
| **[`upgrade`](#upgrade)** | Checks for available software updates. |
| **[`version`](#version)** | Prints binary version, commit SHA, and target architecture. |

---

## `scan`

Executes security scanning rules across project files.

```sh
security-review scan [flags]
```

### Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--config <path>` | `string` | `""` | Path to JSON configuration file. |
| `--root <dir>` | `string` | `"."` | Target project root directory. |
| `--output <path>` | `string` | `""` | Path to write scan report output. |
| `--changed` | `bool` | `false` | Restricts discovery to files modified relative to HEAD. |
| `--staged` | `bool` | `false` | Restricts discovery to files staged in Git index snapshot. |
| `--ci` | `bool` | `false` | Enforces policy threshold exit code `1` on findings. |
| `--fail-on <sev>` | `string` | `""` | Minimum severity threshold (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFO`). |
| `--profile <name>` | `string` | `""` | Performance profile (`fast`, `standard`, `full`, `frontend`). |
| `--scope <name>` | `string` | `""` | Client scope filter (`client`, `server`, `all`). |
| `--format <type>` | `string` | `"json"` | Report format (`json`, `sarif`, `junit`, `terminal`). |
| `--baseline <path>`| `string` | `""` | Path to baseline JSON snapshot file. |
| `--new-only` | `bool` | `false` | Applies CI policy failure only to findings absent from baseline. |
| `--fix` | `bool` | `false` | Applies deterministic auto-fixes directly to source code. |
| `--dry-run` | `bool` | `false` | Previews `--fix` modifications without writing files. |
| `--explain <rule>` | `string` | `""` | Prints remediation guidance for a specific rule ID and exits. |
| `--quiet` | `bool` | `false` | Suppresses terminal summary output. |
| `--verbose` | `bool` | `false` | Prints scanner execution timing and metadata. |
| `--color <mode>` | `string` | `"auto"` | Terminal color mode (`auto`, `always`, `never`). |

---

## `config`

Validates configuration syntax and field constraints.

```sh
security-review config validate <file>
```

---

## `hook`

Manages git hook installation and pre-commit lifecycle.

```sh
security-review hook <subcommand>
```

### Subcommands

- **`install`**: Installs `security-review` pre-commit hook into `.git/hooks/`.
- **`status`**: Displays hook installation state and active `core.hooksPath`.
- **`run`**: Manually executes pre-commit hook logic against staged index.
- **`uninstall`**: Removes `security-review` managed git hook cleanly.

---

## `baseline`

Manages legacy finding baseline snapshots.

```sh
# Generate baseline snapshot
security-review baseline create --output .security-baseline.json

# Check baseline validity
security-review baseline check .security-baseline.json
```

---

## `suppress`

Manages false-positive and risk-accepted rule suppressions.

```sh
security-review suppress add --rule hardcoded-credential --reason "Fixture"
security-review suppress list
security-review suppress check
```

---

## `cache`

Inspects and purges local AST cache storage.

```sh
security-review cache status
security-review cache clear
```

---

## `release` & `upgrade`

```sh
# Check available version upgrades
security-review upgrade --check

# Print version and build details
security-review version
```

---

## Exit Codes

- **`0`**: Scan/command succeeded without policy failures.
- **`1`**: Findings present meeting or exceeding enforced `--ci` / `--fail-on` policy thresholds.
- **`2`**: Invalid CLI flags, missing required tool executables, or fatal configuration error.
