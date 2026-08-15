---
title: "SQLI-001 rule"
description: "For developers remediating SQLI-001: Untrusted value concatenated or formatted into executable SQL at database driver sink"
---

# `SQLI-001` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Untrusted value concatenated or formatted into executable SQL at database driver sink

**Recommendation**: Use parameterized queries ($1, ?, :name) instead of string concatenation or fmt.Sprintf


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go (database/sql)]
// Unsafe example
query := "SELECT * FROM users WHERE id = " + id
row := db.QueryRow(query)

// Safer example
query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRow(query, id)
```

:::

---

[← DBSEC-003](/reference/rules/dbsec-003) · [Rule Catalog](/reference/rule-catalog) · [SQLI-002 →](/reference/rules/sqli-002)
