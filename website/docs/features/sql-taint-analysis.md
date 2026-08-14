---
title: AST & SQL Taint Analysis Subsystem
description: Deep Go AST analysis, intraprocedural string taint propagation, SQL template reconstruction, and dataflow tracing.
---

# AST & SQL Taint Analysis Subsystem

`security-review` includes a native, offline-capable **AST (Abstract Syntax Tree) & SQL Taint Analysis Engine** (`sqltaint`). Unlike simple regex pattern matching, the `sqltaint` engine inspects Go syntax trees, traces variable assignments and format calls, reconstructs dynamic SQL templates, and models database driver sinks with rich dataflow provenance.

---

## Key Capabilities

1. **Go AST Parsing**: Inspects Go syntax trees using `go/ast` and `go/parser` without executing project code or invoking external build hooks.
2. **Intraprocedural Taint Propagation**: Traces untrusted variables through binary additions (`+`), `fmt.Sprintf` calls, and local identifier assignments.
3. **SQL Template & Hole Classification**: Deconstructs SQL query expressions into constant query fragments and dynamic holes (`HoleContextValue`, `HoleContextIdentifier`, `HoleContextClause`).
4. **Dataflow Tracing**: Generates reproducible step-by-step traces (`Source` → `Propagator` → `Sink`) serialized directly in JSON and SARIF reports.
5. **Driver & ORM Modeling**: Supports standard Go `database/sql` driver methods (`Query`, `QueryRow`, `Exec`, `NamedExec`) as well as ORM query builders (`gorm.DB`, `sqlx`).

---

## Supported Rules & Threat Models

### 1. `SQLI-001`: Untrusted Value Reaches Executable SQL

Triggered when dynamic string concatenation or `fmt.Sprintf` formats untrusted parameters into executable SQL queries.

::: code-group
```go [❌ Don't (Unsafe)]
func getUser(db *sql.DB, id string) (*sql.Row, error) {
    // ❌ UNSAFE: Dynamic string concatenation into query sink
    query := "SELECT id, name, email FROM users WHERE id = " + id
    return db.QueryRow(query), nil
}
```

```go [✅ Do (Safe)]
func getUser(db *sql.DB, id string) (*sql.Row, error) {
    // ✅ SAFE: Parameterized bind placeholder ($1 or ?)
    query := "SELECT id, name, email FROM users WHERE id = $1"
    return db.QueryRow(query, id), nil
}
```
:::

---

### 2. `SQLI-002`: Untrusted SQL Identifier Interpolation

Triggered when table names, column names, or `ORDER BY` / `GROUP BY` identifiers are dynamically interpolated without being validated against an explicit allow-list.

::: code-group
```go [❌ Don't (Unsafe)]
func listUsers(db *sql.DB, sortColumn string) (*sql.Rows, error) {
    // ❌ UNSAFE: Identifier interpolation without validation
    query := fmt.Sprintf("SELECT id, name FROM users ORDER BY %s ASC", sortColumn)
    return db.Query(query)
}
```

```go [✅ Do (Safe)]
var allowedColumns = map[string]string{
    "name":       "name",
    "created_at": "created_at",
    "id":         "id",
}

func listUsers(db *sql.DB, sortColumn string) (*sql.Rows, error) {
    col, ok := allowedColumns[sortColumn]
    if !ok {
        return nil, fmt.Errorf("invalid sort column: %q", sortColumn)
    }
    // ✅ SAFE: Validated against strict constant allow-list
    query := fmt.Sprintf("SELECT id, name FROM users ORDER BY %s ASC", col)
    return db.Query(query)
}
```
:::

---

### 3. `SQLI-004`: Unsafe Raw ORM Escape Hatch

Triggered when ORM query-building methods (`db.Where()`, `db.Raw()`, `db.Order()`, `db.Having()`) are called with dynamically formatted strings rather than parameterized placeholders.

::: code-group
```go [❌ Don't (Unsafe)]
func findByRole(db *gorm.DB, role string) ([]User, error) {
    var users []User
    // ❌ UNSAFE: Bypasses ORM parameterization
    err := db.Where(fmt.Sprintf("role = '%s'", role)).Find(&users).Error
    return users, err
}
```

```go [✅ Do (Safe)]
func findByRole(db *gorm.DB, role string) ([]User, error) {
    var users []User
    // ✅ SAFE: Parameterized ORM clause
    err := db.Where("role = ?", role).Find(&users).Error
    return users, err
}
```
:::

---

### 4. `SQLI-008`: Placeholder / Parameter Count Mismatch

Triggered when the number of parameter bind placeholders (`?` or `$1, $2, ...`) in an SQL query string does not match the actual number of arguments passed to the query function.

::: code-group
```go [❌ Don't (Unsafe)]
func getTenantUser(db *sql.DB, userID string) (*sql.Row, error) {
    // ❌ UNSAFE: 2 placeholders (? and ?), but only 1 argument passed
    return db.QueryRow("SELECT id FROM users WHERE id = ? AND tenant_id = ?", userID), nil
}
```

```go [✅ Do (Safe)]
func getTenantUser(db *sql.DB, userID, tenantID string) (*sql.Row, error) {
    // ✅ SAFE: 2 placeholders and 2 arguments passed
    return db.QueryRow("SELECT id FROM users WHERE id = ? AND tenant_id = ?", userID, tenantID), nil
}
```
:::

---

### 5. `SQLSAFE-001`: Unbounded UPDATE or DELETE

Triggered when an `UPDATE` or `DELETE` SQL statement is constructed without a `WHERE` clause or bounding filter, preventing catastrophic table-wide data modification.

::: code-group
```go [❌ Don't (Unsafe)]
func purgeUsers(db *sql.DB) error {
    // ❌ UNSAFE: Destructive query without WHERE clause
    _, err := db.Exec("DELETE FROM users")
    return err
}
```

```go [✅ Do (Safe)]
func purgeUsers(db *sql.DB, cutoff time.Time) error {
    // ✅ SAFE: Bound by explicit timestamp predicate
    _, err := db.Exec("DELETE FROM users WHERE created_at < $1", cutoff)
    return err
}
```
:::

---

## Enabling in Configuration

To enable the `sqltaint` scanner in your `security-review.json` configuration:

```json
{
  "version": 1,
  "scanners": {
    "sqltaint": {
      "enabled": true,
      "required": false
    }
  }
}
```

Or run directly via the CLI:

```bash
security-review scan --mode=full
```
