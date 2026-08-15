---
title: Configuration Builder Contract
description: Architecture, state model, accessibility rules, and privacy contract for the client-side configuration generator.
---

# Configuration Builder Contract

Design specification and client-side architecture contract for the interactive configuration generator.

## Core Design Principles

1. **100% Client-Side Privacy**: All state management, JSON serialization, and clipboard operations occur purely within the browser DOM. No network requests, analytics, or telemetry are transmitted.
2. **Schema Version 1 Compliance**: Serialized output strictly matches Go's `config.Config` struct definition (`version: 1`).
3. **Preset-Driven**: Provides 9 pre-configured templates (`Minimal`, `Go Service`, `Frontend App`, `Monorepo`, `Staged Hook`, `Offline`, `Strict CI`, `External Scanner`, `Gradual Adoption`).
4. **Keyboard & Screen-Reader Accessible**: Built using standard HTML form controls, ARIA labels, and focus indicators.

## State Model

- `version`: `1` (fixed)
- `project`: `string` (default `"minimal-app"`, required non-empty)
- `root`: `string` (default `"."`, cannot contain `..`)
- `mode`: `"full"` | `"changed"` | `"staged"` (default `"full"`)
- `output`: `string` (default `"security_findings.json"`)
- `fail_on`: `"CRITICAL"` | `"HIGH"` | `"MEDIUM"` | `"LOW"` (default `"HIGH"`)
- `workers`: `number` (`1` to `64`, default `4`)
- `include_extensions`: `string[]`
- `exclude_directories`: `string[]`
- `exclude_files`: `string[]`
- `policy`: `map[string]string` (`security`, `reliability`, `hardening`, `quality`, `supply_chain`, `governance`)
- `frontend`: `{ enabled: boolean, frameworks: string[], client_roots: string[], server_roots: string[] }`
- `hooks`: `{ pre_commit: Hook, commit_msg: Hook, pre_push: Hook }`
- `cache`: `{ enabled: boolean, directory: string, max_age: string, max_bytes: number }`

## Client-Side Validation Rules

1. **Non-Empty Project Name**: The generator requires a non-empty project string.
2. **Safe Root Path**: Rejects paths containing path traversal (`..`).
3. **Valid Enum Ranges**: Enforces exact match for `mode` and `fail_on` severities.
4. **Bounded Worker Range**: Validates worker count within `1` to `64`.

## Export & Accessibility Capabilities

- **Copy to Clipboard**: One-click formatted JSON copy using `navigator.clipboard.writeText` with visual feedback and `aria-live="polite"` status announcements.
- **Download File**: Generates and downloads `security-review.json` locally using browser `Blob` object and object URL.
- **Preset Reset & Dirty Tracking**: Visual indicator for customized settings with one-click reset to preset defaults.
