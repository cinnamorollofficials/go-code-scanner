---
title: "tls-insecure-skip-verify rule"
description: "TLS certificate verification is explicitly disabled"
---

# `tls-insecure-skip-verify`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `transport_security`

**Description**: TLS certificate verification is explicitly disabled

**Recommendation**: Enable certificate verification and configure valid trust stores

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// ✅ Do (Recommended)
tr := &http.Transport{
    TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
}
```

---

[← hardcoded-api-url](/reference/rules/hardcoded-api-url) · [Rule Catalog](/reference/rule-catalog) · [wildcard-cors-origin →](/reference/rules/wildcard-cors-origin)
