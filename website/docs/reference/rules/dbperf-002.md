---
title: "DBPERF-002 rule"
description: "Database query executed inside loop (N+1 query anti-pattern)"
---

# `DBPERF-002`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `n-plus-one`

**Description**: Database query executed inside loop (N+1 query anti-pattern)

**Recommendation**: Batch queries using WHERE id IN (...) or JOINs to fetch data in a single roundtrip

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
for _, userID := range userIDs {
    db.QueryRow("SELECT * FROM profiles WHERE user_id = $1", userID)
}

// ✅ Do (Recommended)
db.Query("SELECT * FROM profiles WHERE user_id IN ($1, $2, ...)", userIDs)
```

---

[← DBPERF-001](/reference/rules/dbperf-001) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-001 →](/reference/rules/sqlsafe-001)
