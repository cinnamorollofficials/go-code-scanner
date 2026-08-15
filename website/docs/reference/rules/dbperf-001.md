---
title: "DBPERF-001 rule"
description: "For developers remediating DBPERF-001: Public dataset queried without an explicit LIMIT or pagination boundary"
---

# `DBPERF-001` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `query-performance`

**Description**: Public dataset queried without an explicit LIMIT or pagination boundary

**Recommendation**: Always enforce LIMIT and OFFSET / cursor pagination to prevent unbounded memory allocation and DB stalls


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
db.Query("SELECT * FROM events WHERE created_at > $1", startTime)

// Safer example
db.Query("SELECT * FROM events WHERE created_at > $1 ORDER BY id ASC LIMIT 100", startTime)
```

---

[← DBMIG-003](/reference/rules/dbmig-003) · [Rule Catalog](/reference/rule-catalog) · [DBPERF-002 →](/reference/rules/dbperf-002)
