---
title: "merge-conflict-marker rule"
description: "For developers remediating merge-conflict-marker: Unresolved merge-conflict marker found"
---

# `merge-conflict-marker` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `HIGH`
- **Category**: `repository_hygiene`

**Description**: Unresolved merge-conflict marker found

**Recommendation**: Resolve merge conflict and remove all markers before committing


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
<<<<<<< HEAD
const apiURL = "http://localhost:8080";
=======
const apiURL = "https://api.production.com";
>>>>>>> main

// Safer example
const apiURL = process.env.API_URL || "https://api.production.com";
```

---

[← SQLSAFE-006](/reference/rules/sqlsafe-006) · [Rule Catalog](/reference/rule-catalog) · [javascript-debugger →](/reference/rules/javascript-debugger)
