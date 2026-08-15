---
title: "wildcard-cors-origin rule"
description: "Wildcard CORS origin header found in configuration"
---

# `wildcard-cors-origin`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cors`

**Description**: Wildcard CORS origin header found in configuration

**Recommendation**: Use an explicit CORS origin allowlist tailored for each deployment environment

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
c.Header("Access-Control-Allow-Origin", "*")

// ✅ Do (Recommended)
c.Header("Access-Control-Allow-Origin", "https://app.example.com")
```

---

[← tls-insecure-skip-verify](/reference/rules/tls-insecure-skip-verify) · [Rule Catalog](/reference/rule-catalog) · [go-permissive-file-mode →](/reference/rules/go-permissive-file-mode)
