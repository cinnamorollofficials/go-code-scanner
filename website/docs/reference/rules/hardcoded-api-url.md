---
title: "hardcoded-api-url rule"
description: "For developers remediating hardcoded-api-url: Hardcoded localhost API URL found — load dynamically from environment variable"
---

# `hardcoded-api-url` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `configuration_leak`

**Description**: Hardcoded localhost API URL found — load dynamically from environment variable

**Recommendation**: Configure API endpoints dynamically via environment variables for different environments


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
const API_URL = "http://localhost:8080/api/v1";

// Safer example
const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";
```

---

[← SQLAUTH-004](/reference/rules/sqlauth-004) · [Rule Catalog](/reference/rule-catalog) · [tls-insecure-skip-verify →](/reference/rules/tls-insecure-skip-verify)
