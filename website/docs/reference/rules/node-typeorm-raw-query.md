---
title: "node-typeorm-raw-query rule"
description: "TypeORM raw query with dynamic string interpolation"
---

# `node-typeorm-raw-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: TypeORM raw query with dynamic string interpolation

**Recommendation**: Pass parameters as the second argument array to query() rather than template interpolation

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
await connection.query(`SELECT * FROM users WHERE email = '${email}'`);

// ✅ Do (Recommended)
await connection.query("SELECT * FROM users WHERE email = $1", [email]);
```

---

[← node-prisma-raw-query](/reference/rules/node-prisma-raw-query) · [Rule Catalog](/reference/rule-catalog) · [node-sequelize-raw-query →](/reference/rules/node-sequelize-raw-query)
