# Troubleshooting Guide

Solutions for common operational issues when running `security-review` CLI or agent skills.

## Common Failures

### 1. `security-review` CLI not found
- **Symptom**: `Get-Command security-review` fails or script returns `blocked`.
- **Fix**: Build the local binary using `go build -o security-review.exe ./cmd/security-review` or add `go/bin` to your `PATH`.

### 2. Exit Code 1 (Policy Violation)
- **Symptom**: CLI returns exit code 1 during `--ci` run.
- **Fix**: Inspect generated report artifacts (`security_findings.json`). Apply parameterization fixes or create a baseline (`security-review baseline create`).

### 3. Git Hook installation conflict
- **Symptom**: `security-review hook install` fails because hook file already exists.
- **Fix**: Inspect `git config core.hooksPath` and ensure existing hook files are backed up before re-installing.
