# Security and Release Guide

## Security model

Go Code Scanner analyzes repositories that may be untrusted. Its primary goals
are to avoid executing repository content implicitly, keep scans inside the
configured root, prevent staged scans from reading unstaged data, bound resource
use, and avoid leaking sensitive source material through outputs or caches.

It is a review aid, not a proof that software is secure. Pattern rules can
produce false positives and false negatives. External analyzers inherit their
own threat models and should be pinned and reviewed independently.

## Trust boundaries

| Boundary | Trusted input | Untrusted input and controls |
| --- | --- | --- |
| Configuration | A reviewed schema-v1 JSON file | Unknown fields, invalid paths, ambiguous commands, and invalid limits are rejected. |
| Repository | The selected root and Git index identity | File names, file contents, symlinks, Git metadata, and repository size are untrusted. |
| External scanners | Explicit executable plus argument array | Output is bounded and parsed strictly; stderr is not copied into reports. |
| Reports/cache | Tool-owned output paths | Snippets are removed at artifact/cache boundaries; writes are atomic and least-privilege. |
| Release artifacts | A trusted public key and expected release identity | Archives, manifests, provenance, and signatures are untrusted until verification completes. |
| CI secrets | Tagged release workflow and repository administrators | Pull-request workflows never receive release signing secrets. |

Configuration, rules, suppression files, and baselines are policy inputs. Anyone
who can change them can affect what is reported or blocked. Protect changes to
those files with normal code review and ownership controls.

## Repository and path hardening

- Project-relative output, rule, suppression, baseline, and cache paths are
  resolved against the configured root and cannot escape it with `..`.
- Discovery uses `Lstat` and accepts regular files. Working-tree symlinks are
  not followed outside the repository.
- Staged file contents are read from Git blobs. A staged symlink is treated as
  its link payload, not followed as a filesystem path.
- Snapshot extraction validates every archived path, rejects escaping links,
  omits `.git`, and applies maximum file-count and byte limits.
- Cache directories cannot be symlinks. Cache keys have a fixed validated
  format, so callers cannot select arbitrary paths.
- Report, baseline, suppression, archive, provenance, and cache writers use a
  temporary file followed by rename. Sensitive artifacts use mode `0600` where
  the platform supports Unix permissions.
- Release inputs, checksum subjects, and provenance subjects must be regular
  files with safe basename-only manifest names.

Do not configure outputs inside directories writable by less-trusted users.
Filesystem permissions and ACLs outside the project remain the operator's
responsibility.

## Staged isolation

Staged scans discover names with NUL-delimited Git commands and read `:<path>`
objects from the index. External scanners requesting `workspace: "staged"` run
against a temporary index snapshot. The working tree is not copied into it.

Snapshot cleanup runs after success, findings, failure, timeout, and caller
cancellation. Boundary tests assert cleanup after count/size failures and
cancellation. If a machine terminates the process without allowing cleanup, the
operating system's temporary-directory policy is the final recovery mechanism.

## Redaction and sensitive outputs

The pattern scanner classifies secret-like findings from rule category/tags and
replaces sensitive snippets. Additional entropy-like or credential-shaped text
is also redacted, and ordinary snippets are length-bounded internally.

No source snippet is printed in terminal output. JSON removes snippets again at
the reporter boundary; SARIF and JUnit never include them. Cache writes deep-copy
results and clear snippets. Known secret-bearing external fields, such as
Gitleaks `Secret` and `Match`, are intentionally not mapped into findings.

Scanner process stdout/stderr have independent byte limits. Unexpected command
stderr is not included in scanner status or reports; failures use generic exit
and truncation diagnostics. Provenance and checksum errors report artifact names
and digest mismatch, never artifact contents.

Reports can still contain file paths, rule descriptions, metadata from trusted
parsers, and fingerprints. Treat report artifacts as internal security data and
apply access controls appropriate to the repository.

## External command execution

Command scanners use `exec.CommandContext` with an argument array. Configuration
does not pass through a shell, so shell metacharacters in normal arguments do
not gain command semantics.

Controls include:

- The executable is a PATH name or a clean absolute path. Absolute symlinks and
  non-regular/non-executable files are rejected.
- Only a small platform-safe environment plus explicitly allowlisted variable
  names is forwarded. Arbitrary lowercase/invalid environment names are
  rejected.
- The working directory is the root or staged snapshot selected by policy.
- Scanner deadlines use child contexts. On Unix, cancellation terminates the
  scanner process group; platform-specific process handling is isolated.
- `WaitDelay` bounds waiting for inherited descriptors after cancellation.
- stdout, stderr, and adapter output files are byte-limited before parsing.
- Finding exit codes are configured explicitly and are not confused with
  operational failures.
- Network-requiring scanners are skipped in offline profiles.

Enabling a tool authorizes that binary to read its selected workspace and run
with the allowed environment. Install external binaries from trusted sources,
pin versions in CI, and avoid passing secrets through their allowed variables.

## Resource limits and denial of service

The pattern scanner bounds file and line bytes. Staged snapshots bound file
count and aggregate bytes. External output is bounded. Scanner concurrency,
timeouts, and cache retention are configurable. The release-candidate suite
tests oversized files, excessive counts, output truncation, timeout,
cancellation, cleanup, and exact cache-retention boundaries.

Defaults favor ordinary repositories, not adversarial multi-gigabyte inputs.
Increase limits only after evaluating memory, disk, and CI-time impact.

## Signing-key handling

Release provenance uses Ed25519 detached signatures. Private keys must be PKCS#8
PEM files whose permission bits grant no group or other access. The signing API
rejects broader permissions before reading key material. Public keys use PKIX
PEM.

The tagged GitHub Actions workflow expects two protected secrets:

- `RELEASE_SIGNING_KEY`: Ed25519 PKCS#8 private-key PEM.
- `RELEASE_SIGNING_PUBLIC_KEY`: corresponding PKIX public-key PEM.

The workflow is triggered only by tags, has `contents: read`, does not run on
pull requests, applies `umask 077`, writes keys only under `RUNNER_TEMP`, verifies
the produced signature, and removes both temporary key files with an exit trap.
Checkout credentials do not persist.

Recommended operational controls:

1. Generate and store the private key in a dedicated secret manager.
2. Restrict who can create release tags and modify the release environment.
3. Protect workflow files with code owners and required review.
4. Publish the public key through a separately authenticated channel.
5. Rotate the key after suspected exposure and document which releases use each
   public key. Never commit a private key to the repository.

## Verifying a release

Obtain the archive, `SHA256SUMS`, `provenance.json`, `provenance.sig`, and trusted
public key. Keep them in one directory.

First verify the checksum manifest:

```sh
security-review release checksums verify \
  --manifest SHA256SUMS \
  --directory .
```

Then verify both provenance signature and local subjects:

```sh
security-review release verify \
  --provenance provenance.json \
  --signature provenance.sig \
  --public-key release-public-key.pem \
  --directory .
```

Successful signature verification proves that the holder of the private key
signed the exact provenance bytes. Subject verification proves that local
artifacts match the signed digests. Also inspect provenance `version`, `commit`,
`build_date`, and `builder` against the intended tag and repository.

Exit `1` means a verification mismatch. Exit `2` means invalid arguments,
manifest/provenance format, key format, or unreadable input. Do not execute an
artifact after either result.

## Reproducible builds

Release builds use:

- explicit version, commit, and UTC build date;
- `CGO_ENABLED=0` for target builds;
- `-trimpath` and `-buildvcs=false`;
- deterministic linker metadata;
- a temporary raw-binary directory outside the distribution set;
- explicit archive timestamps;
- normalized tar, gzip, and ZIP metadata;
- stable archive names and sorted checksum/provenance subjects.

`TestReleaseBuildIsReproducible` builds the same target twice with identical
metadata and requires equal archive SHA-256 digests. Reproducibility does not
replace signature verification: both properties are checked.

## Tagged release process

`.github/workflows/release.yml` performs these steps:

1. Check out the tag with credentials disabled and set up Go using actions
   pinned to full commit SHAs.
2. Validate `CHANGELOG.md`.
3. Resolve the commit timestamp and normalize it to UTC.
4. Cross-build Linux, macOS, and Windows targets and package only deterministic
   `.tar.gz`/`.zip` archives.
5. Generate and verify `SHA256SUMS`.
6. Generate provenance containing version, commit, build time, builder identity,
   and every distributable archive; generation validates subjects immediately.
7. Materialize protected signing secrets with restrictive permissions, sign
   provenance, and verify signature plus subjects.
8. Upload only archives, checksum manifest, provenance, and signature using a
   full-SHA-pinned action, explicit retention, and no extra repository write
   permission.

Before tagging, run:

```sh
GOCACHE=/tmp/go-code-scanner-gocache ./scripts/release-candidate.sh
```

The gate includes unit/integration/E2E tests, race, vet, diff checks, cached and
uncached equivalence, release reproducibility, golden contracts, fuzz smoke,
vulnerability scanning when the pinned tool is available, performance budgets,
and self-scan.

## Reporting vulnerabilities

Do not publish suspected vulnerabilities, private keys, tokens, or sensitive
repository content in a public issue. Contact the repository maintainers through
the private security-reporting channel configured by the hosting organization.
Include the affected version/commit, reproduction steps using non-sensitive
fixtures, impact, and any proposed mitigation.
