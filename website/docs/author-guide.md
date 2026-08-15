---
title: Documentation Author Guide
description: Standards, conventions, and style guidelines for contributing content to Go Code Scanner documentation.
---

# Documentation Author Guide

## Content Types and Ownership

Choose one content type before writing:

| Type | Reader question | Content contract |
| :--- | :--- | :--- |
| **Getting Started** | “How do I reach my first successful result?” | Use the shortest supported path and defer optional configuration. |
| **Guide** | “How do I complete this task?” | Start from prerequisites, provide ordered actions, and end with a verifiable outcome. |
| **Concept** | “How does this behavior fit together?” | Explain the mental model without duplicating a full operational procedure. |
| **Reference** | “What is the exact contract?” | List accepted commands, fields, defaults, values, outputs, and failure behavior. |

Each command sequence has one canonical owner. Other pages should summarize the
idea and link to that owner instead of copying a complete procedure. This keeps
examples from drifting when CLI behavior changes.

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
- Never include real secrets, tokens, internal URLs, or workstation-specific
  home-directory paths.
- Use synthetic placeholders (`your-api-key`, `example.org`, `path/to/file.go`).
- Label output or snippets as complete, abbreviated, illustrative, or captured
  from a fixture when that distinction affects whether they can be copied.

## Local Verification

Run the complete documentation gate from `website/` before opening a change:

```sh
npm run docs:verify
```

The same gate runs as part of the production documentation build. It regenerates
source-backed references, verifies their derived pages, checks authored content,
builds the site, inspects rendered links and anchors, and runs browser smoke tests.

| Check | What it protects | Typical failure and response |
| :--- | :--- | :--- |
| `docs:check-source-generated` | Configuration fields, the raw rule reference, and scanner compatibility stay synchronized with Go sources. | `Generated source-backed references were stale` means the command refreshed one or more files. Review those changes and run the check again. |
| `docs:check-generated` | Rule Catalog records and all focused rule pages match the raw generated rule reference. | `Generated Rule Catalog output was stale` lists derived files that need to be reviewed and committed. |
| `docs:check-content` | Frontmatter, unique titles, useful descriptions, one H1, labelled code fences, and portable paths. | The message starts with the affected page and states the missing or duplicate field, heading count, code-fence language, or absolute path. |
| `docs:check-cli` | Copyable `security-review` commands use current subcommands and flags. | `unsupported command shape` or `is not valid` identifies the command that must be corrected against the CLI reference. |
| `docs:check-presets` | Every Config Builder preset passes both browser-side checks and the real CLI validator. | The preset name is followed by `browser validation failed` or `CLI validation failed`; correct the shared preset data rather than the generated output. |
| `docs:check-configs` | Copyable JSON configuration examples pass the real CLI validator. | The page and JSON block number are followed by the validator error. Correct the documented example. |
| `docs:check-deployment` | The internal deployment guide stays portable and agrees with the Docker, Compose, Nginx, and base-path contract shipped under `website/`. | The message names the missing file, outdated command, port mapping, route, or unsupported deployment claim that must be reconciled. |
| `docs:check-links` | Rendered internal destinations and anchors exist and remain inside the documentation base path. | `missing target`, `missing anchor`, or `link escapes documentation base` names the rendered source and broken link. |
| `docs:test-browser` | Critical pages, filters, responsive layouts, keyboard flows, status regions, and WCAG A/AA smoke checks. | Playwright names the failed journey; accessibility failures also include the rule ID and affected element. |

The browser tests use the production output, so run `npm run docs:build` rather
than invoking `docs:test-browser` against an old build. The standalone scripts
remain useful when narrowing down one failure.

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
