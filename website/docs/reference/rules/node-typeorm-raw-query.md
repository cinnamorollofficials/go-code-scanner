---
title: "node-typeorm-raw-query rule"
description: "For developers remediating node-typeorm-raw-query: TypeORM raw query with dynamic string interpolation"
---

# `node-typeorm-raw-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: TypeORM raw query with dynamic string interpolation

**Recommendation**: Pass parameters as the second argument array to query() rather than template interpolation


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
await connection.query(`SELECT * FROM users WHERE email = '${email}'`);

// Safer example
await connection.query("SELECT * FROM users WHERE email = $1", [email]);
```

---

[← node-prisma-raw-query](/reference/rules/node-prisma-raw-query) · [Rule Catalog](/reference/rule-catalog) · [node-sequelize-raw-query →](/reference/rules/node-sequelize-raw-query)
