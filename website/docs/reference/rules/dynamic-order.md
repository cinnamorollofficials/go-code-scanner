---
title: "dynamic-order rule"
description: "For developers remediating dynamic-order: Dynamic ORDER BY clause built via string formatting"
---

# `dynamic-order` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Dynamic ORDER BY clause built via string formatting

**Recommendation**: Validate dynamic column names against an explicit allowlist before building queries


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
db.Order(fmt.Sprintf("%s ASC", sortColumn))

// Safer example
allowedColumns := map[string]bool{"created_at": true, "name": true}
if allowedColumns[sortColumn] {
    db.Order(sortColumn + " ASC")
}
```

---

[← unsafe-inner-html](/reference/rules/unsafe-inner-html) · [Rule Catalog](/reference/rule-catalog) · [api-struct-response →](/reference/rules/api-struct-response)
