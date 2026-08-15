---
title: "SQLSAFE-001 rule"
description: "For developers remediating SQLSAFE-001: Unbounded UPDATE or DELETE query without a WHERE clause"
---

# `SQLSAFE-001` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `destructive-query`

**Description**: Unbounded UPDATE or DELETE query without a WHERE clause

**Recommendation**: Always specify a WHERE clause or explicit target filter to prevent accidental table-wide mutation


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Exec("DELETE FROM users")

// Safer example
db.Exec("DELETE FROM users WHERE expires_at < $1", cutoffTime)
```

:::

---

[← DBPERF-002](/reference/rules/dbperf-002) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-003 →](/reference/rules/sqlsafe-003)
