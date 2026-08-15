---
title: "javascript-dynamic-eval rule"
description: "Dynamic eval execution of untrusted input detected"
---

# `javascript-dynamic-eval`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `unsafe_deserialization`

**Description**: Dynamic eval execution of untrusted input detected

**Recommendation**: Use structured data parsers (JSON.parse) and schema validators instead of code evaluation

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
const config = eval("(" + jsonString + ")");

// ✅ Do (Recommended)
const config = JSON.parse(jsonString);
```

---

[← go-weak-random-secret](/reference/rules/go-weak-random-secret) · [Rule Catalog](/reference/rule-catalog) · [node-prisma-raw-query →](/reference/rules/node-prisma-raw-query)
