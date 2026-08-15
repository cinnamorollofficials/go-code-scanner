---
title: Features Overview
description: Overview of core scanner execution features, report formats, workflows, and client scanning capabilities.
---

# Features Overview

`security-review` offers comprehensive policy-driven, offline-first security scanning capabilities for Go repositories.

## Feature Concepts & Guides

- **[Scan Modes and Isolation](/concepts/scan-modes)**: Full, changed, and staged discovery modes with Git index isolation.
- **[Profiles and Policy](/concepts/profiles-and-policy)**: Performance profiles, 6 canonical domains, and `--fail-on` thresholds.
- **[Reports & Finding Lifecycle](/concepts/reports-and-finding-lifecycle)**: Output formats, deterministic fingerprints, and lifecycle states.
- **[Frontend Scanning](/concepts/frontend-scanning)**: Framework detection (React, Vue, Svelte, Next.js), DOM sanitization, and client security.
- **[AST & SQL Taint Analysis](/concepts/sql-taint-analysis)**: Native AST parsing, interprocedural taint propagation, and auto-fixes.
- **[Developer Workflow Features](/features/developer-workflow-features)**: Git hooks, cache management, dry-run fixes, and rule explanations.
