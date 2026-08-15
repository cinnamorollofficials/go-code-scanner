---
title: "java-spring-jpa-native-query rule"
description: "Spring Data JPA native query built via string concatenation"
---

# `java-spring-jpa-native-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Spring Data JPA native query built via string concatenation

**Recommendation**: Use named parameters (:param) or positional parameters (?1) in native @Query annotations

##### Code Example (Don't vs Do)

```text
// ❌ Don't (Unsafe)
@Query(value = "SELECT * FROM users WHERE role = '" + ROLE + "'", nativeQuery = true)

// ✅ Do (Recommended)
@Query(value = "SELECT * FROM users WHERE role = :role", nativeQuery = true)
```

---

[← python-psycopg-format-query](/reference/rules/python-psycopg-format-query) · [Rule Catalog](/reference/rule-catalog) · [java-hibernate-native-query →](/reference/rules/java-hibernate-native-query)
