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
