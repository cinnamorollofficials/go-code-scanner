---
title: "DBSEC-002 rule"
description: "For developers remediating DBSEC-002: Sensitive credentials or PII fields logged to application tracing stream"
---

# `DBSEC-002` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive credentials or PII fields logged to application tracing stream

**Recommendation**: Redact credentials, tokens, and payment card details before writing to log sinks


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
logger.info("Processing payment for card:", cardToken, secretKey);

// Safer example
logger.info("Processing payment for transaction ID:", transactionId);
```

---

[← java-jdbc-dynamic-query](/reference/rules/java-jdbc-dynamic-query) · [Rule Catalog](/reference/rule-catalog) · [DBSEC-003 →](/reference/rules/dbsec-003)
