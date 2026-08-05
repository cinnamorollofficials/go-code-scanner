# ADR 0001: Documentation Site Architecture

- **Status**: Accepted
- **Date**: 2026-08-05
- **Authors**: Security Review Team
- **Deciders**: Core Maintainers

## Context and Problem Statement

`security-review` (Go Code Scanner) is a policy-driven, offline-first security analysis tool written in Go. As scanner capabilities, detection rules, configuration options, and integration surfaces expand, relying solely on `README.md` and repository-level Markdown files presents maintainability and discoverability challenges.

We need a dedicated, fast, accessible public documentation website to present:
1. Getting-started guides and copyable quick-start paths.
2. Conceptual guides (scan modes, finding lifecycle, baseline management, suppressions).
3. Authoritative, generated reference materials (CLI flags/commands, configuration schemas, rule catalog, scanner/adapter matrix).
4. Developer integration guides and interactive configuration tools.

The documentation system must complement `README.md` (which remains lightweight and GitHub-focused) without fragmenting sources of truth or polluting the core Go codebase with Node.js dependencies.

## Decision Drivers

- **Zero Runtime Product Overhead**: The core `security-review` binary and CLI toolchain must remain 100% Go with zero Node.js/npm dependencies.
- **Single Source of Truth**: Go struct definitions, CLI flag declarations, rule registries, and adapter catalogs in Go source code are authoritative. Reference documentation must be generated from Go code, not hand-copied.
- **Performance and Accessibility**: Exceptional performance (Lighthouse > 95), minimal JS payload, accessible ARIA landmarks, contrast compliance, keyboard navigation, and responsive design.
- **Developer Experience**: Fast local preview, standard Markdown/MDX authoring, deterministic builds, and link validation in CI.
- **Secure & Autonomous CI/CD**: Read-only PR validation, pinned dependency lockfiles, action commit SHA pinning, and automated deployment restricted to the default branch via GitHub Pages.

## Considered Alternatives

1. **Astro Starlight**: Excellent static site generator with zero JS defaults, but uses Astro components rather than Vue 3 components for client-side interactivity (such as our planned configuration builder).
2. **Docusaurus (React)**: Robust documentation ecosystem, but introduces heavy JavaScript bundle defaults, React runtime overhead, and complex theme customization for pure static documentation.
3. **Hugo (Go-based)**: Extremely fast build times native to the Go ecosystem, but lacks out-of-the-box structured documentation features (sidebar navigation, built-in search, accessible theme) without extensive custom theme maintenance.
4. **MkDocs (Material for MkDocs)**: Strong documentation framework, but requires a Python environment setup for repository contributors.

## Decision: VitePress

We choose **VitePress** (powered by Vite and Vue 3) as the static documentation site generator, located isolated within the `website/` directory of the repository.

### Architectural Rules & Conventions

1. **Toolchain Isolation**:
   - Node.js usage and `package.json` / `package-lock.json` are strictly contained within `website/`.
   - The Go build and CLI binary remain completely independent of Node.js.

2. **Content Ownership (Markdown / Vue)**:
   - Hand-authored documentation (guides, tutorials, architecture overviews) lives under `website/docs/`.
   - Content uses standard Markdown (`.md`) with Vue components (`.vue`) when custom interactive components (e.g. badges, cards, configuration generator UI) are required.

3. **Sources of Truth & Reference Generation**:
   - Go definitions in `security-review` (in `cmd/`, `config/`, `rules/`, `scanner/`) are the authoritative sources of truth.
   - Dedicated Go generators (`go generate ./...` or `go run ./cmd/...`) extract metadata and output deterministic Markdown files into `website/docs/reference/`.
   - CI enforces clean working tree checks (`git diff --exit-code`) to prevent generated reference drift.

4. **Deployment Target & Hosting**:
   - Published via **GitHub Pages** using GitHub Actions (`.github/workflows/docs.yml`).
   - PR builds perform strict linting, type checks, link validation, and VitePress site builds (`npm run docs:build`) without deployment credentials or publish permissions.

5. **URL Policy & Base Path**:
   - **Initial Deployment**: Configured with a project path subpath (`https://<owner>.github.io/go-code-scanner/`) using `base: '/go-code-scanner/'` in `website/.vitepress/config.mts`.
   - **Custom Domain Readiness**: Configuration allows seamless transition to a custom domain (e.g. `https://security-review.dev`) by updating `site` and `base` parameters without breaking asset paths or internal links.

6. **Versioning Strategy**:
   - The documentation site reflects the **latest stable release** of `security-review`.
   - Release notes, changelog links, and version badges direct users to historical release details. Single-source content is maintained to avoid multi-version maintenance overhead unless breaking public contract changes dictate explicit versioned docs.

## Content & Data Flow Diagram

```
+-------------------------------------------------------------------+
|                        Go Codebase (Source of Truth)             |
|  - CLI Commands & Flags (cmd/)                                    |
|  - Config Structs & Defaults (config/)                            |
|  - Registered Rules & Metadata (rules/)                           |
|  - Scanner Adapters & Capabilities (scanner/)                    |
+-------------------------------------------------------------------+
                                  |
                                  | `go generate ./...`
                                  v
+-------------------------------------------------------------------+
|                   Generated Reference Files                       |
|  - website/docs/reference/cli.md                              |
|  - website/docs/reference/configuration.md                    |
|  - website/docs/reference/rules.md                            |
|  - website/docs/reference/scanners.md                         |
+-------------------------------------------------------------------+
                                  |
                                  | Combined with Hand-Written Docs
                                  v
+-------------------------------------------------------------------+
|                     VitePress Toolchain                           |
|  - Hand-authored Guides (website/docs/...)                        |
|  - Accessible Theme & Built-in Search                             |
|  - Build & Asset Processing (`npm run docs:build`)                |
+-------------------------------------------------------------------+
                                  |
                                  | GitHub Actions (`docs.yml`)
                                  v
+-------------------------------------------------------------------+
|                 GitHub Pages Static Deployment                    |
|  - Target: https://<owner>.github.io/go-code-scanner/             |
+-------------------------------------------------------------------+
```

## Consequences

### Positive
- Authoritative metadata in Go prevents documentation rot.
- Zero impact on Go binary size, build times, or runtime dependencies.
- VitePress delivers fast Vite HMR, Vue 3 component integration, lightweight static output, and built-in local search out of the box.
- Pre-configured asset base path ensures zero broken links from initial deployment.

### Negative / Trade-offs
- Node.js lockfiles and toolchain must be maintained within `website/`.
- Maintainers modifying CLI flags or configuration schemas must run `go generate` or rely on CI drift checks to update generated references.
