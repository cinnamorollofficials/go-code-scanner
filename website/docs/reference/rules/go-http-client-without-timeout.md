---
title: "go-http-client-without-timeout rule"
description: "For developers remediating go-http-client-without-timeout: HTTP client struct literal does not set an overall request timeout"
---

# `go-http-client-without-timeout` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: HTTP client struct literal does not set an overall request timeout

**Recommendation**: Configure explicit http.Client.Timeout and appropriate transport timeouts


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
client := &http.Client{}

// Safer example
client := &http.Client{Timeout: 10 * time.Second}
```

---

[← go-process-termination](/reference/rules/go-process-termination) · [Rule Catalog](/reference/rule-catalog) · [DBMIG-001 →](/reference/rules/dbmig-001)
