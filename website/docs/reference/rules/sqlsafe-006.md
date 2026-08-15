---
title: "SQLSAFE-006 rule"
description: "For developers remediating SQLSAFE-006: Raw query omits deleted_at IS NULL condition on soft-deletable entity table"
---

# `SQLSAFE-006` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `soft-delete-bypass`

**Description**: Raw query omits deleted_at IS NULL condition on soft-deletable entity table

**Recommendation**: Include 'deleted_at IS NULL' in WHERE clauses when querying tables that use soft deletion


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Query("SELECT * FROM users WHERE email = $1", email)

// Safer example
db.Query("SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", email)
```

:::

---

[← SQLSAFE-005](/reference/rules/sqlsafe-005) · [Rule Catalog](/reference/rule-catalog) · [merge-conflict-marker →](/reference/rules/merge-conflict-marker)
