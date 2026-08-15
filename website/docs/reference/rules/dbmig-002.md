---
title: "DBMIG-002 rule"
description: "Database migration file lacks reversible rollback instructions"
---

# `DBMIG-002`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `migration-safety`

**Description**: Database migration file lacks reversible rollback instructions

**Recommendation**: Always provide corresponding down migrations or automated rollback scripts for disaster recovery

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
-- no-down: Irreversible migration

// ✅ Do (Recommended)
-- Provide matching down.sql migration with schema restore steps
```

---

[← DBMIG-001](/reference/rules/dbmig-001) · [Rule Catalog](/reference/rule-catalog) · [DBMIG-003 →](/reference/rules/dbmig-003)
