---
title: Changelog
description: All notable changes, release notes, and version history for Go Code Scanner.
---

# Changelog

All notable changes to this project will be documented in this file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Native AST & Intraprocedural SQL Taint Analysis Engine (`sqltaint`) for deep vulnerability detection with step-by-step dataflow provenance tracing.
- Detection rules for SQL injection (`SQLI-001`), dynamic identifier interpolation (`SQLI-002`), unsafe ORM escape hatches (`SQLI-004`), bind parameter mismatch (`SQLI-008`), and unbounded destructive queries (`SQLSAFE-001`).
- AI Agent Skill package (`.agents/skills/go-code-scanner`) with automated project detection, config validation, and offline verification scripts.
- Policy-driven scanning across Quality, Reliability, Hardening, Security, Supply Chain, and Governance.
- Safe Git hooks, staged snapshots, external scanner adapters, baselines, suppressions, caching, and CI report formats.
- Cross-platform release builds, checksums, provenance manifests, and Ed25519 signature verification.

### Security

- Strict JSON decoding, path confinement, symlink protection, redaction, resource limits, and process isolation.
