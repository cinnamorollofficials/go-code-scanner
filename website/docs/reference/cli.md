---
title: CLI Command Reference
description: "For operators and automation authors: look up security-review commands, flags, defaults, outputs, and exit codes."
---

# CLI Command Reference

`security-review` provides command-line tools for security scanning, git hooks, baseline management, suppressions, and cache maintenance.

```text
security-review [command] [flags]
```

## Global Command Summary

| Command | Description |
| :--- | :--- |
| **[`scan`](#scan)** | Runs security scans against Go, configuration, and frontend source files. |
| **[`config`](#config)** | Validates `security-review.json` syntax and constraint rules (`config validate <path>`). |
| **[`hook`](#hook)** | Installs, inspects, runs, and uninstalls git hooks (`install`, `uninstall`, `status`, `run`). |
| **[`baseline`](#baseline)** | Manages finding baseline snapshots (`create`, `update`, `status`). |
| **[`suppress`](#suppress)** | Adds reviewed rule suppressions to `.security-ignore` (`add`). |
| **[`cache`](#cache)** | Inspects AST cache statistics and cleans cached scan artifacts (`stats`, `clean`). |
| **[`release`](#release)** | Verifies release artifacts, checksums, and provenance manifests. |
| **[`upgrade`](#upgrade)** | Checks for compatibility contract migrations (`upgrade check`). |
| **[`version`](#version)** | Prints binary version, commit SHA, and target architecture. |

---

## `scan`

Executes security scanning rules across project files. A full working-tree scan is performed by default when neither `--staged` nor `--changed` is supplied.

```sh
security-review scan [flags]
```

### Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--config <path>` | `string` | `""` | Path to JSON configuration file. |
| `--root <dir>` | `string` | `"."` | Target project root directory. |
| `--output <path>` | `string` | `""` | Path to write scan report artifact (defaults to `security_findings.json` in config). |
| `--changed` | `bool` | `false` | Restricts discovery to files modified relative to HEAD. |
| `--staged` | `bool` | `false` | Restricts discovery to files staged in Git index snapshot. |
| `--ci` | `bool` | `false` | Returns exit code `1` when active findings meet the failure threshold. |
| `--fail-on <sev>` | `string` | `""` | Minimum severity threshold (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`). Requires `--ci` to fail the process. |
| `--profile <name>` | `string` | `""` | Performance profile (`fast`, `standard`, `full`, `frontend`). |
| `--scope <name>` | `string` | `""` | Client scope filter (`client`, `server`, `all`). |
| `--format <type>` | `string` | `"json"` | Report artifact format (`json`, `sarif`, `junit`). Terminal output is printed unless `--quiet`. |
| `--baseline <path>`| `string` | `""` | Path to baseline JSON snapshot file. |
| `--new-only` | `bool` | `false` | Applies CI policy evaluation only to findings absent from baseline. |
| `--fix` | `bool` | `false` | Applies deterministic auto-fixes directly to source code (full scan mode only). |
| `--dry-run` | `bool` | `false` | Previews `--fix` modifications without writing files (requires `--fix`). |
| `--explain <rule>` | `string` | `""` | Prints remediation guidance for a specific rule ID and exits. |
| `--quiet` | `bool` | `false` | Suppresses terminal summary output. |
| `--verbose` | `bool` | `false` | Prints scanner execution timing, capabilities, and metadata. |
| `--color <mode>` | `string` | `"auto"` | Terminal color mode (`auto`, `always`, `never`). |

---

## `config`

Validates configuration syntax and field constraints against the canonical schema.

```sh
security-review config validate <path>
```

---

## `hook`

Manages git hook installation and pre-commit lifecycle.

```text
security-review hook <install|uninstall|status|run> [event] [--root <dir>]
```

Supported hook events: `pre-commit` (default), `commit-msg`, `pre-push`.

```sh
# Install pre-commit hook into .git/hooks/
security-review hook install pre-commit

# Inspect hook status and active installation state
security-review hook status pre-commit

# Execute pre-commit checks against staged index
security-review hook run pre-commit

# Remove managed hook cleanly
security-review hook uninstall pre-commit
```

---

## `baseline`

Manages finding baseline snapshots for gradual adoption in existing repositories.

```sh
# Generate a new baseline from a scan report
security-review baseline create --report security_findings.json [--baseline .security-baseline.json] [--dry-run]

# Update an existing baseline from a new scan report
security-review baseline update --report security_findings.json [--baseline .security-baseline.json] [--accept-resolved] [--dry-run]

# Check status of findings against baseline
security-review baseline status --report security_findings.json [--baseline .security-baseline.json]
```

---

## `suppress`

Adds reviewed rule suppressions to the suppression file (`.security-ignore`).

```sh
security-review suppress add --file <path> --reason <text> --expires <YYYY-MM-DD> [options]
```

### Options

- `--file <path>`: *(Required)* Finding file path to suppress.
- `--reason <text>`: Documented audit justification for suppression.
- `--expires <YYYY-MM-DD>`: Expiration date for the suppression.
- `--line <int>`: Target line number (`-1` for any line in file).
- `--rule <id>`: Specific rule ID to suppress.
- `--fingerprint <hash>`: Specific finding fingerprint.
- `--ticket <id>`: Security review or tracking ticket reference.
- `--approved-by <name>`: Approver identity.
- `--suppression-file <path>`: Suppression JSON file (default `.security-ignore`).
- `--dry-run`: Previews suppression change without writing to disk.

---

## `cache`

Inspects and purges local AST cache storage.

```sh
# View cache entry count and disk usage
security-review cache stats [--dir <path>]

# Purge cached ASTs and scan artifacts
security-review cache clean [--dir <path>]
```

---

## `release`

Creates and verifies release archives, checksum manifests, provenance, and
signatures. See the [Security Model](/security) for the complete trusted-release
workflow.

```sh
# Verify every archive named by SHA256SUMS
security-review release checksums verify --manifest SHA256SUMS --directory dist/

# Verify provenance subjects and an Ed25519 signature
security-review release verify \
  --provenance provenance.json \
  --signature provenance.sig \
  --public-key release-public.pem \
  --directory dist/
```

Additional release operations are `archive`, `provenance generate`,
`provenance sign`, and `changelog validate`.

---

## `upgrade`

Checks the current compatibility contract against a previously generated
contract. Exit code `1` means a migration is required.

```sh
security-review upgrade check [--contract <path>]
```

---

## `version`

Prints the version, commit, build date, Go version, target operating system, and
target architecture embedded in the binary.

```sh
security-review version
```

---

## Exit Codes

`security-review` returns deterministic exit codes:

- **`0`**: Scan succeeded without policy failures, or command executed successfully.
- **`1`**: Policy threshold violated when `--ci` is active (findings meet or exceed threshold), or release/contract verification mismatch.
- **`2`**: Invalid CLI flags, missing required arguments, or invalid configuration syntax.
- **`3`**: Operational failure (I/O error, file permission, git repository error, or cache failure).
