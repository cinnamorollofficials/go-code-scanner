---
title: "privacy-sensitive-response rule"
description: "Response construction may expose a sensitive personal field"
---

# `privacy-sensitive-response`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_response`

**Description**: Response construction may expose a sensitive personal field

**Recommendation**: Map the response through an explicit allowlisted DTO and omit sensitive fields

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
res.json({ id: user.id, email: user.email, ssn: user.ssn });

// ✅ Do (Recommended)
res.json({ id: user.id, email: user.email });
```

---

[← privacy-pii-fixture](/reference/rules/privacy-pii-fixture) · [Rule Catalog](/reference/rule-catalog)
