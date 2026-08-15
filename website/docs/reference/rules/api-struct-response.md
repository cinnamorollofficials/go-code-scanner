---
title: "api-struct-response rule"
description: "Internal domain struct may be serialized directly into HTTP response"
---

# `api-struct-response`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Internal domain struct may be serialized directly into HTTP response

**Recommendation**: Map internal domain entities to explicit response DTOs to avoid leaking sensitive fields

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
var user User // Contains HashedPassword, SecretToken
c.JSON(http.StatusOK, user)

// ✅ Do (Recommended)
response := UserResponse{ID: user.ID, Email: user.Email}
c.JSON(http.StatusOK, response)
```

---

[← dynamic-order](/reference/rules/dynamic-order) · [Rule Catalog](/reference/rule-catalog) · [sensitive-json-field →](/reference/rules/sensitive-json-field)
