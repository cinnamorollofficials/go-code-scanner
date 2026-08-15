---
title: Reproducing a Finding
description: Step-by-step operational playbook for isolating, diagnosing, and verifying reported security findings.
---

# Reproducing a Finding

When `security-review` reports a security finding in CI or local CLI execution, follow this 6-step reproduction playbook to isolate, diagnose, and verify the issue.

---

## Step 1: Extract Finding Context

Gather key metadata from your terminal scan output or JSON report:

```json
{
  "rule_id": "go-shell-command",
  "domain": "security",
  "severity": "HIGH",
  "location": {
    "file": "pkg/runner/exec.go",
    "line": 45
  },
  "fingerprint": "a3f89b12e38c...",
  "description": "Unsanitized user input passed to os/exec.Command"
}
```

Key fields for reproduction:
- **`rule_id`**: The detection rule ID (`go-shell-command` or `SQLI-001`).
- **`file` & `line`**: Source location (`pkg/runner/exec.go:45`).
- **`fingerprint`**: Deterministic SHA-256 fingerprint.

---

## Step 2: Run Targeted Isolated Scan

Isolate execution to the target root directory or file path to eliminate noise from unrelated repository files:

```sh
# Run targeted scan on specific directory with verbose output
security-review scan --root pkg/runner --verbose
```

---

## Step 3: Inspect Rule Rationale & Diagnostics

Use the `--explain` flag to retrieve rule recommendations and remediation examples:

```sh
security-review scan --explain go-shell-command
```

Output details include:
- Security risk description & CWE reference.
- Vulnerable vs. Secure code examples.
- Remediation guidance.

---

## Step 4: Verify Baseline & Suppression Status

Determine if the finding is active, suppressed, or baseline-filtered:

1. **Test Raw Scan**:
   ```sh
   # Run scan without baseline to verify raw detection
   security-review scan --root pkg/runner
   ```
2. **Check Inline Annotations**:
   Inspect if `// nolint:<rule_id>` exists on or above the target line.
3. **Check Suppression File**:
   Verify if the file path or fingerprint is registered in `.security-ignore`.

---

## Step 5: Construct Minimal Reproducible Example (MRE)

To confirm whether a finding is a **True Positive (TP)** or a **False Positive (FP)**, isolate the trigger pattern in a minimal test fixture:

```go
// testdata/repro_command_injection.go
package testdata

import "os/exec"

func Repro(userInput string) {
    // Should trigger go-shell-command
    exec.Command("sh", "-c", userInput)
}
```

Scan the fixture directory:
```sh
security-review scan --root testdata
```

- If the MRE triggers as expected: **Confirmed True Positive**.
- If code is safe but flagged: **False Positive**. Document the pattern and register a suppression or submit a rule refinement.

---

## Step 6: Remediation & Fix Verification

Apply the recommended remediation to the source code:

```diff
- exec.Command("sh", "-c", userInput)
+ exec.Command("ls", "-l", safeFileName)
```

Re-run the scan to verify finding resolution:

```sh
security-review scan --root pkg/runner
```

Verification is successful when the scanner reports **0 findings** and exits with code `0`.
