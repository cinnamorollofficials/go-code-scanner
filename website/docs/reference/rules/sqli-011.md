---
title: "SQLI-011 rule"
description: "For developers remediating SQLI-011: Unsafe list or IN clause expansion using strings.Join or manual string interpolation"
---

# `SQLI-011` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `list-expansion`

**Description**: Unsafe list or IN clause expansion using strings.Join or manual string interpolation

**Recommendation**: Use sqlx.In or generate parameterized bind variable lists (?, ?, ...) for slice queries


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", strings.Join(ids, ","))
rows, err := db.Query(query)

// Safer example
query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
query = db.Rebind(query)
rows, err := db.Query(query, args...)
```

:::

---

[← SQLI-008](/reference/rules/sqli-008) · [Rule Catalog](/reference/rule-catalog) · [SQLI-012 →](/reference/rules/sqli-012)
