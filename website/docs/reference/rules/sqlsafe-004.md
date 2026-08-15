---
title: "SQLSAFE-004 rule"
description: "Database operation executes on global connection pool escaping active transaction boundary"
---

# `SQLSAFE-004`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `transaction-loss`

**Description**: Database operation executes on global connection pool escaping active transaction boundary

**Recommendation**: Execute queries using the active transaction handle (tx) to guarantee atomic rollback on error

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
func Transfer(tx *sql.Tx, from, to string, amount int) error {
    db.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}

// ✅ Do (Recommended)
func Transfer(tx *sql.Tx, from, to string, amount int) error {
    tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}
```

:::

---

[← SQLSAFE-003](/reference/rules/sqlsafe-003) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-005 →](/reference/rules/sqlsafe-005)
