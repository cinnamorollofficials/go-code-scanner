# Website Documentation TODO — Second-Pass Audit

This backlog covers only the documentation website under `website/`. It replaces
the first-pass checklist, whose items were marked complete before several of them
had been verified against the rendered site and the current CLI behavior.

Priority legend:

- **P0** — incorrect or misleading documentation; fix before publishing.
- **P1** — structural or navigation issues that materially affect readers.
- **P2** — automated quality, accessibility, SEO, and maintenance improvements.
- **P3** — editorial and visual polish after the documentation is trustworthy.

An item may be checked only when its acceptance criteria have been verified. A
successful VitePress build alone is not sufficient evidence that links, commands,
configuration examples, or accessibility behavior are correct.

## Existing foundations to preserve

- [x] The primary information architecture uses Get Started, Guides, Concepts,
  Reference, and Project groups.
- [x] The homepage provides goal-oriented entry points and a first-scan path.
- [x] Local documentation search is enabled.
- [x] The Config Builder provides reset, copy, download, validation feedback,
  and `aria-live` status announcements.
- [x] The production VitePress build runs successfully.

## P0 — Correct inaccurate documentation

### Installation and command examples

- [x] In `docs/getting-started/installation.md`, replace the unsupported
  checksum verification flag `--checksums` with the current `--manifest` form.
- [x] Run every copyable installation command against the current CLI help and
  release artifact layout.
- [x] Replace handwritten command output with output captured from a stable test
  fixture, or clearly label it as illustrative.
- [x] Replace the homepage's fictional or stale scanner names with output
  produced by the default scan profile.

Acceptance criteria:

- Every installation and first-scan command is accepted by the current CLI.
- The documented output uses scanner names and wording emitted by the product.
- Any intentionally abbreviated output is labelled as abbreviated.

### Suppressions

- [ ] In `docs/guides/suppressions.md`, mark `--reason` and `--expires` as
  required rather than optional.
- [ ] Remove the documented `// nolint:...` inline suppression syntax unless the
  CLI implements and tests it as a public feature.
- [ ] Verify the add, list, remove, and expiry examples against the current
  suppression command behavior.
- [ ] Explain the supported suppression lifecycle without suggesting an
  unsupported source-comment workflow.

Acceptance criteria:

- Every copied suppression command succeeds with the documented inputs.
- Required fields match both CLI help and runtime validation.
- The page contains no unsupported suppression mechanism.

### Scanner configuration reference

- [ ] In `docs/reference/config/scanners.md`, document only supported scanner
  types: the default/pattern form, `command`, and `adapter` as applicable.
- [ ] Remove `builtin` as a supported scanner type unless it becomes a valid
  configuration value.
- [ ] Replace noncanonical policy domains such as `secrets`, `vulnerabilities`,
  `architecture`, and `frontend` with the six canonical domains.
- [ ] Correct the required-scanner failure behavior to use the operational
  failure exit code rather than the invalid-input exit code.
- [ ] Verify every scanner configuration example with the real configuration
  validator.

Acceptance criteria:

- Allowed values, domains, defaults, and exit behavior agree with the product.
- All JSON examples pass the current validator without manual edits.

## P0 — Make the Config Builder export valid configurations

- [ ] In `docs/.vitepress/theme/components/ConfigBuilder.vue`, change the
  external-scanner preset from the unsupported `type: "external"` value to the
  canonical adapter configuration.
- [ ] Validate nested preset data, not only the small set of editable top-level
  fields.
- [ ] Add a fixture test that exports every preset and checks it with the real
  configuration validator.
- [ ] Review preset semantics so names such as `gradual-adoption` and
  `offline-strict` include the settings required to deliver the advertised
  behavior.
- [ ] Warn for confirmation before a preset replaces unsaved edits; the dirty
  badge alone does not prevent accidental loss.
- [ ] Keep the preset count and behavior synchronized across
  `docs/reference/config-builder.md` and
  `docs/reference/config-builder-contract.md`.
- [ ] Document whether the browser-side checks are convenience validation or a
  complete representation of CLI validation.

Acceptance criteria:

- Every preset exports JSON accepted by the current CLI validator.
- Selecting a preset never silently discards edited configuration.
- The builder, user guide, and contract advertise the same capabilities.

## P1 — Finish the `/features/` migration

- [ ] Create a migration table mapping each page in `docs/features/` to its
  canonical replacement in `docs/concepts/`, `docs/guides/`, or `docs/reference/`.
- [ ] Update all inbound links, including the links in
  `docs/getting-started/first-scan.md`, to their canonical destinations.
- [ ] Add redirects for externally linked `/features/*` URLs where VitePress
  hosting supports them.
- [ ] Remove duplicate legacy pages after redirects and inbound links are ready.
- [ ] Remove the `/features/` sidebar mapping and active navigation match from
  `docs/.vitepress/config.mts`.
- [ ] Confirm that the generated sitemap contains only one canonical page for
  each topic.
- [ ] Check that no legacy feature page remains discoverable through search,
  sidebar navigation, or the sitemap.

Acceptance criteria:

- Each topic has one source of truth.
- Existing public `/features/*` links resolve to an intentional destination.
- No duplicate feature/concept pages are indexed by the generated site.

## P1 — Rebuild the Rule Catalog for navigation, not scrolling

- [ ] Replace the single multi-thousand-line `docs/reference/rules.md` output
  with a generated catalog index and smaller rule detail pages.
- [ ] Add client-side filters for rule ID, canonical domain, severity,
  language/ecosystem, and category.
- [ ] Keep a compact rule matrix on the catalog index.
- [ ] Generate rule counts and domain summaries from the same source as rule
  details.
- [ ] Replace Markdown-like text inside raw HTML with real rendered links.
- [ ] Generate working previous/next navigation between rule detail pages.
- [ ] Preserve useful legacy anchors or provide redirects for inbound links.
- [ ] Add automated checks for every detail link, previous/next link, and
  preserved anchor.
- [ ] Verify catalog filtering and the detail layout at 320 px, 768 px, and
  desktop widths.

Acceptance criteria:

- Readers can find a rule without scrolling through one enormous page.
- Every rule, guidance, previous/next, and back-to-catalog link resolves.
- Counts and summaries cannot drift from the generated rule set.

## P1 — Reduce navigation and content duplication

- [ ] Collapse inactive sidebar groups or use section-specific sidebars so each
  page does not expose the full documentation tree at once.
- [ ] Turn the top navigation into predictable Learn/Guides, Reference, and
  Project destinations or dropdowns.
- [ ] Rename the `Unreleased` external link to `Development docs` or replace it
  with a clearer version-status treatment.
- [ ] Expand `docs/getting-started/index.md` and `docs/guides/index.md` into useful
  routing pages with audience, task outcome, and approximate completion time.
- [ ] Keep `docs/guides/troubleshooting.md` focused on diagnosis; move duplicated
  baseline/adoption instructions to the dedicated baseline guide.
- [ ] Apply this content model consistently:
  - Getting Started: shortest path to first success.
  - Guide: steps that complete a user task.
  - Concept: mental model and product behavior.
  - Reference: exact commands, fields, values, and contracts.
- [ ] Prefer cross-links over repeating complete command sequences across
  multiple content layers.

Acceptance criteria:

- A reader can predict where a topic belongs from its page type.
- Core tasks remain reachable within two navigation choices.
- Repeated instructions have one canonical owner.

## P2 — Add documentation quality gates under `website/`

- [ ] Add a `docs:check-links` script that fails on broken internal links and
  missing anchors in the production output.
- [ ] Add a content check for valid frontmatter, unique titles, useful
  descriptions, exactly one H1, and labelled fenced code blocks.
- [ ] Add a command-example check that compares documented CLI flags and
  subcommands with the current command definitions.
- [ ] Validate copyable JSON configurations and every Config Builder preset with
  the real CLI validator.
- [ ] Continue checking generated rule, scanner, and configuration references
  for drift.
- [ ] Add browser smoke tests for the homepage, first-scan guide, CLI reference,
  Rule Catalog, and Config Builder.
- [ ] Add accessibility smoke tests covering keyboard use, focus visibility,
  landmarks, headings, labels, status announcements, and color contrast.
- [ ] Provide one `docs:verify` script that runs content checks, generated-content
  checks, tests, and the production build.
- [ ] Document each check and its expected failure message in the documentation
  author guide.

Acceptance criteria:

- `npm run docs:verify` fails for a deliberately broken link, invalid command,
  invalid preset, duplicate H1, or critical accessibility regression.
- Contributors can run the same verification locally before opening a change.

## P2 — Accessibility, responsive behavior, and metadata

- [ ] Add `prefers-reduced-motion` handling for custom transitions and animated
  controls.
- [ ] Verify the header, mobile menu, sidebar, search, Config Builder, tables,
  and code-copy controls using keyboard-only navigation.
- [ ] Add breadcrumbs or compact section labels to deep configuration and rule
  reference pages.
- [ ] Add an Open Graph image and `twitter:card` metadata.
- [ ] Add canonical URL metadata for generated documentation pages.
- [ ] Add a concise feedback or issue-reporting link only if the destination is
  actively maintained.
- [ ] Confirm that dense tables and generated rule content do not cause unusable
  horizontal or vertical navigation on mobile.

Acceptance criteria:

- Reduced-motion preferences are respected.
- Critical pages pass the selected accessibility smoke test at mobile and
  desktop sizes.
- Shared pages have a useful title, description, preview image, and canonical URL.

## P2 — Clean up website deployment documentation

- [ ] Replace every `file:///Users/...` link in `DEPLOYMENT.md` with a portable
  repository-relative link.
- [ ] Decide whether deployment documentation is an internal Indonesian guide or
  part of the public English documentation, then apply that choice consistently.
- [ ] Lead with the supported deployment path and move optional VPS, reverse
  proxy, and Certbot procedures into clearly labelled advanced sections.
- [ ] Verify Docker, Docker Compose, base-path, and update commands against the
  files currently shipped under `website/`.
- [ ] Remove decorative emoji where they make the operational guide harder to
  scan.

Acceptance criteria:

- The guide contains no workstation-specific path.
- A fresh reader can follow the primary deployment path from the repository
  without guessing which option is supported.

## P3 — Editorial polish

- [ ] Standardize page summaries so they state audience, outcome, and scope.
- [ ] Standardize terminology, capitalization, table headings, punctuation, and
  callout usage.
- [ ] Review page titles and descriptions for search intent and duplicate wording.
- [ ] Label complete, abbreviated, and illustrative examples consistently.
- [ ] Remove claims such as “complete”, “supported”, or “verified” unless a test
  or public contract provides evidence.

## Completion criteria

- [ ] All P0 items and their acceptance criteria are complete.
- [ ] Every topic has one canonical page and the sitemap contains no stale
  `/features/*` duplicates.
- [ ] All Config Builder presets export valid configuration.
- [ ] The Rule Catalog is filterable and no longer depends on a single enormous
  page.
- [ ] Installation, first scan, CI, baseline, suppression, and scanner examples
  match current product behavior.
- [ ] `npm run docs:verify` passes locally and in the website build pipeline.
- [ ] Critical pages pass link, responsive, and accessibility smoke tests.
- [ ] Each checked item has reproducible evidence in a test, generated artifact,
  or documented manual verification result.
