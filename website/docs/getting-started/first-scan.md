---
title: First Scan & Exit Codes
description: Run your first security scan, interpret terminal output, and understand exit code policy enforcement.
---

# First Scan & Exit Codes

Learn how to execute your first repository scan and understand how `security-review` enforces policy exit codes.

## Running Your First Scan

Navigate to any Go repository or project directory and execute:

```sh
security-review scan
```

The scanner automatically discovers Go packages, configuration files, supply chain dependencies, and frontend assets.

## Interpreting Scan Results
 
Scan output displays scanner execution states, aggregate summary, and finding details:

This example was captured from a minimal Go fixture using the default scan
settings. Only the temporary absolute report path has been shortened.

```text
Code review: security-review (full)
  scanner pattern          clean
Findings: 0 | critical=0 high=0 medium=0 low=0 | suppressed=0 stale=0
Report: /path/to/my-project/security_findings.json
```
 
## Exit Code Behavior & Policy Enforcement
 
`security-review` uses deterministic exit codes:
 
- **`0`**: Scan completed without violating enforced CI policy thresholds, or run in local mode without `--ci`.
- **`1`**: Policy threshold violated when `--ci` is active (findings meet or exceed `--fail-on` threshold).
- **`2`**: Invalid CLI flags, missing required arguments, or invalid configuration.
- **`3`**: Operational failure (I/O error, file permission, git repository error, or cache failure).
 
::: important Default Behavior vs. CI Policy Enforcement
By default in local interactive mode, running `security-review scan` prints findings to stdout, writes the report artifact, and returns **exit code `0`** to avoid breaking local developer workflows.
 
To enforce strict failure in CI/CD pipelines, pass the **`--ci`** flag:
 
```sh
# Fails with exit code 1 if ANY active findings exist
security-review scan --ci

# Fails with exit code 1 ONLY if High or Critical findings exist
security-review scan --ci --fail-on HIGH
```
Note: `--fail-on` configures the severity threshold, but process exit code `1` still requires `--ci`.
:::
 
## Next Steps
 
- Explore [Scan Execution & Policy](/features/scan-execution-and-policy) to configure scan modes (staged, full, changed).
- Configure suppressions and baselines in [Reports & Finding Lifecycle](/features/reports-and-finding-lifecycle).
- View complete command flag details in the [CLI Reference](/reference/cli).
