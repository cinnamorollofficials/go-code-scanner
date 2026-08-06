---
name: go-code-scanner
description: Run security scans, validate security-review configuration, inspect code/SQL vulnerabilities, audit findings, set up Git hooks, or remediate security findings in Go and polyglot codebases. Trigger when users ask to scan code, run security review, validate security rules, set up pre-commit hooks, or fix security findings.
---

# Go Code Scanner Agent Skill

This skill enables AI agents to inspect repositories, run static code and SQL security scans using the `security-review` CLI, validate configurations, triage evidence-backed findings, and apply targeted remediations cleanly.

## State Machines

The agent MUST follow one of two explicit state machines based on task type:

### 1. Review Mode (Read-only Analysis & Audit)
```text
PREFLIGHT -> INVENTORY -> SCAN_PLAN -> SCAN -> NORMALIZE -> TRIAGE -> REPORT
```
- Default for ambiguous requests, scan requests, or audit tasks.
- Never modify source files or repository settings in Review Mode.

### 2. Remediation Mode (Authorized Fixes)
```text
PREFLIGHT -> INVENTORY -> SCAN_PLAN -> SCAN -> NORMALIZE -> TRIAGE
  -> AUTHORIZE -> PATCH -> TARGETED_RESCAN -> REGRESSION_SCAN -> REPORT
```
- Used only when the user explicitly requests code fixes or remediation.
- Re-run targeted rescan and regression scan before claiming success.

---

## Detailed Execution Workflow

### Step 1: Preflight & Environment Inspection
Run the detection script to inspect the project layout:
```powershell
powershell -ExecutionPolicy Bypass -File .agents/skills/go-code-scanner/scripts/detect-project.ps1 -TargetDir .
```

### Step 2: Configuration Validation
Before scanning, ensure the project's scanner configuration is valid:
```powershell
powershell -ExecutionPolicy Bypass -File .agents/skills/go-code-scanner/scripts/validate-config.ps1 -ConfigFile security-review.json
```

### Step 3: Run Deterministic Offline Scan
Execute the scanner CLI using the verification script:
```powershell
powershell -ExecutionPolicy Bypass -File .agents/skills/go-code-scanner/scripts/verify-integration.ps1 -TargetDir . -Profile fast
```

Alternatively, invoke native CLI commands directly:
- Full scan: `security-review scan --format json --output artifacts/findings.json`
- Staged commit scan: `security-review scan --staged --profile fast --ci --new-only`
- Baseline comparison: `security-review scan --baseline .security-baseline.json --new-only`

### Step 4: Triage & Evidence Verification
Inspect generated findings:
- Categorize by Domain (`Security`, `Hardening`, `Reliability`, `Quality`, `Supply chain`, `Governance`).
- Check `Confidence` and `Exploitability`.
- Review the `Dataflow` trace (`source` -> `propagator` -> `sanitizer` -> `sink`).
- Group duplicate findings by root cause.

### Step 5: Remediation & Rescan (If Requested)
When performing a fix:
1. Apply the smallest code change using parameterization (e.g., prepared statements `$1`, `?` instead of string concatenation).
2. Re-run `security-review scan` to verify the finding status changed from `candidate`/`open` to `fixed_verified`.
3. Report resolved, still open, and unverified findings separately.

---

## Prohibited Operations & Guardrails

1. **No Secret Exposure**: Never display secret values, token strings, or hardcoded credentials in outputs.
2. **No Unsafe Policy Weakening**: Never disable security rules or add unreviewed suppressions just to pass CI.
3. **No Code Execution**: Parse AST and run static analysis without executing untrusted repository code.
4. **No False Claims**: If verification fails or CLI is unavailable, report status as `not_verified` or `blocked`.

---

## Progressive References

- Detailed CLI flags and exit codes: [`references/integration-contract.md`](references/integration-contract.md)
- Configuration fields and options: [`references/configuration.md`](references/configuration.md)
- Supported frameworks & database drivers: [`references/frameworks.md`](references/frameworks.md)
- Security boundaries & threat model: [`references/threat-model.md`](references/threat-model.md)
- Diagnostic steps for common scan failures: [`references/troubleshooting.md`](references/troubleshooting.md)
- Tested setup & workflow examples: [`references/examples.md`](references/examples.md)
