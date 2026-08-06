# Troubleshooting Guide - `go-code-scanner` Website

Format: **Symptom** -> **Checks** -> **Cause** -> **Remediation**

---

### Issue 1: Dead Link / Broken Link in VitePress Build

- **Symptom**: `npm run docs:build` outputs `[vitepress] found 1 dead link: /features/unknown-page`.
- **Checks**:
  1. Inspect `.vitepress/config.mts` sidebar/nav entries.
  2. Run `powershell -ExecutionPolicy Bypass -File .agents/skills/website-doc-manager/scripts/validate-config.ps1`.
- **Cause**: A link in sidebar configuration or markdown content points to a non-existent file path or missing `.md` suffix target.
- **Remediation**: Correct the link in `.vitepress/config.mts` or create the missing `.md` file under `website/docs/`.

---

### Issue 2: Rule Catalog (`rules.md`) Out of Sync with Go Source

- **Symptom**: `rules.md` is missing newly added Go rules or shows outdated category titles.
- **Checks**:
  1. Check rules defined in `pkg/rules/`.
  2. Check rule catalog generator `cmd/gen-rule-catalog/main.go`.
- **Cause**: Rules were updated in Go code without running the generator script.
- **Remediation**: Run `go run ./cmd/gen-rule-catalog` from repository root to regenerate `website/docs/reference/rules.md`.

---

### Issue 3: CLI Subcommand or Flag Inconsistency

- **Symptom**: Documentation examples show invalid CLI commands (e.g. `security-review hooks install` or `--target`).
- **Checks**:
  1. Inspect CLI flag definitions in `cmd/security-review/main.go`.
  2. Verify command against `references/contract.md`.
- **Cause**: Documentation author used synthetic or outdated CLI command syntax.
- **Remediation**: Update doc examples to use canonical syntax (`security-review hook install pre-commit`, `security-review scan --root <path>`).
