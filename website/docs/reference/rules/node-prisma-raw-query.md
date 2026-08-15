---
title: "node-prisma-raw-query rule"
description: "For developers remediating node-prisma-raw-query: Prisma raw unsafe query executed with potentially untrusted dynamic string"
---

# `node-prisma-raw-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Prisma raw unsafe query executed with potentially untrusted dynamic string

**Recommendation**: Use prisma.$queryRaw with tagged template literals (parameterized) instead of unsafe variants


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
const users = await prisma.$queryRawUnsafe(`SELECT * FROM users WHERE id = '${id}'`);

// Safer example
const users = await prisma.$queryRaw`SELECT * FROM users WHERE id = ${id}`;
```

---

[← javascript-dynamic-eval](/reference/rules/javascript-dynamic-eval) · [Rule Catalog](/reference/rule-catalog) · [node-typeorm-raw-query →](/reference/rules/node-typeorm-raw-query)
