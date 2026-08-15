---
title: "debug-mode-enabled rule"
description: "Debug mode appears to be explicitly enabled in configuration"
---

# `debug-mode-enabled`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `debug_configuration`

**Description**: Debug mode appears to be explicitly enabled in configuration

**Recommendation**: Disable debug mode in production deployment configurations to prevent information disclosure

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
debug := true

// ✅ Do (Recommended)
debug := os.Getenv("APP_ENV") == "development"
```

---

[← go-permissive-file-mode](/reference/rules/go-permissive-file-mode) · [Rule Catalog](/reference/rule-catalog) · [go-insecure-cookie-attribute →](/reference/rules/go-insecure-cookie-attribute)
