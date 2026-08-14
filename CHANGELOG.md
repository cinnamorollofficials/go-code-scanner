# Changelog

All notable changes to this project will be documented in this file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Policy-driven scanning across Quality, Reliability, Hardening, Security, Supply Chain, and Governance.
- Advanced AST & SQL taint analysis engine (`pkg/scanner/sqltaint`) for detecting SQL injection and unbounded queries.
- AI Agent Skill (`.agents/skills/go-code-scanner`) with automated preflight, configuration validation, and scan verification scripts.
- Safe Git hooks, staged snapshots, external scanner adapters, baselines, suppressions, caching, and CI report formats.
- Cross-platform release builds, checksums, provenance manifests, and Ed25519 signature verification.

### Security

- Strict JSON decoding, path confinement, symlink protection, redaction, resource limits, and process isolation.
