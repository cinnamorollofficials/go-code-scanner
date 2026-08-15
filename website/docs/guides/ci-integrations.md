---
title: CI Integration Examples
description: "For CI maintainers: adapt illustrative GitHub Actions and GitLab CI jobs for policy gates, reports, and offline profiles."
---

# CI Integration Examples

Use these examples when adding `security-review` to an existing pipeline. They
assume the repository contains the Go module and scanner source; replace the
`go run` command with your pinned binary installation when scanning another
project.

## GitHub Actions Integration

This illustrative job scans pushes and pull requests, then uploads its SARIF
artifact to GitHub Code Scanning. Review action versions and permissions against
your repository policy before copying it.

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

## GitLab CI Integration

This illustrative job publishes JUnit output to the GitLab test report view:

```yaml
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

## Offline CI

Use a profile declared in `profiles` and listed in `offline_profiles` when the
runner must not invoke network-requiring scanners. This illustrative command
assumes `security-review.json` defines the `offline` profile; its scanners must
still be installed in the runner image:

```sh
security-review scan --config security-review.json --ci --profile offline --fail-on HIGH
```

See [Profiles and Policy](/concepts/profiles-and-policy) for the execution model
and [Scanner and Adapter Compatibility](/reference/scanners) for network
requirements.

## Local Git Hooks

Local hook installation and staged-index behavior have a separate canonical
procedure. Follow [Pre-Commit Hooks](/guides/pre-commit-hooks) instead of copying
CI setup into a developer workstation.
