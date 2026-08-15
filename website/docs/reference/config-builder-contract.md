---
title: Config Builder Contract
description: "For Config Builder maintainers: understand its state, validation boundary, accessibility, export, and privacy contract."
---

# Config Builder Contract

Design specification and client-side architecture contract for the interactive configuration generator.

## Core Design Principles

1. **100% Client-Side Privacy**: All state management, JSON serialization, and clipboard operations occur purely within the browser DOM. No network requests, analytics, or telemetry are transmitted.
2. **Schema Version 1 Compliance**: Serialized output uses `version: 1`; every shipped preset is tested with the real Go configuration validator.
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
- `baseline_file`: relative path to the baseline snapshot when gradual adoption is selected
- `profiles`: map of profile names to scanner ID lists
- `offline_profiles`: list of profiles declared safe for offline execution
- `scanners`: map of supported command or adapter scanner declarations

## Client-Side Validation Rules

1. **Non-Empty Project Name**: The generator requires a non-empty project string.
2. **Safe Paths**: Rejects parent-directory segments in root, output, cache, frontend, and exclusion paths.
3. **Valid Enum Ranges**: Enforces exact values for scan modes, severities, canonical domains, scanner types, adapters, workspaces, and missing-tool behavior.
4. **Bounded Worker Range**: Validates integer worker count within `1` to `64`.
5. **Nested Preset Data**: Checks policy, cache, frontend, scanner, profile, and offline-profile structures before export.
6. **Authoritative Validation**: Browser checks provide fast feedback but do not replace `security-review config validate` for user-edited files.

## Export and Accessibility Capabilities

- **Copy to Clipboard**: One-click formatted JSON copy using `navigator.clipboard.writeText` with visual feedback and `aria-live="polite"` status announcements.
- **Download File**: Generates and downloads `security-review.json` locally using browser `Blob` object and object URL.
- **Preset Reset & Dirty Tracking**: Visual indicator for customized settings with one-click reset to preset defaults.
- **Destructive Change Confirmation**: Changing presets while edits are present requires confirmation and restores the active selection when canceled.

## Automated Preset Contract

`npm run docs:check-presets` imports the same preset objects used by the Vue
component, runs browser-side validation, builds the current CLI, and validates
all nine exported JSON files through `security-review config validate`.
