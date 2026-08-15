---
title: "node-mysql-dynamic-query rule"
description: "For developers remediating node-mysql-dynamic-query: mysql2 query executed with dynamic template string interpolation"
---

# `node-mysql-dynamic-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: mysql2 query executed with dynamic template string interpolation

**Recommendation**: Use query placeholders (?) and pass arguments in the parameter array


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
await pool.query(`SELECT * FROM products WHERE category = '${category}'`);

// Safer example
await pool.query("SELECT * FROM products WHERE category = ?", [category]);
```

---

[← node-pg-dynamic-query](/reference/rules/node-pg-dynamic-query) · [Rule Catalog](/reference/rule-catalog) · [python-sqlalchemy-raw-sql →](/reference/rules/python-sqlalchemy-raw-sql)
