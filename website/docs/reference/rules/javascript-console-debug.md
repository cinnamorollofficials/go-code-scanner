---
title: "javascript-console-debug rule"
description: "For developers remediating javascript-console-debug: Console debug statement left in code"
---

# `javascript-console-debug` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `debug_code`

**Description**: Console debug statement left in code

**Recommendation**: Remove debug statements or use an application logger with proper log level


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
function handleLogin(user: User) {
  console.log("User logged in:", user);
}

// Safer example
function handleLogin(user: User) {
  logger.info("User logged in", { userId: user.id });
}
```

---

[← mixed-indentation](/reference/rules/mixed-indentation) · [Rule Catalog](/reference/rule-catalog) · [privacy-pii-log →](/reference/rules/privacy-pii-log)
