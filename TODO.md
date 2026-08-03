# Implementation Roadmap Status

This file reconciles the original M0–M8 roadmap against the implementation and
tests as of the task 20 release-candidate gate. The old checklist was written
before most packages existed and is preserved in Git history. Current release
work is tracked in [`REMAINING_TASKS.md`](REMAINING_TASKS.md).

## Completed milestones

- [x] **M0 — Domain and configuration contracts.** Six finding domains,
  lifecycle states, rule metadata, profiles, per-domain policy, hook settings,
  strict version-1 configuration decoding, and compatibility manifests are
  implemented in `finding/`, `rules/`, `config/`, `policy/`, and
  `compatibility/`.
- [x] **M1 — Scanner orchestration.** Enabled/required/timeout behavior,
  deterministic concurrency, structured failure kinds, panic recovery,
  cancellation, partial reports, and policy separation are covered by
  `securityreview_test.go` and race tests.
- [x] **M2 — Git hooks and staged workspaces.** NUL-safe Git discovery,
  effective hooks-path handling, index snapshots, resource limits, cleanup,
  safe hook ownership, all hook commands, and stable exit codes are implemented
  and exercised through temporary repositories and a real release binary.
- [x] **M3 — Built-in fast checks.** Quality, reliability, hardening, security,
  supply-chain, and governance rule packs have positive/negative fixtures,
  bounded input handling, context-aware redaction, and an enforced fast-profile
  performance budget.
- [x] **M4 — External command scanners.** Argument-array execution, sanitized
  environments, executable validation, root/staged workspaces, process-group
  cancellation, output limits, exit-code mapping, structured parsers, and the
  supported adapter presets are implemented under `scanner/command` and
  `scanner/adapters`.
- [x] **M5 — Baseline and incremental policy.** Versioned atomic baselines,
  stable fingerprinting, relocation behavior, create/update/status commands,
  new-only policy, dry-run updates, and suppression separation are covered by
  baseline, lifecycle, CLI, and policy tests.
- [x] **M6 — Developer experience and reports.** Actionable terminal output,
  TTY/`NO_COLOR` handling, JSON, SARIF 2.1.0, JUnit XML, rule explanation,
  deterministic fixes, safe report replacement, and public contract tests are
  implemented.
- [x] **M7 — Supply chain and governance.** Optional vulnerability/security
  adapters, dependency and license policy, immutable-action and container
  checks, commit-message policy, required files/headers, ownership,
  architecture boundaries/cycles, and offline-profile isolation are present.
- [x] **M8 — Cache, hardening, and distribution.** Content-addressed cache,
  locking/retention/redaction, path and symlink hardening, fuzz/race/vet/self
  scanning, cross-platform deterministic archives, checksums, signed
  provenance, compatibility checks, and the release-candidate gate are complete.

## Acceptance evidence

- [x] Unit, integration, end-to-end, cancellation, and boundary tests pass.
- [x] Staged-safe/working-unsafe and staged-unsafe/working-safe isolation is
  verified for every scanner implementation that claims staged support.
- [x] Terminal, JSON, SARIF, JUnit, cache, provenance-error, and external-tool
  diagnostics have secret-canary redaction coverage.
- [x] Linux, macOS, and Windows build/test jobs use pinned actions.
- [x] Discovery, pattern scanning, baseline comparison, cache hits, and the fast
  pre-commit path have enforceable, non-racy performance budgets.
- [x] Cached and uncached results are equivalent and content changes invalidate
  cache entries.
- [x] A release binary safely installs, executes, reports, and uninstalls all
  supported hooks without replacing an unmanaged hook.
- [x] JSON, SARIF, JUnit, compatibility, checksum, and provenance contracts are
  protected by strict structure checks or deterministic golden files.
- [x] `scripts/release-candidate.sh` runs tests, race, vet, diff checks,
  acceptance tests, fuzz smoke, vulnerability scanning when available,
  performance budgets, reproducibility, and self-scan.

## Deferred ideas

These items were suggestions in the original roadmap, not acceptance blockers
for the current release. They remain intentionally open for a future roadmap:

First-class browser client scanning is planned separately in
[`FRONTEND_SCANNING_TODO.md`](FRONTEND_SCANNING_TODO.md). That backlog is split
into atomic numbered tasks, with one required commit per task.

- [ ] Add captured adapter-output fixtures from multiple upstream tool versions;
  current parsers have representative fixtures and fuzz coverage.
- [ ] Establish a measured false-positive corpus across several external
  repositories; current rule packs use positive and negative fixtures.
- [ ] Add an explicit `--format terminal` alias; terminal output is currently
  always produced unless `--quiet`, while `--format` selects artifact formats.
- [ ] Add preview modes for hook installation and the suppression helper;
  baseline update and deterministic source fixes already support dry-run.
- [ ] Record the remote vulnerability-database revision separately from scanner
  version when upstream tools expose a stable contract for it.

## Documentation handoff

- [x] Complete English user documentation — tracked by tasks 22–25 in
  [`REMAINING_TASKS.md`](REMAINING_TASKS.md).
