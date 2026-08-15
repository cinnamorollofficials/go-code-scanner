---
title: "wildcard-cors-origin rule"
description: "For developers remediating wildcard-cors-origin: Wildcard CORS origin header found in configuration"
---

# `wildcard-cors-origin` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cors`

**Description**: Wildcard CORS origin header found in configuration

**Recommendation**: Use an explicit CORS origin allowlist tailored for each deployment environment


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
c.Header("Access-Control-Allow-Origin", "*")

// Safer example
c.Header("Access-Control-Allow-Origin", "https://app.example.com")
```

---

[← tls-insecure-skip-verify](/reference/rules/tls-insecure-skip-verify) · [Rule Catalog](/reference/rule-catalog) · [go-permissive-file-mode →](/reference/rules/go-permissive-file-mode)
