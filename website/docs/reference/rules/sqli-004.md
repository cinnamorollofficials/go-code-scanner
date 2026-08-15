---
title: "SQLI-004 rule"
description: "For developers remediating SQLI-004: Unsafe raw ORM escape hatch called with dynamic or concatenated string"
---

# `SQLI-004` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `orm-escape-hatch`

**Description**: Unsafe raw ORM escape hatch called with dynamic or concatenated string

**Recommendation**: Pass parameters as separate arguments to ORM clauses (e.g. db.Where("name = ?", val)) rather than dynamic string formatting


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go (GORM)]
// Unsafe example
db.Where(fmt.Sprintf("role = '%s'", role)).Find(&users)

// Safer example
db.Where("role = ?", role).Find(&users)
```

:::

---

[← SQLI-002](/reference/rules/sqli-002) · [Rule Catalog](/reference/rule-catalog) · [SQLI-008 →](/reference/rules/sqli-008)
