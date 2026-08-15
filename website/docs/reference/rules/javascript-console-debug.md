---
title: "javascript-console-debug rule"
description: "Console debug statement left in code"
---

# `javascript-console-debug`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `debug_code`

**Description**: Console debug statement left in code

**Recommendation**: Remove debug statements or use an application logger with proper log level

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
function handleLogin(user: User) {
  console.log("User logged in:", user);
}

// ✅ Do (Recommended)
function handleLogin(user: User) {
  logger.info("User logged in", { userId: user.id });
}
```

---

[← mixed-indentation](/reference/rules/mixed-indentation) · [Rule Catalog](/reference/rule-catalog) · [privacy-pii-log →](/reference/rules/privacy-pii-log)
