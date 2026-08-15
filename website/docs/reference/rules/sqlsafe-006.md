---
title: "SQLSAFE-006 rule"
description: "Raw query omits deleted_at IS NULL condition on soft-deletable entity table"
---

# `SQLSAFE-006`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `soft-delete-bypass`

**Description**: Raw query omits deleted_at IS NULL condition on soft-deletable entity table

**Recommendation**: Include 'deleted_at IS NULL' in WHERE clauses when querying tables that use soft deletion

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
db.Query("SELECT * FROM users WHERE email = $1", email)

// ✅ Do (Recommended)
db.Query("SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", email)
```

:::

---

[← SQLSAFE-005](/reference/rules/sqlsafe-005) · [Rule Catalog](/reference/rule-catalog) · [merge-conflict-marker →](/reference/rules/merge-conflict-marker)
