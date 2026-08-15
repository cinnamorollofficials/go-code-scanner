---
title: "SQLSAFE-001 rule"
description: "Unbounded UPDATE or DELETE query without a WHERE clause"
---

# `SQLSAFE-001`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `destructive-query`

**Description**: Unbounded UPDATE or DELETE query without a WHERE clause

**Recommendation**: Always specify a WHERE clause or explicit target filter to prevent accidental table-wide mutation

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
db.Exec("DELETE FROM users")

// ✅ Do (Recommended)
db.Exec("DELETE FROM users WHERE expires_at < $1", cutoffTime)
```

:::

---

[← DBPERF-002](/reference/rules/dbperf-002) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-003 →](/reference/rules/sqlsafe-003)
