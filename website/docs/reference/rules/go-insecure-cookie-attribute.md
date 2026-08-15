---
title: "go-insecure-cookie-attribute rule"
description: "Cookie configured with explicitly insecure security attributes"
---

# `go-insecure-cookie-attribute`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cookie_security`

**Description**: Cookie configured with explicitly insecure security attributes

**Recommendation**: Enable Secure and HttpOnly flags and set an appropriate SameSite policy for session cookies

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
cookie := &http.Cookie{Name: "session", Value: token, Secure: false}

// ✅ Do (Recommended)
cookie := &http.Cookie{Name: "session", Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
```

---

[← debug-mode-enabled](/reference/rules/debug-mode-enabled) · [Rule Catalog](/reference/rule-catalog) · [go-multipart-memory →](/reference/rules/go-multipart-memory)
