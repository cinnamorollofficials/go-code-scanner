---
title: "node-prisma-raw-query rule"
description: "Prisma raw unsafe query executed with potentially untrusted dynamic string"
---

# `node-prisma-raw-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Prisma raw unsafe query executed with potentially untrusted dynamic string

**Recommendation**: Use prisma.$queryRaw with tagged template literals (parameterized) instead of unsafe variants

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
const users = await prisma.$queryRawUnsafe(`SELECT * FROM users WHERE id = '${id}'`);

// ✅ Do (Recommended)
const users = await prisma.$queryRaw`SELECT * FROM users WHERE id = ${id}`;
```

---

[← javascript-dynamic-eval](/reference/rules/javascript-dynamic-eval) · [Rule Catalog](/reference/rule-catalog) · [node-typeorm-raw-query →](/reference/rules/node-typeorm-raw-query)
