# Changelog

All notable changes to this project will be documented in this file. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and releases use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Policy-driven scanning across Quality, Reliability, Hardening, Security, Supply Chain, and Governance.
- Advanced AST & SQL taint analysis engine (`pkg/scanner/sqltaint`) with interprocedural call graph tracing and HTTP router entrypoint models (Gin, Echo, Chi, Fiber, Gorilla Mux, `net/http`).
- Multi-tenant and authorization rule taxonomy (`SQLAUTH-001` through `SQLAUTH-004`).
- Concurrency and transactional safety guardrails (`SQLSAFE-003` through `SQLSAFE-006`).
- Multi-language security and ORM taint models across Node.js/TypeScript (Prisma, TypeORM, Sequelize, `pg`, `mysql2`), Python (SQLAlchemy, Django, `psycopg`), and Java/Kotlin (Spring Data JPA, Hibernate, JDBC).
- Database migration & schema integrity guardrails (`DBMIG-001` through `DBMIG-003`).
- Performance and error privacy guardrails (`DBPERF-001`, `DBPERF-002`, `DBSEC-002`, `DBSEC-003`).
- Safe automated AST code remediation (`--fix`) for SQL injection vulnerabilities (`SQLI-001`).
- AI Agent Skill (`.agents/skills/go-code-scanner`) with automated preflight, configuration validation, and scan verification scripts.
- Safe Git hooks, staged snapshots, external scanner adapters, baselines, suppressions, caching, and CI report formats.
- Cross-platform release builds, checksums, provenance manifests, and Ed25519 signature verification.

### Security

- Strict JSON decoding, path confinement, symlink protection, redaction, resource limits, and process isolation.
