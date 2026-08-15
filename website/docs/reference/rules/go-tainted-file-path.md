---
title: "go-tainted-file-path rule"
description: "Untrusted request parameter used directly in file system operation"
---

# `go-tainted-file-path`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `path_traversal`

**Description**: Untrusted request parameter used directly in file system operation

**Recommendation**: Normalize paths, enforce base directory boundaries, and use allowlisted identifiers

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
filePath := r.URL.Query().Get("file")
data, _ := os.ReadFile(filePath)

// ✅ Do (Recommended)
filename := filepath.Base(r.URL.Query().Get("file"))
safePath := filepath.Join("/var/app/storage", filename)
data, _ := os.ReadFile(safePath)
```

---

[← go-weak-cryptographic-hash](/reference/rules/go-weak-cryptographic-hash) · [Rule Catalog](/reference/rule-catalog) · [go-weak-random-secret →](/reference/rules/go-weak-random-secret)
