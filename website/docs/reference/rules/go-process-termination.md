---
title: "go-process-termination rule"
description: "For developers remediating go-process-termination: Application path may terminate entire process unexpectedly"
---

# `go-process-termination` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `process_termination`

**Description**: Application path may terminate entire process unexpectedly

**Recommendation**: Propagate errors to request boundaries and perform controlled shutdown instead of calling panic/log.Fatal


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
if err != nil {
    panic(err)
}

// Safer example
if err != nil {
    return fmt.Errorf("process request: %w", err)
}
```

---

[← go-discarded-error](/reference/rules/go-discarded-error) · [Rule Catalog](/reference/rule-catalog) · [go-http-client-without-timeout →](/reference/rules/go-http-client-without-timeout)
