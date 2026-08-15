---
title: "debug-mode-enabled rule"
description: "For developers remediating debug-mode-enabled: Debug mode appears to be explicitly enabled in configuration"
---

# `debug-mode-enabled` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `debug_configuration`

**Description**: Debug mode appears to be explicitly enabled in configuration

**Recommendation**: Disable debug mode in production deployment configurations to prevent information disclosure


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
debug := true

// Safer example
debug := os.Getenv("APP_ENV") == "development"
```

---

[← go-permissive-file-mode](/reference/rules/go-permissive-file-mode) · [Rule Catalog](/reference/rule-catalog) · [go-insecure-cookie-attribute →](/reference/rules/go-insecure-cookie-attribute)
