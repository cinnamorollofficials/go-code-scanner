---
title: Managing Suppressions
description: Register, audit, and manage reviewed false-positive and risk-accepted rule exceptions with security-review.
---

# Managing Suppressions

A **suppression** is an explicit, audited exception for a false-positive detection or an approved business risk acceptance.

---

## Suppressions vs. Baselines

- **Baselines (`.security-baseline.json`)**: Temporary snapshot of legacy findings used for gradual adoption.
- **Suppressions (`.security-ignore`)**: Reviewed, accountable exceptions for specific files or rules with required audit justifications.

---

## Adding a Suppression

Use the `security-review suppress add` CLI command:

```sh
security-review suppress add \
  --file "testdata/mock_credentials.go" \
  --rule "hardcoded-credential" \
  --reason "Synthetic mock key for integration test fixture" \
  --approved-by "appsec-team" \
  --expires "2026-12-31"
```

### Supported Options

| Flag | Description | Required? |
| :--- | :--- | :---: |
| `--file <path>` | Relative file path of the finding. | **Yes** |
| `--reason <text>` | Documented business or technical justification. | Optional |
| `--expires <YYYY-MM-DD>` | Expiration date for temporary risk acceptance. | Optional |
| `--rule <id>` | Specific rule ID to suppress (e.g. `SQLI-001`). | Optional |
| `--line <int>` | Target line number (`-1` for entire file). | Optional |
| `--fingerprint <hash>` | Deterministic finding SHA-256 fingerprint. | Optional |
| `--ticket <id>` | Tracking ticket reference (e.g. `SEC-1234`). | Optional |
| `--approved-by <name>` | Security reviewer or engineering lead identity. | Optional |
| `--suppression-file <path>` | Path to suppression file (defaults to `.security-ignore`). | Optional |
| `--dry-run` | Preview suppression addition without modifying file. | Optional |

---

## Storage Schema (`.security-ignore`)

Suppressions are recorded as JSON in `.security-ignore`:

```json
{
  "version": 1,
  "suppressions": [
    {
      "rule_id": "hardcoded-credential",
      "file": "testdata/mock_credentials.go",
      "line": -1,
      "reason": "Synthetic mock key for integration test fixture",
      "approved_by": "appsec-team",
      "expires": "2026-12-31"
    }
  ]
}
```

---

## Inline Code Comments

For localized exceptions, you can also annotate source code directly using inline comments:

```go
// nolint:hardcoded-credential // synthetic fixture
const mockAPIKey = "AKIAEXAMPLEKEY123456"
```
