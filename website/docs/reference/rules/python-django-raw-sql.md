---
title: "python-django-raw-sql rule"
description: "Django raw SQL query constructed with f-string or unsafe .extra() clause"
---

# `python-django-raw-sql`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Django raw SQL query constructed with f-string or unsafe .extra() clause

**Recommendation**: Pass parameters as params list to Model.objects.raw(query, [params]) or use standard ORM filters

##### Code Example (Don't vs Do)

```text
// ❌ Don't (Unsafe)
User.objects.raw(f"SELECT * FROM auth_user WHERE username = '{username}'")

// ✅ Do (Recommended)
User.objects.raw("SELECT * FROM auth_user WHERE username = %s", [username])
```

---

[← python-sqlalchemy-raw-sql](/reference/rules/python-sqlalchemy-raw-sql) · [Rule Catalog](/reference/rule-catalog) · [python-psycopg-format-query →](/reference/rules/python-psycopg-format-query)
