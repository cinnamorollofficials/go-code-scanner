---
title: "go-weak-random-secret rule"
description: "For developers remediating go-weak-random-secret: Security-sensitive value generated using pseudo-random math/rand package"
---

# `go-weak-random-secret` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `insecure_randomness`

**Description**: Security-sensitive value generated using pseudo-random math/rand package

**Recommendation**: Use crypto/rand for generating tokens, nonces, session identifiers, and secret keys


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
sessionToken := fmt.Sprintf("%d", rand.Intn(1000000))

// Safer example
tokenBytes := make([]byte, 32)
cryptoRand.Read(tokenBytes)
sessionToken := hex.EncodeToString(tokenBytes)
```

---

[← go-tainted-file-path](/reference/rules/go-tainted-file-path) · [Rule Catalog](/reference/rule-catalog) · [javascript-dynamic-eval →](/reference/rules/javascript-dynamic-eval)
