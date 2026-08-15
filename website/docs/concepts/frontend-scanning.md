---
title: Frontend Scanning
description: "For full-stack developers: understand frontend discovery, AST analyzers, framework checks, and client-server boundaries."
---

# Frontend Scanning

In modern full-stack web applications, security boundaries span both backend APIs and client-side frontends. `security-review` includes native AST and syntax analysis for modern frontend frameworks (React, Vue, Svelte, Next.js, and Nuxt).

---

## Supported Ecosystems and Frameworks

| Framework or ecosystem | Recognized file types | Key security checks |
| :--- | :--- | :--- |
| **React / Next.js** | `.jsx`, `.tsx`, `.js`, `.ts` | `dangerouslySetInnerHTML`, Client Secret Leaks, Server Actions |
| **Vue / Nuxt** | `.vue`, `.js`, `.ts` | `v-html` Injection, SSR State Pollution, Window Object Leaks |
| **Svelte** | `.svelte`, `.js`, `.ts` | Directives (`@html`), Unsanitized Event Handlers |

---

## Key Frontend Vulnerabilities Detected

### 1. Client-Side Cross-Site Scripting (DOM XSS)
Detects raw HTML interpolation directives and dynamic DOM sinks without proper sanitization:
- **React**: `<span v-pre><code>dangerouslySetInnerHTML={&#123; __html: userInput &#125;}</code></span>`
- **Vue**: `<span v-pre><code>&lt;div v-html="userInput" /&gt;</code></span>`
- **Svelte**: `<span v-pre><code>{&#64;html userInput}</code></span>`

### 2. Client-Side Secret Exposure
Flags high-entropy strings, API tokens, and private keys declared in client-side bundles or `public/` directory assets.

### 3. Circular Import and Boundary Violations
Identifies circular component dependencies and prohibited client-to-server module imports.

---

## Configuration

Frontend scanning can be enabled and tailored in `security-review.json`:

```json
{
  "frontend": {
    "enabled": true,
    "frameworks": ["react", "nextjs", "vue"],
    "client_roots": ["src/client", "pages", "components"],
    "server_roots": ["src/server", "api"]
  }
}
```

To run a frontend-only scan:

```sh
security-review scan --scope client
```
