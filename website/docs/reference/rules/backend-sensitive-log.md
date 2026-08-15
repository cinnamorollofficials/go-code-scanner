---
title: "backend-sensitive-log rule"
description: "Backend log statement may expose sensitive credentials or keys"
---

# `backend-sensitive-log`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Backend log statement may expose sensitive credentials or keys

**Recommendation**: Redact sensitive parameters before writing to application log streams

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
log.Printf("Connecting to DB with secret: %s", dbSecret)

// ✅ Do (Recommended)
log.Printf("Connecting to DB host: %s", dbHost)
```

---

[← frontend-sensitive-log](/reference/rules/frontend-sensitive-log) · [Rule Catalog](/reference/rule-catalog) · [sql-string-format →](/reference/rules/sql-string-format)
