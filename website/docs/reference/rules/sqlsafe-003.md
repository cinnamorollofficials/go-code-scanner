---
title: "SQLSAFE-003 rule"
description: "Non-atomic read-modify-write pattern detected on balance/inventory field without row locking"
---

# `SQLSAFE-003`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `concurrency-hazard`

**Description**: Non-atomic read-modify-write pattern detected on balance/inventory field without row locking

**Recommendation**: Use SELECT ... FOR UPDATE within a transaction or perform atomic SQL mutations

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
var balance int
db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&balance)
balance += 100
db.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", balance, id)

// ✅ Do (Recommended)
tx, _ := db.Begin()
var balance int
tx.QueryRow("SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", id).Scan(&balance)
balance += 100
tx.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", balance, id)
tx.Commit()
```

:::

---

[← SQLSAFE-001](/reference/rules/sqlsafe-001) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-004 →](/reference/rules/sqlsafe-004)
