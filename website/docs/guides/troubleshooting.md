---
title: Troubleshooting
description: "For operators diagnosing failures: use exit codes and messages to resolve input, scanner, timeout, cache, and migration errors."
---

# Troubleshooting

Diagnose `security-review` failures from the process exit code, scanner status,
and error message. For rollout steps, use [Gradual Adoption with Baselines](/guides/baselines);
for reviewed exceptions, use [Managing Suppressions](/guides/suppressions).

## Start with the Exit Code

| Exit code | Meaning | First check |
| :---: | :--- | :--- |
| `1` | A CI policy threshold was violated, or a verification comparison failed. | Review active findings or verification mismatch output. |
| `2` | The command, flag, argument, or configuration is invalid. | Compare the command with `security-review help` and validate the configuration. |
| `3` | An operational action failed after valid input was accepted. | Inspect failed required scanners, file permissions, Git state, and cache/report writes. |

## Common Operational Failures

### 1. Missing External Scanner Executable (`Exit Code 3` when required)
- **Symptom**: `executable "govulncheck" not found in $PATH`
- **Cause**: A required scanner adapter is not installed and its `on_missing`
  behavior is `fail`.
- **Remediation**: Install the executable, or make the scanner optional and set
  `"on_missing": "skip"` when skipping is acceptable. Optional failures appear
  as warnings rather than changing the process exit code to `3`.

### 2. Scanner Timeout Exceeded
- **Symptom**: `scanner execution timed out after 30s`
- **Remediation**: Increase the scanner timeout limit in configuration (`"timeout": "2m"`) or use `--profile fast` in pre-commit hooks.

### 3. File Size Exceeded and Partial Scans
- **Symptom**: `file exceeds pattern_max_file_bytes limit`
- **Remediation**: Increase `"pattern_max_file_bytes"` in `security-review.json` or add generated vendor files to `"exclude_directories"`.

### 4. Cache Stale or Corrupted
- **Symptom**: Inconsistent scan results across local runs.
- **Remediation**: Purge local AST cache storage:
  ```sh
  security-review cache clean
  ```

### 5. Schema Migration and Upgrade Verification
- **Symptom**: Configuration decoding error after upgrading CLI.
- **Remediation**: Validate configuration syntax against current schema version:
  ```sh
  security-review config validate security-review.json
  security-review upgrade check
  ```

If a problem remains after these checks, capture the full command, exit code,
scanner status lines, tool version, operating system, and a redacted
configuration before opening an issue.
