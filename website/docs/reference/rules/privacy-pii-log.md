---
title: "privacy-pii-log rule"
description: "For developers remediating privacy-pii-log: Logging statement may expose personally identifiable information"
---

# `privacy-pii-log` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_log`

**Description**: Logging statement may expose personally identifiable information

**Recommendation**: Remove the PII field or log a non-reversible, access-controlled reference identifier


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
log.Printf("User registered with email: %s, phone: %s", email, phone)

// Safer example
log.Printf("User registered with ID: %s", userID)
```

---

[← javascript-console-debug](/reference/rules/javascript-console-debug) · [Rule Catalog](/reference/rule-catalog) · [privacy-pii-url →](/reference/rules/privacy-pii-url)
