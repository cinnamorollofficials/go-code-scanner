---
title: Guides Overview
description: Task-oriented integration, operational playbooks, and troubleshooting guides for security-review.
---

# Guides Overview

Use these guides when you already have a working `security-review` binary and
want to complete a specific repository or CI task.

| Guide | Primary audience | Outcome | Typical time |
| :--- | :--- | :--- | :---: |
| **[Pre-Commit Hooks](/guides/pre-commit-hooks)** | Developers and repository maintainers | An isolated staged-file check before every commit. | 5–10 min |
| **[GitHub Actions / GitLab CI](/guides/ci-integrations)** | CI maintainers | A pull-request policy gate and supported report upload. | 10–15 min |
| **[Gradual Adoption with Baselines](/guides/baselines)** | Teams introducing scanning to an existing codebase | Existing findings recorded while new findings are enforced. | 10–20 min |
| **[Managing Suppressions](/guides/suppressions)** | Developers and security reviewers | A reviewed, expiring exception stored in `.security-ignore`. | 5 min |
| **[Reproducing a Finding](/guides/reproducing-findings)** | Developers investigating a report | A minimal reproduction and evidence for remediation or suppression. | 10–30 min |
| **[Troubleshooting](/guides/troubleshooting)** | Developers and CI operators | A diagnosis based on the CLI exit code and scanner state. | As needed |

Need the first successful scan instead? Start in [Getting Started](/getting-started/).
Need exact syntax rather than a workflow? Open the [CLI Reference](/reference/cli)
or [Configuration Reference](/reference/configuration).
