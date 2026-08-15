---
title: "go-process-termination rule"
description: "Application path may terminate entire process unexpectedly"
---

# `go-process-termination`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `process_termination`

**Description**: Application path may terminate entire process unexpectedly

**Recommendation**: Propagate errors to request boundaries and perform controlled shutdown instead of calling panic/log.Fatal

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
if err != nil {
    panic(err)
}

// ✅ Do (Recommended)
if err != nil {
    return fmt.Errorf("process request: %w", err)
}
```

---

[← go-discarded-error](/reference/rules/go-discarded-error) · [Rule Catalog](/reference/rule-catalog) · [go-http-client-without-timeout →](/reference/rules/go-http-client-without-timeout)
