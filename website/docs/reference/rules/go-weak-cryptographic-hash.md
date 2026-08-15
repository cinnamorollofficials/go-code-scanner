---
title: "go-weak-cryptographic-hash rule"
description: "Weak cryptographic hash algorithm (MD5/SHA1) detected"
---

# `go-weak-cryptographic-hash`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `weak_cryptography`

**Description**: Weak cryptographic hash algorithm (MD5/SHA1) detected

**Recommendation**: Use SHA-256 or stronger algorithms; use bcrypt/argon2 for password hashing

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
hasher := md5.New()
hasher.Write([]byte(password))

// ✅ Do (Recommended)
hasher := sha256.New()
hasher.Write([]byte(password))
```

```ts [TypeScript / Node.js]
// ❌ Don't (Unsafe)
const hash = crypto.createHash("md5").update(password).digest("hex");

// ✅ Do (Recommended)
const hash = crypto.createHash("sha256").update(password).digest("hex");
```

```python [Python]
# ❌ Don't (Unsafe)
hash_val = hashlib.md5(password.encode()).hexdigest()

# ✅ Do (Recommended)
hash_val = hashlib.sha256(password.encode()).hexdigest()
```

:::

---

[← go-shell-command](/reference/rules/go-shell-command) · [Rule Catalog](/reference/rule-catalog) · [go-tainted-file-path →](/reference/rules/go-tainted-file-path)
