---
title: "go-weak-cryptographic-hash rule"
description: "For developers remediating go-weak-cryptographic-hash: Weak cryptographic hash algorithm (MD5/SHA1) detected"
---

# `go-weak-cryptographic-hash` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `weak_cryptography`

**Description**: Weak cryptographic hash algorithm (MD5/SHA1) detected

**Recommendation**: Use SHA-256 or stronger algorithms; use bcrypt/argon2 for password hashing


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
hasher := md5.New()
hasher.Write([]byte(password))

// Safer example
hasher := sha256.New()
hasher.Write([]byte(password))
```

```ts [TypeScript / Node.js]
// Unsafe example
const hash = crypto.createHash("md5").update(password).digest("hex");

// Safer example
const hash = crypto.createHash("sha256").update(password).digest("hex");
```

```python [Python]
# Unsafe example
hash_val = hashlib.md5(password.encode()).hexdigest()

# Safer example
hash_val = hashlib.sha256(password.encode()).hexdigest()
```

:::

---

[← go-shell-command](/reference/rules/go-shell-command) · [Rule Catalog](/reference/rule-catalog) · [go-tainted-file-path →](/reference/rules/go-tainted-file-path)
