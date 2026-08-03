# Frontend Client Scanning Implementation Tasks

This document is the implementation backlog for first-class browser-client
scanning. It extends the completed Go Code Scanner roadmap without changing the
existing six finding domains or weakening staged-content, redaction, offline,
and compatibility guarantees.

The target scope is browser-executed JavaScript and TypeScript, including
vanilla applications, React and Next.js client components, Vue and Nuxt,
Svelte and SvelteKit, and HTML entry points. Node.js services and framework
server modules remain outside the client scope unless a boundary rule needs to
inspect them.

## Working agreement

- Complete tasks in order unless a task explicitly names a different dependency.
- Create exactly one atomic commit for each numbered task using the suggested
  subject. Do not combine adjacent tasks into one commit.
- Every implementation commit includes its direct positive, negative,
  invalid-input, cancellation, and boundary tests as applicable.
- Keep the default fast profile deterministic and offline.
- Never report or persist source snippets, secrets, tokens, PII, raw analyzer
  output, or arbitrary command stderr.
- A scanner may advertise `staged` support only after both isolation directions
  are proven: staged-safe/working-unsafe and staged-unsafe/working-safe.
- Preserve report schema `1.0`, finding domains, exit codes, fingerprints, and
  existing configuration meaning unless a task explicitly documents a reviewed
  migration.
- Use stable rule IDs prefixed with `frontend/` and descriptions that express
  confidence rather than claiming a vulnerability is proven.
- After every task, run its listed focused checks, `go vet ./...`, and
  `git diff --check`. After every five tasks, run the full verification suite.
- Mark a checkbox complete only after its commit exists and the required checks
  pass. Record the commit hash next to the task when reconciling this file.

## Phase 1 — Product contract and configuration

- [x] **1. Define the frontend scanning and threat-boundary contract.** Document
  what qualifies as client, server, and shared code; supported ecosystems;
  trust boundaries; severity principles; sanitizer assumptions; and explicitly
  deferred data-flow behavior. Add representative safe and unsafe examples so
  later rules do not silently broaden the contract. Context: documentation and
  public behavior only; no runtime change. Verification: `git diff --check`.
  Commit: `docs(frontend): define client scanning contract` (8759768).

- [x] **2. Add strict frontend configuration.** Add an optional nested
  `frontend` configuration with `enabled`, `frameworks`, `client_roots`,
  `server_roots`, `shared_roots`, `include_extensions`,
  `recognize_sanitizers`, `detect_import_cycles`, and
  `detect_client_server_boundaries`. Define deterministic defaults, normalize
  safe project-relative paths, reject duplicates and unsupported frameworks,
  update strict-decode tests and `CONFIGURATION.md`, and review the
  compatibility manifest. An omitted block must preserve existing behavior.
  Context: `config`, configuration docs, compatibility tests. Verification:
  `go test ./config ./compatibility`. Depends on task 1. Commit: `feat(config): add frontend scanning policy` (e17d298).

- [x] **3. Expand frontend file discovery safely.** Support `.mjs`, `.cjs`,
  `.mts`, `.cts`, `.html`, `.vue`, and `.svelte` without scanning generated
  output or dependency directories. Preserve NUL-safe Git discovery, symlink
  rejection, sorted results, full/changed/staged semantics, file-count limits,
  and the current handling of manifests excluded from line scanning. Context:
  `config`, `discovery`, and discovery tests. Verification:
  `go test ./config ./discovery`. Depends on task 2. Commit: `feat(discovery): include frontend client sources` (283b847).

- [ ] **4. Implement deterministic client-scope classification.** Detect
  client, server, shared, and unknown files using configured roots, framework
  file conventions, package metadata, and markers such as `"use client"`.
  Explicit roots take precedence over automatic conventions. Classification
  must read through `scanner.Source`, remain bounded, work with staged blobs,
  and return a safe unknown result for malformed inputs. Cover vanilla,
  React/Next.js, Vue/Nuxt, and SvelteKit fixtures. Context: new
  `scanner/frontend` classifier package. Verification:
  `go test ./scanner/frontend`. Depends on tasks 2–3. Commit:
  `feat(frontend): classify browser client sources`.

## Phase 2 — Native scanner foundation

- [ ] **5. Add the bounded frontend scanner lifecycle.** Implement the scanner
  shell, descriptor, deterministic worker orchestration, resource limits,
  cancellation, partial/failure states, and safe diagnostics. Register it only
  when frontend scanning is enabled; do not add blocking rules yet. Verify
  enabled/disabled, required, timeout, panic, cache, and profile behavior.
  Context: `scanner/frontend`, `securityreview.go`, orchestration tests. Checks:
  `go test ./scanner/frontend ./... -run 'Frontend|frontend'`. Depends on task
  4. Commit: `feat(frontend): add native scanner lifecycle`.

- [ ] **6. Add lightweight JavaScript and template lexical context.** Build a
  bounded, dependency-free tokenizer sufficient to distinguish code, comments,
  string literals, template literals, JSX attributes, and embedded script or
  template regions. It is not a complete TypeScript parser. Malformed or
  oversized syntax must produce deterministic partial/skipped behavior, never a
  panic. Add fuzz seeds for escaping, nested templates, and truncated files.
  Context: `scanner/frontend` lexer and tests. Verification:
  `go test ./scanner/frontend` and the focused frontend fuzz smoke. Depends on
  task 5. Commit: `feat(frontend): add lexical source context`.

- [ ] **7. Establish frontend finding metadata and rule registry.** Add an
  internal rule registry that maps stable IDs to domain, category, severity,
  recommendation, documentation, framework, confidence, sink, and source
  metadata. Reuse the public `finding.Finding` contract and existing tags; do
  not add a new domain or expose snippets. Make rules discoverable through
  `--explain` and include registry content in the rule-set hash. Context:
  `scanner/frontend`, rule explanation, cache/rule hashing tests. Verification:
  `go test ./scanner/frontend ./cache ./... -run 'Explain|RuleSetHash'`. Depends
  on tasks 5–6. Commit: `feat(frontend): register client rule metadata`.

## Phase 3 — High-confidence browser rules

- [ ] **8. Detect DOM injection sinks.** Add rules for unsafe dynamic values
  reaching `innerHTML`, `outerHTML`, `insertAdjacentHTML`, `document.write`, and
  equivalent direct DOM sinks. Recognize configured sanitizer calls and Trusted
  Types patterns, ignore comments and harmless literals, and emit one stable
  finding per sink location. Context: native frontend security rules. Checks:
  `go test ./scanner/frontend -run 'DOM|Injection|Sanitizer'`. Depends on tasks
  6–7. Commit: `feat(frontend): detect DOM injection sinks`.

- [ ] **9. Detect dynamic execution and unsafe messaging.** Add rules for
  variable input passed to `eval`, `Function`, string-based timers, wildcard
  `postMessage`, and message handlers that consume sensitive data without a
  recognizable origin check. Avoid flagging constant safe messages and test
  fixtures by default. Context: native frontend security/hardening rules.
  Verification: `go test ./scanner/frontend -run 'Execution|Message'`. Depends
  on tasks 6–7. Commit:
  `feat(frontend): detect unsafe execution and messaging`.

- [ ] **10. Detect client-side secret and token exposure.** Detect credential-
  like values stored in localStorage, sessionStorage, or IndexedDB; secret-like
  names exposed through `NEXT_PUBLIC_`, `VITE_`, or `REACT_APP_`; and embedded
  private keys or high-confidence tokens. Do not classify ordinary public API
  URLs or public identifiers as secrets. Preserve canary redaction through
  terminal, artifacts, and cache. Context: frontend rules and redaction tests.
  Checks: `go test ./scanner/frontend ./reporter ./cache -run 'Secret|Token|Redact'`.
  Depends on tasks 6–7. Commit:
  `feat(frontend): detect client credential exposure`.

- [ ] **11. Detect unsafe navigation and transport usage.** Add rules for
  untrusted values assigned to browser navigation sinks, `javascript:` URLs,
  production HTTP API endpoints, unsafe `window.open` usage, and sensitive
  fields placed in URL query parameters. Do not flag localhost development URLs
  unless policy explicitly enables them. Context: frontend security,
  hardening, and governance rules. Verification:
  `go test ./scanner/frontend -run 'Navigation|Transport|Query'`. Depends on
  tasks 6–7. Commit:
  `feat(frontend): detect unsafe navigation and transport`.

- [ ] **12. Detect privacy leakage in logs and telemetry.** Extend PII checks to
  browser logs, analytics calls, error-reporting context, and telemetry events.
  Report only field names and safe metadata, never literal values. Allow
  reviewed organization-specific PII names through rule configuration rather
  than hardcoded project terms. Context: frontend governance rules and
  configuration tests. Verification:
  `go test ./scanner/frontend ./config -run 'Privacy|Telemetry|PII'`. Depends on
  tasks 2 and 6–7. Commit:
  `feat(frontend): detect client telemetry privacy leaks`.

## Phase 4 — Framework-aware checks

- [ ] **13. Add React and Next.js client rules.** Detect dynamic
  `dangerouslySetInnerHTML`, client components importing Node/server-only
  modules, private environment access from client modules, and unsafe link or
  redirect constructs. Account for Server Components so server-only findings
  are not emitted outside a client dependency path. Context: frontend framework
  rules and fixtures. Verification:
  `go test ./scanner/frontend -run 'React|Next'`. Depends on tasks 4 and 8–11.
  Commit: `feat(frontend): scan React and Next client code`.

- [ ] **14. Add Vue and Nuxt client rules.** Detect unsafe `v-html`, sensitive
  private runtime config used by client code, unsafe dynamic URL bindings, and
  imports that cross Nuxt client/server boundaries. Cover single-file component
  script and template regions without treating generated output as source.
  Context: frontend framework rules and fixtures. Verification:
  `go test ./scanner/frontend -run 'Vue|Nuxt'`. Depends on tasks 4, 6, and 8–11.
  Commit: `feat(frontend): scan Vue and Nuxt client code`.

- [ ] **15. Add Svelte and SvelteKit client rules.** Detect unsafe `{@html}` use,
  imports from private environment modules in client code, and violations
  between `+page`, `+layout`, and their `.server` counterparts. Cover script and
  template regions and recognize sanitized expressions. Context: frontend
  framework rules and fixtures. Verification:
  `go test ./scanner/frontend -run 'Svelte'`. Depends on tasks 4, 6, and 8–11.
  Commit: `feat(frontend): scan SvelteKit client code`.

## Phase 5 — External frontend analyzers

- [ ] **16. Add an ESLint JSON adapter.** Add an adapter preset that runs an
  argument-array command, consumes bounded JSON output, maps rule IDs and
  severities deterministically, normalizes paths, and never includes source or
  message excerpts. Support root and staged workspaces with captured fixtures,
  malformed-output tests, cancellation, missing executable behavior, and fuzz
  coverage. Context: `scanner/adapters` and command-scanner integration tests.
  Verification: `go test ./scanner/adapters ./scanner/command`. Depends on task
  5. Commit: `feat(adapters): add ESLint scanner`.

- [ ] **17. Add a TypeScript type-check adapter.** Add a preset for
  `tsc --noEmit --pretty false` with bounded deterministic parsing and safe
  generic diagnostics. Define exact finding exit codes and skip cleanly when no
  applicable TypeScript project exists. Prove staged workspace isolation.
  Context: adapters and command scanner tests. Verification:
  `go test ./scanner/adapters ./scanner/command`. Depends on task 5. Commit:
  `feat(adapters): add TypeScript checker`.

- [ ] **18. Add an optional Biome adapter.** Support Biome lint/format JSON
  output as a quality and reliability analyzer without making it a required
  runtime dependency. Preserve user-provided arguments and environment
  allowlisting. Add fixtures for at least two supported output shapes or reject
  unknown shapes safely. Context: adapters. Verification:
  `go test ./scanner/adapters ./scanner/command`. Depends on task 5. Commit:
  `feat(adapters): add Biome scanner`.

- [ ] **19. Bundle an offline Semgrep frontend ruleset.** Add reviewed local
  Semgrep rules for analysis that requires AST or source-to-sink matching and
  configure the existing adapter to use local files without downloading a
  registry pack. Pin public rule IDs, map findings to the six existing domains,
  and test multiple supported Semgrep output versions. Context: bundled rules,
  Semgrep adapter, parser fixtures, and packaging tests. Verification:
  `go test ./scanner/adapters ./release ./scripts -run 'Semgrep|Archive|Package'`.
  Depends on tasks 8–15. Commit:
  `feat(frontend): bundle offline Semgrep rules`.

## Phase 6 — Frontend architecture analysis

- [ ] **20. Extract JavaScript and TypeScript import edges.** Parse static
  imports, exports, CommonJS requires, and literal dynamic imports using bounded
  lexical context. Preserve file and line location, ignore external packages
  when building the local graph, and sort edges deterministically. Context:
  frontend import graph. Verification:
  `go test ./scanner/frontend -run 'Import|Graph'`. Depends on task 6. Commit:
  `feat(frontend): extract client import graph`.

- [ ] **21. Resolve local aliases and workspace packages.** Resolve relative
  modules, supported extensions, `index` files, `tsconfig`/`jsconfig` base URL
  and path aliases, and local npm workspace package mappings. Reject unsafe
  paths and ambiguous aliases rather than guessing. Context: frontend resolver
  and manifest parsing. Verification:
  `go test ./scanner/frontend -run 'Resolve|Alias|Workspace'`. Depends on task
  20. Commit: `feat(frontend): resolve TypeScript project imports`.

- [ ] **22. Enforce client/server dependency boundaries.** Report client paths
  that transitively reach configured server roots, Node built-ins, private
  environment modules, or known server-only framework modules. Include a safe
  dependency path in metadata without source text. Allow reviewed shared
  modules and avoid treating UI authorization as server-side authorization.
  Context: frontend governance scanner. Verification:
  `go test ./scanner/frontend -run 'Boundary|ServerOnly'`. Depends on tasks 4
  and 20–21. Commit:
  `feat(frontend): enforce client server boundaries`.

- [ ] **23. Detect frontend import cycles.** Add deterministic cycle detection
  for local client/shared modules, deduplicate equivalent cycles, respect the
  configuration toggle, and preserve full/changed/staged behavior using the
  repository inventory. Context: frontend architecture scanner. Verification:
  `go test ./scanner/frontend -run 'Cycle'`. Depends on tasks 20–21. Commit:
  `feat(frontend): detect client import cycles`.

## Phase 7 — JavaScript supply-chain policy

- [ ] **24. Support all common JavaScript lockfiles.** Extend manifest-without-
  lockfile checks to npm, pnpm, Yarn, and Bun; account for monorepo roots and
  nested projects; and avoid requiring a lockfile in every workspace package.
  Preserve bounded parsing and staged repository inventory semantics. Context:
  pattern file policy and fixtures. Verification:
  `go test ./scanner/pattern -run 'Manifest|Lockfile'`. Depends on task 3.
  Commit: `feat(supply-chain): support JavaScript lockfiles`.

- [ ] **25. Harden JavaScript dependency policies.** Detect wildcard/latest
  versions, mutable Git references, unpinned URL dependencies, and suspicious
  lifecycle scripts with reviewed severity and confidence. Support dependencies,
  devDependencies, optionalDependencies, peerDependencies, and workspace
  protocols without misclassifying valid local workspace references. Context:
  pattern file policy. Verification:
  `go test ./scanner/pattern -run 'Package|Dependency|Lifecycle'`. Depends on
  task 24. Commit:
  `feat(supply-chain): audit JavaScript dependency policy`.

- [ ] **26. Check remote browser resources.** Detect unversioned remote scripts,
  missing Subresource Integrity for cross-origin scripts and styles, dynamic
  remote imports, and insecure resource URLs. Avoid requiring SRI where the
  browser contract does not support it. Context: HTML/template frontend rules.
  Verification: `go test ./scanner/frontend -run 'Remote|Integrity|Resource'`.
  Depends on tasks 6 and 14–15. Commit:
  `feat(frontend): check remote resource integrity`.

## Phase 8 — CLI, profiles, and compatibility

- [ ] **27. Add client scope selection to the CLI.** Add
  `--scope client|server|all` with `all` preserving current behavior. Define
  interactions with config, changed/staged modes, hooks, and invalid arguments;
  update help and CLI tests without changing existing exit codes. Context:
  `cmd/security-review`, config override handling, help tests. Verification:
  `go test ./cmd/security-review ./config -run 'Scope|Help|Arguments'`. Depends
  on tasks 2 and 4–5. Commit: `feat(cli): add client scan scope`.

- [ ] **28. Add ecosystem-aware frontend profiles.** Update fast, standard, and
  full profile behavior so native frontend checks run offline, while ESLint,
  TypeScript, Biome, Semgrep, and vulnerability tools are selected only when
  configured and applicable. A frontend-only repository must not fail because
  `go.mod` or a Go-only tool is absent. Context: configuration defaults,
  orchestration, hooks, and profile tests. Verification:
  `go test ./config ./hook ./... -run 'Profile|Applicable|Frontend'`. Depends on
  tasks 5 and 16–19. Commit:
  `feat(profiles): select applicable frontend scanners`.

- [ ] **29. Preserve frontend findings across cache and lifecycle features.**
  Verify cache keys include frontend config, framework detection, sanitizer
  policy, and native rule versions. Prove baseline relocation, new-only policy,
  suppression matching, fingerprint stability, and cached/uncached equivalence
  for representative frontend findings. Context: cache, baseline, suppression,
  lifecycle, and integration tests. Verification:
  `go test ./cache ./baseline ./suppression ./... -run 'Frontend|Fingerprint|Equivalence'`.
  Depends on tasks 7–15 and 22–23. Commit:
  `test(frontend): preserve finding lifecycle contracts`.

- [ ] **30. Validate public report and compatibility contracts.** Add frontend
  findings to terminal, JSON, SARIF, and JUnit contract fixtures; confirm safe
  metadata, stable rule IDs, no snippets, and unchanged schema versions. Update
  the compatibility manifest only for intentional additive capabilities and
  document any migration requirement before changing a public contract.
  Context: reporters and compatibility tests. Verification:
  `go test ./reporter ./compatibility ./... -run 'Frontend|Golden|Contract'`.
  Depends on tasks 7–29. Commit:
  `test(frontend): validate public report contracts`.

## Phase 9 — Acceptance, performance, and handoff

- [ ] **31. Add end-to-end staged isolation coverage.** Build a real binary and
  prove both staged isolation directions for the native frontend scanner,
  import graph, ESLint, TypeScript, Biome, and local Semgrep when each claims
  staged support. Verify temporary workspace cleanup after success, failure,
  timeout, and cancellation. Context: staged isolation and release-binary E2E
  tests. Verification: targeted staged E2E tests and `go test ./...`. Depends on
  tasks 5, 16–19, and 20–23. Commit:
  `test(frontend): verify staged scanner isolation`.

- [ ] **32. Establish a cross-framework false-positive corpus.** Add sanitized
  fixtures covering real-world safe and unsafe patterns in vanilla, React,
  Next.js, Vue, Nuxt, Svelte, and SvelteKit. Record expected findings by stable
  rule ID and enforce a reviewed noise budget. Include comments, tests, mocks,
  sanitizers, generated files, malformed syntax, and monorepo layouts. Context:
  test corpus and acceptance harness. Verification: corpus test command plus
  `git diff --check`. Depends on tasks 8–26. Commit:
  `test(frontend): add false positive corpus`.

- [ ] **33. Enforce frontend resource and performance budgets.** Benchmark
  classification, lexical scanning, native rules, import extraction, graph
  traversal, and cached scans. Add non-flaky budgets for file size, line size,
  file count, graph size, traversal depth, command output, and fast pre-commit
  runtime. Context: benchmarks, resource-boundary tests, performance script.
  Verification: frontend benchmarks, boundary tests, and
  `scripts/performance-budget.sh`. Depends on tasks 5–26 and 32. Commit:
  `test(frontend): enforce scanning budgets`.

- [ ] **34. Add frontend fuzz and redaction gates.** Fuzz configuration,
  classifier, lexer, manifest parsing, import resolution, and adapter parsers.
  Add secret/PII canaries covering terminal, JSON, SARIF, JUnit, cache, baseline,
  scanner status, and operational errors. Integrate bounded frontend fuzz smoke
  into the canonical verification script. Context: fuzz targets, redaction
  integration tests, verification scripts. Verification:
  `scripts/fuzz-smoke.sh` and focused redaction tests. Depends on tasks 2–30.
  Commit: `test(frontend): fuzz parsers and verify redaction`.

- [ ] **35. Complete frontend user and contributor documentation.** Document
  supported file types and frameworks, scope classification, configuration,
  profiles, optional analyzer installation, offline behavior, rule catalog,
  severity, suppressions, staged limitations, troubleshooting, and the process
  for adding a frontend rule or framework. Update README, configuration,
  contribution, security, and migration documentation as applicable. Context:
  documentation only. Verification: documentation link checks where available
  and `git diff --check`. Depends on tasks 1–34. Commit:
  `docs(frontend): document client scanning`.

- [ ] **36. Certify the frontend scanning release candidate.** Run unit,
  integration, end-to-end, race, vet, fuzz smoke, vulnerability, golden,
  self-scan, cached/uncached equivalence, staged isolation, performance, and
  release reproducibility checks. Fix discovered defects in separate atomic
  commits before marking this task complete; this task's own commit only records
  passing acceptance evidence and reconciles the roadmap. Context: roadmap and
  release evidence. Verification:
  `GOCACHE=/tmp/go-code-scanner-gocache ./scripts/release-candidate.sh`. Depends
  on tasks 1–35. Commit:
  `test(frontend): certify client scanning release`.

## Milestone completion criteria

The frontend scanning milestone is complete when all 36 tasks are checked, each
task maps to one focused commit, and the release-candidate gate passes. The
result must scan applicable browser client code in full, changed, and staged
modes; remain useful offline; integrate with profiles, hooks, baseline,
suppression, cache, policy, and reports; and preserve the project's redaction,
resource-boundary, determinism, and compatibility guarantees.
