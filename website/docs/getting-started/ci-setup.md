---
title: Five-Minute CI Setup
description: "For CI maintainers: add a security-review pull-request gate to GitHub Actions or GitLab CI."
---

# Five-Minute CI Setup

Use this short setup when the repository contains the scanner source and a Go
toolchain is already available. The jobs are illustrative; adapt permissions,
action pins, and artifact retention to your repository policy.

## GitHub Actions

Create `.github/workflows/security.yml` in your repository:

```yaml
name: Security Review Gate

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

permissions:
  contents: read
  security-events: write

jobs:
  security-scan:
    name: Security Analysis
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false

      - name: Set up Go
        uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version-file: go.mod

      - name: Run Go Code Scanner
        run: |
          go run ./cmd/security-review scan \
            --ci \
            --fail-on HIGH \
            --format sarif \
            --output results.sarif

      - name: Upload SARIF to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@df409f7d9260372aa5c192dd7520b22a00c6d2d4 # v3.28.10
        if: always()
        with:
          sarif_file: results.sarif
```

::: tip Exit Codes and Gating
- Running with `--ci` returns **exit code `1`** if active findings meet or exceed `--fail-on HIGH` (or your configured policy).
- The upload step publishes a successfully generated SARIF file to your repository's **Security > Code scanning** alerts tab.
:::

---

## GitLab CI/CD

The following illustrative stage publishes JUnit results from a repository that
contains the scanner source:

```yaml
stages:
  - test

security_review:
  stage: test
  image: golang:1.25
  script:
    - go run ./cmd/security-review scan --ci --fail-on HIGH --format junit --output junit-report.xml
  artifacts:
    reports:
      junit: junit-report.xml
    when: always
```

---

## Next Steps

- Check out [Pre-Commit Hooks](/guides/pre-commit-hooks) to catch vulnerabilities before commits are pushed.
- Set up [Gradual Adoption with Baselines](/guides/baselines) for existing repositories with legacy debt.
- Customize security thresholds in [Configuration Reference](/reference/configuration).
