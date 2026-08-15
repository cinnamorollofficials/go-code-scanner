---
title: "DBMIG-003 rule"
description: "For developers remediating DBMIG-003: Security-sensitive key column defined in table definition"
---

# `DBMIG-003` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `schema-integrity`

**Description**: Security-sensitive key column defined in table definition

**Recommendation**: Enforce explicit FOREIGN KEY, UNIQUE, or CHECK constraints on tenant and account scoping columns


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```text
// Unsafe example
CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID);

// Safer example
CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE);
```

---

[← DBMIG-002](/reference/rules/dbmig-002) · [Rule Catalog](/reference/rule-catalog) · [DBPERF-001 →](/reference/rules/dbperf-001)
