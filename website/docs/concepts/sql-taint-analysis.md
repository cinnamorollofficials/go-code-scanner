---
title: SQL Taint Analysis
description: Deep dive into Abstract Syntax Tree (AST) parsing, interprocedural taint propagation, call graphs, and automated remediation.
---

# AST & SQL Taint Analysis

`security-review` features a native, offline-capable **AST (Abstract Syntax Tree) & SQL Taint Analysis Engine**. Rather than relying purely on naive regex pattern matching, the engine analyzes Go AST syntax structures, models interprocedural parameter flows, reconstructs dynamic query templates, and tracks dataflow provenance from untrusted inputs down to database sinks.

---

## Interprocedural Call Graph Modeling

Unlike intraprocedural linters that inspect only a single function body, `security-review` traces dataflow across helper functions, repository layers, and data access objects (DAOs).

```text
HTTP Handler (Untrusted Source)
      │
      ▼
Helper Function / Middleware
      │
      ▼
Repository Layer
      │
      ▼
Database Driver (Sink: db.Query, db.Exec)
```

If an HTTP parameter is received from a router (Gin, Echo, Chi, Fiber, `net/http`) and passed down through multiple functions before entering a database query, the taint engine tracks each step and reports the full propagation trace.

---

## Core Detection Rules

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| **`SQLI-001`** | `HIGH` | `sql-injection` | Untrusted value concatenated or formatted into executable SQL. |
| **`SQLI-002`** | `HIGH` | `sql-injection` | Untrusted table or column identifier dynamically interpolated. |
| **`SQLI-004`** | `HIGH` | `orm-escape-hatch` | Unsafe raw ORM escape hatch called with dynamic string. |
| **`SQLI-008`** | `MEDIUM` | `bind-mismatch` | Query specifies $N$ bind placeholders but different argument count passed. |

---

## Automated AST Code Remediation (`--fix`)

For standard parameterized queries, `security-review` includes an automated AST rewriter:

```sh
# Preview auto-fix modifications without writing
security-review scan --fix --dry-run

# Apply deterministic fixes directly to source code
security-review scan --fix
```

Example transformation applied:

```diff
- query := "SELECT * FROM users WHERE id = " + id
- row := db.QueryRow(query)
+ query := "SELECT * FROM users WHERE id = $1"
+ row := db.QueryRow(query, id)
```
