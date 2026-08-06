# Documentation Author Guide

This guide outlines standards and conventions for contributing content to the **Go Code Scanner** documentation site.

## Frontmatter & Headings

Every page must start with YAML frontmatter specifying title and description:

```yaml
---
title: Page Title
description: Concise description of the page purpose.
---
```

Use a single `<h1>` (`# Page Title`) per document. Subsections should follow a strict hierarchy (`## Heading 2`, `### Heading 3`).

## Code Examples

- Always declare the code block language (e.g. `sh`, `json`, `go`, `yaml`).
- Never include real secrets, tokens, internal URLs, or local absolute paths (`C:\Users\<user>\...` or `/home/...`).
- Use synthetic placeholders (`your-api-key`, `example.org`, `path/to/file.go`).

## Admonitions (Callouts)

VitePress supports built-in container callouts:

::: tip
Helpful suggestions, efficiency tips, or best practices.
:::

::: warning
Potential pitfalls, breaking changes, or caution notices.
:::

::: danger
Critical warnings regarding security risks or destructive operations.
:::

## Terminology & Style

- Use **`security-review`** when referring to the CLI executable binary.
- Use **Go Code Scanner** when referring to the product or suite.
- Use relative internal links without file extensions (e.g. `[CLI Reference](/reference/cli)`).
