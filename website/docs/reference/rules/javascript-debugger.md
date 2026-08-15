---
title: "javascript-debugger rule"
description: "For developers remediating javascript-debugger: JavaScript debugger statement found"
---

# `javascript-debugger` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `MEDIUM`
- **Category**: `debug_code`

**Description**: JavaScript debugger statement found

**Recommendation**: Remove debugger statement before committing


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Example

```ts
// Unsafe example
function calculateTotal(items: Item[]) {
  debugger; // Leftover debug statement
  return items.reduce((acc, item) => acc + item.price, 0);
}

// Safer example
function calculateTotal(items: Item[]) {
  return items.reduce((acc, item) => acc + item.price, 0);
}
```

---

[← merge-conflict-marker](/reference/rules/merge-conflict-marker) · [Rule Catalog](/reference/rule-catalog) · [trailing-whitespace →](/reference/rules/trailing-whitespace)
