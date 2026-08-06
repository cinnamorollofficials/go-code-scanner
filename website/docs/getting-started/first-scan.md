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

Scan output is structured by severity and domain:

```
[HIGH] secret/gcp-api-key: GCP API Key identified in config/secrets.go:12
[MEDIUM] sast/sql-concat: Unsanitized SQL query string concatenation in pkg/db/user.go:45
[LOW] governance/missing-license: LICENSE file missing in repository root

Scan Summary: 3 finding(s) across 3 domain(s) [1 High, 1 Medium, 1 Low]
```

## Exit Code Behavior & Policy Enforcement

`security-review` uses deterministic exit codes:

- **`0`**: Scan succeeded without violating enforced policy thresholds.
- **`1`**: Policy threshold violated (e.g. `--ci` flag active, or `--fail-on` threshold met).
- **`2`**: Invalid CLI flags, missing configuration file, or execution failure.

::: important Default Behavior vs. CI Policy Enforcement
By default in local interactive mode, running `security-review scan` prints findings to terminal stdout and returns **exit code `0`** to avoid blocking interactive local iteration.

To enforce strict blocking in CI/CD pipelines, pass the **`--ci`** flag or specify **`--fail-on HIGH`**:

```sh
# Enforces exit code 1 if ANY findings exist
security-review scan --ci

# Enforces exit code 1 ONLY if High or Critical findings exist
security-review scan --fail-on HIGH
```
:::

## Next Steps

- Explore [Scan Execution & Policy](/features/) to configure scan modes (staged, full, changed).
- Configure suppressions and baselines in [Reports & Finding Lifecycle](/features/).
- View complete command flag details in the [CLI Reference](/reference/cli).
