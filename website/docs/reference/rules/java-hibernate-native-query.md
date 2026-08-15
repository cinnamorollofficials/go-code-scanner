---
title: "java-hibernate-native-query rule"
description: "Hibernate createNativeQuery executed with dynamic string concatenation"
---

# `java-hibernate-native-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Hibernate createNativeQuery executed with dynamic string concatenation

**Recommendation**: Use parameterized placeholders and bind parameters via query.setParameter()

##### Code Example (Don't vs Do)

```text
// ❌ Don't (Unsafe)
session.createNativeQuery("SELECT * FROM orders WHERE status = '" + status + "'")

// ✅ Do (Recommended)
session.createNativeQuery("SELECT * FROM orders WHERE status = :status").setParameter("status", status)
```

---

[← java-spring-jpa-native-query](/reference/rules/java-spring-jpa-native-query) · [Rule Catalog](/reference/rule-catalog) · [java-jdbc-dynamic-query →](/reference/rules/java-jdbc-dynamic-query)
