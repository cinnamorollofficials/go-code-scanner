---
title: "javascript-dynamic-eval rule"
description: "For developers remediating javascript-dynamic-eval: Dynamic eval execution of untrusted input detected"
---

# `javascript-dynamic-eval` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `unsafe_deserialization`

**Description**: Dynamic eval execution of untrusted input detected

**Recommendation**: Use structured data parsers (JSON.parse) and schema validators instead of code evaluation


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
const config = eval("(" + jsonString + ")");

// Safer example
const config = JSON.parse(jsonString);
```

---

[← go-weak-random-secret](/reference/rules/go-weak-random-secret) · [Rule Catalog](/reference/rule-catalog) · [node-prisma-raw-query →](/reference/rules/node-prisma-raw-query)
