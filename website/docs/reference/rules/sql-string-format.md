---
title: "sql-string-format rule"
description: "For developers remediating sql-string-format: Potential SQL injection using formatted strings"
---

# `sql-string-format` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Potential SQL injection using formatted strings

**Recommendation**: Use parameterized queries or prepared statements instead of string formatting


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)
rows, err := db.Query(query)

// Safer example
query := "SELECT * FROM users WHERE email = $1"
rows, err := db.Query(query, userEmail)
```

```ts [TypeScript / JavaScript]
// Unsafe example
const query = `SELECT * FROM users WHERE email = '${userEmail}'`;
const result = await client.query(query);

// Safer example
const query = "SELECT * FROM users WHERE email = $1";
const result = await client.query(query, [userEmail]);
```

```python [Python]
# Unsafe example
query = f"SELECT * FROM users WHERE email = '{user_email}'"
cursor.execute(query)

# Safer example
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_email,))
```

:::

---

[← backend-sensitive-log](/reference/rules/backend-sensitive-log) · [Rule Catalog](/reference/rule-catalog) · [hardcoded-credential →](/reference/rules/hardcoded-credential)
