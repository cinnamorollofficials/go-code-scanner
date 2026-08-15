---
title: "weak-secret rule"
description: "For developers remediating weak-secret: Default or weak secret value found"
---

# `weak-secret` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Default or weak secret value found

**Recommendation**: Replace default/placeholder secrets with cryptographically strong random values from secure configuration


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
jwtSecret := []byte("change-me-in-production")

// Safer example
jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))
```

```ts [TypeScript / JavaScript]
// Unsafe example
const jwtSecret = "change-me-in-production";

// Safer example
const jwtSecret = process.env.JWT_SECRET_KEY;
```

```python [Python]
# Unsafe example
JWT_SECRET = "change-me-in-production"

# Safer example
JWT_SECRET = os.environ.get("JWT_SECRET_KEY")
```

:::

---

[← permission-bypass](/reference/rules/permission-bypass) · [Rule Catalog](/reference/rule-catalog) · [frontend-sensitive-log →](/reference/rules/frontend-sensitive-log)
