---
title: "node-pg-dynamic-query rule"
description: "node-postgres query executed with template string interpolation"
---

# `node-pg-dynamic-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: node-postgres query executed with template string interpolation

**Recommendation**: Use parameterized query format ($1, $2) and pass values in the values parameter array

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
await client.query(`SELECT * FROM accounts WHERE id = '${id}'`);

// ✅ Do (Recommended)
await client.query("SELECT * FROM accounts WHERE id = $1", [id]);
```

---

[← node-sequelize-raw-query](/reference/rules/node-sequelize-raw-query) · [Rule Catalog](/reference/rule-catalog) · [node-mysql-dynamic-query →](/reference/rules/node-mysql-dynamic-query)
