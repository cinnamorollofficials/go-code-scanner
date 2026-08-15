---
title: "SQLI-012 rule"
description: "For developers remediating SQLI-012: Tainted SQL query template passed into statement preparation method db.Prepare()"
---

# `SQLI-012` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `prepared-statement`

**Description**: Tainted SQL query template passed into statement preparation method db.Prepare()

**Recommendation**: Keep the SQL query string passed to db.Prepare strictly constant and bind dynamic values via stmt.Query / stmt.Exec


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
stmt, err := db.Prepare("SELECT * FROM users WHERE status = " + filter)

// Safer example
stmt, err := db.Prepare("SELECT * FROM users WHERE status = $1")
rows, err := stmt.Query(filter)
```

:::

---

[← SQLI-011](/reference/rules/sqli-011) · [Rule Catalog](/reference/rule-catalog) · [SQLAUTH-001 →](/reference/rules/sqlauth-001)
