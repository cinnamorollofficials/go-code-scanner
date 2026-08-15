---
title: "privacy-sensitive-response rule"
description: "For developers remediating privacy-sensitive-response: Response construction may expose a sensitive personal field"
---

# `privacy-sensitive-response` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_response`

**Description**: Response construction may expose a sensitive personal field

**Recommendation**: Map the response through an explicit allowlisted DTO and omit sensitive fields


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
res.json({ id: user.id, email: user.email, ssn: user.ssn });

// Safer example
res.json({ id: user.id, email: user.email });
```

---

[← privacy-pii-fixture](/reference/rules/privacy-pii-fixture) · [Rule Catalog](/reference/rule-catalog)
