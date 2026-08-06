---
title: Features Overview
description: Overview of core scanner execution features, report formats, workflows, and client scanning capabilities.
---

# Features Overview

`security-review` offers comprehensive policy-driven, offline-first security scanning capabilities for Go repositories.

## Feature Guides

- **[Scan Execution & Policy](/features/scan-execution-and-policy)**: Full, changed, and staged discovery modes, index isolation, scope filtering, performance profiles, and `--fail-on` policy thresholds.
- **[Reports & Finding Lifecycle](/features/reports-and-finding-lifecycle)**: Terminal, JSON, SARIF, and JUnit output formats, deterministic fingerprints, baselines, and suppressions.
- **[How It Works & Reproducing Findings](/features/analysis-and-reproduction)**: Multi-engine detection architecture (Go AST, Pattern, Frontend, Supply Chain) and step-by-step finding reproduction playbook.
- **[Developer Workflow Features](/features/developer-workflow-features)**: Git hooks lifecycle, local cache management, dry-run auto-fixes, rule explanations, and config validation.
- **[Frontend & Client Scanning](/features/client-scanning)**: Framework detection (React, Vue, Svelte), threat boundaries, DOM injection sanitization, and secret exposure prevention.
