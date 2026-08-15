---
title: "SQLAUTH-001 rule"
description: "For developers remediating SQLAUTH-001: Multi-tenant entity queried without tenant_id or organization_id scoping constraint"
---

# `SQLAUTH-001` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `multi-tenant-isolation`

**Description**: Multi-tenant entity queried without tenant_id or organization_id scoping constraint

**Recommendation**: Enforce explicit tenant_id or organization_id filtering on all multi-tenant queries to prevent cross-tenant data access


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func getAccounts(db *sql.DB) (*sql.Rows, error) {
    return db.Query("SELECT * FROM accounts WHERE status = 'active'")
}

// Safer example
func getAccounts(db *sql.DB, tenantID string) (*sql.Rows, error) {
    return db.Query("SELECT * FROM accounts WHERE tenant_id = $1 AND status = 'active'", tenantID)
}
```

:::

---

[← SQLI-012](/reference/rules/sqli-012) · [Rule Catalog](/reference/rule-catalog) · [SQLAUTH-002 →](/reference/rules/sqlauth-002)
