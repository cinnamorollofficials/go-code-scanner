---
title: Frontend Policy Configuration
description: Field reference for client/server roots, framework detection, and DOM sanitizer rules.
---

# Frontend Policy Configuration

Configure frontend framework detection, boundary isolation, and sanitizer recognition rules.

## Schema

```json
{
  "frontend": {
    "enabled": true,
    "frameworks": ["react", "vue", "svelte", "nextjs", "nuxt", "sveltekit"],
    "client_roots": ["src/client", "app"],
    "server_roots": ["src/server", "api"],
    "recognize_sanitizers": ["DOMPurify.sanitize", "sanitizeHtml"],
    "detect_import_cycles": true,
    "detect_client_server_boundaries": true
  }
}
```

### Fields

- **`enabled`** (`bool`): Enables native frontend AST and pattern scanning.
- **`frameworks`** (`string[]`): Ecosystem frameworks allowed during auto-detection.
- **`client_roots`** (`string[]`): Directory paths designated exclusively as client-side code.
- **`server_roots`** (`string[]`): Directory paths designated exclusively as server-side code.
- **`recognize_sanitizers`** (`string[]`): Custom function names recognized as safe HTML sanitizers.
- **`detect_import_cycles`** (`bool`): Flag enabling circular import detection across JavaScript/TypeScript module graphs.
- **`detect_client_server_boundaries`** (`bool`): Flag enforcing client/server import separation rules.
