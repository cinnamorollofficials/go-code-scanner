---
title: "sql-string-format rule"
description: "Potential SQL injection using formatted strings"
---

# `sql-string-format`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Potential SQL injection using formatted strings

**Recommendation**: Use parameterized queries or prepared statements instead of string formatting

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)
rows, err := db.Query(query)

// ✅ Do (Recommended)
query := "SELECT * FROM users WHERE email = $1"
rows, err := db.Query(query, userEmail)
```

```ts [TypeScript / JavaScript]
// ❌ Don't (Unsafe)
const query = `SELECT * FROM users WHERE email = '${userEmail}'`;
const result = await client.query(query);

// ✅ Do (Recommended)
const query = "SELECT * FROM users WHERE email = $1";
const result = await client.query(query, [userEmail]);
```

```python [Python]
# ❌ Don't (Unsafe)
query = f"SELECT * FROM users WHERE email = '{user_email}'"
cursor.execute(query)

# ✅ Do (Recommended)
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_email,))
```

:::

---

[← backend-sensitive-log](/reference/rules/backend-sensitive-log) · [Rule Catalog](/reference/rule-catalog) · [hardcoded-credential →](/reference/rules/hardcoded-credential)
