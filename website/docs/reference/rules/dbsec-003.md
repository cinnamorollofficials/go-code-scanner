---
title: "DBSEC-003 rule"
description: "Internal database driver error exposed directly in HTTP client response"
---

# `DBSEC-003`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `information_exposure`

**Description**: Internal database driver error exposed directly in HTTP client response

**Recommendation**: Log the internal database error securely on the server and return a sanitized, generic error message to the client

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

// ✅ Do (Recommended)
c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
```

---

[← DBSEC-002](/reference/rules/dbsec-002) · [Rule Catalog](/reference/rule-catalog) · [SQLI-001 →](/reference/rules/sqli-001)
