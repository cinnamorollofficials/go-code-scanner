# SentraSec Agent Skill Implementation Plan

## Objective

Build a distributable `implement-sentrasec` skill that enables an AI coding agent to inspect a repository, choose the correct SentraSec integration path, implement the smallest safe change, verify it, and report any unverified assumptions.

The skill is considered ready when it can complete representative integration and troubleshooting tasks without inventing API details, exposing secrets, weakening security policy, or claiming success without evidence.

For advanced code and SQL vulnerability analysis, use the companion architecture and backlog in [`advanced-code-sql-analysis-plan.md`](advanced-code-sql-analysis-plan.md). Its deterministic scanner, evidence model, and quality gates are prerequisites before the skill may claim advanced analysis capability.

### Current codebase status & constraints

- The workspace contains a fully functional Go CLI tool (`security-review` under `cmd/security-review`) and scanner packages under `pkg/`.
- Built-in pattern scanning, external adapters (`gosec`, `semgrep`, `trivy`, `gofmt`, `govulncheck`), staged Git index isolation, baselines, suppressions, and SARIF/JSON reporting are already implemented.
- The AI Agent Skill will be located under `.agents/skills/go-code-scanner/` (and alias `implement-sentrasec`) to wrap and interface directly with the `security-review` CLI.
- Helper scripts (`detect-project.ps1`, `validate-config.ps1`, `verify-integration.ps1`) will invoke native CLI commands (such as `security-review config validate` and `security-review scan`).

## Scope

### MVP

- Detect the target project's language, framework, package manager, and CI provider.
- Wrap `security-review` CLI commands to perform offline scanning, staged scans (`--staged`), and baseline comparisons (`--baseline`, `--new-only`).
- Validate configuration via `security-review config validate` without displaying secret values.
- Support CI integration for GitHub Actions / primary CI provider.
- Diagnose common authentication, configuration, network, and scan failures.
- Produce an evidence-based handoff report.

### Later releases

- Additional frameworks and CI providers.
- Version migration workflows (`security-review upgrade check`).
- Automated remediation (`security-review scan --fix`) for explicitly supported finding types.
- Offline fixtures and broader cross-platform support.
- Telemetry for anonymous skill-quality measurement, only if product policy permits it.
- Evidence-backed code and SQL vulnerability analysis using the staged rule engine defined in [`advanced-code-sql-analysis-plan.md`](advanced-code-sql-analysis-plan.md).

### Explicitly out of scope for MVP

- Changing production settings or external permissions automatically.
- Creating or rotating real credentials.
- Disabling security rules to make CI pass.
- Automatically classifying findings as false positives.
- Uploading source code unless the documented product contract requires it and the user has authorized it.

## Deliverable structure

```text
.agents/skills/go-code-scanner/
|-- SKILL.md
|-- agents/
|   `-- openai.yaml
|-- references/
|   |-- integration-contract.md
|   |-- configuration.md
|   |-- frameworks.md
|   |-- threat-model.md
|   |-- troubleshooting.md
|   `-- examples.md
`-- scripts/
    |-- detect-project.ps1
    |-- validate-config.ps1
    `-- verify-integration.ps1
```

Do not add a README, changelog, quick-reference file, or duplicate user-facing documentation inside the skill.

## Execution plan

### Phase 0 - Acquire the source of truth

Goal: eliminate undocumented assumptions before writing implementation guidance.

- [ ] Add the documentation source or provide its public URL.
- [ ] Add or identify the SDK/CLI repository and supported version.
- [ ] Identify the primary framework and CI provider for the MVP.
- [ ] Identify the canonical package names and installation commands.
- [ ] Record authentication methods and exact secret names.
- [ ] Record configuration schema, defaults, precedence, and validation rules.
- [ ] Record CLI commands, flags, output formats, and exit codes.
- [ ] Record network behavior, data uploaded, retention, redaction, retry, and timeout rules.
- [ ] Obtain at least one known-good integration repository or fixture.
- [ ] Assign a product/security reviewer who can approve the extracted contract.

Exit criteria:

- No placeholder remains for package names, commands, environment variables, or endpoints.
- The extracted contract is approved by the product or SDK owner.

### Phase 1 - Define behavior and evaluation cases

Goal: describe observable skill behavior before building it.

- [ ] Write 8-12 realistic user prompts covering installation, CI, diagnosis, migration, read-only review, and unsafe requests.
- [ ] For every prompt, define allowed mutations and expected evidence.
- [ ] Define the expected refusal behavior for secret exposure, rule disabling, and unauthorized production changes.
- [ ] Define a standard final-report format: changes, verification, assumptions, and user actions.
- [ ] Create sanitized fixture repositories for the supported framework and failure cases.
- [ ] Establish an MVP scorecard.

Minimum scorecard:

| Metric | MVP target |
| --- | ---: |
| Correct trigger selection | 100% of evaluation prompts |
| Successful primary-framework integration | 3/3 clean fixtures |
| Secret leakage | 0 occurrences |
| Invented package/config/API values | 0 occurrences |
| Unsafe policy weakening | 0 occurrences |
| False success claims | 0 occurrences |
| Useful diagnosis of known failures | At least 90% |

Exit criteria:

- Each evaluation prompt has expected outcomes and prohibited actions.
- Fixtures contain no real credentials or customer data.

### Phase 2 - Build deterministic resources first

Goal: move fragile detection and validation away from free-form agent reasoning.

#### `detect-project.ps1`

- [ ] Detect language and framework from manifests and source structure.
- [ ] Detect package manager from lockfiles.
- [ ] Detect supported CI providers from workflow files.
- [ ] Detect existing SentraSec dependencies and configuration.
- [ ] Return stable JSON with `status`, `detections`, `evidence`, and `warnings`.
- [ ] Remain read-only.
- [ ] Add tests for empty, supported, ambiguous, and already-integrated repositories.

#### `validate-config.ps1`

- [ ] Validate required fields and types against the approved schema.
- [ ] Validate referenced paths without reading secret values.
- [ ] Detect literal credentials and unsafe placeholders.
- [ ] Detect contradictory or deprecated options.
- [ ] Return stable JSON diagnostics with remediation messages.
- [ ] Add valid and invalid configuration fixtures.

#### `verify-integration.ps1`

- [ ] Check installed package or CLI versions.
- [ ] Run configuration validation.
- [ ] Run documented dry-run or local scan when available.
- [ ] Make network checks opt-in and clearly label them.
- [ ] Preserve upstream exit codes or map them in a documented way.
- [ ] Distinguish `passed`, `failed`, `blocked`, and `not_verified`.
- [ ] Add tests for authentication, timeout, malformed output, and unavailable CLI failures.

Exit criteria:

- All scripts run on a clean fixture without mutation.
- Outputs are machine-readable and contain no secret values.
- Automated tests cover success and known failure modes.

### Phase 3 - Write progressive references

Goal: supply exact product knowledge without overloading the skill body.

- [ ] Write `integration-contract.md` from approved API/CLI documentation.
- [ ] Write `configuration.md` with field types, defaults, precedence, sensitivity, and examples.
- [ ] Write `frameworks.md` as a framework-to-files-to-verification matrix.
- [ ] Write `threat-model.md` with trust boundaries, data flow, authorization boundaries, and prohibited operations.
- [ ] Write `troubleshooting.md` in symptom -> checks -> cause -> remediation form.
- [ ] Write `examples.md` with complete, tested examples only.
- [ ] Add a table of contents to any reference longer than 100 lines.
- [ ] Ensure detailed facts appear in one reference only; link instead of duplicating.

Exit criteria:

- Every factual command and configuration example is traceable to the approved contract or a passing fixture.
- `SKILL.md` can point directly to each reference without nested reference chains.

### Phase 4 - Implement the skill workflow

Goal: teach the agent when to read references, when to run scripts, and when to stop.

- [ ] Initialize the skill with the official skill initialization script.
- [ ] Write frontmatter containing only `name` and a comprehensive trigger `description`.
- [ ] Keep `SKILL.md` procedural and below 500 lines.
- [ ] Require repository-instruction and dirty-worktree inspection before mutation.
- [ ] Route the agent based on task type: install, CI, diagnose, migrate, remediate, or review.
- [ ] Require a brief change plan before editing.
- [ ] Require the smallest change consistent with repository conventions.
- [ ] Require relevant test, lint, type-check, build, and SentraSec verification.
- [ ] Require explicit `not_verified` reporting when verification is unavailable.
- [ ] Require approval before credentials, production state, permissions, or external uploads are involved.
- [ ] Reject requests to expose secrets or weaken security policy merely to pass CI.
- [ ] Define the final handoff format.
- [ ] Implement separate read-only review and authorized-remediation state machines.
- [ ] Require scan-plan preview, rulepack provenance, coverage review, and evidence gating before confirming findings.
- [ ] Require targeted rescan plus regression scan before marking a remediation verified.
- [ ] Treat scanner failure, partial parse coverage, unsupported dialects, and stale finding hashes as explicit non-clean states.
- [ ] Generate `agents/openai.yaml` from the completed skill content.

Exit criteria:

- The skill triggers on all positive evaluation prompts and not on unrelated tasks.
- An agent can locate the appropriate reference and script without guessing.

### Phase 5 - Validate and forward-test

Goal: prove the skill generalizes beyond its authoring context.

- [ ] Run the official skill validator and fix all errors.
- [ ] Run script unit and fixture tests.
- [ ] Forward-test each evaluation prompt in a fresh agent context.
- [ ] Give test agents only the skill, fixture, and realistic user request; do not reveal expected answers.
- [ ] Capture raw outputs, diffs, logs, and failures.
- [ ] Check every run against the scorecard.
- [ ] Fix ambiguous instructions or missing deterministic checks.
- [ ] Repeat failed scenarios from a clean fixture.
- [ ] Ask the product/security reviewer for final approval.

Exit criteria:

- All safety metrics pass with zero violations.
- Functional metrics meet the MVP targets.
- The approved primary integration works from a clean fixture.

### Phase 6 - Package and maintain

Goal: make releases reproducible and keep the skill aligned with the product.

- [ ] Choose the distribution location: repository-owned skill, personal skill, or plugin package.
- [ ] Add CI checks for skill validation, script tests, fixture tests, and secret scanning.
- [ ] Pin the compatible SentraSec product/SDK range in the contract reference.
- [ ] Define an owner for product releases and an owner for security review.
- [ ] Re-run evaluation when configuration, CLI behavior, authentication, or supported frameworks change.
- [ ] Track skill version compatibility without placing a changelog inside the skill folder.

Exit criteria:

- Installation and discovery work in a clean Codex environment.
- Maintenance ownership and release triggers are documented outside the skill package.

## Prioritized TODO backlog

### P0 - Blocks implementation

- [ ] Provide documentation source or URL.
- [ ] Provide SDK/CLI source and supported version.
- [ ] Select one MVP framework.
- [ ] Select one MVP CI provider.
- [ ] Approve the authentication and data-flow contract.
- [ ] Provide or create one known-good sanitized fixture.

### P1 - Required for MVP

- [ ] Extract and approve the integration contract.
- [ ] Create evaluation prompts and expected outcomes.
- [ ] Implement and test `detect-project.ps1`.
- [ ] Implement and test `validate-config.ps1`.
- [ ] Implement and test `verify-integration.ps1`.
- [ ] Create the six reference files.
- [ ] Create `SKILL.md` and `agents/openai.yaml`.
- [ ] Validate the skill package.
- [ ] Run clean-context forward tests.
- [ ] Complete product and security review.

### P2 - After MVP

- [ ] Add the second framework.
- [ ] Add the second CI provider.
- [ ] Add version migration support.
- [ ] Add supported remediation recipes.
- [ ] Add Linux/macOS script equivalents if required by target users.
- [ ] Add regression cases from real agent failures.
- [ ] Integrate normalized JSON/SARIF findings from the advanced code/SQL analyzer.
- [ ] Add the SQL-analysis workflow and evidence requirements to `SKILL.md`.

## Recommended ownership

| Workstream | Accountable role |
| --- | --- |
| API, SDK, and configuration facts | Product/SDK owner |
| Threat model and guardrails | Security owner |
| Scripts and fixtures | Tooling engineer |
| Skill workflow and references | Agent/AI engineer |
| Forward-test acceptance | Product + security reviewers |

## Suggested delivery sequence

Assuming one primary implementer and timely product review:

| Stage | Indicative effort | Dependency |
| --- | ---: | --- |
| Contract extraction and approval | 1-2 days | Documentation and SDK access |
| Evaluation cases and fixtures | 1-2 days | Approved primary framework |
| Scripts and tests | 2-4 days | Approved contract and fixtures |
| References and `SKILL.md` | 1-2 days | Scripts and contract |
| Forward-testing and fixes | 2-3 days | Complete candidate skill |
| Packaging and release checks | 1 day | Acceptance approval |

These are planning estimates, not commitments; missing or contradictory product documentation will extend Phase 0.

## Definition of done

The MVP is done only when all of the following are true:

- [ ] The skill package passes the official validator.
- [ ] All included scripts have passing automated tests.
- [ ] The primary framework and CI integration succeed from clean fixtures.
- [ ] Every package, command, configuration field, and environment variable is sourced and tested.
- [ ] No forward test exposes secrets or weakens security controls.
- [ ] Failed or unavailable verification is reported as `failed`, `blocked`, or `not_verified`, never as success.
- [ ] The skill handles install, CI, diagnosis, read-only review, and unsafe-request scenarios.
- [ ] Product and security owners approve the release candidate.
- [ ] A maintenance owner and compatibility policy are assigned.
