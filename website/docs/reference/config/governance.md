---
title: Governance Policy Configuration
description: "For configuration authors: look up required files, license headers, ownership attribution, and ticket fields."
---

# Governance Policy Configuration

Enforce repository governance compliance, file existence, copyright header patterns, and code ownership rules.

## Schema

```json
{
  "governance": {
    "required_files": ["README.md", "LICENSE", "SECURITY.md"],
    "ownership_file": "CODEOWNERS",
    "required_headers": [
      {
        "id": "apache-2.0-header",
        "paths": ["*.go"],
        "pattern": "Copyright \\d{4} Security Review Team",
        "max_lines": 5,
        "severity": "MEDIUM"
      }
    ]
  }
}
```

### Fields

- **`required_files`** (`string[]`): Mandatory filenames that must exist in root repository directory.
- **`ownership_file`** (`string`): Path to code ownership file (e.g. `CODEOWNERS`).
- **`required_headers`** (`HeaderRule[]`): Array of header regex pattern rules enforcing copyright or license statements in source code.
