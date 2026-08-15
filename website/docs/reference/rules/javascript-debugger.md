---
title: "javascript-debugger rule"
description: "JavaScript debugger statement found"
---

# `javascript-debugger`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `quality`
- **Severity**: `MEDIUM`
- **Category**: `debug_code`

**Description**: JavaScript debugger statement found

**Recommendation**: Remove debugger statement before committing

##### Code Example (Don't vs Do)

```ts
// ❌ Don't (Unsafe)
function calculateTotal(items: Item[]) {
  debugger; // Leftover debug statement
  return items.reduce((acc, item) => acc + item.price, 0);
}

// ✅ Do (Recommended)
function calculateTotal(items: Item[]) {
  return items.reduce((acc, item) => acc + item.price, 0);
}
```

---

[← merge-conflict-marker](/reference/rules/merge-conflict-marker) · [Rule Catalog](/reference/rule-catalog) · [trailing-whitespace →](/reference/rules/trailing-whitespace)
