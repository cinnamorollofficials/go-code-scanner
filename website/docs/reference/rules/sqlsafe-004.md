---
title: "SQLSAFE-004 rule"
description: "For developers remediating SQLSAFE-004: Database operation executes on global connection pool escaping active transaction boundary"
---

# `SQLSAFE-004` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `transaction-loss`

**Description**: Database operation executes on global connection pool escaping active transaction boundary

**Recommendation**: Execute queries using the active transaction handle (tx) to guarantee atomic rollback on error


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func Transfer(tx *sql.Tx, from, to string, amount int) error {
    db.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}

// Safer example
func Transfer(tx *sql.Tx, from, to string, amount int) error {
    tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}
```

:::

---

[← SQLSAFE-003](/reference/rules/sqlsafe-003) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-005 →](/reference/rules/sqlsafe-005)
