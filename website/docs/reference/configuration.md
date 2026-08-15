---
title: Configuration Reference
description: "For configuration authors: find schema version 1 entry points, validation behavior, and focused field references."
---

# Configuration Reference

`security-review` is configured via JSON files (`security-review.json` or `.security-review.json`).

## Schema Version 1

Every valid configuration file must declare `"version": 1`.

```json
{
  "version": 1,
  "project": "my-go-service",
  "mode": "full",
  "fail_on": "HIGH"
}
```

## Reference Navigation

Select a section below to view detailed field definitions, defaults, and constraints:

- **[Input & Paths](/reference/config/input-and-paths)**: Root directory, included extensions, excluded files/directories, and file byte limits.
- **[Profiles & Policy](/reference/config/profiles-and-policy)**: Performance profiles, domain policy thresholds, `--fail-on` severity gates.
- **[Scanner Definitions](/reference/config/scanners)**: Scanner declarations, custom command scanners, timeouts, and adapter configs.
- **[Git Hooks](/reference/config/hooks)**: Pre-commit, commit-msg, and pre-push hook behavior and validation rules.
- **[Frontend Policy](/reference/config/frontend)**: Client/server root classification, framework detection, and sanitizer rules.
- **[Supply Chain Policy](/reference/config/supply-chain)**: Dependency allowlist/denylist and open-source license policy checks.
- **[Governance Policy](/reference/config/governance)**: Required repository files, license header rules, ownership attribution, and ticket requirements.
- **[Architecture Policy](/reference/config/architecture)**: Package layering rules, forbidden import paths, and cycle detection.
- **[Cache Policy](/reference/config/cache)**: AST cache enablement, retention max-age, and storage byte limits.
