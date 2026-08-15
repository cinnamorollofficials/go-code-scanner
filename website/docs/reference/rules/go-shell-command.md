---
title: "go-shell-command rule"
description: "Shell command interpreter executed via os/exec"
---

# `go-shell-command`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `command_injection`

**Description**: Shell command interpreter executed via os/exec

**Recommendation**: Execute binary commands directly with argument arrays and sanitize untrusted input

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
cmd := exec.Command("sh", "-c", "ls " + userInput)

// ✅ Do (Recommended)
cmd := exec.Command("ls", "--", validatedPath)
```

```ts [TypeScript / Node.js]
// ❌ Don't (Unsafe)
child_process.exec("ls " + userInput);

// ✅ Do (Recommended)
child_process.execFile("ls", ["--", validatedPath]);
```

```python [Python]
# ❌ Don't (Unsafe)
subprocess.Popen("ls " + user_input, shell=True)

# ✅ Do (Recommended)
subprocess.Popen(["ls", "--", validated_path], shell=False)
```

:::

---

[← sensitive-json-field](/reference/rules/sensitive-json-field) · [Rule Catalog](/reference/rule-catalog) · [go-weak-cryptographic-hash →](/reference/rules/go-weak-cryptographic-hash)
