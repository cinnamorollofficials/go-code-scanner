---
title: "hardcoded-api-url rule"
description: "Hardcoded localhost API URL found — load dynamically from environment variable"
---

# `hardcoded-api-url`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `configuration_leak`

**Description**: Hardcoded localhost API URL found — load dynamically from environment variable

**Recommendation**: Configure API endpoints dynamically via environment variables for different environments

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
const API_URL = "http://localhost:8080/api/v1";

// ✅ Do (Recommended)
const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";
```

---

[← SQLAUTH-004](/reference/rules/sqlauth-004) · [Rule Catalog](/reference/rule-catalog) · [tls-insecure-skip-verify →](/reference/rules/tls-insecure-skip-verify)
