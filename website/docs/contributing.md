---
title: Contributing Guide
description: Guidelines, development setup, change design rules, tests, and release checklists for contributing to Go Code Scanner.
---

# Contributing to Go Code Scanner

Thank you for improving Go Code Scanner. Changes should preserve deterministic, offline-safe behavior, staged-content isolation, output compatibility, and redaction guarantees.

## Development Setup

Requirements:

- Go version declared by `go.mod` (currently 1.25 or newer).
- Git in `PATH` for repository, staged, hook, and release tests.
- A Unix shell for canonical scripts. Cross-platform Go tests also run on Windows CI.
- Optional `govulncheck v1.1.4` for the vulnerability gate.

Clone the repository and warm the test/build cache:

```sh
git clone <repository-url>
cd go-code-scanner
GOCACHE=/tmp/go-code-scanner-gocache go test ./...
```

Do not commit generated scan reports, cache directories, release artifacts, or fuzz work directories.

## Required Checks

For an ordinary change, run the affected package tests, vet, and whitespace validation:

```sh
GOCACHE=/tmp/go-code-scanner-gocache go test ./path/to/package
GOCACHE=/tmp/go-code-scanner-gocache go vet ./...
git diff --check
```

Before handoff or release, run the canonical gate:

```sh
GOCACHE=/tmp/go-code-scanner-gocache ./scripts/release-candidate.sh
```

The gate runs:

- all unit and integration tests;
- race detection;
- `go vet` and `git diff --check`;
- cached/uncached equivalence;
- release reproducibility, release E2E, and real-binary hook lifecycle;
- public-format golden/structure tests;
- bounded fuzz smoke for config, rules, suppression, baseline, and adapters;
- pinned vulnerability scanning when available;
- performance budgets and repository self-scan.

If the default Go cache is not writable, use the `/tmp` `GOCACHE` shown above.

## Repository Structure

| Path | Responsibility |
| --- | --- |
| `cmd/security-review` | CLI parsing, output, and exit-code mapping |
| `securityreview.go` | scanner orchestration and report lifecycle |
| `config` | defaults, strict decoding, validation, path resolution |
| `finding`, `policy` | public finding model and allow/block decisions |
| `discovery`, `gitrepo`, `workspace` | filesystem/Git inputs and index snapshots |
| `rules`, `scanner/pattern` | built-in/custom rules and fast checks |
| `scanner/command`, `scanner/adapters` | external process execution and parsers |
| `baseline`, `suppression` | incremental adoption and reviewed exceptions |
| `reporter` | terminal, JSON, SARIF, and JUnit output |
| `hook`, `cache`, `release` | hooks, result cache, and release integrity |
| `scripts` | verification and deterministic release tooling |

## Change Design Rules

1. Keep default behavior offline and deterministic.
2. Never read working-tree content during a staged scan unless the public contract explicitly says so and tests prove the boundary.
3. Treat repository files, paths, external output, and manifests as untrusted.
4. Bound file, output, process, snapshot, and cache resources.
5. Never print or persist raw secret-bearing snippets.
6. Preserve schema and exit-code compatibility unless a reviewed migration is part of the change.
7. Add positive, negative, invalid-input, cancellation, and boundary tests in proportion to the behavior changed.

## Adding a Built-In Rule

Choose the domain file under `rules/defaults_<domain>.go`. A rule should have:

- a stable unique ID (prefer `<domain>/<name>` for new public IDs);
- a valid domain, category, severity, and RE2-compatible pattern;
- a description that communicates confidence without claiming proof;
- a concrete recommendation;
- relevant extension filters and tags;
- `Fixable: true` only when an existing deterministic fixer safely supports it.

Add at least one positive and one negative fixture to the matching defaults test. Include redaction coverage for categories/tags involving credentials, secrets, authorization, PII, or other sensitive content. Avoid regexes with broad common tokens that create noisy results.

For repository-level checks (required files, dependencies, headers, ownership), extend the pattern scanner's bounded file-policy path rather than forcing a line-oriented regex.

## Adding a Custom-Rule Field or Schema

`rules.Set` schema is public. Optional additive fields may remain in version 1 only when old files and consumers retain their meaning. Required, removed, or semantically incompatible fields require a schema bump, strict decode tests, golden updates, and a migration entry in `MIGRATIONS.md`.

Custom rule decoding must continue to reject unknown fields, trailing JSON, duplicate IDs, invalid domains/severities, and invalid regular expressions.

## Adding a Scanner

Implement `scanner.Scanner`:

```go
type Scanner interface {
    ID() string
    Scan(context.Context, Request) Result
}
```

Implement `scanner.Described` when the scanner exposes domain, version, capabilities, supported modes, or network requirements. Requirements:

- Return one valid state: clean, findings, partial, failed, or skipped.
- Use a structured `FailureKind` for non-clean operational outcomes.
- Honor context cancellation promptly and close readers/processes.
- Return deterministic finding/status order for identical input.
- Use request `Sources`, `Files`, and `RepositoryFiles` according to their documented scope.
- If `staged` is advertised, add both staged-safe/working-unsafe and staged-unsafe/working-safe tests.
- Do not place source snippets or arbitrary external diagnostics in messages.

Register built-in scanners in `securityreview.New`, preserve the built-in pattern scanner contract, and test enabled, required, timeout, profile, cache, panic, and cancellation behavior where applicable.

## Adding an External Adapter

Add a named preset in `scanner/adapters` and return a `scanner/command.Spec`. Specify:

- executable and argument array (never a shell command string);
- domain, severity, category, and safe generic description;
- exact finding exit codes;
- parser/output format and whether output is read on success;
- network requirement;
- redacted tool flags when available.

Parser rules:

- Accept byte input within the command output limit.
- Never map raw secrets, matches, source excerpts, tokens, or command stderr.
- Normalize paths through command-scanner validation.
- Return deterministic metadata and finding order.
- Add representative fixtures, malformed-input tests, and fuzz coverage.

Adapters should default to `on_missing: skip` only through explicit configuration; security-sensitive required workflows can choose `fail`.

## Configuration Changes

Every new `config.Config` or nested public field needs:

- a JSON tag and deterministic default;
- validation, including duplicates and unsafe paths;
- strict decode and invalid-input tests;
- documentation in `CONFIGURATION.md`;
- compatibility review against `compatibility.Current()`.

Do not silently reinterpret an existing field. See `MIGRATIONS.md` before changing schema versions or defaults that affect policy outcomes.

## Reports and Public Contracts

The following are compatibility surfaces:

- config, report, rule, suppression, and baseline schemas;
- fingerprint algorithm version;
- hook marker version;
- cache key version;
- provenance schema;
- JSON, SARIF, JUnit, checksum, and provenance formats;
- CLI command names and exit codes.

Update golden files only after intentional review. A golden mismatch is not a formatting inconvenience; it signals a consumer-visible change.

## Commit Conventions

Use focused commits with an imperative Conventional Commit-style subject:

```text
feat(scanner): add bounded analyzer
fix(cache): preserve entry permissions
test(release): cover tampered signature
docs(config): document scanner field
ci(security): pin vulnerability scanner
```

Keep unrelated refactors separate from behavior changes. Each numbered roadmap task should remain atomic. Do not amend public contracts, generated artifacts, and implementation in an opaque catch-all commit.

## Pull-Request Checklist

- Scope and threat-boundary impact are described.
- Behavior changes have tests and documentation.
- Staged-mode behavior is tested when relevant.
- Secret redaction and least-privilege output are preserved.
- Public compatibility is unchanged or has an approved migration.
- Package tests, vet, race/release gate as appropriate, and `git diff --check` pass.
- Workflow actions are pinned to full commit SHAs.

## Release Checklist and Ownership

Release maintainers own the final tag and signing operation. Security reviewers own signing-key policy and changes to trust boundaries. Format owners review schema/golden changes. Repository administrators own protected tags, environments, secrets, and required CI checks.

Before creating a tag:

1. Confirm the working tree is clean and the intended commit is reviewed.
2. Run `scripts/release-candidate.sh` with pinned `govulncheck` available.
3. Validate `CHANGELOG.md` and compatibility/migration notes.
4. Confirm performance budgets and reproducibility pass.
5. Confirm protected release secrets/public key are current.
6. Create a semantic `vMAJOR.MINOR.PATCH` tag on the certified commit.
7. Observe the tagged workflow through checksum, provenance generation, signing, verification, and upload.
8. Independently download and verify the uploaded artifact set.

Never bypass a failed release gate by updating golden files, raising resource limits, or relaxing budgets without an explained and reviewed reason.
