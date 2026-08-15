---
title: "node-sequelize-raw-query rule"
description: "Sequelize raw query executed with template string interpolation"
---

# `node-sequelize-raw-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Sequelize raw query executed with template string interpolation

**Recommendation**: Use replacements or bind options in sequelize.query for safe parameter binding

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
await sequelize.query(`SELECT * FROM users WHERE status = '${status}'`);

// ✅ Do (Recommended)
await sequelize.query("SELECT * FROM users WHERE status = :status", { replacements: { status } });
```

---

[← node-typeorm-raw-query](/reference/rules/node-typeorm-raw-query) · [Rule Catalog](/reference/rule-catalog) · [node-pg-dynamic-query →](/reference/rules/node-pg-dynamic-query)
