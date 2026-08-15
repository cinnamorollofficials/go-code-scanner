---
title: "DBMIG-002 rule"
description: "For developers remediating DBMIG-002: Database migration file lacks reversible rollback instructions"
---

# `DBMIG-002` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `migration-safety`

**Description**: Database migration file lacks reversible rollback instructions

**Recommendation**: Always provide corresponding down migrations or automated rollback scripts for disaster recovery


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
-- no-down: Irreversible migration

// Safer example
-- Provide matching down.sql migration with schema restore steps
```

---

[← DBMIG-001](/reference/rules/dbmig-001) · [Rule Catalog](/reference/rule-catalog) · [DBMIG-003 →](/reference/rules/dbmig-003)
