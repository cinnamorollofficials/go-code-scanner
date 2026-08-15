---
title: "browser-token-storage rule"
description: "Token stored in localStorage — vulnerable to XSS token theft"
---

# `browser-token-storage`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Token stored in localStorage — vulnerable to XSS token theft

**Recommendation**: Store authentication tokens in HttpOnly, Secure, SameSite cookies instead of localStorage

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
localStorage.setItem("access_token", response.token);

// ✅ Do (Recommended)
await fetch("/api/login", { credentials: "include", method: "POST", body });
```

---

[← mock-token](/reference/rules/mock-token) · [Rule Catalog](/reference/rule-catalog) · [permission-bypass →](/reference/rules/permission-bypass)
