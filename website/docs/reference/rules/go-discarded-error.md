---
title: "go-discarded-error rule"
description: "For developers remediating go-discarded-error: Returned error value is explicitly ignored with blank identifier"
---

# `go-discarded-error` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `error_handling`

**Description**: Returned error value is explicitly ignored with blank identifier

**Recommendation**: Check and handle returned errors or document valid reason for ignoring


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
_ = db.Close()

// Safer example
if err := db.Close(); err != nil {
    log.Printf("Failed to close DB connection: %v", err)
}
```

---

[← go-unbounded-request-read](/reference/rules/go-unbounded-request-read) · [Rule Catalog](/reference/rule-catalog) · [go-process-termination →](/reference/rules/go-process-termination)
