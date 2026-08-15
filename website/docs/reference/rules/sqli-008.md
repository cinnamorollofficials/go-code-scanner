---
title: "SQLI-008 rule"
description: "SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed"
---

# `SQLI-008`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `bind-mismatch`

**Description**: SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed

**Recommendation**: Ensure the number of bind placeholders ($1, ?) matches the count of passed query arguments exactly

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id)

// ✅ Do (Recommended)
db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id, tenantID)
```

:::

---

[← SQLI-004](/reference/rules/sqli-004) · [Rule Catalog](/reference/rule-catalog) · [SQLI-011 →](/reference/rules/sqli-011)
