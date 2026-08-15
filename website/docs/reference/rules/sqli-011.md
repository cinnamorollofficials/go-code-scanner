---
title: "SQLI-011 rule"
description: "Unsafe list or IN clause expansion using strings.Join or manual string interpolation"
---

# `SQLI-011`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `list-expansion`

**Description**: Unsafe list or IN clause expansion using strings.Join or manual string interpolation

**Recommendation**: Use sqlx.In or generate parameterized bind variable lists (?, ?, ...) for slice queries

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", strings.Join(ids, ","))
rows, err := db.Query(query)

// ✅ Do (Recommended)
query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
query = db.Rebind(query)
rows, err := db.Query(query, args...)
```

:::

---

[← SQLI-008](/reference/rules/sqli-008) · [Rule Catalog](/reference/rule-catalog) · [SQLI-012 →](/reference/rules/sqli-012)
