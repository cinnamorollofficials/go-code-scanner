---
title: "go-multipart-memory rule"
description: "Ensure multipart request processing configures explicit memory limits"
---

# `go-multipart-memory`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Ensure multipart request processing configures explicit memory limits

**Recommendation**: Set explicit memory limit with ParseMultipartForm or MaxBytesReader to prevent memory exhaustion

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
c.Request.ParseMultipartForm(100 << 20) // Unbounded 100MB buffer

// ✅ Do (Recommended)
c.Request.ParseMultipartForm(10 << 20) // Controlled 10MB memory limit
```

---

[← go-insecure-cookie-attribute](/reference/rules/go-insecure-cookie-attribute) · [Rule Catalog](/reference/rule-catalog) · [go-http-default-server →](/reference/rules/go-http-default-server)
