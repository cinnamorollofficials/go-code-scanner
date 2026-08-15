---
title: Input and Path Configuration
description: "For configuration authors: look up the root, file extensions, exclusions, and file or line byte limits."
---

# Input and Path Configuration

Configure workspace discovery boundaries, file extension filtering, directory exclusions, and line/file size safety thresholds.

## Field Reference

### `root` (`string`)
- **Default**: `"."`
- **Description**: Target workspace directory for scanner execution. Relative paths resolve relative to the configuration file location.

### `include_extensions` (`string[]`)
- **Default**: `[".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts", ".html", ".vue", ".svelte", ".yaml", ".yml", ".json"]`
- **Description**: List of file extensions evaluated during discovery.

### `exclude_directories` (`string[]`)
- **Default**: `[".git", "node_modules", "vendor", "dist", "build", ".next", ".nuxt", ".svelte-kit", ".output", "out", "bin"]`
- **Description**: Directory names or paths ignored during recursive file traversal.

### `exclude_files` (`string[]`)
- **Default**: `["security_findings.json", "package-lock.json"]`
- **Description**: Specific filenames excluded from scan discovery.

### `rule_files` (`string[]`)
- **Default**: `[]`
- **Description**: Paths to custom external JSON rule definition files.

### `pattern_max_file_bytes` (`int64`)
- **Default**: `1048576` (1 MB)
- **Description**: Maximum file size in bytes evaluated by pattern-matching engines. Files exceeding this limit trigger partial scan warnings.
