---
title: "backend-sensitive-log rule"
description: "For developers remediating backend-sensitive-log: Backend log statement may expose sensitive credentials or keys"
---

# `backend-sensitive-log` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Backend log statement may expose sensitive credentials or keys

**Recommendation**: Redact sensitive parameters before writing to application log streams


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
log.Printf("Connecting to DB with secret: %s", dbSecret)

// Safer example
log.Printf("Connecting to DB host: %s", dbHost)
```

---

[← frontend-sensitive-log](/reference/rules/frontend-sensitive-log) · [Rule Catalog](/reference/rule-catalog) · [sql-string-format →](/reference/rules/sql-string-format)
