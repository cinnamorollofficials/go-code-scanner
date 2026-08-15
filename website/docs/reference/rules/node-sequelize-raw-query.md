---
title: "node-sequelize-raw-query rule"
description: "For developers remediating node-sequelize-raw-query: Sequelize raw query executed with template string interpolation"
---

# `node-sequelize-raw-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Sequelize raw query executed with template string interpolation

**Recommendation**: Use replacements or bind options in sequelize.query for safe parameter binding


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
await sequelize.query(`SELECT * FROM users WHERE status = '${status}'`);

// Safer example
await sequelize.query("SELECT * FROM users WHERE status = :status", { replacements: { status } });
```

---

[← node-typeorm-raw-query](/reference/rules/node-typeorm-raw-query) · [Rule Catalog](/reference/rule-catalog) · [node-pg-dynamic-query →](/reference/rules/node-pg-dynamic-query)
