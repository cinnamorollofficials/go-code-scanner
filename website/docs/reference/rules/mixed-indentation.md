---
title: "mixed-indentation rule"
description: "Mixed tabs and spaces used for indentation on the same line"
---

# `mixed-indentation`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Mixed tabs and spaces used for indentation on the same line

**Recommendation**: Use a consistent indentation style throughout the project

##### Code Example (Don't vs Do)

```go
// ❌ Don't (Unsafe)
func process() {
	  var x = 10 // Mixed tabs and spaces
}

// ✅ Do (Recommended)
func process() {
	var x = 10 // Consistent tab indentation
}
```

---

[← trailing-whitespace](/reference/rules/trailing-whitespace) · [Rule Catalog](/reference/rule-catalog) · [javascript-console-debug →](/reference/rules/javascript-console-debug)
