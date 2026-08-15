---
title: "java-hibernate-native-query rule"
description: "For developers remediating java-hibernate-native-query: Hibernate createNativeQuery executed with dynamic string concatenation"
---

# `java-hibernate-native-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Hibernate createNativeQuery executed with dynamic string concatenation

**Recommendation**: Use parameterized placeholders and bind parameters via query.setParameter()


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```text
// Unsafe example
session.createNativeQuery("SELECT * FROM orders WHERE status = '" + status + "'")

// Safer example
session.createNativeQuery("SELECT * FROM orders WHERE status = :status").setParameter("status", status)
```

---

[← java-spring-jpa-native-query](/reference/rules/java-spring-jpa-native-query) · [Rule Catalog](/reference/rule-catalog) · [java-jdbc-dynamic-query →](/reference/rules/java-jdbc-dynamic-query)
