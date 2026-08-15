---
title: "SQLAUTH-004 rule"
description: "For developers remediating SQLAUTH-004: Database query assumes Row-Level Security but explicitly switches to superuser or bypass role"
---

# `SQLAUTH-004` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `rls-misconfiguration`

**Description**: Database query assumes Row-Level Security but explicitly switches to superuser or bypass role

**Recommendation**: Connect and execute application queries using least-privilege non-superuser roles to enforce database Row-Level Security


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Exec("SET ROLE postgres")
db.Query("SELECT * FROM sensitive_documents")

// Safer example
db.Exec("SET LOCAL app.current_tenant_id = $1", tenantID)
db.Query("SELECT * FROM sensitive_documents")
```

:::

---

[← SQLAUTH-003](/reference/rules/sqlauth-003) · [Rule Catalog](/reference/rule-catalog) · [hardcoded-api-url →](/reference/rules/hardcoded-api-url)
