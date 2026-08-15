---
title: "SQLI-004 rule"
description: "Unsafe raw ORM escape hatch called with dynamic or concatenated string"
---

# `SQLI-004`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `orm-escape-hatch`

**Description**: Unsafe raw ORM escape hatch called with dynamic or concatenated string

**Recommendation**: Pass parameters as separate arguments to ORM clauses (e.g. db.Where("name = ?", val)) rather than dynamic string formatting

##### Code Examples (Don't vs Do)

::: code-group

```go [Go (GORM)]
// ❌ Don't (Unsafe)
db.Where(fmt.Sprintf("role = '%s'", role)).Find(&users)

// ✅ Do (Recommended)
db.Where("role = ?", role).Find(&users)
```

:::

---

[← SQLI-002](/reference/rules/sqli-002) · [Rule Catalog](/reference/rule-catalog) · [SQLI-008 →](/reference/rules/sqli-008)
