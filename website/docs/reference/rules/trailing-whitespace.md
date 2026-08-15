---
title: "trailing-whitespace rule"
description: "Trailing whitespace found at end of line"
---

# `trailing-whitespace`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Trailing whitespace found at end of line

**Recommendation**: Remove trailing whitespace at line end

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
const username = "john_doe";

// ✅ Do (Recommended)
const username = "john_doe";
```

---

[← javascript-debugger](/reference/rules/javascript-debugger) · [Rule Catalog](/reference/rule-catalog) · [mixed-indentation →](/reference/rules/mixed-indentation)
