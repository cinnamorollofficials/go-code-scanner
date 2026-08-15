---
title: "go-insecure-cookie-attribute rule"
description: "For developers remediating go-insecure-cookie-attribute: Cookie configured with explicitly insecure security attributes"
---

# `go-insecure-cookie-attribute` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cookie_security`

**Description**: Cookie configured with explicitly insecure security attributes

**Recommendation**: Enable Secure and HttpOnly flags and set an appropriate SameSite policy for session cookies


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
cookie := &http.Cookie{Name: "session", Value: token, Secure: false}

// Safer example
cookie := &http.Cookie{Name: "session", Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
```

---

[← debug-mode-enabled](/reference/rules/debug-mode-enabled) · [Rule Catalog](/reference/rule-catalog) · [go-multipart-memory →](/reference/rules/go-multipart-memory)
