---
title: "privacy-pii-url rule"
description: "For developers remediating privacy-pii-url: Personally identifiable information may be placed in a URL query string"
---

# `privacy-pii-url` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_url`

**Description**: Personally identifiable information may be placed in a URL query string

**Recommendation**: Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
urlParams.append("email", userEmail);

// Safer example
// Transmit sensitive parameters in authenticated POST request body
```

---

[← privacy-pii-log](/reference/rules/privacy-pii-log) · [Rule Catalog](/reference/rule-catalog) · [privacy-pii-fixture →](/reference/rules/privacy-pii-fixture)
