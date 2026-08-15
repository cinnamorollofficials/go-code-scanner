---
title: "SQLI-012 rule"
description: "Tainted SQL query template passed into statement preparation method db.Prepare()"
---

# `SQLI-012`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `prepared-statement`

**Description**: Tainted SQL query template passed into statement preparation method db.Prepare()

**Recommendation**: Keep the SQL query string passed to db.Prepare strictly constant and bind dynamic values via stmt.Query / stmt.Exec

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
stmt, err := db.Prepare("SELECT * FROM users WHERE status = " + filter)

// ✅ Do (Recommended)
stmt, err := db.Prepare("SELECT * FROM users WHERE status = $1")
rows, err := stmt.Query(filter)
```

:::

---

[← SQLI-011](/reference/rules/sqli-011) · [Rule Catalog](/reference/rule-catalog) · [SQLAUTH-001 →](/reference/rules/sqlauth-001)
