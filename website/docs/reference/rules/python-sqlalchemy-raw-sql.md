---
title: "python-sqlalchemy-raw-sql rule"
description: "For developers remediating python-sqlalchemy-raw-sql: SQLAlchemy raw text expression formatted with dynamic Python f-string or format()"
---

# `python-sqlalchemy-raw-sql` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: SQLAlchemy raw text expression formatted with dynamic Python f-string or format()

**Recommendation**: Use bound parameters (:param_name) with session.execute(text("..."), {"param_name": val})


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```text
// Unsafe example
session.execute(text(f"SELECT * FROM users WHERE username = '{username}'"))

// Safer example
session.execute(text("SELECT * FROM users WHERE username = :u"), {"u": username})
```

---

[← node-mysql-dynamic-query](/reference/rules/node-mysql-dynamic-query) · [Rule Catalog](/reference/rule-catalog) · [python-django-raw-sql →](/reference/rules/python-django-raw-sql)
