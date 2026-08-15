---
title: "go-http-default-server rule"
description: "Default HTTP server does not configure defensive timeouts"
---

# `go-http-default-server`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: Default HTTP server does not configure defensive timeouts

**Recommendation**: Use custom http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout, and IdleTimeout

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
http.ListenAndServe(":8080", handler)

// ✅ Do (Recommended)
server := &http.Server{
    Addr: ":8080", Handler: handler,
    ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
}
server.ListenAndServe()
```

---

[← go-multipart-memory](/reference/rules/go-multipart-memory) · [Rule Catalog](/reference/rule-catalog) · [go-unbounded-request-read →](/reference/rules/go-unbounded-request-read)
