---
title: "go-unbounded-request-read rule"
description: "Request body may be read without explicit size limits"
---

# `go-unbounded-request-read`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Request body may be read without explicit size limits

**Recommendation**: Limit request body with http.MaxBytesReader or io.LimitReader before reading into memory

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
body, err := io.ReadAll(r.Body)

// ✅ Do (Recommended)
body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB size limit
```

---

[← go-http-default-server](/reference/rules/go-http-default-server) · [Rule Catalog](/reference/rule-catalog) · [go-discarded-error →](/reference/rules/go-discarded-error)
