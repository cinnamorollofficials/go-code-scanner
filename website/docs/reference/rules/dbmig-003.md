---
title: "DBMIG-003 rule"
description: "Security-sensitive key column defined in table definition"
---

# `DBMIG-003`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `schema-integrity`

**Description**: Security-sensitive key column defined in table definition

**Recommendation**: Enforce explicit FOREIGN KEY, UNIQUE, or CHECK constraints on tenant and account scoping columns

##### Code Example (Don't vs Do)

```text
// ❌ Don't (Unsafe)
CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID);

// ✅ Do (Recommended)
CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE);
```

---

[← DBMIG-002](/reference/rules/dbmig-002) · [Rule Catalog](/reference/rule-catalog) · [DBPERF-001 →](/reference/rules/dbperf-001)
