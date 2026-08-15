---
title: "unsafe-inner-html rule"
description: "For developers remediating unsafe-inner-html: dangerouslySetInnerHTML used — potential DOM XSS vulnerability"
---

# `unsafe-inner-html` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `xss`

**Description**: dangerouslySetInnerHTML used — potential DOM XSS vulnerability

**Recommendation**: Sanitize raw HTML using DOMPurify before injecting into the DOM


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// Safer example
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(userInput) }} />
```

---

[← hardcoded-credential](/reference/rules/hardcoded-credential) · [Rule Catalog](/reference/rule-catalog) · [dynamic-order →](/reference/rules/dynamic-order)
