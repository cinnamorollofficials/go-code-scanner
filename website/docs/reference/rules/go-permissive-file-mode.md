---
title: "go-permissive-file-mode rule"
description: "For developers remediating go-permissive-file-mode: File or directory created with permissive world-writable file permissions (0777)"
---

# `go-permissive-file-mode` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `file_permission`

**Description**: File or directory created with permissive world-writable file permissions (0777)

**Recommendation**: Use minimum required file permissions such as 0600 for files or 0750 for directories


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
os.WriteFile("config.json", data, 0777)

// Safer example
os.WriteFile("config.json", data, 0600)
```

---

[← wildcard-cors-origin](/reference/rules/wildcard-cors-origin) · [Rule Catalog](/reference/rule-catalog) · [debug-mode-enabled →](/reference/rules/debug-mode-enabled)
