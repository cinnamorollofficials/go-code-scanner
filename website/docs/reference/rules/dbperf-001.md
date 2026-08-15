---
title: "DBPERF-001 rule"
description: "Public dataset queried without an explicit LIMIT or pagination boundary"
---

# `DBPERF-001`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `query-performance`

**Description**: Public dataset queried without an explicit LIMIT or pagination boundary

**Recommendation**: Always enforce LIMIT and OFFSET / cursor pagination to prevent unbounded memory allocation and DB stalls

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
db.Query("SELECT * FROM events WHERE created_at > $1", startTime)

// ✅ Do (Recommended)
db.Query("SELECT * FROM events WHERE created_at > $1 ORDER BY id ASC LIMIT 100", startTime)
```

---

[← DBMIG-003](/reference/rules/dbmig-003) · [Rule Catalog](/reference/rule-catalog) · [DBPERF-002 →](/reference/rules/dbperf-002)
