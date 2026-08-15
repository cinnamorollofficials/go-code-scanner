---
title: "frontend-sensitive-log rule"
description: "Frontend log statement may expose sensitive credentials or PII"
---

# `frontend-sensitive-log`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Frontend log statement may expose sensitive credentials or PII

**Recommendation**: Sanitize log parameters and remove sensitive tokens or user identifiers from console logs

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
console.log("User auth failed for password:", password);

// ✅ Do (Recommended)
console.error("User authentication failed", { username });
```

---

[← weak-secret](/reference/rules/weak-secret) · [Rule Catalog](/reference/rule-catalog) · [backend-sensitive-log →](/reference/rules/backend-sensitive-log)
