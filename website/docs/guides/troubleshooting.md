---
title: Adoption & Troubleshooting Playbook
description: Operational playbook for gradual scanner adoption, baseline migration, and troubleshooting common CLI issues.
---

# Adoption & Troubleshooting Playbook

Operational guide for adopting `security-review` across engineering teams and diagnosing common execution errors.

## Gradual Adoption Strategy

When introducing `security-review` into an existing codebase, follow this 3-step gradual rollout plan to prevent blocking developer velocity:

### Step 1: Capture Legacy Baseline
Generate a baseline snapshot to record pre-existing technical debt:

```sh
# 1. Run full scan and save findings report
security-review scan --output security_findings.json

# 2. Create baseline snapshot from report
security-review baseline create --report security_findings.json --baseline .security-baseline.json
```

### Step 2: Enforce Policy on New Code Only
Configure your CI pipeline to fail **only** when *new* findings are introduced relative to the baseline snapshot:

```sh
security-review scan --ci --baseline .security-baseline.json --new-only
```

### Step 3: Progressive Remediation & Suppressions
Audit baseline issues over time. For legitimate false positives or risk-accepted exceptions, record explicit suppressions in `.security-ignore`:

```sh
security-review suppress add \
  --file pkg/secrets/fixture_test.go \
  --rule hardcoded-credential \
  --reason "Synthetic unit test fixture" \
  --expires 2026-12-31
```

---

## Troubleshooting Common Operational Failures

### 1. Missing External Scanner Executable (`Exit Code 2`)
- **Symptom**: `executable "govulncheck" not found in $PATH`
- **Cause**: A scanner adapter declared as `"required": true` in configuration is not installed on the system PATH.
- **Remediation**: Install the missing tool (`go install golang.org/x/vuln/cmd/govulncheck@latest`) or set `"required": false` in configuration.

### 2. Scanner Timeout Exceeded
- **Symptom**: `scanner execution timed out after 30s`
- **Remediation**: Increase the scanner timeout limit in configuration (`"timeout": "2m"`) or use `--profile fast` in pre-commit hooks.

### 3. File Size Exceeded & Partial Scans
- **Symptom**: `file exceeds pattern_max_file_bytes limit`
- **Remediation**: Increase `"pattern_max_file_bytes"` in `security-review.json` or add generated vendor files to `"exclude_directories"`.

### 4. Cache Stale or Corrupted
- **Symptom**: Inconsistent scan results across local runs.
- **Remediation**: Purge local AST cache storage:
  ```sh
  security-review cache clean
  ```

### 5. Schema Migration & Upgrade Verification
- **Symptom**: Configuration decoding error after upgrading CLI.
- **Remediation**: Validate configuration syntax against current schema version:
  ```sh
  security-review config validate security-review.json
  security-review upgrade check
  ```
