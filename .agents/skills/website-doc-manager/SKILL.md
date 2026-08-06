---
name: website-doc-manager
description: Manage, update, build, and verify the go-code-scanner website documentation (VitePress) and rule catalog docs. Trigger when working on go-code-scanner documentation, website pages, sidebar navigation, or rule catalog generator.
---

# Website Documentation Manager (`website-doc-manager`)

Use this skill when modifying, adding, or auditing documentation for `go-code-scanner` under the `website/` directory or rule catalog documentation generators in `cmd/gen-rule-catalog`.

## Core Guardrails & Requirements

1. **Pre-inspection & Repository State**:
   - Inspect the git status before editing to identify dirty worktrees (`git status`).
   - Read relevant existing docs and `.vitepress/config.mts` before making changes.
   - Ensure CLI commands and flags mentioned in docs match actual Go CLI flags in `cmd/security-review/main.go`.

2. **Security & Threat Model**:
   - Never embed secret keys, real API tokens, or internal credentials in documentation code samples.
   - Refer to `references/threat-model.md` for trust boundaries and prohibited operations.

3. **Reference Routing**:
   - Refer to `references/contract.md` for VitePress sidebar structure, rule domain taxonomy, and CLI command contracts.
   - Refer to `references/troubleshooting.md` when encountering VitePress build failures or dead links.
   - Refer to `references/examples.md` for approved document layout and component templates.

---

## Workflow Steps

### Step 1: Detect Environment
Run the detection script to confirm the project structure and toolchain availability:
```powershell
powershell -ExecutionPolicy Bypass -File .agents/skills/website-doc-manager/scripts/detect-project.ps1
```

### Step 2: Formulate Change Plan
Before editing, formulate a minimal change plan detailing:
- Target markdown files or configuration files (`.vitepress/config.mts`).
- Any new sidebar nav entries required.
- Verification steps.

### Step 3: Execute Edits
- Edit markdown files in `website/docs/`.
- Ensure accurate headings, standard frontmatter (if needed), and precise hyperlinks.
- If modifying rule catalog descriptions or built-in rules, update Go definitions in `pkg/rules` or generator in `cmd/gen-rule-catalog/main.go`.

### Step 4: Validate & Verify
Run the validation and verification scripts:
```powershell
powershell -ExecutionPolicy Bypass -File .agents/skills/website-doc-manager/scripts/validate-config.ps1
powershell -ExecutionPolicy Bypass -File .agents/skills/website-doc-manager/scripts/verify-integration.ps1
```

### Step 5: Final Handoff
Report execution summary including:
- Files modified/added.
- Verification result (`0 broken links`, `docs:build` status).
- Any explicit `not_verified` items if verification tools could not be run.
