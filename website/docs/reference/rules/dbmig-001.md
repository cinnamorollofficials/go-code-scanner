---
title: "DBMIG-001 rule"
description: "Destructive schema migration detected without guarded rollout or deprecation phase"
---

# `DBMIG-001`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `destructive-migration`

**Description**: Destructive schema migration detected without guarded rollout or deprecation phase

**Recommendation**: Follow the expand-contract migration pattern and avoid immediate column/table drops in live environments

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
ALTER TABLE users DROP COLUMN phone_number;

// ✅ Do (Recommended)
-- Phase 1: Mark column deprecated in application code; Phase 2: Drop after code deployment
```

---

[← go-http-client-without-timeout](/reference/rules/go-http-client-without-timeout) · [Rule Catalog](/reference/rule-catalog) · [DBMIG-002 →](/reference/rules/dbmig-002)
