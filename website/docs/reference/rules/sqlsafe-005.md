---
title: "SQLSAFE-005 rule"
description: "Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence"
---

# `SQLSAFE-005`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `logic-operator-precedence`

**Description**: Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence

**Recommendation**: Explicitly group logical expressions with parentheses to avoid inadvertent filter bypass or tenant leakage

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
query := "SELECT * FROM orders WHERE tenant_id = $1 AND status = 'active' OR is_admin = true"

// ✅ Do (Recommended)
query := "SELECT * FROM orders WHERE tenant_id = $1 AND (status = 'active' OR is_admin = true)"
```

:::

---

[← SQLSAFE-004](/reference/rules/sqlsafe-004) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-006 →](/reference/rules/sqlsafe-006)
