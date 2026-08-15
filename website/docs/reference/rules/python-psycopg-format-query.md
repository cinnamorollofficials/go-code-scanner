---
title: "python-psycopg-format-query rule"
description: "psycopg database cursor executed with Python string formatting instead of query parameters"
---

# `python-psycopg-format-query`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: psycopg database cursor executed with Python string formatting instead of query parameters

**Recommendation**: Pass query parameters as the second tuple argument to cursor.execute(query, (param,))

##### Code Example (Don't vs Do)

```text
// ❌ Don't (Unsafe)
cursor.execute(f"SELECT * FROM items WHERE owner_id = '{owner_id}'")

// ✅ Do (Recommended)
cursor.execute("SELECT * FROM items WHERE owner_id = %s", (owner_id,))
```

---

[← python-django-raw-sql](/reference/rules/python-django-raw-sql) · [Rule Catalog](/reference/rule-catalog) · [java-spring-jpa-native-query →](/reference/rules/java-spring-jpa-native-query)
