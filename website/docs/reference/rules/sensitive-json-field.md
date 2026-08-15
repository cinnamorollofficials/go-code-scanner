---
title: "sensitive-json-field rule"
description: "For developers remediating sensitive-json-field: Sensitive struct field may be exposed in JSON serialization"
---

# `sensitive-json-field` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive struct field may be exposed in JSON serialization

**Recommendation**: Use json:"-" struct tag or custom serializer to exclude sensitive attributes


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"password_hash"`
}

// Safer example
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"-"`
}
```

---

[← api-struct-response](/reference/rules/api-struct-response) · [Rule Catalog](/reference/rule-catalog) · [go-shell-command →](/reference/rules/go-shell-command)
