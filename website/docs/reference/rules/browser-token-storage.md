---
title: "browser-token-storage rule"
description: "For developers remediating browser-token-storage: Token stored in localStorage — vulnerable to XSS token theft"
---

# `browser-token-storage` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Token stored in localStorage — vulnerable to XSS token theft

**Recommendation**: Store authentication tokens in HttpOnly, Secure, SameSite cookies instead of localStorage


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
localStorage.setItem("access_token", response.token);

// Safer example
await fetch("/api/login", { credentials: "include", method: "POST", body });
```

---

[← mock-token](/reference/rules/mock-token) · [Rule Catalog](/reference/rule-catalog) · [permission-bypass →](/reference/rules/permission-bypass)
