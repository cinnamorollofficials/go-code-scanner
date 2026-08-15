---
title: "privacy-pii-fixture rule"
description: "For developers remediating privacy-pii-fixture: Fixture may contain a literal personally identifiable value"
---

# `privacy-pii-fixture` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `MEDIUM`
- **Category**: `privacy_fixture`

**Description**: Fixture may contain a literal personally identifiable value

**Recommendation**: Use clearly synthetic, reserved test data and keep production-derived records out of the repository


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```json
// Unsafe example
{"email": "real_person_1985@gmail.com", "ssn": "123-45-6789"}

// Safer example
{"email": "user@example.com", "ssn": "000-00-0000"}
```

---

[← privacy-pii-url](/reference/rules/privacy-pii-url) · [Rule Catalog](/reference/rule-catalog) · [privacy-sensitive-response →](/reference/rules/privacy-sensitive-response)
