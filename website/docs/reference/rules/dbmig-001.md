---
title: "DBMIG-001 rule"
description: "For developers remediating DBMIG-001: Destructive schema migration detected without guarded rollout or deprecation phase"
---

# `DBMIG-001` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `destructive-migration`

**Description**: Destructive schema migration detected without guarded rollout or deprecation phase

**Recommendation**: Follow the expand-contract migration pattern and avoid immediate column/table drops in live environments


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
ALTER TABLE users DROP COLUMN phone_number;

// Safer example
-- Phase 1: Mark column deprecated in application code; Phase 2: Drop after code deployment
```

---

[← go-http-client-without-timeout](/reference/rules/go-http-client-without-timeout) · [Rule Catalog](/reference/rule-catalog) · [DBMIG-002 →](/reference/rules/dbmig-002)
