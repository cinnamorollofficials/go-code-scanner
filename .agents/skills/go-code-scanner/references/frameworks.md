# Framework & Database Model Matrix

Supported languages, database drivers, and frameworks for static AST & taint analysis.

## Primary Frameworks & Drivers

| Language | Database Driver / ORM | Supported Sink APIs | Analyzed Rules |
| :--- | :--- | :--- | :--- |
| **Go** | `database/sql` | `db.Query`, `db.QueryRow`, `db.Exec`, `tx.Exec` | `SQLI-001`, `SQLI-002`, `SQLI-008`, `SQLSAFE-001` |
| **Go** | `gorm.io/gorm` | `db.Raw`, `db.Exec`, `db.Where` | `SQLI-001`, `SQLI-004`, `SQLSAFE-001` |
| **Go** | `jmoiron/sqlx` | `db.NamedQuery`, `db.MustExec` | `SQLI-001`, `SQLI-008` |
| **JS/TS** | `node-postgres` (`pg`) | `client.query` | `SQLI-001` |
