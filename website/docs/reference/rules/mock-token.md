---
title: "mock-token rule"
description: "For developers remediating mock-token: Hardcoded mock token found — remove before production deployment"
---

# `mock-token` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Hardcoded mock token found — remove before production deployment

**Recommendation**: Remove hardcoded mock tokens and load credentials from environment variables or key vaults


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
const authHeader = "Bearer google-mock-jwt-token-12345"

// Safer example
authHeader := fmt.Sprintf("Bearer %s", os.Getenv("AUTH_TOKEN"))
```

```ts [TypeScript / JavaScript]
// Unsafe example
const AUTH_HEADER = "Bearer google-mock-jwt-token-12345";

// Safer example
const AUTH_HEADER = `Bearer ${process.env.AUTH_TOKEN}`;
```

```python [Python]
# Unsafe example
AUTH_HEADER = "Bearer google-mock-jwt-token-12345"

# Safer example
auth_header = f"Bearer {os.environ.get('AUTH_TOKEN')}"
```

:::

---

[Rule Catalog](/reference/rule-catalog) · [browser-token-storage →](/reference/rules/browser-token-storage)
