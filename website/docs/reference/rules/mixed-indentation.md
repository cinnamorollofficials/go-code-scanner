---
title: "mixed-indentation rule"
description: "For developers remediating mixed-indentation: Mixed tabs and spaces used for indentation on the same line"
---

# `mixed-indentation` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Mixed tabs and spaces used for indentation on the same line

**Recommendation**: Use a consistent indentation style throughout the project


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```go
// Unsafe example
func process() {
	  var x = 10 // Mixed tabs and spaces
}

// Safer example
func process() {
	var x = 10 // Consistent tab indentation
}
```

---

[← trailing-whitespace](/reference/rules/trailing-whitespace) · [Rule Catalog](/reference/rule-catalog) · [javascript-console-debug →](/reference/rules/javascript-console-debug)
