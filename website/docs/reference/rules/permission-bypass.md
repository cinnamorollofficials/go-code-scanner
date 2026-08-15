---
title: "permission-bypass rule"
description: "For developers remediating permission-bypass: Hardcoded permission bypass found in application logic"
---

# `permission-bypass` rule

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `security_misconfiguration`

**Description**: Hardcoded permission bypass found in application logic

**Recommendation**: Remove permission bypass conditions and enforce strict authorization checks


The examples below are illustrative and focus on the pattern relevant to this rule.

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func CheckPermission(user User) bool {
    if user.Role == "admin" || bypassPermission {
        return true
    }
    return false
}

// Safer example
func CheckPermission(ctx context.Context, user User, resource string) bool {
    return authzService.CanAccess(ctx, user.ID, resource)
}
```

```ts [TypeScript / JavaScript]
// Unsafe example
function checkPermission(user: User): boolean {
    if (user.role === 'admin' || process.env.BYPASS_PERMISSIONS === 'true') {
        return true;
    }
    return false;
}

// Safer example
async function checkPermission(user: User, resource: string): Promise<boolean> {
    return await authzService.canAccess(user.id, resource);
}
```

```python [Python]
# Unsafe example
def check_permission(user):
    if user.role == "admin" or bypass_permission:
        return True
    return False

# Safer example
def check_permission(user, resource):
    return authz_service.can_access(user.id, resource)
```

:::

---

[← browser-token-storage](/reference/rules/browser-token-storage) · [Rule Catalog](/reference/rule-catalog) · [weak-secret →](/reference/rules/weak-secret)
