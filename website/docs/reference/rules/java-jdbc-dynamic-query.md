---
title: "java-jdbc-dynamic-query rule"
description: "Spring JdbcTemplate executed with concatenated SQL string"
---

# `java-jdbc-dynamic-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Spring JdbcTemplate executed with concatenated SQL string

**Recommendation**: Pass query parameters as separate Object[] or varargs to jdbcTemplate

##### Code Example (Don't vs Do)

```text
// ❌ Don't (Unsafe)
jdbcTemplate.query("SELECT * FROM users WHERE id = " + id, rowMapper)

// ✅ Do (Recommended)
jdbcTemplate.query("SELECT * FROM users WHERE id = ?", rowMapper, id)
```

---

[← java-hibernate-native-query](/reference/rules/java-hibernate-native-query) · [Rule Catalog](/reference/rule-catalog) · [DBSEC-002 →](/reference/rules/dbsec-002)
