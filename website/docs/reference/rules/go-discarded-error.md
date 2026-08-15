---
title: "go-discarded-error rule"
description: "Returned error value is explicitly ignored with blank identifier"
---

# `go-discarded-error`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `error_handling`

**Description**: Returned error value is explicitly ignored with blank identifier

**Recommendation**: Check and handle returned errors or document valid reason for ignoring

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
_ = db.Close()

// ✅ Do (Recommended)
if err := db.Close(); err != nil {
    log.Printf("Failed to close DB connection: %v", err)
}
```

---

[← go-unbounded-request-read](/reference/rules/go-unbounded-request-read) · [Rule Catalog](/reference/rule-catalog) · [go-process-termination →](/reference/rules/go-process-termination)
