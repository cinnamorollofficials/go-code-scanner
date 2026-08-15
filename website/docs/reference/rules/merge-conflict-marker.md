---
title: "merge-conflict-marker rule"
description: "Unresolved merge-conflict marker found"
---

# `merge-conflict-marker`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `HIGH`
- **Category**: `repository_hygiene`

**Description**: Unresolved merge-conflict marker found

**Recommendation**: Resolve merge conflict and remove all markers before committing

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
<<<<<<< HEAD
const apiURL = "http://localhost:8080";
=======
const apiURL = "https://api.production.com";
>>>>>>> main

// ✅ Do (Recommended)
const apiURL = process.env.API_URL || "https://api.production.com";
```

---

[← SQLSAFE-006](/reference/rules/sqlsafe-006) · [Rule Catalog](/reference/rule-catalog) · [javascript-debugger →](/reference/rules/javascript-debugger)
