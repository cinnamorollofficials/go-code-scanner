# SQLI-008: SQL Placeholder / Bind Parameter Mismatch

| Property | Value |
| :--- | :--- |
| **Rule ID** | `SQLI-008` |
| **Domain** | `security` |
| **Category** | `bind-mismatch` |
| **Severity** | `MEDIUM` |
| **Confidence** | `HIGH` |
| **Exploitability** | `UNLIKELY` |
| **CWE** | [CWE-687: Function Call With Incorrect Number of Arguments](https://cwe.mitre.org/data/definitions/687.html) |

## Description

The number of parameter bind placeholders (e.g. `?`, `$1, $2`) in a parameterized SQL query literal does not match the number of parameter arguments passed into the execution method.

This typically causes runtime database driver errors, unexpected null bindings, or skipped predicate evaluation.

## Vulnerable Example

```go
func getActiveUser(db *sql.DB, id string) (*User, error) {
    // UNSAFE: Query has 2 placeholders (? and ?), but only 1 argument is provided
    row := db.QueryRow("SELECT id, name FROM users WHERE id = ? AND is_active = ?", id)
    // ...
}
```

## Remediation

Ensure the count of bind placeholders matches the number of passed arguments exactly.

```go
func getActiveUser(db *sql.DB, id string) (*User, error) {
    // SAFE: Query has 2 placeholders and 2 arguments
    row := db.QueryRow("SELECT id, name FROM users WHERE id = ? AND is_active = ?", id, true)
    // ...
}
```
