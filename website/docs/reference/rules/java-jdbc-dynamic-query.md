---
title: "java-jdbc-dynamic-query rule"
description: "For developers remediating java-jdbc-dynamic-query: Spring JdbcTemplate executed with concatenated SQL string"
---

# `java-jdbc-dynamic-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Spring JdbcTemplate executed with concatenated SQL string

**Recommendation**: Pass query parameters as separate Object[] or varargs to jdbcTemplate


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```text
// Unsafe example
jdbcTemplate.query("SELECT * FROM users WHERE id = " + id, rowMapper)

// Safer example
jdbcTemplate.query("SELECT * FROM users WHERE id = ?", rowMapper, id)
```

---

[← java-hibernate-native-query](/reference/rules/java-hibernate-native-query) · [Rule Catalog](/reference/rule-catalog) · [DBSEC-002 →](/reference/rules/dbsec-002)
