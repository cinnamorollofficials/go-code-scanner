---
title: Five-Minute CI Setup
description: Fast track guide to adding security-review commit and pull request gates to GitHub Actions and GitLab CI.
---

# Five-Minute CI Setup

Add automated security analysis and commit gates to your CI/CD pipelines in under 5 minutes.

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
      - name: Checkout Code
        uses: actions/checkout@v4
        with:
          persist-credentials: false

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run Go Code Scanner
        run: |
          go run ./cmd/security-review scan \
            --ci \
            --fail-on HIGH \
            --format sarif \
            --output results.sarif

      - name: Upload SARIF to GitHub Security Tab
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: results.sarif
```

::: tip Exit Code & Gating
- Running with `--ci` returns **exit code `1`** if active findings meet or exceed `--fail-on HIGH` (or your configured policy).
- Results are automatically published to your repository's **Security > Code scanning** alerts tab.
:::

---

## GitLab CI/CD

Add the following stage to your `.gitlab-ci.yml`:

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
