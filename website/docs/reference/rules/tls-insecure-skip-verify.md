---
title: "tls-insecure-skip-verify rule"
description: "For developers remediating tls-insecure-skip-verify: TLS certificate verification is explicitly disabled"
---

# `tls-insecure-skip-verify` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `transport_security`

**Description**: TLS certificate verification is explicitly disabled

**Recommendation**: Enable certificate verification and configure valid trust stores


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// Safer example
tr := &http.Transport{
    TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
}
```

---

[← hardcoded-api-url](/reference/rules/hardcoded-api-url) · [Rule Catalog](/reference/rule-catalog) · [wildcard-cors-origin →](/reference/rules/wildcard-cors-origin)
