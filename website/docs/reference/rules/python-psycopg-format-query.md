---
title: "python-psycopg-format-query rule"
description: "For developers remediating python-psycopg-format-query: psycopg database cursor executed with Python string formatting instead of query parameters"
---

# `python-psycopg-format-query` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: psycopg database cursor executed with Python string formatting instead of query parameters

**Recommendation**: Pass query parameters as the second tuple argument to cursor.execute(query, (param,))


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```text
// Unsafe example
cursor.execute(f"SELECT * FROM items WHERE owner_id = '{owner_id}'")

// Safer example
cursor.execute("SELECT * FROM items WHERE owner_id = %s", (owner_id,))
```

---

[← python-django-raw-sql](/reference/rules/python-django-raw-sql) · [Rule Catalog](/reference/rule-catalog) · [java-spring-jpa-native-query →](/reference/rules/java-spring-jpa-native-query)
