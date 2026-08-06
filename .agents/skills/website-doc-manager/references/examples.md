# Approved Documentation Examples & Templates

## Example 1: New Feature Documentation Page

Location: `website/docs/features/new-feature.md`

```markdown
# Feature Title

Brief 1-2 sentence description of what the feature does and why it is useful.

## Overview

Detailed context and architecture details.

## Usage

```bash
security-review scan --root ./my-project --format table
```

## Configuration Options

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `enable` | `boolean` | `true` | Enables or disables the feature |
| `threshold` | `string` | `"high"` | Minimum severity threshold |

## Related Documentation

- [CLI Reference](/reference/cli)
- [Rule Catalog](/reference/rules)
```

---

## Example 2: Adding Item to VitePress Sidebar (`.vitepress/config.mts`)

```typescript
{
  text: 'Features',
  items: [
    { text: 'Scan Execution & Policy', link: '/features/scan-execution-and-policy' },
    { text: 'Developer Workflow', link: '/features/developer-workflow-features' },
    { text: 'Reports & Findings Lifecycle', link: '/features/reports-and-finding-lifecycle' },
    { text: 'New Feature Title', link: '/features/new-feature' }
  ]
}
```
