---
title: "hardcoded-credential rule"
description: "For developers remediating hardcoded-credential: Hardcoded credential or API secret key found"
---

# `hardcoded-credential` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `secret_leak`

**Description**: Hardcoded credential or API secret key found

**Recommendation**: Extract credentials to environment variables or secret management services


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
apiKey := "synthetic_secret_api_key_12345"

// Safer example
apiKey := os.Getenv("STRIPE_API_KEY")
```

```ts [TypeScript / JavaScript]
// Unsafe example
const apiKey = "synthetic_secret_api_key_12345";

// Safer example
const apiKey = process.env.STRIPE_API_KEY;
```

```python [Python]
# Unsafe example
api_key = "synthetic_secret_api_key_12345"

# Safer example
api_key = os.environ.get("STRIPE_API_KEY")
```

```java [Java]
// Unsafe example
String apiKey = "synthetic_secret_api_key_12345";

// Safer example
String apiKey = System.getenv("STRIPE_API_KEY");
```

:::

---

[← sql-string-format](/reference/rules/sql-string-format) · [Rule Catalog](/reference/rule-catalog) · [unsafe-inner-html →](/reference/rules/unsafe-inner-html)
