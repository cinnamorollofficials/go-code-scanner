---
title: Concepts Overview
description: "For policy owners and developers: understand Go Code Scanner scan modes, profiles, findings, frontend analysis, and SQL taint analysis."
---

# Concepts Overview

Learn the foundational concepts and design principles behind `security-review` (Go Code Scanner).

---

## Concept Pages

<div class="tip custom-block" style="padding-top: 8px">

- [**Scan Modes and Isolation**](/concepts/scan-modes) — Full, changed, and staged Git index snapshot isolation mechanics.
- [**Profiles and Policy**](/concepts/profiles-and-policy) — Performance profiles, 6 canonical domains, and severity thresholds.
- [**Reports and Finding Lifecycle**](/concepts/reports-and-finding-lifecycle) — Deterministic fingerprints, report artifacts (JSON, SARIF, JUnit), and lifecycle states.
- [**Frontend Scanning**](/concepts/frontend-scanning) — Client-side asset inspection, JSX/Vue/Svelte analysis, and import cycle detection.
- [**SQL Taint Analysis**](/concepts/sql-taint-analysis) — AST-based interprocedural taint propagation, call graphs, and automated remediation.

</div>

---

## 6 Canonical Finding Domains

`security-review` categorizes all security findings into 6 policy domains:

1. **Security (`security`)**: Secrets exposure, SQL injection, command injection, and tainted dataflows.
2. **Hardening (`hardening`)**: Defensive configuration, file permissions, and strict transport security.
3. **Reliability (`reliability`)**: Resource exhaustion, goroutine leaks, and unhandled errors.
4. **Quality (`quality`)**: Dead code, anti-patterns, and bad practices.
5. **Supply Chain (`supply_chain`)**: Vulnerable dependencies and untrusted third-party packages.
6. **Governance (`governance`)**: License compliance, unresolved merge conflict markers, and architectural constraints.
