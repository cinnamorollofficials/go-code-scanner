---
title: "SQLAUTH-003 rule"
description: "Raw query bypasses standard ORM authorization scopes and permission filters"
---

# `SQLAUTH-003`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `raw-query-bypass`

**Description**: Raw query bypasses standard ORM authorization scopes and permission filters

**Recommendation**: Ensure raw queries replicate all security barriers, role restrictions, and tenant scopes provided by ORM repositories

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
db.Raw("SELECT * FROM users")

// ✅ Do (Recommended)
db.Raw("SELECT * FROM users WHERE organization_id = ? AND role <= ?", orgID, maxRole)
```

:::

---

[← SQLAUTH-002](/reference/rules/sqlauth-002) · [Rule Catalog](/reference/rule-catalog) · [SQLAUTH-004 →](/reference/rules/sqlauth-004)
