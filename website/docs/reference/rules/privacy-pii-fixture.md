---
title: "privacy-pii-fixture rule"
description: "Fixture may contain a literal personally identifiable value"
---

# `privacy-pii-fixture`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `governance`
- **Severity**: `MEDIUM`
- **Category**: `privacy_fixture`

**Description**: Fixture may contain a literal personally identifiable value

**Recommendation**: Use clearly synthetic, reserved test data and keep production-derived records out of the repository

##### Code Example (Don't vs Do)

```json
// ❌ Don't (Unsafe)
{"email": "real_person_1985@gmail.com", "ssn": "123-45-6789"}

// ✅ Do (Recommended)
{"email": "user@example.com", "ssn": "000-00-0000"}
```

---

[← privacy-pii-url](/reference/rules/privacy-pii-url) · [Rule Catalog](/reference/rule-catalog) · [privacy-sensitive-response →](/reference/rules/privacy-sensitive-response)
