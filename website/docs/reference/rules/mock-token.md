---
title: "mock-token rule"
description: "Hardcoded mock token found — remove before production deployment"
---

# `mock-token`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Hardcoded mock token found — remove before production deployment

**Recommendation**: Remove hardcoded mock tokens and load credentials from environment variables or key vaults

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
const authHeader = "Bearer google-mock-jwt-token-12345"

// ✅ Do (Recommended)
authHeader := fmt.Sprintf("Bearer %s", os.Getenv("AUTH_TOKEN"))
```

```ts [TypeScript / JavaScript]
// ❌ Don't (Unsafe)
const AUTH_HEADER = "Bearer google-mock-jwt-token-12345";

// ✅ Do (Recommended)
const AUTH_HEADER = `Bearer ${process.env.AUTH_TOKEN}`;
```

```python [Python]
# ❌ Don't (Unsafe)
AUTH_HEADER = "Bearer google-mock-jwt-token-12345"

# ✅ Do (Recommended)
auth_header = f"Bearer {os.environ.get('AUTH_TOKEN')}"
```

:::

---

[Rule Catalog](/reference/rule-catalog) · [browser-token-storage →](/reference/rules/browser-token-storage)
