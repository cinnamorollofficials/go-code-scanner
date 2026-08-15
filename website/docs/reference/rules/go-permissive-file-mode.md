---
title: "go-permissive-file-mode rule"
description: "File or directory created with permissive world-writable file permissions (0777)"
---

# `go-permissive-file-mode`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `file_permission`

**Description**: File or directory created with permissive world-writable file permissions (0777)

**Recommendation**: Use minimum required file permissions such as 0600 for files or 0750 for directories

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
os.WriteFile("config.json", data, 0777)

// ✅ Do (Recommended)
os.WriteFile("config.json", data, 0600)
```

---

[← wildcard-cors-origin](/reference/rules/wildcard-cors-origin) · [Rule Catalog](/reference/rule-catalog) · [debug-mode-enabled →](/reference/rules/debug-mode-enabled)
