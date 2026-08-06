---
title: Frontend & Client Scanning Guide
description: Client/server file classification, framework detection (React, Vue, Svelte), threat boundaries, and safe code patterns.
---

# Frontend & Client Scanning Guide

Learn how `security-review` detects frontend frameworks, classifies client versus server code boundaries, and enforces security policies across JavaScript and TypeScript ecosystems.

## Supported Ecosystems & Framework Detection

`security-review` automatically inspects `package.json` and project files to identify frontend frameworks:

- **React & Next.js**: App Router, Pages Router, Server Components (`'use server'`), Client Components (`'use client'`).
- **Vue & Nuxt**: Single File Components (`.vue`), Nuxt server routes (`server/api/`), client composables.
- **Svelte & SvelteKit**: `.svelte` components, SvelteKit server endpoints (`+page.server.ts`).

---

## Threat Boundaries & Native Rules

### 1. Unsanitized HTML & DOM Injection

Flagged when raw unescaped HTML strings are rendered into the client DOM.

#### Unsafe React Example

```tsx
// ❌ UNSAFE: Direct HTML injection without sanitization
function UserComment({ rawHtml }: { rawHtml: string }) {
  return <div dangerouslySetInnerHTML={{ __html: rawHtml }} />;
}
```

#### Safe React Example

```tsx
// ✅ SAFE: Sanitized using DOMPurify before rendering
import DOMPurify from 'dompurify';

function UserComment({ rawHtml }: { rawHtml: string }) {
  const cleanHtml = DOMPurify.sanitize(rawHtml);
  return <div dangerouslySetInnerHTML={{ __html: cleanHtml }} />;
}
```

---

### 2. Client-Side Secret Exposure

Client build tools expose environment variables prefixed with `NEXT_PUBLIC_`, `VITE_`, or `PUBLIC_`. `security-review` flags secret patterns (such as private keys or database passwords) assigned to public client prefixes.

#### Unsafe Next.js Config

```sh
# ❌ UNSAFE: Exposing private database secret to client bundle
NEXT_PUBLIC_DB_PASSWORD=synthetic_secret_password_12345
```

#### Safe Next.js Config

```sh
# ✅ SAFE: Restricted to server runtime only
DATABASE_PASSWORD=synthetic_secret_password_12345
```

---

## Running Client Scans

To run a scan targeting frontend assets and client codebases:

```sh
security-review scan --scope client --profile frontend
```
