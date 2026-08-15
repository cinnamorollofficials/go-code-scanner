---
title: "hardcoded-credential rule"
description: "Hardcoded credential or API secret key found"
---

# `hardcoded-credential`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `secret_leak`

**Description**: Hardcoded credential or API secret key found

**Recommendation**: Extract credentials to environment variables or secret management services

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
apiKey := "synthetic_secret_api_key_12345"

// ✅ Do (Recommended)
apiKey := os.Getenv("STRIPE_API_KEY")
```

```ts [TypeScript / JavaScript]
// ❌ Don't (Unsafe)
const apiKey = "synthetic_secret_api_key_12345";

// ✅ Do (Recommended)
const apiKey = process.env.STRIPE_API_KEY;
```

```python [Python]
# ❌ Don't (Unsafe)
api_key = "synthetic_secret_api_key_12345"

# ✅ Do (Recommended)
api_key = os.environ.get("STRIPE_API_KEY")
```

```java [Java]
// ❌ Don't (Unsafe)
String apiKey = "synthetic_secret_api_key_12345";

// ✅ Do (Recommended)
String apiKey = System.getenv("STRIPE_API_KEY");
```

:::

---

[← sql-string-format](/reference/rules/sql-string-format) · [Rule Catalog](/reference/rule-catalog) · [unsafe-inner-html →](/reference/rules/unsafe-inner-html)
