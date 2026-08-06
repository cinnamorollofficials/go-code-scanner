# Integration Contract

## CLI Invocation & Commands

The `security-review` CLI is the primary execution binary for running security scans, managing baselines, and configuring Git hooks.

### Core Commands

| Command | Usage | Description |
| :--- | :--- | :--- |
| `security-review scan` | `security-review scan [--config <file>] [--profile <name>] [--staged]` | Executes offline or profile-based security review scan. |
| `security-review config validate` | `security-review config validate <file>` | Validates configuration syntax and domain policy. |
| `security-review baseline create` | `security-review baseline create --report <file> --baseline <file>` | Generates a new baseline file from scan report. |
| `security-review baseline update` | `security-review baseline update --report <file> --baseline <file> [--accept-resolved]` | Updates baseline entries with resolved or new items. |
| `security-review suppress add` | `security-review suppress add --file <path> --rule <id> --reason <text> --expires <date>` | Adds a reviewed suppression record. |
| `security-review hook install` | `security-review hook install pre-commit --root .` | Installs Git pre-commit hook. |

### Exit Codes

| Exit Code | Meaning | Action required |
| :---: | :--- | :--- |
| `0` | Success / Policy Passed | No action required. |
| `1` | Policy Mismatch / Finding Threshold Reached | Triage findings, resolve security issues or accept baselines. |
| `2` | Argument or Configuration Error | Fix invalid configuration or missing CLI options. |
| `3` | Operational Error | Check file permissions, unreadable input files, or binary PATH. |
