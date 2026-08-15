---
title: "SQLI-008 rule"
description: "For developers remediating SQLI-008: SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed"
---

# `SQLI-008` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `bind-mismatch`

**Description**: SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed

**Recommendation**: Ensure the number of bind placeholders ($1, ?) matches the count of passed query arguments exactly


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id)

// Safer example
db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id, tenantID)
```

:::

---

[← SQLI-004](/reference/rules/sqli-004) · [Rule Catalog](/reference/rule-catalog) · [SQLI-011 →](/reference/rules/sqli-011)
