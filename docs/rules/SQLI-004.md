# SQLI-004: Unsafe Raw ORM Escape Hatch

| Property | Value |
| :--- | :--- |
| **Rule ID** | `SQLI-004` |
| **Domain** | `security` |
| **Category** | `orm-escape-hatch` |
| **Severity** | `HIGH` |
| **Confidence** | `HIGH` |
| **Exploitability** | `LIKELY` |
| **CWE** | [CWE-89: Improper Neutralization of Special Elements used in an SQL Command ('SQL Injection')](https://cwe.mitre.org/data/definitions/89.html) |

## Description

An ORM query clause method such as `db.Where()`, `db.Raw()`, `db.Order()`, or `db.Having()` (e.g. in GORM, Ent, or SQLX) is invoked with dynamically formatted strings instead of passing arguments as separate parameter placeholders.

Even when using an ORM, calling `.Where("field = " + val)` bypasses ORM parameterization protections and opens the application to SQL injection.

## Vulnerable Example

```go
func findByRole(db *gorm.DB, role string) ([]User, error) {
    var users []User
    // UNSAFE: String interpolation inside ORM .Where clause
    err := db.Where(fmt.Sprintf("role = '%s'", role)).Find(&users).Error
    return users, err
}
```

## Remediation

Pass query expressions with `?` placeholders and supply dynamic variables as additional arguments to the ORM method.

```go
func findByRole(db *gorm.DB, role string) ([]User, error) {
    var users []User
    // SAFE: Parameterized ORM clause
    err := db.Where("role = ?", role).Find(&users).Error
    return users, err
}
```
