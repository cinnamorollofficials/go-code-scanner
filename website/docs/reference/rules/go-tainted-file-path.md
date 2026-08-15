---
title: "go-tainted-file-path rule"
description: "For developers remediating go-tainted-file-path: Untrusted request parameter used directly in file system operation"
---

# `go-tainted-file-path` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `path_traversal`

**Description**: Untrusted request parameter used directly in file system operation

**Recommendation**: Normalize paths, enforce base directory boundaries, and use allowlisted identifiers


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
filePath := r.URL.Query().Get("file")
data, _ := os.ReadFile(filePath)

// Safer example
filename := filepath.Base(r.URL.Query().Get("file"))
safePath := filepath.Join("/var/app/storage", filename)
data, _ := os.ReadFile(safePath)
```

---

[← go-weak-cryptographic-hash](/reference/rules/go-weak-cryptographic-hash) · [Rule Catalog](/reference/rule-catalog) · [go-weak-random-secret →](/reference/rules/go-weak-random-secret)
