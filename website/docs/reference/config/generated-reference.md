---
title: Configuration Field Reference
description: "For configuration authors: look up schema fields, types, defaults, and requirements generated from Go definitions."
---

# Configuration Field Reference

This page is automatically generated from Go struct definitions in `pkg/config`. Do not edit manually.

| Field Path | Type | Default | Required | Description |
| :--- | :--- | :--- | :--- | :--- |
| `baseline_file` | `string` | `` | No | Path to baseline snapshot JSON file. |
| `cache.directory` | `string` | `` | No | Path to local cache storage directory. |
| `cache.enabled` | `bool` | `false` | No | Enables local AST and scan caching. |
| `cache.max_age` | `string` | `` | No | Maximum cache entry retention duration string. |
| `cache.max_bytes` | `int64` | `0` | No | Maximum byte size threshold for local cache directory. |
| `exclude_directories` | `[]string` | `[.git node_modules vendor dist build .next .nuxt .svelte-kit .output out bin]` | No | Directory names or paths excluded during discovery. |
| `exclude_files` | `[]string` | `[security_findings.json package-lock.json]` | No | Filenames excluded during discovery. |
| `fail_on` | `string` | `CRITICAL` | No | Global severity threshold that triggers exit code 1. |
| `frontend.client_roots` | `[]string` | `[]` | No | Client-side root directory paths. |
| `frontend.enabled` | `bool` | `false` | No | Enables native frontend AST and pattern scanning. |
| `frontend.frameworks` | `[]string` | `[]` | No | Framework identifiers for frontend detection. |
| `frontend.server_roots` | `[]string` | `[]` | No | Server-side root directory paths. |
| `hooks.pre_commit.enabled` | `bool` | `false` | No | Enables pre-commit git hook execution. |
| `hooks.pre_commit.new_only` | `bool` | `false` | No | Evaluates pre-commit findings against baseline snapshot. |
| `hooks.pre_commit.profile` | `string` | `fast` | No | Performance profile used by pre-commit hook. |
| `hooks.pre_commit.staged_only` | `bool` | `false` | No | Restricts pre-commit scan to git staged index snapshot. |
| `include_extensions` | `[]string` | `[.go .ts .tsx .js .jsx .mjs .cjs .mts .cts .html .vue .svelte .yaml .yml .json]` | No | List of file extensions included during discovery. |
| `mode` | `string` | `full` | No | Default scan discovery mode. |
| `offline_profiles` | `[]string` | `[]` | No | List of profiles runnable offline without network access. |
| `output` | `string` | `security_findings.json` | No | Default report output filename. |
| `pattern_max_file_bytes` | `int64` | `1048576` | No | Maximum file size in bytes for pattern scanner. |
| `pattern_max_line_bytes` | `int` | `4096` | No | Maximum line buffer length in bytes for pattern scanner. |
| `profiles` | `map[string][]string` | `map[]` | No | Custom named profile scanner mappings. |
| `project` | `string` | `security-review` | No | Project or repository identifier. |
| `root` | `string` | `.` | No | Target root workspace directory. |
| `rule_files` | `[]string` | `[]` | No | External JSON rule file paths. |
| `scanners` | `map[string]Scanner` | `map[]` | No | Map of declared scanner definitions. |
| `suppression_file` | `string` | `` | No | Path to suppressions JSON file. |
| `version` | `int` | `1` | Yes | Configuration schema version (must be 1). |
| `workers` | `int` | `4` | No | Maximum concurrent worker goroutines. |
