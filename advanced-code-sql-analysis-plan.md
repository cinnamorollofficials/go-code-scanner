# Advanced Code and SQL Analysis Plan

## Outcome

Build an advanced, evidence-producing analysis subsystem for bugs and SQL-related vulnerabilities, then expose it safely through the future `implement-sentrasec` agent skill.

This plan does not claim that the scanner already exists. It defines the architecture, rule contract, safety boundaries, quality gates, and prioritized implementation backlog required to build it.

## Product principles

1. Produce a finding only when it can cite reproducible code evidence.
2. Separate deterministic detection from AI-assisted interpretation.
3. Keep severity, confidence, exploitability, and remediation confidence as separate values.
4. Prefer data-flow and semantic evidence over text matching.
5. Treat parameterization as context-specific; escaping is not a universal sanitizer.
6. Never weaken a rule merely to make CI pass.
7. Allow suppression only with ownership, reason, and expiry.
8. Scan untrusted repositories without executing their code by default.

## Target architecture

```text
Repository
  -> inventory and trust-boundary discovery
  -> language/framework adapters
  -> AST + control-flow + call graph
  -> normalized security IR
  -> SQL template reconstruction + dialect parser
  -> rule engine
       Tier 1: syntax and structural rules
       Tier 2: intraprocedural data flow
       Tier 3: interprocedural/context-sensitive taint
       Tier 4: reachability, authorization, and configuration context
  -> finding normalizer and deduplicator
  -> confidence/severity policy
  -> JSON/SARIF output
  -> AI validation and explanation layer
  -> human review, CI gate, or authorized remediation
```

### Component boundaries

#### 1. Repository inventory

Detect without executing project code:

- Languages, frameworks, ORM/database libraries, and versions.
- Package managers and lockfiles.
- HTTP, RPC, GraphQL, message-queue, CLI, and scheduled-job entry points.
- Database clients and raw-query escape hatches.
- Schema, migrations, generated code, tests, vendored code, and build output.
- Multi-tenant indicators, authorization middleware, and data-classification hints.
- CI provider and existing security configuration.

Return evidence for each detection, including the file and manifest field that caused it.

#### 2. Language adapters

Each adapter must provide:

- Parsed AST with stable source locations.
- Control-flow graph.
- Symbol resolution and type information when safely available.
- Call graph with explicit unknown/dynamic-call edges.
- Constant and string-template evaluation.
- Basic alias, field, collection, and return-value propagation.
- Framework models for request sources, database sinks, and safe bind APIs.

Do not silently treat unresolved calls as safe. Record an analysis-gap diagnostic.

#### 3. Normalized security IR

Represent at least:

- `source`: untrusted or privilege-sensitive data origin.
- `propagator`: assignment, return, field, collection, template, or builder flow.
- `transform`: validation, normalization, escaping, or encoding operation.
- `sanitizer`: a context-specific operation proven safe for a specific sink position.
- `sink`: query execution, dynamic SQL construction, logging, error output, or privileged data operation.
- `barrier`: authorization, tenant scoping, trusted constant, or proven allow-list.
- `unknown`: unresolved semantic behavior that lowers confidence but does not imply safety.

Every taint step must retain a source location and a short semantic explanation.

#### 4. SQL template reconstruction and dialect parsing

Reconstruct query state instead of treating every query as an opaque string:

```text
Const("SELECT * FROM users WHERE id = ")
Hole(source=request.id, trust=untrusted, context=value)
BoundParam(name=tenantId, context=value)
```

After constant folding, parse the reconstructed template with the detected SQL dialect. Classify every dynamic hole as `value`, `identifier`, `table`, `column`, `order-direction`, `operator`, `keyword`, `clause`, `list-expansion`, or `unknown`.

Model prepared-query state explicitly:

```text
raw template -> prepared template -> parameters bound -> executed
```

Binding a value after tainted text has already entered the prepared template does not remove that taint. Parameter binding is a barrier only for a value-position hole supported by the exact driver/dialect model.

Use function summaries for cross-call analysis. A summary must capture argument/receiver/global taint, return taint, field writes, query-builder mutation, persistent writes, and emitted sink events. Prefer call-site-sensitive, field-sensitive, type-aware analysis; use branch sensitivity for finite allow-lists. Record unresolved dynamic dispatch as a coverage gap.

## Detection tiers

| Tier | Analysis | Example | Default CI behavior |
| --- | --- | --- | --- |
| 1 | Syntax/structure | Hardcoded DB password, raw query API | Informational or warning |
| 2 | Local data flow | Request parameter concatenated into a query in one function | Warning; block only at high precision |
| 3 | Interprocedural taint | Input flows through helpers/DTOs into a query sink | Candidate for blocking |
| 4 | Context and reachability | Internet-reachable source, missing tenant boundary, destructive DB privilege | Risk-based blocking |

Tier 1 must not claim exploitability when only a dangerous API is present. Tier 4 must retain the complete evidence inherited from lower tiers.

## Advanced SQL rule taxonomy

### A. Injection and query construction

#### SQLI-001 - Untrusted value reaches an executable SQL string

- Sources: request parameters, headers, cookies, path values, RPC/GraphQL arguments, queue messages, CLI input, uploaded files, and data previously stored from an untrusted source.
- Propagators: concatenation, interpolation, formatting, join operations, string builders, helper returns, DTO/object fields, collections, serialization, and ORM expression wrappers.
- Sinks: driver query/execute methods, raw ORM methods, stored-procedure invocation with dynamic SQL, and migration/runtime execution helpers.
- Safe case: a documented driver bind mechanism where the tainted value remains a value parameter through execution.
- Unsafe case: interpolation occurs before a nominally parameterized API is called.

#### SQLI-002 - Untrusted identifier or SQL fragment

Detect taint in table names, column names, direction keywords, operators, clauses, and expressions. Bind parameters generally do not protect identifier positions. Require an exact allow-list mapping from external values to trusted constants.

#### SQLI-003 - Dynamic `ORDER BY`, `GROUP BY`, `LIMIT`, or pagination clause

Detect both direct interpolation and query-builder APIs that accept raw fragments. Distinguish numeric bounds validated as integers from arbitrary SQL fragments.

#### SQLI-004 - Unsafe raw ORM escape hatch

Model raw/unsafe APIs for each supported ORM. Presence alone is a Tier 1 warning; promote only when tainted data or an unsafe fragment reaches it.

#### SQLI-005 - Second-order SQL injection

Track untrusted data written to persistent storage and later read into dynamic SQL construction. Mark incomplete storage modeling explicitly and require higher evidence before blocking.

#### SQLI-006 - Dynamic SQL inside stored procedures

Detect `EXEC`, `EXECUTE`, or equivalent dynamic SQL assembled from procedure inputs. Do not assume a stored procedure is safe merely because it is parameterized at the application boundary.

#### SQLI-007 - Multi-statement or stacked-query exposure

Detect enabling flags and execution paths that permit multiple statements with untrusted construction. Raise severity when destructive statements or high-privilege connections are reachable.

#### SQLI-008 - Placeholder/bind mismatch

Detect missing, extra, incorrectly ordered, or incorrectly expanded bind values where statically provable. Include framework-specific list/array expansion behavior.

#### SQLI-009 - Context-inappropriate escaping

Detect generic HTML/URL/shell escaping used as an alleged SQL sanitizer, manual quote replacement, or database escaping applied for the wrong connection/encoding/context.

#### SQLI-010 - Unsafe pattern construction

Detect untrusted `LIKE`/regex patterns when wildcard control creates an authorization or data-exposure problem. Classify wildcard-only behavior as a correctness/authorization issue unless it can alter SQL syntax.

#### SQLI-011 - Unsafe list or `IN` expansion

Detect arrays joined into SQL text, raw list fragments, and framework helpers that interpolate list values. Accept only driver/ORM-supported list binding or an exact constant allow-list appropriate to the dialect.

#### SQLI-012 - Tainted prepared-query template

Detect query text that becomes tainted before statement preparation even when later values use a binding API. Report the prepared-statement state transition as evidence.

### B. Authorization and data isolation

#### SQLAUTH-001 - Missing tenant constraint

For declared multi-tenant models, detect read/update/delete queries lacking a tenant predicate or a proven repository-level tenant scope. Require a project configuration that identifies tenant keys and approved scoping abstractions.

#### SQLAUTH-002 - User-controlled object lookup without ownership scope

Detect direct lookup by external ID when the reachable path lacks an authorization barrier or ownership/tenant predicate. Treat this as a cross-layer rule requiring route-to-query evidence.

#### SQLAUTH-003 - Authorization filter dropped during raw-query fallback

Detect code paths where a scoped ORM query is replaced or supplemented by a raw query that does not preserve authorization predicates.

#### SQLAUTH-004 - Row-level security assumption mismatch

Detect application paths that assume database row-level security while using a connection role or session state that does not enforce it. Enable only when deployment/database policy inputs are available.

### C. Destructive and integrity bugs

#### SQLSAFE-001 - Unbounded update or delete

Detect missing or trivially true predicates, ORM bulk operations without a filter, and control-flow paths that can produce an empty filter.

#### SQLSAFE-002 - Accidental mass assignment to persistence model

Detect untrusted objects spread or bound into create/update operations without a field allow-list, especially privilege, tenant, ownership, or lifecycle fields.

#### SQLSAFE-003 - Non-atomic read-modify-write

Detect security- or money-sensitive state transitions performed without a transaction, atomic update, version check, or lock. Require domain configuration before blocking.

#### SQLSAFE-004 - Transaction boundary loss

Detect database calls that escape the intended transaction/client context or asynchronous work launched before commit.

#### SQLSAFE-005 - Incorrect `AND`/`OR` precedence

Detect authorization or tenant predicates whose grouping changes meaning because of operator precedence or query-builder nesting.

#### SQLSAFE-006 - Soft-delete bypass

Detect raw queries or alternate repositories that omit a declared soft-delete condition.

#### SQLSAFE-007 - Nondeterministic limited result

Detect security-sensitive or pagination queries that use `LIMIT`/`TOP` without a deterministic order. Default to bug/reliability severity unless it affects authorization or processing integrity.

#### SQLSAFE-008 - Null and three-valued logic error

Detect incorrect equality comparisons to null and security predicates whose unknown result changes access behavior.

### D. Secrets, privacy, and operational exposure

#### DBSEC-001 - Hardcoded database credential or connection secret

Use secret scanning plus semantic connection-string recognition. Redact the matched value in all output.

#### DBSEC-002 - Sensitive query parameters or result data logged

Trace credentials, tokens, personal data, or declared sensitive columns into logging, tracing, exception, or analytics sinks.

#### DBSEC-003 - Detailed database error exposed to an untrusted client

Trace driver errors, query strings, schema names, or stack traces into HTTP/RPC responses.

#### DBSEC-004 - Excessive database privilege

Analyze infrastructure and database-role configuration when available. Do not infer runtime privilege solely from a username.

#### DBSEC-005 - Transport/security option disabled

Detect explicit disabling of TLS or certificate verification in database connection configuration. Account for documented local/test environments without blanket suppression.

### E. Availability and performance bugs

#### DBPERF-001 - Unbounded user-controlled result set

Detect externally reachable queries without a limit where input controls filters or pagination. Contextualize based on endpoint type and streaming behavior.

#### DBPERF-002 - N+1 query in an externally controlled loop

Use loop/control-flow and database-call evidence. Raise risk when attacker-controlled collection size or page size drives the iteration.

#### DBPERF-003 - User-controlled expensive query shape

Detect user-controlled sort fields, wildcard prefixes, complex filters, recursive queries, or regex operations without bounds. Require database/framework-specific models.

#### DBPERF-004 - Missing timeout or cancellation propagation

Detect request-bound database operations that ignore supported cancellation/deadline mechanisms.

### F. Schema and migration safety

#### DBMIG-001 - Destructive migration without guarded rollout

Detect drop/rename/type-narrowing/non-null changes that lack an expand-migrate-contract sequence or an explicit reviewed exception.

#### DBMIG-002 - Irreversible or inconsistent migration

Detect a missing/down migration only when the project's migration policy requires reversibility.

#### DBMIG-003 - Constraint/index gap on security-critical key

Detect missing uniqueness, foreign-key, or tenant-composite constraints only when schema intent is declared. Avoid guessing business invariants.

## Framework model registry

Keep framework behavior versioned outside generic rule definitions:

```text
models/
|-- javascript/
|   |-- node-postgres.yml
|   |-- mysql2.yml
|   |-- prisma.yml
|   |-- sequelize.yml
|   `-- typeorm.yml
|-- python/
|   |-- db-api.yml
|   |-- django.yml
|   `-- sqlalchemy.yml
|-- java/
|   |-- jdbc.yml
|   |-- jpa-hibernate.yml
|   `-- spring-jdbc.yml
|-- dotnet/
|   |-- ado-net.yml
|   `-- entity-framework.yml
|-- go/
|   `-- database-sql.yml
|-- php/
|   |-- pdo.yml
|   `-- laravel.yml
`-- ruby/
    |-- active-record.yml
    `-- sequel.yml
```

Each model must declare version range, sources, sinks, bind semantics, raw APIs, sanitizers, propagation helpers, analysis limitations, and fixture references.

## Rule specification contract

Use a versioned, machine-validatable schema. A conceptual rule record:

```yaml
schema_version: 1
id: SQLI-001
version: 1.0.0
title: Untrusted value reaches executable SQL
category: injection
languages: [javascript]
framework_models: [node-postgres]
analysis_tier: 3
default_severity: high
default_confidence: high
mappings:
  cwe: [CWE-89]
sources: []
propagators: []
sanitizers: []
sinks: []
preconditions: []
exceptions: []
message: "..."
remediation: "..."
tests:
  positive: []
  negative: []
owner: security-rules
```

Requirements:

- Give every rule and framework model a semantic version.
- Reject unknown schema fields during CI validation.
- Require positive and negative fixtures before enabling a rule.
- Require an explicit owner and review date.
- Keep mappings such as CWE informational; mappings do not determine severity.
- Record analysis limitations and unsupported dynamic behavior.
- Declare a performance budget and minimum engine version.
- Require a signed manifest containing the rulepack version, content hashes, and compatibility information before production use.
- Keep rule IDs immutable. Use a rule revision for behavior changes and SemVer for the complete rulepack.

Rule maturity must progress through `experimental`, `beta`, `stable`, `deprecated`, and `retired`. A rule author cannot be the only approver promoting a rule to `stable`. Pin exact rulepack versions in CI; never use a floating production tag.

## Finding output contract

Emit normalized JSON and SARIF. Each finding must contain:

```json
{
  "finding_id": "stable-fingerprint",
  "rule_id": "SQLI-001",
  "rule_version": "1.0.0",
  "severity": "high",
  "confidence": "high",
  "exploitability": "likely",
  "status": "open",
  "finding_state": "confirmed",
  "primary_location": {},
  "source": {},
  "sink": {},
  "dataflow": [],
  "preconditions": [],
  "analysis_gaps": [],
  "evidence": [],
  "remediation": {},
  "verification": {},
  "fingerprint_inputs": []
}
```

Rules for output:

- Keep severity independent from confidence.
- Identify whether the result is pattern-only, proven flow, or context-enriched.
- Redact secrets and bound literal values.
- Limit source excerpts and never include unrelated customer data.
- Use semantic fingerprints based on rule, symbol, sink, and normalized flow—not absolute line number alone.
- Deduplicate multiple syntactic paths that represent the same root cause while preserving alternate flows.
- Report unresolved calls or missing framework models as analysis gaps.

Use these finding states consistently:

- `candidate`: raw engine result not yet checked against evidence requirements.
- `probable`: strong flow with a bounded unresolved gap.
- `confirmed`: complete required evidence and supported framework semantics.
- `needs_context`: business or deployment facts are required.
- `dismissed_with_evidence`: explicit counter-evidence invalidates the candidate.
- `fixed_verified`: targeted rescan, relevant regression scan, and tests prove the fix.
- `fixed_not_verified`: code changed but complete verification could not run.

Never call a result a false positive without counter-evidence. Never mark a finding `fixed_verified` when the fingerprint only disappeared because coverage decreased, a rule was disabled, or a suppression was added.

## Scoring and CI policy

### Dimensions

- **Impact:** consequence if exploited or triggered.
- **Exploitability:** required access, attacker control, reachability, and database capability.
- **Confidence:** strength and completeness of static evidence.
- **Exposure:** public endpoint, internal service, CLI, background worker, or unreachable/dead path.
- **Remediation confidence:** likelihood that the proposed fix preserves behavior.

Do not collapse these values permanently into one opaque score. A UI may calculate a priority score, but the raw dimensions must remain visible.

### Suggested gates

| Finding class | Default behavior |
| --- | --- |
| Critical/high impact + high confidence + reachable | Block after rollout period |
| High impact + medium confidence | Warn and require triage |
| Dangerous API without tainted flow | Informational/warn |
| Analysis gap | Warn; never claim safe |
| New finding in changed code | Higher priority than unchanged baseline |
| Existing accepted finding | Track until suppression expiry |

A finding may block only when the rule is stable, impact is high/critical, confidence is high, and the finding is new relative to a compatible approved baseline. Compare baselines only when engine, rulepack, configuration, and scope compatibility keys match.

## Suppression governance

Every suppression must contain:

- Rule and stable finding fingerprint.
- Technical reason.
- Owner/approver.
- Creation and expiry date.
- Tracking ticket or risk-acceptance reference.
- Scope restricted to the narrowest location or semantic symbol.

Policy:

- Reject empty reasons and wildcard suppressions by default.
- Set a maximum default lifetime, such as 30-90 days according to severity.
- Notify before expiry and reopen automatically at expiry.
- Re-evaluate suppressions when rule major versions or relevant code change.
- Do not let the AI create a suppression unless the user explicitly requests risk acceptance and has authority.
- Treat a safe wrapper or framework model as a reviewed rule-model change, not an inline suppression repeated across the codebase.
- Limit suppression lifetimes by default: critical 7 days, high 30 days, medium 90 days, and low 180 days.
- Require security approval for critical findings and any directory/global suppression.
- Convert every adjudicated false positive into a permanent negative regression fixture.
- Invalidate a suppression after a semantic source/sink change, incompatible fingerprint change, or major rule revision.

## Scanner safety model

- Parse source without executing it by default.
- Do not install dependencies or run build hooks without approval.
- Disable network access for analysis unless an authorized operation requires it.
- Enforce repository boundaries and safe symlink handling.
- Apply file-size, recursion, archive, and generated-code limits.
- Ignore vendor/generated paths only through evidence-backed detection or explicit configuration.
- Redact secret-like data before persistence, logs, or AI context.
- Never send full files to an AI when a minimal slice and data-flow summary is sufficient.
- Record tool version, rule bundle version, framework models, and scan configuration for reproducibility.

## AI-agent orchestration

The agent skill must follow this sequence:

1. Establish whether the request is read-only analysis, configuration, or authorized remediation.
2. Inspect repository instructions, dirty state, language/framework inventory, and scanner availability.
3. Run the deterministic scanner in read-only/offline mode first.
4. Parse normalized JSON/SARIF rather than scraping prose output.
5. Validate top findings against the cited source-to-sink path and repository context.
6. State analysis gaps; never infer that unmodeled code is safe.
7. Group duplicate findings by root cause.
8. Propose fixes that preserve public behavior and use the framework's documented safe API.
9. Edit only when requested or clearly included in the user's task.
10. Re-run the exact rule and relevant repository tests after a fix.
11. Report resolved, still open, regressed, blocked, and not-verified findings separately.

Use two explicit state machines:

```text
Review:
PREFLIGHT -> INVENTORY -> SCAN_PLAN -> SCAN -> NORMALIZE -> TRIAGE -> REPORT

Remediation:
PREFLIGHT -> INVENTORY -> SCAN_PLAN -> SCAN -> NORMALIZE -> TRIAGE
-> AUTHORIZE -> PATCH -> TARGETED_RESCAN -> REGRESSION_SCAN -> REPORT
```

Default an ambiguous request to read-only review. Do not patch until the requested scope and selected findings are explicit. In read-only mode, keep caches and output in a temporary directory outside the repository.

Stop and report `blocked` when scanner/rulepack integrity fails, output schema is incompatible, network/source upload lacks authorization, production credentials or databases would be touched, a finding has become stale, or safe remediation requires an unapproved auth/schema/public-API boundary change. Report `partial` for timeouts, parse failures, skipped files, or unsupported language/dialect slices; do not turn partial coverage into a clean bill of health.

The AI may:

- Explain a deterministic finding.
- Inspect contextual preconditions.
- Suggest a narrow remediation.
- Identify likely duplicate/root-cause relationships.
- Recommend a new rule-model fixture when the scanner lacks framework knowledge.

The AI must not:

- Create a high-confidence finding without code evidence.
- Downgrade or suppress solely because a developer says it is a false positive.
- expose secret values or sensitive query results.
- Execute an exploit against a real system.
- Claim remediation success without re-scan evidence.

## Evaluation system

### Fixture matrix

For each rule/framework/version combination, include:

- Minimal positive fixture.
- Minimal negative/safe fixture.
- Realistic multi-file positive fixture.
- Near-miss negative fixture.
- Sanitizer-valid fixture.
- Sanitizer-invalid/wrong-context fixture.
- Dynamic/unresolved fixture expected to yield an analysis gap.
- Fixed variant proving the recommended remediation.

Maintain three distinct corpora:

1. Paired rule fixtures for fast, exhaustive rule development.
2. Buildable integration fixtures with multi-file wrappers, async flows, and real framework versions.
3. A frozen holdout corpus, split by repository and unseen during rule authoring.

Store provenance, license, reviewers, framework/driver version, expected span, expected trace, and checksum for every case. Deduplicate corpus samples by normalized AST hash. Require two independent labels for blocking-rule cases and adjudicate disagreements through a third security reviewer.

### Test layers

1. **Schema tests:** rule/model/finding records validate strictly.
2. **Parser tests:** source locations and AST facts remain stable.
3. **Flow tests:** expected source-to-sink paths match golden output.
4. **Mutation tests:** inject known unsafe changes into safe fixtures and require detection.
5. **Regression tests:** every confirmed false positive/negative gets a permanent fixture.
6. **Performance tests:** large monorepo, generated code, deep call graph, and path explosion.
7. **Security tests:** malicious repositories, symlinks, huge files, parser crashes, secret redaction.
8. **Agent tests:** fresh-context review and remediation prompts against raw scan artifacts.
9. **Metamorphic tests:** rename symbols, extract helpers, add wrappers, or change equivalent syntax without changing the expected result.
10. **Safe-transform tests:** parameter binding, identifier allow-list, typed APIs, and strict numeric parsing must kill the intended finding.

### Quality gates

Initial targets for the curated benchmark:

| Metric | Blocking rules | Warning rules |
| --- | ---: | ---: |
| Precision | >= 98% | >= 90% |
| Recall on modeled patterns | >= 90% | >= 85% |
| Secret leakage | 0 | 0 |
| Unexplained findings | 0 | 0 |
| Stable fingerprint after line-only edits | >= 99% | >= 99% |
| Deterministic output across identical runs | 100% | 100% |

Additional production gates for blocking rules:

- Report 95% Wilson confidence intervals, not only point estimates.
- Require precision lower bound of at least 93% and recall lower bound of at least 80%.
- Require hard-negative specificity of at least 99.5%.
- Allow no more than one adjudicated false positive per 100,000 executable lines.
- Require exact primary-location accuracy of at least 95% and complete traces for at least 90% of taint findings.
- Keep duplicate rate below 1%.
- Achieve deterministic output over three identical runs.
- Keep scanner crash rate at or below 0.1%.
- Enforce performance budgets per target repository size before rollout.

Measure all metrics per rule, language, framework, and driver. Do not hide a weak slice behind a global average, and do not promote a rule from a statistically inadequate sample.

Measure precision and recall per rule and framework, not only as a global average. Do not enable blocking when the sample size is too small to support the target.

### Rollout stages

1. Disabled: schema and fixture development only.
2. Experimental: opt-in local scans.
3. Audit: CI reports without developer notification or blocking.
4. Warn: visible findings with feedback collection.
5. Block-new: block only qualifying findings introduced by the change.
6. Block-all: optional policy for mature, high-confidence rules.

Rollback a rule to warning when a framework update, precision regression, or performance regression breaches its gate.

Require a shadow period before blocking and deploy organization-wide through canary stages such as 5%, 25%, 50%, and 100%. Treat scanner failure as a separate status; on protected/release branches it must never be interpreted as `completed_no_findings`.

## Prioritized implementation backlog

### P0 - Architecture and contracts (Go Code Scanner Alignment)

- [x] Primary target language selected: **Go** (`go/ast`, `golang.org/x/tools/go/analysis`) with `database/sql`, `gorm`, `sqlx`, and `pgx`.
- [ ] Extend `pkg/finding/report.go` to include `DataflowStep` (`Source`, `Propagator`, `Sanitizer`, `Sink`), `Confidence`, `Exploitability`, and `FindingState`.
- [ ] Define `pkg/scanner/sqltaint` engine interface and SQL template/hole reconstruction model.
- [ ] Define strict schemas for rules, framework models, findings, and suppressions (integrating with `pkg/suppression` and `pkg/baseline`).
- [ ] Define repository threat model and non-execution sandbox.
- [ ] Define severity, confidence, exploitability, and CI-gate policy (aligning with `pkg/policy`).

### P1 - First high-confidence slice (Go & SQL Taint)

- [ ] Implement Go AST/source location parser in `pkg/scanner/sqltaint`.
- [ ] Implement local intraprocedural string taint propagation for Go functions.
- [ ] Implement SQL template reconstruction & prepared-statement state tracking for `database/sql` / `sqlx` / `gorm`.
- [ ] Implement `SQLI-001`, `SQLI-002`, `SQLI-004`, `SQLI-008`, and `SQLSAFE-001` in `pkg/rules/defaults_security.go` or `pkg/scanner/sqltaint`.
- [ ] Add `SQLI-011` and `SQLI-012` for complete list expansion and prepared-query coverage.
- [ ] Ensure JSON and SARIF output writers ([pkg/reporter](file:///c:/Users/Gositus%20Hadi/code/go-code-scanner/pkg/reporter)) serialize dataflow traces with secret redaction.
- [ ] Align stable semantic fingerprints with `pkg/finding/report.go`.
- [ ] Build positive/negative Go fixture matrix for the initial rules.

### P2 - Cross-layer context & Agent Skill Integration

- [ ] Add HTTP/RPC/CLI entry-point reachability for Go routers (`net/http`, `gin`, `chi`, `fiber`).
- [ ] Add authorization and tenant-barrier models.
- [ ] Implement `SQLAUTH-001` through `SQLAUTH-003` behind project configuration.
- [ ] Add persistent-source modeling for `SQLI-005`.
- [ ] Add transaction/atomicity rules (`SQLSAFE-003`, `SQLSAFE-004`).
- [ ] Integrate normalized JSON/SARIF taint findings into `.agents/skills/go-code-scanner/`.

### P3 - Breadth and scale

- [ ] Add more languages and framework models one at a time.
- [ ] Add database migration analysis.
- [ ] Add availability/performance rules.
- [ ] Add incremental and changed-code scanning.
- [ ] Add cached summaries and path-explosion controls.
- [ ] Add cross-service data-flow hints without claiming complete global taint.
- [ ] Add rule telemetry only with approved privacy controls.

## Definition of done for the first rule bundle

- [ ] Five P1 rules work for one explicitly versioned framework model.
- [ ] Every rule has positive, negative, near-miss, invalid-sanitizer, and fixed fixtures.
- [ ] Findings contain a reproducible source-to-sink trace or are labeled pattern-only.
- [ ] JSON and SARIF outputs validate and are deterministic.
- [ ] Secrets are redacted in scanner and agent outputs.
- [ ] Semantic fingerprints survive line-only edits.
- [ ] The scanner does not execute repository code or access the network by default.
- [ ] Blocking rules meet the precision gate on an adequately sized benchmark.
- [ ] The agent distinguishes findings, analysis gaps, and unverified assumptions.
- [ ] Production SARIF contains rule metadata, code flows, semantic fingerprints, baseline state, suppressions, redaction, and deterministic ordering.
- [ ] Security and product owners approve the initial bundle.
