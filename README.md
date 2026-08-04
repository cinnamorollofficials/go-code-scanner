# Go Code Scanner

Go Code Scanner is a policy-driven commit gate and static review orchestrator.
It combines fast built-in checks with optional external analyzers, evaluates
findings across six domains, and produces deterministic reports for local
development, Git hooks, and CI.

The project is both a Go library and the `security-review` CLI. It uses only the
Go standard library at runtime; external scanners are opt-in.

## What it checks

Every finding belongs to one of six domains:

| Domain | Examples |
| --- | --- |
| Quality | formatting, conflict markers, debug statements, generated files |
| Reliability | ignored errors, missing timeouts, unbounded reads, unsafe retries |
| Hardening | insecure TLS, permissive modes, wildcard CORS, root containers |
| Security | secrets, injection, path traversal, weak cryptography |
| Supply chain | vulnerable dependencies, mutable actions/images, license policy |
| Governance | commit policy, required files, ownership, architecture boundaries |

Built-in pattern, file-policy, and architecture scanners work offline. Optional
adapters are available for `gofmt`, `go vet`, `go test`, `govulncheck`, `gosec`,
Gitleaks, Trivy, OSV-Scanner, and Semgrep.

## Requirements and installation

- Go 1.25 or newer when building from source.
- Git for changed/staged scans and hook management.
- Optional analyzer binaries only when their scanners are enabled.

Install with Go:

```sh
go install github.com/cinnamorollofficials/go-code-scanner/cmd/security-review@latest
```

Or build a local binary:

```sh
go build -trimpath -o security-review ./cmd/security-review
./security-review version
```

Tagged releases contain `.tar.gz` archives for Unix targets and `.zip` archives
for Windows, plus checksums and signed provenance. See
[`SECURITY.md`](SECURITY.md) for independent verification.

## Quick start

Scan the current directory and write `security_findings.json`:

```sh
security-review scan
```

Run the fast staged commit gate:

```sh
git add .
security-review scan --staged --profile fast --ci --new-only
```

Scan a configured project and produce SARIF:

```sh
security-review config validate security-review.json
security-review scan --config security-review.json --format sarif --output artifacts/security.sarif --ci
```

The CLI also accepts scan flags without the `scan` subcommand for backward
compatibility.

## Scan modes and staged isolation

- `full` walks the working tree.
- `--changed` reads files changed from `HEAD`; in a repository without `HEAD`,
  it uses the index.
- `--staged` discovers and reads blobs from the Git index.

Staged mode never substitutes an unstaged working-tree version. External tools
configured with `workspace: "staged"` run inside a temporary materialization of
the index that excludes `.git` and unstaged content. Snapshots have configurable
file/byte limits and are removed after success, failure, timeout, or cancellation.

`--changed` and `--staged` are mutually exclusive.

## Profiles and policy

Profiles select scanners:

- `fast`: offline built-in checks suitable for pre-commit.
- `standard`: broader checks suitable for pre-push.
- `full`: the complete configured CI set.
- `frontend`: offline native browser client checks and optional frontend adapters (`tsc`, `biome`, `eslint`, `semgrep`).

Scan scope can be selected via `--scope client|server|all` (default `all`).

`offline_profiles` prevents scanners marked `requires_network` from running in
those profiles. Default fast and frontend profiles are offline.

Policy thresholds can be global (`fail_on`) or per domain (`policy`). A finding
blocks `--ci` when its severity reaches the relevant threshold. `--fail-on`
overrides domain policy for that invocation. Suppressed findings never block.
With `--new-only`, findings already present in the baseline do not block.

## Common scan options

| Option | Meaning |
| --- | --- |
| `--config <file>` | Load strict JSON configuration. |
| `--root <dir>` | Project root when no config file is used. |
| `--changed`, `--staged` | Select Git-aware input mode. |
| `--profile <name>` | Select a configured scanner profile. |
| `--ci` | Return exit 1 for a policy violation. |
| `--fail-on <severity>` | Global threshold override. |
| `--baseline <file>` | Baseline used for lifecycle comparison. |
| `--new-only` | Apply policy only to findings absent from the baseline. |
| `--format json\|sarif\|junit` | Artifact format. |
| `--output <file>` | Artifact path relative to project root unless absolute. |
| `--quiet` | Suppress terminal summary. |
| `--verbose` | Include scanner metadata and timing in terminal output. |
| `--color auto\|always\|never` | Control terminal color; `NO_COLOR` is honored. |
| `--explain <rule-id>` | Print configured rule metadata and exit. |
| `--fix` | Apply supported deterministic fixes, then rescan. |
| `--dry-run` | Preview `--fix` without writing. |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Command succeeded; scan policy allowed the result. |
| `1` | Policy mismatch, checksum/signature mismatch, or validation finding. |
| `2` | Invalid arguments/configuration or unreadable verification input. |
| `3` | Operational scan, hook, discovery, or report-writing failure. |

Scan findings return 1 only when `--ci` is enabled. A report may still be
available after an operational scanner error.

## Git hooks

Supported events are `pre-commit`, `commit-msg`, and `pre-push`:

```sh
security-review hook install pre-commit --root .
security-review hook install commit-msg --root .
security-review hook install pre-push --root .

security-review hook status pre-commit --root .
security-review hook uninstall pre-commit --root .
```

Hooks use Git's effective `core.hooksPath`. Installation is idempotent and
refuses to replace an unmanaged file. Uninstall removes only an exact managed
hook. The default pre-commit hook uses staged content, the fast profile, and
new-only policy. Commit-message and pre-push behavior is configured under
`hooks`.

## Reports

JSON uses report schema `1.0`. SARIF uses version 2.1.0 and JUnit separates
policy findings from operational scanner errors. Public formats have golden or
strict structure tests.

The terminal shows domain, severity, rule, location, recommendation, explain,
fix, and suppression hints. It never prints source snippets. Artifact writers
also remove snippets at the output boundary and use atomic replacement with
mode `0600` where supported.

## Baselines

Create a baseline from a JSON report:

```sh
security-review baseline create \
  --report security_findings.json \
  --baseline .security-baseline.json
```

Inspect or update it:

```sh
security-review baseline status --report security_findings.json --baseline .security-baseline.json
security-review baseline update --report security_findings.json --baseline .security-baseline.json --dry-run
security-review baseline update --report security_findings.json --baseline .security-baseline.json --accept-resolved
```

An update that removes resolved entries requires `--accept-resolved`. Baselines
are versioned, strictly decoded, and atomically replaced. They are not
suppressions: they classify debt as new, existing, or resolved.

## Suppressions

Add a reviewed, expiring suppression:

```sh
security-review suppress add \
  --suppression-file .security-ignore \
  --file internal/store.go \
  --rule security/sql-injection \
  --fingerprint '<report fingerprint>' \
  --reason 'Reviewed false positive' \
  --expires 2026-12-31 \
  --ticket SEC-123 \
  --approved-by security-team
```

Use `--dry-run` to preview. Every suppression requires a file, reason, and
expiry. Governance policy can require a ticket and approver for selected rules.
Expired entries stop suppressing findings and are reported as stale.

## Cache

Enable the content-addressed scanner cache in configuration. Cache keys include
content, scanner version, configuration hash, and rule-set hash. Cache entries
are atomically written, locked across processes, pruned by age/size, and never
store source snippets.

```sh
security-review cache stats --dir .go-code-scanner-cache
security-review cache clean --dir .go-code-scanner-cache
```

## Release verification

Verify downloaded checksums:

```sh
security-review release checksums verify \
  --manifest SHA256SUMS \
  --directory .
```

Verify the detached Ed25519 signature and every local provenance subject:

```sh
security-review release verify \
  --provenance provenance.json \
  --signature provenance.sig \
  --public-key release-public-key.pem \
  --directory .
```

Omit `--directory` for signature-only compatibility. Release-maintainer commands
also support deterministic archive creation, provenance generation/signing, and
changelog validation.

## Other commands

```text
security-review config validate <file>
security-review upgrade check [--contract previous-contract.json]
security-review release changelog validate [--file CHANGELOG.md]
security-review version
security-review help
```

`upgrade check` prints the current compatibility manifest when no previous
contract is provided, or reports public-contract changes against one.

## Architecture

```text
configuration
    -> filesystem/Git discovery
    -> working tree or staged snapshot
    -> built-in and optional external scanners
    -> normalization, fingerprinting, and deduplication
    -> suppression and baseline lifecycle
    -> report artifacts and terminal output
    -> domain policy decision
```

Important packages include `config`, `discovery`, `workspace`, `scanner`,
`rules`, `baseline`, `suppression`, `policy`, `reporter`, `hook`, `cache`, and
`release`.

## Library use

```go
cfg := config.Default()
cfg.Root = "/path/to/project"

reviewer, err := securityreview.New(cfg)
if err != nil {
    return err
}
report, runErr := reviewer.Run(context.Background())
// report can be non-nil when runErr describes an operational failure.
```

Add custom implementations through `securityreview.WithScanner` or
`securityreview.WithRequiredScanner`.

## Troubleshooting

- **A staged scan differs from the editor:** run `git diff --cached`; staged
  mode intentionally ignores unstaged edits.
- **An optional scanner is skipped:** install its binary, check `PATH`, and make
  sure the selected profile is not offline when the scanner requires network.
- **Hook installation reports a conflict:** inspect the effective hooks path
  with `git config --get core.hooksPath`; unmanaged hooks are never overwritten.
- **A scan times out or truncates output:** raise the scanner `timeout` or
  `max_output_bytes` after reviewing resource risk.
- **A baseline update refuses resolved entries:** review the status/dry-run and
  pass `--accept-resolved` intentionally.
- **A release verification returns 1:** treat the artifact as untrusted; fetch
  the archive, checksum manifest, provenance, signature, and public key again
  from authoritative sources.

For every configuration field, see [`CONFIGURATION.md`](CONFIGURATION.md). For
threat boundaries and release operations, see [`SECURITY.md`](SECURITY.md). For
development and upgrades, see [`CONTRIBUTING.md`](CONTRIBUTING.md) and
[`MIGRATIONS.md`](MIGRATIONS.md).
