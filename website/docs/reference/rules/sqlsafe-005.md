---
title: "SQLSAFE-005 rule"
description: "For developers remediating SQLSAFE-005: Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence"
---

# `SQLSAFE-005` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `logic-operator-precedence`

**Description**: Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence

**Recommendation**: Explicitly group logical expressions with parentheses to avoid inadvertent filter bypass or tenant leakage


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := "SELECT * FROM orders WHERE tenant_id = $1 AND status = 'active' OR is_admin = true"

// Safer example
query := "SELECT * FROM orders WHERE tenant_id = $1 AND (status = 'active' OR is_admin = true)"
```

:::

---

[← SQLSAFE-004](/reference/rules/sqlsafe-004) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-006 →](/reference/rules/sqlsafe-006)
