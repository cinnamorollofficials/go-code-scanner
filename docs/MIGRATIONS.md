# Compatibility and Migration Guide

## Compatibility policy

Go Code Scanner records public compatibility in
`compatibility/testdata/contract.json`. The manifest contains:

- configuration schema;
- report schema;
- rule schema;
- suppression schema;
- baseline schema;
- fingerprint version;
- hook marker version;
- cache key version;
- provenance schema.

Within a schema version, changes must be backward-compatible. Usually safe
changes include optional fields with stable defaults, new scanner/rule IDs, and
new CLI commands that do not reinterpret existing arguments. Compatibility
tests and golden artifacts must remain unchanged unless a migration is intended.

Potentially breaking changes include:

- removing or renaming a field;
- making an optional field required;
- changing a default that alters scanning or blocking behavior;
- changing severity/domain semantics;
- changing fingerprint identity;
- changing hook ownership markers;
- changing cache key inputs;
- changing JSON/SARIF/JUnit/checksum/provenance structure;
- reassigning an exit code or existing command meaning.

## Checking an upgrade

Print the current contract:

```sh
security-review upgrade check > current-contract.json
```

Compare a previous released contract:

```sh
security-review upgrade check --contract previous-contract.json
```

Exit `0` means the public versions match. Exit `1` reports versioned changes.
Exit `2` means the previous contract is invalid or unreadable. The command uses
strict decoding and rejects unknown/trailing content.

Contract equality does not replace release notes: additive behavior within a
version should still be documented when it changes configuration choices,
findings, performance, or operational requirements.

## Migration workflow for maintainers

When an incompatible change is necessary:

1. Describe consumers and the reason an additive change is insufficient.
2. Choose the smallest affected public version. Do not bump unrelated schemas.
3. Implement old-version reading where practical and write the new version
   explicitly.
4. Add migration tests using fixtures from the previous release.
5. Update the compatibility golden manifest and relevant public-format goldens.
6. Document before/after examples and automated/manual conversion steps here.
7. Add a changelog entry under `Unreleased` with compatibility impact.
8. Run `security-review upgrade check` against the previous release contract.
9. Run the complete release-candidate gate.
10. Use semantic-versioning impact appropriate to the public break.

Do not reuse an old version number for a new meaning.

## Configuration schema

Current schema: `1`.

Version 1 uses strict JSON decoding but supports backward-compatible optional
additions. Loading overlays JSON on deterministic defaults. A future version
must define whether old files are accepted directly or require conversion.

Recommended conversion process for a future schema:

1. Validate the old file with the old release.
2. Back it up outside the project output paths.
3. Convert fields with a dedicated deterministic tool or documented mapping.
4. Validate the new file with `security-review config validate`.
5. Run full, changed, and staged scans and compare policy decisions.
6. Review profile, hook, external command, environment, and path semantics.

The complete current field reference is in `CONFIGURATION.md`.

## Report schema

Current schema: `1.0`.

Consumers should inspect `schema_version`, tolerate optional additive fields,
and avoid assuming map iteration order. Do not parse terminal output as a report
API. JSON golden tests protect field names and representation; SARIF declares
2.1.0; JUnit is validated structurally.

For a future report version, maintainers must document field mappings and
whether dual-output support exists. Consumers should update before relying on
new fields or lifecycle meanings.

## Rule schema

Current schema: `1`.

Custom rule files are strict and merge with defaults. IDs must remain unique.
New rules can create new findings without changing the rule schema; that is a
policy-content change and belongs in release notes. Renaming a rule ID affects
baselines, suppressions, and integrations and requires an alias/migration plan.

## Suppression schema

Current schema: `1`.

Suppressions require file, reason, and expiry. Rule/fingerprint/line narrow the
match; governance may require ticket/approver. Never migrate a baseline into a
suppression automatically: they represent different review decisions.

If rule IDs or fingerprints change, migrate suppression entries with explicit
review and preserve reason, approver, ticket, and expiry.

## Baseline and fingerprint migrations

Current baseline schema: `1`. Current fingerprint version is recorded in the
compatibility manifest and every report/baseline.

Baseline loading rejects fingerprint-version mismatches. If the fingerprint
algorithm changes:

1. Produce a report using the new scanner.
2. Compare old and new findings by rule, path, normalized context, and symbol.
3. Review every entry that cannot be mapped confidently.
4. Create a new baseline from the reviewed report.
5. Do not silently relabel old fingerprints as the new version.

Use `baseline status` and `baseline update --dry-run`; removing resolved entries
requires `--accept-resolved`.

## Hook marker migrations

Managed hooks contain a marker version and exact expected content. Install and
uninstall refuse to modify an unmanaged or unexpected file. A new marker version
must include an explicit upgrade path:

1. Report the old managed state without treating arbitrary content as owned.
2. Require intentional replacement or provide a safe dedicated upgrade command.
3. Preserve existing unmanaged hooks.
4. Test effective `core.hooksPath`, idempotence, execution, and rollback.

Never broaden ownership matching to a loose substring.

## Cache-key migrations

Cache keys include a version plus scanner/content/config/rule identity. A key
version bump should create misses, not reinterpret old entries. Old files may be
left for retention cleanup or removed by `cache clean`; they must never be read
as new-version results.

Cached and uncached findings and summaries must remain equivalent.

## Provenance migrations

Current provenance schema: `go-code-scanner/provenance/v1`.

Verification rejects unknown schemas and fields. A new provenance schema needs:

- a distinct schema identifier;
- canonical deterministic serialization rules;
- subject-name and digest safety rules;
- signing/verifying support and tampering tests;
- documentation for independent verifiers;
- a transition plan for release public keys if signing policy also changes.

Never accept a new document under the v1 identifier by ignoring unknown fields.

## CLI and exit-code migrations

Existing command meanings and exit codes are public. Additive flags should have
defaults that preserve behavior. If a command or exit code must change, provide
a versioned or compatibility alias where possible and document automation
updates.

Current general codes are `0` success, `1` policy/verification mismatch, `2`
invalid input, and `3` operational failure. Individual commands document any
narrower interpretation in `README.md`.

## Release upgrade checklist for consumers

1. Read `CHANGELOG.md` and this file.
2. Download and verify checksums plus signed provenance.
3. Run `upgrade check` against the contract saved from the deployed version.
4. Validate configuration and custom rules.
5. Run cached and uncached scans on representative repositories.
6. Run staged isolation fixtures and hook status in a temporary repository.
7. Review new rule findings and update baselines/suppressions intentionally.
8. Roll out to CI before enforcing local hooks broadly.
9. Retain the previous verified binary and configuration for rollback.

## Current migration notes

There are no required schema migrations for the current release candidate. All
public compatibility versions remain those recorded in
`compatibility/testdata/contract.json`. New release, hardening, and documentation
features were added without changing existing schema meanings.
