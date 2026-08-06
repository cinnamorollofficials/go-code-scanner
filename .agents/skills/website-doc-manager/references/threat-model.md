# Threat Model & Safety Guardrails - `go-code-scanner` Website

## Trust Boundaries & Isolation

1. **Local File System Scope**:
   - Operations must be strictly scoped to the repository workspace (`c:\Users\Gositus Hadi\code\go-code-scanner`).
   - Never attempt to read or modify files outside the workspace directory boundary.

2. **Secret & Credential Boundaries**:
   - **PROHIBITED**: Writing real AWS keys, GCP API tokens, JWT secrets, passwords, or private SSH keys inside documentation code blocks or fixtures.
   - **REQUIRED**: Use synthetic placeholders for examples (e.g., `AKIAIOSFODNN7EXAMPLE`, `example_secret_token_12345`).

3. **Execution Guardrails**:
   - Never execute untrusted third-party binaries or unknown shell commands.
   - Build commands (`npm run docs:build`, `go run ./cmd/gen-rule-catalog`) must be executed with standard local environment flags and without network uploads.

4. **Security Policy Guardrails**:
   - Documentation must accurately reflect built-in security defaults (`pkg/rules`).
   - Do not instruct users to disable security controls or bypass rule verification in production setups without explicit warnings.
