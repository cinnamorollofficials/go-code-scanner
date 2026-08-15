---
title: Gradual Adoption with Baselines
description: "For maintainers adopting security-review: create and update a baseline while continuing to block newly introduced findings."
---

# Gradual Adoption with Baselines

Adopting security tooling in established codebases often produces hundreds of pre-existing findings. `security-review` supports **finding baselines** to prevent newly introduced vulnerabilities while allowing teams to remediate existing debt progressively.

---

## Step 1: Capture Initial Baseline Snapshot

Run a full scan across your project and record findings into a baseline file:

```sh
# 1. Generate full scan JSON report
security-review scan --output security_findings.json

# 2. Create baseline snapshot from report
security-review baseline create \
  --report security_findings.json \
  --baseline .security-baseline.json
```

Commit the generated `.security-baseline.json` into your Git repository so all team members and CI runners share the same snapshot.

---

## Step 2: Enforce Policy on New Findings Only

Configure your CI pipeline or local hooks to pass `--baseline` and **`--new-only`**:

```sh
# Fails with exit code 1 ONLY if NEW findings are introduced
security-review scan --ci --baseline .security-baseline.json --new-only
```

When `--new-only` is active, any pre-existing finding recorded in `.security-baseline.json` is ignored during policy evaluation. Only newly added violations will fail CI.

---

## Step 3: Inspect Baseline Status

Check the status of current repository findings compared to the recorded baseline:

```sh
# 1. Run a new scan
security-review scan --output security_findings.json

# 2. Compare against baseline
security-review baseline status \
  --report security_findings.json \
  --baseline .security-baseline.json
```

Illustrative output:
```text
Baseline: new=0 existing=14 resolved=3
```

---

## Step 4: Updating the Baseline

When your team fixes legacy vulnerabilities, update the baseline snapshot to lock in improvements:

```sh
# Preview baseline updates with --dry-run
security-review baseline update \
  --report security_findings.json \
  --baseline .security-baseline.json \
  --dry-run

# Apply update and accept resolved findings removal
security-review baseline update \
  --report security_findings.json \
  --baseline .security-baseline.json \
  --accept-resolved
```
