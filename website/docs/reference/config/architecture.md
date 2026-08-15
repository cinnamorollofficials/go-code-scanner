---
title: Architecture Policy Configuration
description: "For configuration authors: look up layer isolation, forbidden dependency, and package cycle-detection fields."
---

# Architecture Policy Configuration

Define package layering rules, forbidden import paths between components, and package dependency cycle detection.

## Schema

```json
{
  "architecture": {
    "detect_cycles": true,
    "layers": [
      { "name": "cmd", "paths": ["cmd/*"] },
      { "name": "pkg", "paths": ["pkg/*"] }
    ],
    "forbidden_dependencies": [
      { "from": "pkg", "to": "cmd" }
    ]
  }
}
```

### Fields

- **`detect_cycles`** (`bool`): Enables circular package import detection across Go package graph.
- **`layers`** (`ArchitectureLayer[]`): Named architecture layer declarations and matching path globs.
- **`forbidden_dependencies`** (`ForbiddenDependency[]`): Disallowed import directions between layers (e.g. `pkg` importing `cmd`).
