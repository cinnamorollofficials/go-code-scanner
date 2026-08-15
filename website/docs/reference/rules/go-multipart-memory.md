---
title: "go-multipart-memory rule"
description: "For developers remediating go-multipart-memory: Ensure multipart request processing configures explicit memory limits"
---

# `go-multipart-memory` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Ensure multipart request processing configures explicit memory limits

**Recommendation**: Set explicit memory limit with ParseMultipartForm or MaxBytesReader to prevent memory exhaustion


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
c.Request.ParseMultipartForm(100 << 20) // Unbounded 100MB buffer

// Safer example
c.Request.ParseMultipartForm(10 << 20) // Controlled 10MB memory limit
```

---

[← go-insecure-cookie-attribute](/reference/rules/go-insecure-cookie-attribute) · [Rule Catalog](/reference/rule-catalog) · [go-http-default-server →](/reference/rules/go-http-default-server)
