---
title: "DBPERF-002 rule"
description: "For developers remediating DBPERF-002: Database query executed inside loop (N+1 query anti-pattern)"
---

# `DBPERF-002` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `n-plus-one`

**Description**: Database query executed inside loop (N+1 query anti-pattern)

**Recommendation**: Batch queries using WHERE id IN (...) or JOINs to fetch data in a single roundtrip


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
for _, userID := range userIDs {
    db.QueryRow("SELECT * FROM profiles WHERE user_id = $1", userID)
}

// Safer example
db.Query("SELECT * FROM profiles WHERE user_id IN ($1, $2, ...)", userIDs)
```

---

[← DBPERF-001](/reference/rules/dbperf-001) · [Rule Catalog](/reference/rule-catalog) · [SQLSAFE-001 →](/reference/rules/sqlsafe-001)
