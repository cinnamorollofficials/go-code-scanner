---
title: "frontend-sensitive-log rule"
description: "For developers remediating frontend-sensitive-log: Frontend log statement may expose sensitive credentials or PII"
---

# `frontend-sensitive-log` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Frontend log statement may expose sensitive credentials or PII

**Recommendation**: Sanitize log parameters and remove sensitive tokens or user identifiers from console logs


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
console.log("User auth failed for password:", password);

// Safer example
console.error("User authentication failed", { username });
```

---

[← weak-secret](/reference/rules/weak-secret) · [Rule Catalog](/reference/rule-catalog) · [backend-sensitive-log →](/reference/rules/backend-sensitive-log)
