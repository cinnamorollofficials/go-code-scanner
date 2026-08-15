---
title: "SQLAUTH-004 rule"
description: "Database query assumes Row-Level Security but explicitly switches to superuser or bypass role"
---

# `SQLAUTH-004`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `rls-misconfiguration`

**Description**: Database query assumes Row-Level Security but explicitly switches to superuser or bypass role

**Recommendation**: Connect and execute application queries using least-privilege non-superuser roles to enforce database Row-Level Security

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
db.Exec("SET ROLE postgres")
db.Query("SELECT * FROM sensitive_documents")

// ✅ Do (Recommended)
db.Exec("SET LOCAL app.current_tenant_id = $1", tenantID)
db.Query("SELECT * FROM sensitive_documents")
```

:::

---

[← SQLAUTH-003](/reference/rules/sqlauth-003) · [Rule Catalog](/reference/rule-catalog) · [hardcoded-api-url →](/reference/rules/hardcoded-api-url)
