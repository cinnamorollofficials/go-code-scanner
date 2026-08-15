---
title: Changelog
description: "For users tracking releases: review notable Go Code Scanner changes, release notes, and version history."
---

# Changelog

All notable changes to this project will be documented in this file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Native AST & SQL Taint Analysis Engine (`sqltaint`) with Interprocedural Call Graph and HTTP Router Models (Gin, Echo, Chi, Fiber, Mux, `net/http`) for deep vulnerability detection with step-by-step dataflow provenance tracing.
- Detection rules for SQL injection (`SQLI-001`), dynamic identifier interpolation (`SQLI-002`), unsafe ORM escape hatches (`SQLI-004`), bind parameter mismatch (`SQLI-008`), list expansion (`SQLI-011`), tainted prepare templates (`SQLI-012`), and unbounded destructive queries (`SQLSAFE-001`).
- Multi-tenant and authorization rule taxonomy (`SQLAUTH-001` through `SQLAUTH-004`).
- Integrity and transactional safety guardrails (`SQLSAFE-003` through `SQLSAFE-006`).
- Multi-language security and ORM taint models across Node.js/TypeScript (Prisma, TypeORM, Sequelize, `pg`, `mysql2`), Python (SQLAlchemy, Django, `psycopg`), and Java/Kotlin (Spring Data JPA, Hibernate, JDBC).
- Database migration & schema integrity guardrails (`DBMIG-001` through `DBMIG-003`).
- Performance and error privacy guardrails (`DBPERF-001`, `DBPERF-002`, `DBSEC-002`, `DBSEC-003`).
- Safe automated AST code remediation (`--fix`) for SQL injection vulnerabilities (`SQLI-001`).
- AI Agent Skill package (`.agents/skills/go-code-scanner`) with automated project detection, config validation, and offline verification scripts.
- Policy-driven scanning across Quality, Reliability, Hardening, Security, Supply Chain, and Governance.
- Safe Git hooks, staged snapshots, external scanner adapters, baselines, suppressions, caching, and CI report formats.
- Cross-platform release builds, checksums, provenance manifests, and Ed25519 signature verification.

### Security

- Strict JSON decoding, path confinement, symlink protection, redaction, resource limits, and process isolation.
