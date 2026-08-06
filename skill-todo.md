# Go Code Scanner & Agent Skill Implementation TODO

This tracking document outlines the unified TODO backlog for implementing the AI Agent Skill (`.agents/skills/go-code-scanner`) and the Advanced AST & SQL Vulnerability Engine (`pkg/scanner/sqltaint`) for the `go-code-scanner` repository.

---

## 1. Core Go Scanner Engine & AST Taint Subsystem

### Phase 1.1 - Finding Data Model & Schema Alignment (`pkg/finding/report.go`)
- [x] Add `DataflowStep` struct (`Source`, `Propagator`, `Sanitizer`, `Sink`) to represent taint paths.
- [x] Add `Confidence` (`high`, `medium`, `low`) and `Exploitability` (`likely`, `unlikely`, `unknown`) fields.
- [x] Add `FindingState` (`candidate`, `probable`, `confirmed`, `dismissed_with_evidence`, `fixed_verified`).
- [x] Ensure JSON (1.0) and SARIF (2.1.0) reporters in `pkg/reporter` serialize dataflow traces safely with secret redaction.

### Phase 1.2 - Go AST & SQL Taint Analyzer (`pkg/scanner/sqltaint`)
- [x] Implement Go AST parser using `go/ast` and `golang.org/x/tools/go/analysis`.
- [x] Implement intraprocedural string concatenation and variable taint propagation.
- [x] Implement SQL query template reconstruction & driver hole classification (`value`, `identifier`, `clause`, `unknown`).
- [x] Implement tracking of prepared-statement states (`raw` -> `prepared` -> `bound` -> `executed`).
- [x] Implement rules: `SQLI-001`, `SQLI-002`, `SQLI-004` (ORM escape hatch), `SQLI-008` (bind mismatch), and `SQLSAFE-001` (unbounded update/delete).
- [x] Implement positive/negative Go fixture test suite in `pkg/scanner/sqltaint/sqltaint_test.go`.
- [x] Register `sqltaint` scanner into `securityreview.go` default reviewers list.

---

## 2. AI Agent Skill Package (`.agents/skills/go-code-scanner/`)

### Phase 2.1 - Script Helpers & Native CLI Integration
- [x] Implement `scripts/detect-project.ps1`: Detect Go/JS/Python/etc. manifests, package managers, and CI workflows.
- [x] Implement `scripts/validate-config.ps1`: Validate `security-review.json` via native `security-review config validate`.
- [x] Implement `scripts/verify-integration.ps1`: Execute offline scan audit or dry-run via `security-review scan`.

### Phase 2.2 - Progressive Reference Files (`references/`)
- [x] Write `references/integration-contract.md`: Document CLI flags (`--staged`, `--baseline`, `--new-only`, `--ci`) and exit codes.
- [x] Write `references/configuration.md`: Document configuration fields, domain thresholds, and offline profile policies.
- [x] Write `references/frameworks.md`: Matrix of supported Go drivers (`database/sql`, `gorm`, `sqlx`, `pgx`) and adapters.
- [x] Write `references/threat-model.md`: Document non-execution sandbox, read-only guarantees, secret redaction, and local isolation.
- [x] Write `references/troubleshooting.md`: Diagnostic procedures for common exit codes, missing binaries, or unmanaged Git hook conflicts.
- [x] Write `references/examples.md`: Complete working examples for local commit gates and GitHub Actions workflows.

### Phase 2.3 - Skill Workflow & State Machine (`SKILL.md`)
- [x] Initialize `SKILL.md` with YAML frontmatter (`name: go-code-scanner`, comprehensive trigger description).
- [x] Implement **Review Mode State Machine**: `PREFLIGHT` -> `INVENTORY` -> `SCAN_PLAN` -> `SCAN` -> `NORMALIZE` -> `TRIAGE` -> `REPORT`.
- [x] Implement **Remediation Mode State Machine**: `PREFLIGHT` -> `INVENTORY` -> `SCAN_PLAN` -> `SCAN` -> `NORMALIZE` -> `TRIAGE` -> `AUTHORIZE` -> `PATCH` -> `TARGETED_RESCAN` -> `REGRESSION_SCAN` -> `REPORT`.
- [x] Enforce safety guardrails: Reject exposing secrets, never weaken policy solely to pass CI, report unverified state as `not_verified`.
- [x] Generate `agents/openai.yaml` from completed `SKILL.md`.

---

## 3. Prioritized Backlog Summary

### P0 - Core Contracts & Base Integration (Completed)
- [x] Align plans with actual Go codebase (`cmd/security-review` and `pkg/`).
- [x] Update `pkg/finding/report.go` with taint trace fields.
- [x] Create `.agents/skills/go-code-scanner/` structure and helper scripts.

### P1 - MVP Release Target (Completed)
- [x] Complete Go AST SQL Taint scanner (`SQLI-001` through `SQLI-008`).
- [x] Complete `SKILL.md` and six reference files.

### P2 - Post-MVP Enhancements
- [ ] Add interprocedural taint analysis across Go package call graphs.
- [ ] Add HTTP router entry-point reachability (`chi`, `gin`, `fiber`).
- [ ] Add automated remediation workflows (`security-review scan --fix`).
