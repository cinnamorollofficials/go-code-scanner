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
- `project`: `string` (default `"my-app"`)
- `root`: `string` (default `"."`)
- `mode`: `"full"` | `"changed"` | `"staged"` (default `"full"`)
- `output`: `string` (default `"security_findings.json"`)
- `fail_on`: `"CRITICAL"` | `"HIGH"` | `"MEDIUM"` | `"LOW"` | `"INFO"`
- `workers`: `number` (default `4`)
- `include_extensions`: `string[]`
- `exclude_directories`: `string[]`
- `exclude_files`: `string[]`
- `frontend`: `{ enabled: boolean, frameworks: string[], client_roots: string[], server_roots: string[] }`
- `cache`: `{ enabled: boolean, directory: string, max_age: string, max_bytes: number }`

## Export Capabilities

- **Copy to Clipboard**: One-click formatted JSON copy using `navigator.clipboard.writeText`.
- **Download File**: Downloads `security-review.json` locally using `Blob` URL.
