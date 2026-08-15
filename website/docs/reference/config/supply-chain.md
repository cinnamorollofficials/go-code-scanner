---
title: Supply Chain Policy Configuration
description: "For configuration authors: look up dependency audit, license allowlist, and package denylist policy fields."
---

# Supply Chain Policy Configuration

Configure dependency auditing, license compliance, and package denylist policies.

## Schema

```json
{
  "supply_chain": {
    "dependency_denylist": ["crypto/md5"],
    "license_allowlist": ["MIT", "Apache-2.0", "BSD-3-Clause"],
    "license_denylist": ["GPL-3.0-only", "AGPL-3.0-only"]
  }
}
```

### Fields

- **`dependency_allowlist`** (`string[]`): Explicit list of allowed external module paths.
- **`dependency_denylist`** (`string[]`): Forbidden module import patterns or vulnerable packages.
- **`license_allowlist`** (`string[]`): Approved SPDX license identifiers for third-party dependencies.
- **`license_denylist`** (`string[]`): Prohibited SPDX license identifiers.
