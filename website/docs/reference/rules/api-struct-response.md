---
title: "api-struct-response rule"
description: "For developers remediating api-struct-response: Internal domain struct may be serialized directly into HTTP response"
---

# `api-struct-response` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Internal domain struct may be serialized directly into HTTP response

**Recommendation**: Map internal domain entities to explicit response DTOs to avoid leaking sensitive fields


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
var user User // Contains HashedPassword, SecretToken
c.JSON(http.StatusOK, user)

// Safer example
response := UserResponse{ID: user.ID, Email: user.Email}
c.JSON(http.StatusOK, response)
```

---

[← dynamic-order](/reference/rules/dynamic-order) · [Rule Catalog](/reference/rule-catalog) · [sensitive-json-field →](/reference/rules/sensitive-json-field)
