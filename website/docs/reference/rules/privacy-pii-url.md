---
title: "privacy-pii-url rule"
description: "Personally identifiable information may be placed in a URL query string"
---

# `privacy-pii-url`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_url`

**Description**: Personally identifiable information may be placed in a URL query string

**Recommendation**: Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
urlParams.append("email", userEmail);

// ✅ Do (Recommended)
// Transmit sensitive parameters in authenticated POST request body
```

---

[← privacy-pii-log](/reference/rules/privacy-pii-log) · [Rule Catalog](/reference/rule-catalog) · [privacy-pii-fixture →](/reference/rules/privacy-pii-fixture)
