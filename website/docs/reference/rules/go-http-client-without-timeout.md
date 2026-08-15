---
title: "go-http-client-without-timeout rule"
description: "HTTP client struct literal does not set an overall request timeout"
---

# `go-http-client-without-timeout`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: HTTP client struct literal does not set an overall request timeout

**Recommendation**: Configure explicit http.Client.Timeout and appropriate transport timeouts

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
client := &http.Client{}

// ✅ Do (Recommended)
client := &http.Client{Timeout: 10 * time.Second}
```

---

[← go-process-termination](/reference/rules/go-process-termination) · [Rule Catalog](/reference/rule-catalog) · [DBMIG-001 →](/reference/rules/dbmig-001)
