---
title: "go-unbounded-request-read rule"
description: "For developers remediating go-unbounded-request-read: Request body may be read without explicit size limits"
---

# `go-unbounded-request-read` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Request body may be read without explicit size limits

**Recommendation**: Limit request body with http.MaxBytesReader or io.LimitReader before reading into memory


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
body, err := io.ReadAll(r.Body)

// Safer example
body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB size limit
```

---

[← go-http-default-server](/reference/rules/go-http-default-server) · [Rule Catalog](/reference/rule-catalog) · [go-discarded-error →](/reference/rules/go-discarded-error)
