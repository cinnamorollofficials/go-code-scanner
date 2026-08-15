---
title: "DBSEC-002 rule"
description: "Sensitive credentials or PII fields logged to application tracing stream"
---

# `DBSEC-002`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive credentials or PII fields logged to application tracing stream

**Recommendation**: Redact credentials, tokens, and payment card details before writing to log sinks

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
logger.info("Processing payment for card:", cardToken, secretKey);

// ✅ Do (Recommended)
logger.info("Processing payment for transaction ID:", transactionId);
```

---

[← java-jdbc-dynamic-query](/reference/rules/java-jdbc-dynamic-query) · [Rule Catalog](/reference/rule-catalog) · [DBSEC-003 →](/reference/rules/dbsec-003)
