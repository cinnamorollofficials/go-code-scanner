---
title: "sensitive-json-field rule"
description: "Sensitive struct field may be exposed in JSON serialization"
---

# `sensitive-json-field`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive struct field may be exposed in JSON serialization

**Recommendation**: Use json:"-" struct tag or custom serializer to exclude sensitive attributes

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"password_hash"`
}

// ✅ Do (Recommended)
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"-"`
}
```

---

[← api-struct-response](/reference/rules/api-struct-response) · [Rule Catalog](/reference/rule-catalog) · [go-shell-command →](/reference/rules/go-shell-command)
