# SQLSAFE-001: Unbounded UPDATE or DELETE Query

| Property | Value |
| :--- | :--- |
| **Rule ID** | `SQLSAFE-001` |
| **Domain** | `reliability` |
| **Category** | `destructive-query` |
| **Severity** | `HIGH` |
| **Confidence** | `HIGH` |
| **Exploitability** | `LIKELY` |
| **CWE** | [CWE-670: Always-Incorrect Control Flow Implementation](https://cwe.mitre.org/data/definitions/670.html) |

## Description

An `UPDATE` or `DELETE` SQL statement is constructed without a `WHERE` clause or bounding predicate.

Executing an unbounded update or delete results in table-wide data modification or complete data loss.

## Vulnerable Example

```go
func purgeUsers(db *sql.DB) error {
    // UNSAFE: Destructive query without WHERE clause
    _, err := db.Exec("DELETE FROM users")
    return err
}
```

## Remediation

Always specify an explicit `WHERE` filter or condition. If a full table purge is genuinely intended, use a dedicated truncation method or explicit batch condition.

```go
func purgeExpiredSessions(db *sql.DB, cutoff time.Time) error {
    // SAFE: Filtered predicate
    _, err := db.Exec("DELETE FROM sessions WHERE expires_at < $1", cutoff)
    return err
}
```
