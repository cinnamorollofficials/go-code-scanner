---
title: "SQLAUTH-003 rule"
description: "For developers remediating SQLAUTH-003: Raw query bypasses standard ORM authorization scopes and permission filters"
---

# `SQLAUTH-003` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `raw-query-bypass`

**Description**: Raw query bypasses standard ORM authorization scopes and permission filters

**Recommendation**: Ensure raw queries replicate all security barriers, role restrictions, and tenant scopes provided by ORM repositories


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Raw("SELECT * FROM users")

// Safer example
db.Raw("SELECT * FROM users WHERE organization_id = ? AND role <= ?", orgID, maxRole)
```

:::

---

[← SQLAUTH-002](/reference/rules/sqlauth-002) · [Rule Catalog](/reference/rule-catalog) · [SQLAUTH-004 →](/reference/rules/sqlauth-004)
