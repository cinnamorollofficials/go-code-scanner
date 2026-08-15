---
title: "permission-bypass rule"
description: "Hardcoded permission bypass found in application logic"
---

# `permission-bypass`

[← Rule Catalog](/reference/rule-catalog)

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `security_misconfiguration`

**Description**: Hardcoded permission bypass found in application logic

**Recommendation**: Remove permission bypass conditions and enforce strict authorization checks

##### Code Examples (Don't vs Do)

::: code-group

```go [Go]
// ❌ Don't (Unsafe)
func CheckPermission(user User) bool {
    if user.Role == "admin" || bypassPermission {
        return true
    }
    return false
}

// ✅ Do (Recommended)
func CheckPermission(ctx context.Context, user User, resource string) bool {
    return authzService.CanAccess(ctx, user.ID, resource)
}
```

```ts [TypeScript / JavaScript]
// ❌ Don't (Unsafe)
function checkPermission(user: User): boolean {
    if (user.role === 'admin' || process.env.BYPASS_PERMISSIONS === 'true') {
        return true;
    }
    return false;
}

// ✅ Do (Recommended)
async function checkPermission(user: User, resource: string): Promise<boolean> {
    return await authzService.canAccess(user.id, resource);
}
```

```python [Python]
# ❌ Don't (Unsafe)
def check_permission(user):
    if user.role == "admin" or bypass_permission:
        return True
    return False

# ✅ Do (Recommended)
def check_permission(user, resource):
    return authz_service.can_access(user.id, resource)
```

:::

---

[← browser-token-storage](/reference/rules/browser-token-storage) · [Rule Catalog](/reference/rule-catalog) · [weak-secret →](/reference/rules/weak-secret)
