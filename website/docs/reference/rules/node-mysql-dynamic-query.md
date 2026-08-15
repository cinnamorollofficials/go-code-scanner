---
title: "node-mysql-dynamic-query rule"
description: "mysql2 query executed with dynamic template string interpolation"
---

# `node-mysql-dynamic-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: mysql2 query executed with dynamic template string interpolation

**Recommendation**: Use query placeholders (?) and pass arguments in the parameter array

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
await pool.query(`SELECT * FROM products WHERE category = '${category}'`);

// ✅ Do (Recommended)
await pool.query("SELECT * FROM products WHERE category = ?", [category]);
```

---

[← node-pg-dynamic-query](/reference/rules/node-pg-dynamic-query) · [Rule Catalog](/reference/rule-catalog) · [python-sqlalchemy-raw-sql →](/reference/rules/python-sqlalchemy-raw-sql)
