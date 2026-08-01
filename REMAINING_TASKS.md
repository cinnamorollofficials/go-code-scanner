# Remaining Implementation Tasks

This document is the handoff checklist for continuing the project in a new session. It reflects the implementation state after commit `f7543b3` and supersedes the unchecked state in `TODO.md`, whose checkboxes have not yet been reconciled with the code.

## Working agreement

- Complete tasks in the listed order unless a dependency requires otherwise.
- Create one atomic commit per numbered task.
- Preserve backward compatibility unless a task explicitly introduces a versioned migration.
- Keep all checks deterministic and offline by default.
- Never print secrets or sensitive source snippets.
- Add or update tests for every behavior change.
- After every task, run the relevant package tests, `go vet`, and `git diff --check`.
- After every five tasks, run the full verification suite.
- Write the complete English user documentation only after implementation tasks 1–20 are complete.

## Phase 1 — Complete the release pipeline

- [x] **1. Integrate deterministic archives into the release build.** Update the release build flow to package Unix binaries as `.tar.gz` and Windows binaries as `.zip` using `release.ArchiveBinary`. Ensure raw binaries are not accidentally included in the final distribution set. Commit suggestion: `feat(release): package distribution archives`.
- [x] **2. Add a release archive CLI command.** Provide a stable command for creating an archive from a binary with an explicit timestamp. Reject symlinks, unsupported extensions, missing timestamps, and unexpected positional arguments. Commit suggestion: `feat(cli): create release archives`.
- [x] **3. Add a checksum verification CLI command.** Expose `release.VerifyChecksums` through `security-review release checksums verify`. Use exit code `0` for success, `1` for verification mismatch, and `2` for invalid arguments or unreadable inputs. Commit suggestion: `feat(cli): verify release checksums`.
- [x] **4. Extend provenance verification to validate subjects.** Update the existing release verification command so signature validation and local subject digest validation can be requested together. Preserve signature-only compatibility. Commit suggestion: `feat(cli): verify provenance artifacts`.
- [x] **5. Add provenance generation and signing commands.** Expose deterministic provenance generation and optional Ed25519 detached signing without weakening private-key permission checks. Commit suggestion: `feat(cli): generate signed provenance`.

## Phase 2 — Publishable and reproducible releases

- [x] **6. Generate provenance in the tagged release workflow.** Include version, commit, UTC build timestamp, builder identity, and every distributable archive. Validate the generated document before continuing. Commit suggestion: `ci(release): generate artifact provenance`.
- [x] **7. Sign provenance in the release workflow.** Use a documented secret input or keyless mechanism, avoid writing key material with broad permissions, and verify the signature before publication. Keep untrusted pull-request workflows unable to access signing credentials. Commit suggestion: `ci(release): sign artifact provenance`.
- [x] **8. Upload verified release artifacts.** Pin the upload action to a full commit SHA, upload archives, `SHA256SUMS`, provenance, and signature, and set explicit retention. The workflow must remain least-privilege. Commit suggestion: `ci(release): upload verified artifacts`.
- [x] **9. Add reproducible-build verification.** Build the same target twice with identical metadata and assert equal digests. Normalize every timestamp or metadata source that prevents reproducibility. Commit suggestion: `test(release): verify reproducible builds`.
- [x] **10. Add release end-to-end tests.** Exercise build, archive, checksum, provenance, signing, and verification in a temporary directory, including tampered artifact and tampered signature cases. Commit suggestion: `test(release): cover end-to-end pipeline`.

## Phase 3 — Final quality and security gates

- [x] **11. Add fuzz-smoke execution to verification.** Run existing fuzz targets for a short bounded duration covering config, rules, suppression, baseline, and external adapter parsing. Ensure the command is stable in CI. Commit suggestion: `ci(test): run parser fuzz smoke tests`.
- [ ] **12. Add vulnerability scanning.** Run `govulncheck ./...` when available, pin installation/version behavior, and distinguish tool unavailability from discovered vulnerabilities. Commit suggestion: `ci(security): scan Go vulnerabilities`.
- [ ] **13. Add cross-platform CI.** Verify build and platform-sensitive tests on Linux, macOS, and Windows. Keep expensive race tests on supported runners only and ensure path/permission expectations are portable. Commit suggestion: `ci: verify supported platforms`.
- [ ] **14. Add artifact and report redaction tests.** Assert secrets cannot appear in terminal, JSON, SARIF, JUnit, cache entries, provenance errors, or external scanner diagnostics. Commit suggestion: `test(security): verify output redaction`.
- [ ] **15. Add resource-boundary integration tests.** Cover oversized files, excessive file counts, command output limits, timeouts, cancellation, snapshot cleanup, and cache retention under boundary conditions. Commit suggestion: `test(hardening): cover resource boundaries`.

## Phase 4 — Product acceptance

- [ ] **16. Verify hook installation using a release binary.** Build a real binary, install each supported hook into a temporary Git repository, execute it, inspect status, uninstall it, and confirm existing hooks remain intact. Commit suggestion: `test(hook): verify release binary lifecycle`.
- [ ] **17. Verify staged-content isolation end to end.** For every scanner claiming staged support, prove that staged safe content is not replaced by an unsafe unstaged working-tree version and vice versa. Commit suggestion: `test(scanner): verify staged isolation`.
- [ ] **18. Validate all public output contracts.** Add or complete golden/schema checks for JSON, SARIF 2.1.0, JUnit XML, compatibility manifests, checksum manifests, and provenance. Commit suggestion: `test(contract): validate public formats`.
- [ ] **19. Benchmark performance budgets.** Measure discovery, pattern scanning, baseline comparison, cache hits, and the fast pre-commit profile. Record enforceable thresholds carefully enough to avoid flaky CI. Commit suggestion: `test(performance): enforce commit gate budgets`.
- [ ] **20. Run and fix the release-candidate gate.** Execute unit, integration, end-to-end, race, vet, fuzz-smoke, vulnerability, golden, self-scan, cached/uncached equivalence, and release reproducibility checks. Fix failures in separate atomic commits if required. Commit suggestion: `test(release): certify release candidate`.

## Phase 5 — Documentation and project handoff

- [ ] **21. Reconcile `TODO.md` with the implementation.** Audit every roadmap checkbox against code and tests. Mark completed items, retain genuinely open items, and link deferred ideas to this file or a future roadmap. Do not mark acceptance criteria complete without evidence. Commit suggestion: `docs(roadmap): reconcile implementation status`.
- [ ] **22. Rewrite the README in English.** Cover purpose, architecture, supported domains, installation, quick start, exit codes, profiles, hooks, staged behavior, reports, baseline, suppression, cache, release verification, and troubleshooting. Commit suggestion: `docs: complete project README`.
- [ ] **23. Add a complete configuration reference.** Document every config field, default, validation rule, compatibility behavior, scanner type, adapter, environment allowlist, workspace mode, policy threshold, and hook option with minimal and full examples. Commit suggestion: `docs(config): add complete reference`.
- [ ] **24. Add security and release documentation.** Document threat model, trust boundaries, redaction, symlink/path hardening, external command execution, signing-key handling, checksum/provenance verification, reproducible builds, and the release process. Commit suggestion: `docs(security): document hardening and releases`.
- [ ] **25. Add contribution and migration guides.** Explain development setup, tests, adding rules/scanners/adapters, commit conventions, compatibility contracts, schema migrations, upgrade checks, and release checklist ownership. Commit suggestion: `docs: add contribution and migration guides`.

## Mandatory verification commands

Use a writable Go cache when the environment restricts the default user cache:

```sh
GOCACHE=/tmp/go-code-scanner-gocache go test ./...
GOCACHE=/tmp/go-code-scanner-gocache go test -race ./...
GOCACHE=/tmp/go-code-scanner-gocache go vet ./...
git diff --check
```

Also run the repository verification entry point when the environment permits it:

```sh
GOCACHE=/tmp/go-code-scanner-gocache ./scripts/verify.sh
```

## Completion condition

The project is complete when all 25 tasks are checked, every required verification command passes, tagged release artifacts can be independently verified from checksum and signed provenance, the release binary can safely manage hooks on supported platforms, and the English documentation describes all public behavior without relying on undocumented defaults.
