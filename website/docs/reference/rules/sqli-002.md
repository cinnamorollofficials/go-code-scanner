---
title: "SQLI-002 rule"
description: "Untrusted table, column, or identifier dynamically interpolated into SQL"
---

# `SQLI-002`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Untrusted table, column, or identifier dynamically interpolated into SQL

**Recommendation**: Validate SQL identifiers against an explicit allow-list of known safe column/table names before interpolation

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
query := fmt.Sprintf("SELECT * FROM %s WHERE active = 1", tableName)
rows, err := db.Query(query)

// ✅ Do (Recommended)
allowed := map[string]string{"users": "users", "admins": "admins"}
table, ok := allowed[tableName]
if !ok { return nil, errors.New("invalid table") }
query := fmt.Sprintf("SELECT * FROM %s WHERE active = 1", table)
rows, err := db.Query(query)
```

:::

---

[← SQLI-001](/reference/rules/sqli-001) · [Rule Catalog](/reference/rule-catalog) · [SQLI-004 →](/reference/rules/sqli-004)
