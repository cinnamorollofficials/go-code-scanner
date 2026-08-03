# Frontend Client Scanning and Threat-Boundary Contract

This document defines the contract, threat boundaries, classification model, severity principles, sanitizer assumptions, and explicit non-goals for browser-client scanning in Go Code Scanner.

It extends the core Go Code Scanner roadmap without altering the six finding domains (`Quality`, `Reliability`, `Hardening`, `Security`, `Supply chain`, `Governance`) or weakening staged-content, redaction, offline, and compatibility guarantees.

---

## 1. Code Scope Classification

Go Code Scanner categorizes source files into three distinct scopes:

### Client Code
- **Definition**: JavaScript, TypeScript, HTML, Vue, and Svelte code compiled into or directly executed within browser environments (DOM, Web Workers, Service Workers).
- **Includes**:
  - Vanilla JS/TS (`.js`, `.jsx`, `.ts`, `.tsx`, `.mjs`, `.cjs`, `.mts`, `.cts`) executed in browser contexts.
  - React/Next.js Client Components (e.g. modules containing `"use client"` directives or imported by client trees).
  - Vue/Nuxt Single-File Components (`.vue`) and client plugins.
  - Svelte/SvelteKit components (`.svelte`) and client scripts (`+page.ts`, `+layout.ts`).
  - HTML entry points (`.html`).
- **Targeting**: Primary subject of frontend security, hardening, quality, and privacy scanning.

### Server Code
- **Definition**: Code executed exclusively on the server (Node.js runtime, Edge runtime, serverless functions).
- **Includes**:
  - Node.js backend services, Express/Koa handlers, database access modules.
  - Next.js Server Components, API routes (`app/api/**`, `pages/api/**`), Server Actions.
  - Nuxt server routes (`server/api/**`, `server/routes/**`).
  - SvelteKit server endpoints and page loaders (`+page.server.ts`, `+layout.server.ts`, `+server.ts`).
- **Targeting**: Excluded from client-side rule evaluations unless explicit client/server boundary rules inspect them (e.g. detecting server module imports in client code).

### Shared Code
- **Definition**: Utilities, types, constants, and validation schemas shared between client and server environments.
- **Includes**:
  - Common validators (e.g., Zod schemas), helper functions, interface definitions.
- **Targeting**: Evaluated under client rules if imported or transitively reachable by client modules.

---

## 2. Supported Ecosystems

The scanner supports the following browser-client ecosystems out-of-the-box:
1. **Vanilla JavaScript & TypeScript**: Standard ECMAScript modules, CommonJS client scripts, HTML files.
2. **React & Next.js**: Single-page React applications and Next.js App/Pages router client modules.
3. **Vue & Nuxt**: Vue 2/3 Single-File Components (`.vue`) and Nuxt 3 client modules.
4. **Svelte & SvelteKit**: Svelte 3/4/5 components (`.svelte`) and SvelteKit client routes.

---

## 3. Trust Boundaries and Threat Model

### Untrusted Inputs
In a browser client, the following sources are untrusted input vectors:
- **DOM / User Input**: `location.search`, `location.hash`, `window.name`, form inputs, `document.referrer`.
- **Browser Storage**: `localStorage`, `sessionStorage`, `IndexedDB`, cookies (if populated by untrusted responses).
- **Inter-Window / Worker Messages**: `window.addEventListener('message', ...)` payload data.
- **External Web Content**: Unauthenticated API payloads, third-party iframe responses, WebSocket messages.

### Critical Sinks
Destinations where untrusted input must not pass without proper validation or sanitization:
- **DOM Injection**: `innerHTML`, `outerHTML`, `insertAdjacentHTML`, `document.write`.
- **Dynamic Code Execution**: `eval(...)`, `new Function(...)`, `setTimeout(string, ...)`, `setInterval(string, ...)`.
- **Navigation / URL Assignment**: `location.href = ...`, `window.open(...)`, `javascript:` protocol links.
- **Client Storage / Telemetry**: Storing sensitive credentials/tokens in browser storage, or sending PII to analytics/logging services.

---

## 4. Severity Principles

Findings follow Go Code Scanner's standardized severity levels:

| Severity | Definition | Example Findings |
| --- | --- | --- |
| **Critical** | High-impact direct vulnerability allowing remote code execution or immediate account takeover. | Direct `eval` of unsanitized user input; exposed private key embedded in frontend source. |
| **High** | Direct client-side injection or dangerous boundary bypass. | DOM XSS via `innerHTML`; wildcard origin in sensitive `postMessage` listener; secret token in `NEXT_PUBLIC_` / `VITE_` env vars. |
| **Medium** | Insecure storage, unsafe transport, or risky browser API usage. | Storing credentials/tokens in `localStorage`; `javascript:` navigation URL; missing Subresource Integrity (SRI) on cross-origin scripts. |
| **Low** | Privacy leak, verbose logging, or mild hardening issue. | PII (email/phone) logged in client console; target `_blank` missing `rel="noopener"`. |
| **Info** | Informational compliance or structural recommendation. | Non-standard environment variable naming; minor framework deprecation. |

---

## 5. Sanitizer and Safe Pattern Assumptions

Native frontend rules account for sanitizers and safe static patterns to prevent false positives:
- **Recognized Sanitizers**: Standard HTML sanitization libraries (e.g. `DOMPurify.sanitize(...)`, `sanitizeHtml(...)`, Trusted Types wrappers) applied to dynamic arguments are recognized as safe for DOM injection sinks.
- **Static Literals**: String literals, compile-time constants, and boolean/numeric expressions reaching sinks are considered safe.
- **Configured Sanitizers**: Custom sanitizer function names defined in configuration (`frontend.recognize_sanitizers`) are treated as valid sanitizing wrappers.

---

## 6. Deferred Data-Flow Behavior

To ensure deterministic, offline, and high-speed execution:
- **Lexical / AST Scope**: Native rules perform intra-file lexical parsing and local AST pattern matching.
- **Deferred Inter-procedural Taint**: Deep multi-file dynamic taint tracking across complex asynchronous function boundaries is explicitly deferred to external analyzers (e.g. Semgrep AST rules). Native rules express confidence based on visible local sinks and sources.

---

## 7. Representative Safe and Unsafe Examples

### DOM Injection (`innerHTML`)
- **UNSAFE**:
  ```javascript
  const params = new URLSearchParams(window.location.search);
  document.getElementById('app').innerHTML = params.get('query');
  ```
- **SAFE**:
  ```javascript
  const params = new URLSearchParams(window.location.search);
  const cleanHTML = DOMPurify.sanitize(params.get('query'));
  document.getElementById('app').innerHTML = cleanHTML;
  ```

### Dynamic Code Execution (`eval`)
- **UNSAFE**:
  ```javascript
  const code = getQueryParams('code');
  eval(code);
  ```
- **SAFE**:
  ```javascript
  const command = getQueryParams('action');
  if (command === 'play') handlePlay();
  ```

### Exposure of Client Secrets (`NEXT_PUBLIC_`)
- **UNSAFE**:
  ```env
  NEXT_PUBLIC_STRIPE_SECRET_KEY=sk_live_51Hz...
  ```
- **SAFE**:
  ```env
  NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_51Hz...
  ```

### Client / Server Boundary Crossing
- **UNSAFE** (in client component):
  ```typescript
  "use client";
  import { db } from "@/lib/db.server"; // Importing server-only module into client component
  ```
- **SAFE** (in client component):
  ```typescript
  "use client";
  import { formatName } from "@/lib/utils"; // Importing safe shared utility
  ```

### Privacy Telemetry Leaks
- **UNSAFE**:
  ```javascript
  console.log("User login failed:", { email: user.email, password: user.password });
  ```
- **SAFE**:
  ```javascript
  console.log("User login failed:", { userId: user.id });
  ```
