---
title: "go-shell-command rule"
description: "For developers remediating go-shell-command: Shell command interpreter executed via os/exec"
---

# `go-shell-command` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `command_injection`

**Description**: Shell command interpreter executed via os/exec

**Recommendation**: Execute binary commands directly with argument arrays and sanitize untrusted input


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
cmd := exec.Command("sh", "-c", "ls " + userInput)

// Safer example
cmd := exec.Command("ls", "--", validatedPath)
```

```ts [TypeScript / Node.js]
// Unsafe example
child_process.exec("ls " + userInput);

// Safer example
child_process.execFile("ls", ["--", validatedPath]);
```

```python [Python]
# Unsafe example
subprocess.Popen("ls " + user_input, shell=True)

# Safer example
subprocess.Popen(["ls", "--", validated_path], shell=False)
```

:::

---

[← sensitive-json-field](/reference/rules/sensitive-json-field) · [Rule Catalog](/reference/rule-catalog) · [go-weak-cryptographic-hash →](/reference/rules/go-weak-cryptographic-hash)
