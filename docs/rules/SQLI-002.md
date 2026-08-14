# SQLI-002: Untrusted SQL Identifier Interpolation

| Property | Value |
| :--- | :--- |
| **Rule ID** | `SQLI-002` |
| **Domain** | `security` |
| **Category** | `sql-injection` |
| **Severity** | `HIGH` |
| **Confidence** | `HIGH` |
| **Exploitability** | `LIKELY` |
| **CWE** | [CWE-89: Improper Neutralization of Special Elements used in an SQL Command ('SQL Injection')](https://cwe.mitre.org/data/definitions/89.html) |

## Description

Untrusted user input is interpolated into SQL identifier positions such as table names, column names, or `ORDER BY` / `GROUP BY` clauses.

Because SQL query drivers typically do not allow bind parameters for table or column identifiers (e.g. `SELECT * FROM $1` is invalid syntax in SQL), developers often resort to string formatting without validating against an allow-list.

## Vulnerable Example

```go
func getSortedUsers(db *sql.DB, sortBy string) (*sql.Rows, error) {
    // UNSAFE: Direct interpolation of column name
    query := fmt.Sprintf("SELECT id, name FROM users ORDER BY %s ASC", sortBy)
    return db.Query(query)
}
```

## Remediation

Validate incoming identifier parameters against a strict allow-list of known safe constant identifiers.

```go
var allowedSortColumns = map[string]string{
    "name":       "name",
    "created_at": "created_at",
    "id":         "id",
}

func getSortedUsers(db *sql.DB, sortBy string) (*sql.Rows, error) {
    column, ok := allowedSortColumns[sortBy]
    if !ok {
        return nil, errors.New("invalid sort column")
    }
    query := fmt.Sprintf("SELECT id, name FROM users ORDER BY %s ASC", column)
    return db.Query(query)
}
```
