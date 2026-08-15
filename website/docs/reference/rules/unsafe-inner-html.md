---
title: "unsafe-inner-html rule"
description: "dangerouslySetInnerHTML used — potential DOM XSS vulnerability"
---

# `unsafe-inner-html`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `xss`

**Description**: dangerouslySetInnerHTML used — potential DOM XSS vulnerability

**Recommendation**: Sanitize raw HTML using DOMPurify before injecting into the DOM

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// ✅ Do (Recommended)
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(userInput) }} />
```

---

[← hardcoded-credential](/reference/rules/hardcoded-credential) · [Rule Catalog](/reference/rule-catalog) · [dynamic-order →](/reference/rules/dynamic-order)
