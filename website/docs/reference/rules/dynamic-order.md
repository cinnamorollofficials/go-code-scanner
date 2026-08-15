---
title: "dynamic-order rule"
description: "Dynamic ORDER BY clause built via string formatting"
---

# `dynamic-order`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Dynamic ORDER BY clause built via string formatting

**Recommendation**: Validate dynamic column names against an explicit allowlist before building queries

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
db.Order(fmt.Sprintf("%s ASC", sortColumn))

// ✅ Do (Recommended)
allowedColumns := map[string]bool{"created_at": true, "name": true}
if allowedColumns[sortColumn] {
    db.Order(sortColumn + " ASC")
}
```

---

[← unsafe-inner-html](/reference/rules/unsafe-inner-html) · [Rule Catalog](/reference/rule-catalog) · [api-struct-response →](/reference/rules/api-struct-response)
