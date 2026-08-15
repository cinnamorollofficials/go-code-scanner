---
title: Cache Policy Configuration
description: Field reference for AST cache enablement, retention max-age, and storage byte limits.
---

# Cache Policy Configuration

Configure local AST and scan artifact caching parameters.

## Schema

```json
{
  "cache": {
    "enabled": true,
    "directory": ".go-code-scanner-cache",
    "max_age": "168h",
    "max_bytes": 104857600
  }
}
```

### Fields

- **`enabled`** (`bool`): Enables or disables AST caching.
- **`directory`** (`string`): Path to cache storage directory (relative to root or absolute).
- **`max_age`** (`string`): Maximum age for cached entries before eviction (e.g. `"24h"`, `"7d"`).
- **`max_bytes`** (`int64`): Maximum total byte size for cache directory before oldest entries are pruned.
