---
title: "weak-secret rule"
description: "Default or weak secret value found"
---

# `weak-secret`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Default or weak secret value found

**Recommendation**: Replace default/placeholder secrets with cryptographically strong random values from secure configuration

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
jwtSecret := []byte("change-me-in-production")

// ✅ Do (Recommended)
jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))
```

```ts [TypeScript / JavaScript]
// ❌ Don't (Unsafe)
const jwtSecret = "change-me-in-production";

// ✅ Do (Recommended)
const jwtSecret = process.env.JWT_SECRET_KEY;
```

```python [Python]
# ❌ Don't (Unsafe)
JWT_SECRET = "change-me-in-production"

# ✅ Do (Recommended)
JWT_SECRET = os.environ.get("JWT_SECRET_KEY")
```

:::

---

[← permission-bypass](/reference/rules/permission-bypass) · [Rule Catalog](/reference/rule-catalog) · [frontend-sensitive-log →](/reference/rules/frontend-sensitive-log)
