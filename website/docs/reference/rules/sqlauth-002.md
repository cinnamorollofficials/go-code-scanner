---
title: "SQLAUTH-002 rule"
description: "Sensitive resource queried solely by object ID without user ownership scoping (IDOR risk)"
---

# `SQLAUTH-002`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `authorization-idor`

**Description**: Sensitive resource queried solely by object ID without user ownership scoping (IDOR risk)

**Recommendation**: Scope entity lookups by both the object ID and authenticated user/account ID

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
func getOrder(db *sql.DB, orderID string) (*sql.Row, error) {
    return db.QueryRow("SELECT * FROM orders WHERE id = $1", orderID), nil
}

// ✅ Do (Recommended)
func getOrder(db *sql.DB, orderID, userID string) (*sql.Row, error) {
    return db.QueryRow("SELECT * FROM orders WHERE id = $1 AND user_id = $2", orderID, userID), nil
}
```

:::

---

[← SQLAUTH-001](/reference/rules/sqlauth-001) · [Rule Catalog](/reference/rule-catalog) · [SQLAUTH-003 →](/reference/rules/sqlauth-003)
