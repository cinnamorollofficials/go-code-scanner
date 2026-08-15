---
title: "privacy-pii-log rule"
description: "Logging statement may expose personally identifiable information"
---

# `privacy-pii-log`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_log`

**Description**: Logging statement may expose personally identifiable information

**Recommendation**: Remove the PII field or log a non-reversible, access-controlled reference identifier

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
log.Printf("User registered with email: %s, phone: %s", email, phone)

// ✅ Do (Recommended)
log.Printf("User registered with ID: %s", userID)
```

---

[← javascript-console-debug](/reference/rules/javascript-console-debug) · [Rule Catalog](/reference/rule-catalog) · [privacy-pii-url →](/reference/rules/privacy-pii-url)
