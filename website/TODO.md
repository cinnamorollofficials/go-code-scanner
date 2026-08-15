# Website Documentation TODO

This checklist covers only the documentation website under `website/`. Complete
the correctness work before reorganizing or polishing the UI so readers never
receive a cleaner presentation of inaccurate instructions.

Priority legend:

- **P0** — incorrect or misleading documentation; fix before publishing.
- **P1** — navigation and task-flow problems that materially affect usability.
- **P2** — maintainability, accessibility, and content-quality improvements.
- **P3** — optional polish after the core documentation is trustworthy.

## P0 — Restore documentation accuracy

### Canonical CLI commands and behavior

- [x] Audit every shell command in `docs/**/*.md` against the current
  `security-review` CLI.
- [x] Replace `--mode staged` with `--staged` and `--mode changed` with
  `--changed`; document that a full scan is the default when neither flag is
  supplied.
- [x] Remove or replace unsupported examples using `--scanners=sqltaint`.
- [x] Remove `--format terminal`; explain that terminal output is shown unless
  `--quiet`, while `--format` selects the artifact format.
- [x] Clarify that `--fail-on` changes the threshold but exit code `1` still
  requires `--ci`.
- [x] Correct baseline commands to use `create`, `update`, and `status` with the
  required `--report` and `--baseline` arguments.
- [x] Correct suppression examples to use the supported `suppress add` fields,
  including file, reason, and expiry.
- [x] Replace unsupported cache commands with `cache stats` and `cache clean`.
- [x] Replace `upgrade --check` with `upgrade check`.
- [x] Correct hook examples so each command includes a supported hook event.
- [x] Remove references to unsupported hook flags such as `--overwrite`.
- [x] Document all exit codes consistently: `0` success/allowed, `1` policy or
  verification mismatch, `2` invalid input/configuration, and `3` operational
  failure.
- [x] Update the sample terminal output so its wording and layout match actual
  `security-review scan` output.

Affected pages:

- `docs/index.md`
- `docs/getting-started/first-scan.md`
- `docs/reference/cli.md`
- `docs/features/scan-execution-and-policy.md`
- `docs/features/reports-and-finding-lifecycle.md`
- `docs/features/developer-workflow-features.md`
- `docs/features/analysis-and-reproduction.md`
- `docs/features/sql-taint-analysis.md`
- `docs/guides/ci-integrations.md`
- `docs/guides/troubleshooting.md`

### Canonical product terminology

- [x] Use the canonical finding domains everywhere: `quality`, `reliability`,
  `hardening`, `security`, `supply_chain`, and `governance`.
- [x] Use only supported severities: `CRITICAL`, `HIGH`, `MEDIUM`, and `LOW`.
- [x] Remove `INFO` from configuration references, the CLI reference, and the
  interactive builder.
- [x] Distinguish domains from categories and capabilities. Terms such as
  “Secrets”, “SAST”, “Frontend”, and “Vulnerabilities” may describe capabilities
  but must not be presented as policy domains.
- [x] Use **Go Code Scanner** for the product and `security-review` for the CLI,
  following `docs/author-guide.md`.

### Installation, release, and trust information

- [x] Change the Go prerequisite in `docs/getting-started/installation.md` to the
  version declared by the repository's `go.mod`.
- [x] Replace the hard-coded `v1.0.0` navigation label with the actual released
  version, or display `Pre-release`/`Unreleased` until a stable tag exists.
- [x] Ensure release-download commands use artifact names that are actually
  published; hide the precompiled-binary path until artifacts exist.
- [x] Keep `docs/changelog.md`, the navigation version, and installation page in
  sync.
- [x] Link the footer's MIT License statement to an actual license document, or
  remove the claim until that document exists.
- [x] Replace the vague security-reporting instruction with a concrete private
  reporting URL or contact method.

## P0 — Make the Config Builder truthful and functional

- [x] Decide whether the builder is a complete configuration generator or a
  small starter; align its title and description with that scope.
- [x] Make the preset list, preset implementation, and
  `docs/reference/config-builder-contract.md` agree on the same preset count.
- [x] Implement every selectable preset so choosing one never silently does
  nothing.
- [x] Remove the invalid `INFO` threshold.
- [x] Add the promised file-download action, or remove it from the contract.
- [x] Add client-side validation for required values, number ranges, enum values,
  and unsafe/empty paths before copy or download.
- [x] Show clear success and failure states for clipboard and download actions.
- [x] Announce copy/download status through an `aria-live` region.
- [x] Add a reset action and a clear indication when changing presets will
  replace unsaved edits.
- [x] If the builder remains “complete”, expose the high-value policy, profile,
  scanner, hook, frontend, and cache settings instead of only four editable
  fields.
- [x] Test exported JSON with the project's real configuration validator.

Affected files:

- `docs/.vitepress/theme/components/ConfigBuilder.vue`
- `docs/reference/config-builder.md`
- `docs/reference/config-builder-contract.md`

## P1 — Reorganize navigation around reader goals

- [ ] Replace the current sidebar groups with the following information
  architecture:

  ```text
  Get Started
    Installation
    First Scan
    Five-Minute CI Setup

  Guides
    Pre-Commit Hooks
    GitHub Actions / GitLab CI
    Gradual Adoption with Baselines
    Managing Suppressions
    Reproducing a Finding
    Troubleshooting

  Concepts
    Scan Modes and Isolation
    Profiles and Policy
    Reports and Finding Lifecycle
    Frontend Scanning
    SQL Taint Analysis

  Reference
    CLI
    Configuration
    Scanner Compatibility
    Rule Catalog
    Config Builder

  Project
    Security Model
    Changelog
    Contributing
    Documentation Author Guide
  ```

- [ ] Rename the misleading “Development” group to “Project”.
- [ ] Move the Documentation Author Guide out of end-user guides.
- [ ] Split “Adoption & Troubleshooting” into separate task-oriented pages.
- [ ] Split “How It Works & Reproducing Findings” into a concept page and an
  operational reproduction guide.
- [ ] Remove bilingual parenthetical headings unless the whole site adopts a
  documented localization strategy.
- [ ] Add the SQL Taint Analysis page to the Features/Concepts overview.
- [ ] Either expand thin overview pages with useful routing context or link
  directly to the first meaningful page.
- [ ] Make “Documentation”, “Guides”, “Reference”, and “Project” top-level paths
  predictable and apply consistent active navigation states.

Affected files:

- `docs/.vitepress/config.mts`
- `docs/getting-started/index.md`
- `docs/features/index.md`
- `docs/guides/index.md`
- `docs/reference/index.md`

## P1 — Improve the homepage onboarding path

- [ ] Keep the hero concise and place a “Choose your goal” section immediately
  after it: local scan, pre-commit gate, CI integration, existing-project
  adoption, and reference lookup.
- [ ] Correct all three quick-start commands.
- [ ] Add a realistic expected-output excerpt to the local-scan path.
- [ ] Explain that the default scan writes `security_findings.json`.
- [ ] Explain local versus `--ci` exit behavior next to the first scan command.
- [ ] Link every quick-start path to a complete task-oriented guide.
- [ ] Present the six canonical domains consistently and move detailed marketing
  claims below the first successful user workflow.
- [ ] Avoid absolute claims such as “enterprise-grade” and “100%” unless the
  relevant guarantee is explicitly scoped and verifiable.

Affected file: `docs/index.md`.

## P1 — Make the Rule Catalog navigable

- [ ] Replace the single 2,000+ line catalog page with a generated index and
  smaller detail pages, preferably grouped by canonical domain.
- [ ] Add filters for rule ID, domain, severity, language/ecosystem, and category.
- [ ] Keep a compact rule matrix above detailed remediation guidance.
- [ ] Preserve existing rule anchors or add redirects so inbound links continue
  working after the split.
- [ ] Ensure rule counts and domain summaries are generated rather than manually
  maintained.
- [ ] Add previous/next navigation between rule details.
- [ ] Verify every “View Guidance” link resolves to an existing rule.

Affected file: `docs/reference/rules.md` and its generator output structure.

## P2 — Strengthen reference content

- [ ] Generate the CLI reference from the actual command and flag definitions.
- [ ] Keep generated configuration defaults synchronized with the configuration
  metadata source.
- [ ] Add clear “Required?”, “Default”, “Allowed values”, and “Example” fields to
  each configuration reference section.
- [ ] Cross-link related concepts, such as `--ci` with `fail_on`, and baselines
  with `--new-only`.
- [ ] Clearly label generated pages and provide a link to their source of truth.
- [ ] Explain whether code examples are complete, abbreviated, or illustrative.
- [ ] Avoid describing internal scanner implementation details as user-visible
  guarantees unless they are covered by tests and public contracts.

## P2 — Accessibility and responsive behavior

- [ ] Verify keyboard navigation for the header, mobile menu, sidebar, local
  search, Config Builder, and code-copy controls.
- [ ] Check visible focus in light and dark modes.
- [ ] Confirm color contrast for brand buttons, links, callouts, tables, and code
  blocks in both themes.
- [ ] Add accessible names and status announcements to interactive controls.
- [ ] Test dense tables at 320 px, 768 px, and desktop widths.
- [ ] Ensure the Rule Catalog does not create an unusably long mobile table of
  contents.
- [ ] Respect reduced-motion preferences for any added transitions.
- [ ] Confirm heading order and landmark structure on every page template.

## P2 — Documentation quality gates

- [ ] Add CI checks for broken internal links and missing anchors.
- [ ] Require valid frontmatter with a unique title and useful description.
- [ ] Check for exactly one H1 per content page.
- [ ] Verify fenced code blocks specify an appropriate language.
- [ ] Validate every documented CLI command against current command definitions.
- [ ] Run safe copyable examples in temporary fixtures where practical.
- [ ] Verify configuration JSON examples with `security-review config validate`.
- [ ] Check that generated rule, scanner, and configuration references are not
  stale.
- [ ] Fail CI when the displayed release version disagrees with the release tag
  or changelog.
- [ ] Run the VitePress production build in CI.
- [ ] Add an accessibility smoke test for the homepage, first-scan guide, CLI
  reference, Rule Catalog, and Config Builder.
- [ ] Keep `docs/.vitepress/dist` and `docs/.vitepress/cache` untracked.

## P3 — Visual and editorial polish

- [ ] Add small, consistent goal cards to the homepage rather than increasing
  the number of marketing feature cards.
- [ ] Add breadcrumbs or a compact section label above deep reference pages.
- [ ] Use consistent labels for “Overview”, “Guide”, “Concept”, and “Reference”.
- [ ] Standardize table terminology, capitalization, punctuation, and emoji use.
- [ ] Add a concise “Was this page helpful?” or issue-reporting link if there is
  a maintained feedback channel.
- [ ] Add Open Graph image metadata for shared documentation links.
- [ ] Review page titles and descriptions for search intent and remove duplicate
  wording.

## Completion criteria

- [ ] Every copyable command is supported by the current CLI and has the expected
  exit behavior.
- [ ] Domain, severity, version, installation, and release information is
  consistent across the website.
- [ ] All Config Builder presets and advertised actions work and export valid
  configuration.
- [ ] A new user can move from installation to a successful local scan, then to
  pre-commit or CI setup, without searching outside the website.
- [ ] The Rule Catalog is searchable or filterable without scrolling through a
  single multi-thousand-line page.
- [ ] Production build, link checks, example validation, and accessibility smoke
  tests pass in CI.
