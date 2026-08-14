# SQLI-001: Untrusted Value Reaches Executable SQL

| Property | Value |
| :--- | :--- |
| **Rule ID** | `SQLI-001` |
| **Domain** | `security` |
| **Category** | `sql-injection` |
| **Severity** | `HIGH` |
| **Confidence** | `HIGH` |
| **Exploitability** | `LIKELY` |
| **CWE** | [CWE-89: Improper Neutralization of Special Elements used in an SQL Command ('SQL Injection')](https://cwe.mitre.org/data/definitions/89.html) |

## Description

Untrusted user input or dynamic string variables are directly concatenated or formatted using `fmt.Sprintf` / `+` into an SQL query string before being passed to a database execution method (e.g. `db.Query`, `db.Exec`, `db.QueryRow`).

This pattern allows an attacker to manipulate the structure of the SQL command, bypass authentication, exfiltrate data, or modify database state.

## Vulnerable Example

```go
func getUserByID(db *sql.DB, userID string) (*User, error) {
    // UNSAFE: Dynamic string concatenation
    query := "SELECT id, name, email FROM users WHERE id = " + userID
    row := db.QueryRow(query)
    // ...
}
```

## Remediation

Always use parameterized queries with bind variables (`$1`, `?`, `:name` depending on your SQL driver).

```go
func getUserByID(db *sql.DB, userID string) (*User, error) {
    // SAFE: Parameterized query
    query := "SELECT id, name, email FROM users WHERE id = $1"
    row := db.QueryRow(query, userID)
    // ...
}
```
