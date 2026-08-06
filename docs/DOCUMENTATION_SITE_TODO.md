# Documentation Site Implementation Roadmap

This roadmap tracks the implementation of a public documentation website for
Go Code Scanner. It is intentionally separate from `TODO.md` and
`REMAINING_TASKS.md`, which describe the scanner and release implementation.

The target is a static VitePress site built from repository-owned content,
with generated CLI, configuration, and rule references. Each numbered task must
be delivered as one atomic commit.

## Working agreement

- Complete tasks in order unless a task explicitly says it can run in parallel.
- Keep one numbered task in one commit; do not combine unrelated cleanup.
- Keep `README.md` useful on GitHub. The website complements it rather than
  replacing it.
- Use English for published documentation and stable identifiers. Indonesian or
  other translations can be added after the English information architecture is
  stable.
- Treat Go definitions as the source of truth for CLI commands, configuration,
  rules, defaults, and validation constraints.
- Do not hand-copy generated reference data into multiple pages.
- Never publish source snippets, secrets, local absolute paths, signing material,
  or scanner output containing sensitive values.
- All example configuration files must pass `security-review config validate`.
- All example commands must use supported public CLI syntax.
- Pin CI actions to full commit SHAs and pin the documentation toolchain through
  the package lockfile.
- Run `git diff --check` after every task. Run the full documentation verification
  after every phase.

## Definition of done for every commit

- [ ] The task's acceptance criteria are met.
- [ ] New or changed behavior has automated coverage where practical.
- [ ] Internal and external links introduced by the task resolve.
- [ ] Code blocks declare the correct language and do not contain local paths.
- [ ] The production documentation build succeeds.
- [ ] `git diff --check` succeeds.

## Phase 1 — Foundation and deployment

- [x] **1. Record the documentation architecture decision.** Add an ADR covering
  VitePress, Markdown/Vue ownership, Node usage limited to documentation,
  generated references, deployment target, URL policy, versioning strategy, and
  rejected alternatives. Define whether the initial site uses a project path or
  custom domain so asset URLs are correct from the first deployment.
  Acceptance: the ADR identifies sources of truth and contains a content/data
  flow diagram. Commit: `docs(adr): define documentation site architecture`.

- [x] **2. Scaffold the VitePress application.** Create `website/` with pinned
  dependencies, lockfile, VitePress configuration (`.vitepress/config.mts`), TypeScript
  configuration, base styles, public assets, and placeholder home page. Add local
  scripts for development, build, preview, and VitePress validation. Do not migrate
  product content yet. Acceptance: a clean install and production build succeed from a
  fresh checkout. Commit: `docs(site): scaffold VitePress application`.

- [x] **3. Establish brand and accessibility foundations.** Add the product
  wordmark or text logo, favicon, theme colors, typography, visible focus states,
  skip navigation, accessible contrast, social metadata defaults, and responsive
  layout rules. Avoid inventing security certification claims. Acceptance: the
  landing placeholder is usable at narrow and wide viewports and has no critical
  automated accessibility violations. Commit: `style(docs): establish accessible site theme`.

- [x] **4. Add documentation CI.** Add a least-privilege workflow that installs
  locked dependencies, checks formatting, runs VitePress validation, builds the site,
  and checks internal links. Cache only safe dependency/build data. Acceptance:
  pull requests fail on broken builds or links, and workflow actions are pinned
  to full SHAs. Commit: `ci(docs): validate documentation site`.

- [x] **5. Deploy a previewable production site.** Add the GitHub Pages build and
  deployment workflow, correct `site` and `base` configuration, canonical URLs,
  artifact upload, concurrency cancellation, and documented repository settings.
  Keep pull requests read-only and without deployment credentials. Acceptance:
  the default branch deploys successfully and nested routes/assets work when
  opened directly. Commit: `ci(docs): deploy site to GitHub Pages`.

## Phase 2 — Information architecture and core content

- [x] **6. Implement the site navigation and content conventions.** Configure
  top navigation, sidebar groups, previous/next navigation, breadcrumbs, edit
  links, last-updated metadata, and a 404 page. Add an author guide describing
  titles, descriptions, headings, code examples, admonitions, terminology, and
  link style. Acceptance: every planned top-level section has a stable route and
  placeholder index page. Commit: `docs(site): define information architecture`.

- [x] **7. Build the product landing page.** Explain the policy-driven commit
  gate, six finding domains, offline-first behavior, local/hook/CI workflows, and
  report formats. Add installation and three copyable quick-start paths: local
  scan, staged pre-commit scan, and CI SARIF scan. Acceptance: a new user can
  install and run a first scan without opening another page. Commit:
  `docs(site): add product landing page`.

- [x] **8. Publish getting-started guides.** Add requirements, installation from
  source and releases, first scan, result interpretation, exit codes, and a
  guided next-step page. Explain that findings only return exit code 1 under
  policy-enforcing conditions such as `--ci`. Acceptance: commands match current
  CLI behavior on Windows and Unix where syntax differs. Commit:
  `docs(getting-started): publish first-scan guides`.

- [x] **9. Document scan modes, scope, profiles, and policy.** Cover full,
  changed, and staged discovery; index isolation; client/server/all scope;
  fast/standard/full/frontend profiles; offline profiles; domain thresholds;
  `--fail-on`; and `--new-only`. Include a decision table for choosing modes and
  profiles. Acceptance: staged-versus-working-tree guarantees and mutually
  exclusive flags are explicit. Commit: `docs(features): explain scan execution and policy`.

- [x] **10. Document reports and finding lifecycle.** Cover terminal, JSON,
  SARIF, and JUnit output; deterministic ordering; redaction boundaries;
  fingerprints; new/existing/resolved lifecycle; baselines; and suppressions.
  Include safe sample output with synthetic values only. Acceptance: users can
  distinguish baselines from suppressions and know which artifacts are suitable
  for CI systems. Commit: `docs(features): document reports and finding lifecycle`.

- [x] **11. Document hooks, cache, fixes, and utility commands.** Explain hook
  install/status/run/uninstall behavior, effective `core.hooksPath`, managed-file
  safety, cache identity and retention, deterministic fixes and dry runs, rule
  explanation, configuration validation, and upgrade checks. Acceptance: every
  example includes prerequisites and expected side effects. Commit:
  `docs(features): cover developer workflow features`.

- [x] **12. Publish frontend scanning documentation.** Convert the existing
  frontend contract into task-oriented pages covering framework detection,
  client/server/shared classification, supported ecosystems, threat boundaries,
  native rules, sanitizers, import cycles, secret exposure, telemetry privacy,
  and optional adapters. Preserve the contract as the authoritative behavioral
  specification and link to it. Acceptance: React/Next.js, Vue/Nuxt, and
  Svelte/SvelteKit examples include safe and unsafe patterns without real
  secrets. Commit: `docs(frontend): publish client scanning guide`.

## Phase 3 — Configuration documentation

- [x] **13. Split the configuration reference into navigable pages.** Preserve
  all material from `CONFIGURATION.md` while creating overview, paths and input,
  profiles and policy, scanner definitions, hooks, frontend, supply chain,
  governance, architecture, cache, and complete-example pages. Keep the original
  file as a concise pointer or generated aggregate to prevent two independent
  references. Acceptance: every currently documented version-1 field appears in
  exactly one canonical reference location. Commit:
  `docs(config): restructure configuration reference`.

- [x] **14. Add validated configuration recipes.** Add minimal, Go service,
  frontend application, monorepo, staged hook, offline, strict CI, external
  scanner, and gradual-adoption examples under `examples/config/`. Add a test
  that validates every complete example with the built CLI. Acceptance: examples
  use synthetic paths, state their tradeoffs, and pass strict decoding.
  Commit: `docs(config): add validated configuration recipes`.

- [x] **15. Introduce machine-readable configuration metadata.** Define a stable
  generator input or derive metadata from Go configuration definitions for field
  path, type, default, allowed values, requirement, description, and version.
  Fail generation on duplicate or undocumented public fields. Acceptance:
  metadata covers every serialized public configuration field and has a golden
  test. Commit: `feat(docs): expose configuration reference metadata`.

- [x] **16. Generate the complete configuration reference.** Build a deterministic
  Go generator that converts configuration metadata into Markdown tables and
  detail sections. Add `go generate` or an explicit documented command and a CI
  clean-tree check. Acceptance: two runs are byte-identical and CI detects stale
  generated pages. Commit: `docs(config): generate field reference from Go metadata`.

## Phase 4 — Generated product references

- [x] **17. Generate the CLI command reference.** Produce pages for scan,
  baseline, suppress, cache, config, hooks, release, upgrade, and version command
  groups from the command definitions or structured help metadata. Include usage,
  arguments, flags, defaults, exit behavior, and examples. Do not parse unstable
  terminal formatting when structured metadata can be exposed. Acceptance: every
  public command and flag appears once and generation is deterministic.
  Commit: `docs(cli): generate command reference`.

- [x] **18. Generate the rule catalog.** Expose safe rule metadata including ID,
  title, domain, severity, description, recommendation, supported scope, explain
  text, and fix availability. Generate index and detail pages with filters encoded
  in static data. Acceptance: rule IDs are unique, no detection pattern leaks
  sensitive test fixtures, and the catalog matches the registered default rules.
  Commit: `docs(rules): generate rule catalog`.

- [x] **19. Generate scanner and adapter compatibility pages.** Document built-in,
  command, and adapter scanners; executable requirements; network behavior;
  workspace support; parser format; missing-tool policy; supported profiles; and
  availability by operating system. Acceptance: every registered adapter is
  represented and offline/network claims have automated consistency checks.
  Commit: `docs(scanners): generate compatibility reference`.

- [x] **20. Add a unified reference generation gate.** Provide one cross-platform
  command that generates configuration, CLI, rule, and scanner references. Add a
  CI step that runs it and fails on a dirty tree, with actionable output for
  contributors. Acceptance: reference drift cannot merge unnoticed and the
  command works without network access after dependencies are installed.
  Commit: `ci(docs): prevent generated reference drift`.

## Phase 5 — Task-oriented guides and discoverability

- [x] **21. Add local and CI integration guides.** Publish guides for pre-commit,
  pre-push, GitHub Actions, generic CI, SARIF upload, artifact retention, and
  offline execution. Pin example actions, grant minimum permissions, and explain
  safe handling of untrusted pull requests. Acceptance: snippets are syntactically
  validated where tooling permits. Commit: `docs(guides): add local and CI integrations`.

- [ ] **22. Add adoption, troubleshooting, and migration guides.** Cover staged
  rollout with baselines, tuning thresholds, reviewing suppressions, missing
  external tools, timeouts, partial scans, path/symlink failures, cache issues,
  exit-code diagnosis, schema migration, and upgrade checks. Acceptance: every
  common operational failure exposed by the CLI has a troubleshooting route.
  Commit: `docs(guides): add adoption and troubleshooting playbooks`.

- [ ] **23. Add search, SEO, and social metadata.** Enable local static search,
  sitemap, robots policy, canonical metadata, Open Graph cards, descriptive page
  metadata, and meaningful heading structure. Exclude duplicate/generated data
  routes when appropriate. Acceptance: production search indexes core content and
  generated references, and canonical URLs contain the configured base path.
  Commit: `feat(docs): add search and discovery metadata`.

- [ ] **24. Add documentation quality tests.** Check internal links, duplicate
  headings/IDs, missing descriptions, malformed code fences, forbidden local
  paths, accidental secret-like examples, unreferenced pages, and spelling of
  product terminology. Maintain a narrow allowlist for intentional fake tokens.
  Acceptance: failures identify the page and remediation. Commit:
  `test(docs): enforce content quality`.

## Phase 6 — Interactive configuration and release readiness

- [ ] **25. Design the configuration builder contract.** Define supported fields,
  state model, defaults, presets, serialization rules, URL persistence policy,
  accessibility behavior, and explicit non-goals. The browser must not execute
  scanners, upload repositories, or send configuration data to a server.
  Acceptance: generated JSON semantics match version-1 overlay/default behavior.
  Commit: `docs(builder): define configuration generator contract`.

- [ ] **26. Implement the client-side configuration builder.** Add progressive,
  accessible controls for project basics, scan mode, profiles, policy thresholds,
  frontend settings, external scanners, governance, architecture, and cache.
  Generate copyable/downloadable JSON locally. Acceptance: keyboard-only use
  works, invalid combinations show actionable errors, and no user data leaves the
  browser. Commit: `feat(docs): add configuration builder`.

- [ ] **27. Validate builder output against Go configuration behavior.** Create
  shared fixtures covering presets, edge values, disabled sections, custom
  scanners, and invalid combinations. Test serialized output with Go's strict
  config loader and protect expected JSON with golden tests. Acceptance: every
  builder preset passes `config validate` and semantic defaults are documented.
  Commit: `test(docs): verify configuration builder output`.

- [ ] **28. Add documentation version and release integration.** Display the
  documented scanner release, define latest-versus-versioned URL behavior, add
  release notes and migration links, and update the release workflow or checklist
  to publish compatible documentation. Avoid cloning every version unless public
  compatibility requires it. Acceptance: users can identify which scanner version
  a reference describes and reach older compatibility information.
  Commit: `docs(release): integrate site with version lifecycle`.

- [ ] **29. Run documentation release acceptance.** Perform clean checkout/build,
  link, accessibility, responsive, browser smoke, generated-reference drift,
  example validation, security-header guidance, and GitHub Pages route checks.
  Fix discovered defects in focused prerequisite commits; use this commit only
  for the final acceptance automation and recorded checklist. Acceptance: all
  mandatory checks pass in CI and the published site matches the release commit.
  Commit: `test(docs): certify documentation site release`.

## Mandatory verification commands

Run from the repository root unless stated otherwise:

```sh
git diff --check
go test ./...
go vet ./...
go run ./cmd/security-review config validate examples/config/minimal.json
```

Run the documentation checks from `website/` using the selected package manager:

```sh
npm ci
npm run docs:dev
npm run docs:build
npm run docs:preview
```

Run the unified generator after task 20:

```sh
go generate ./...
git diff --exit-code
```

On Windows, provide equivalent PowerShell-friendly scripts through `npm run` or
Go commands rather than requiring Bash for documentation development.

## Launch checklist

- [ ] Installation and first scan are reachable from the home page.
- [ ] All six finding domains and supported scan modes are documented.
- [ ] Every public CLI command, flag, config field, rule, scanner, and adapter is
  present in a canonical reference.
- [ ] Examples are validated and contain no real secrets or machine-local paths.
- [ ] Search, direct nested routes, canonical URLs, sitemap, and 404 behavior work
  on the production base path.
- [ ] Keyboard navigation, focus visibility, contrast, headings, and landmarks
  meet the agreed accessibility target.
- [ ] Pull requests cannot deploy or access release credentials.
- [ ] Generated references are deterministic and CI rejects drift.
- [ ] The site identifies its compatible scanner release.
- [ ] README, website, security policy, migration guide, and contribution guide
  cross-link without duplicating independent sources of truth.

## Completion condition

The documentation site is complete when tasks 1–29 and the launch checklist are
checked, a clean checkout can reproduce the production site, all public behavior
is discoverable through either task-oriented guides or generated references, and
the published site remains synchronized with the Go implementation through CI.
