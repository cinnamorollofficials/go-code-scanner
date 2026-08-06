---
title: Local & CI Integrations Guide
description: Step-by-step integration guide for pre-commit hooks, GitHub Actions, GitLab CI, SARIF upload, and offline CI pipelines.
---

# Local & CI Integrations Guide

Integrate `security-review` seamlessly into local developer hooks and continuous integration (CI/CD) pipelines.

## GitHub Actions Integration

Use GitHub Actions to automatically scan pull requests and upload SARIF security results directly to GitHub's **Security > Code Scanning** tab.

```yaml
name: Security Audit

on:
  push:
    branches: [ main ]
  pull_request:

permissions:
  contents: read
  security-events: write

jobs:
  scan:
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

      - name: Run security-review scanner
        run: |
          go run ./cmd/security-review scan --ci --fail-on HIGH --format sarif --output results.sarif

      - name: Upload SARIF report
        uses: github/codeql-action/upload-sarif@df409f7d9260372aa5c192dd7520b22a00c6d2d4 # v3.28.10
        if: always()
        with:
          sarif_file: results.sarif
```

---

## GitLab CI Integration

```yaml
security_review:
  stage: test
  image: golang:1.22
  script:
    - go run ./cmd/security-review scan --ci --fail-on HIGH --format junit --output junit-report.xml
  artifacts:
    reports:
      junit: junit-report.xml
    when: always
```

---

## Pre-Commit Git Hook Integration

Automate sub-second local checks before every git commit:

```sh
# Install hook into .git/hooks/pre-commit
security-review hook install
```

When installed, committing staged changes triggers:

```sh
security-review scan --mode staged --profile fast
```
