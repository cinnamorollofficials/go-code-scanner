# Security & Threat Model

`go-code-scanner` and its AI Agent Skill enforce strict isolation, non-execution sandboxing, and secret redaction boundaries.

## Security Principles

1. **Non-Execution Sandbox**: Scans analyze AST and text patterns without building or executing untrusted repository code.
2. **Local Isolation**: Staged mode (`--staged`) materializes index snapshots in isolated temporary directories without altering working tree state.
3. **Secret Redaction**: Credentials, API tokens, and secret strings matched by pattern scanners are redacted before printing or emitting SARIF/JSON artifacts.
4. **No Unreviewed Suppressions**: AI agents must never create wildcard suppressions or remove security gates without security owner approval.
