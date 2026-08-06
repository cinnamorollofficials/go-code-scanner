---
title: Rule Catalog
description: Complete catalog of default built-in security, secret, governance, and quality rules grouped by domain.
---

# Built-In Rule Catalog

Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog is organized into functional policy domains.

## Domain Overview

| Domain | Icon | Total Rules | Scope & Focus |
| :--- | :---: | :---: | :--- |
| **Security Rules** | 🔒 | 17 | Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws. |
| **Hardening Rules** | 🛡️ | 6 | Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings. |
| **Reliability Rules** | ⚡ | 6 | Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes. |
| **Quality Rules** | 🧹 | 5 | Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements. |
| **Governance Rules** | 📜 | 4 | Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints. |

---

## 🔒 Security Rules

Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`mock-token`](#mock-token) | `CRITICAL` | `secret_leak` | Hardcoded mock token ditemukan — hapus sebelum production |
| [`browser-token-storage`](#browser-token-storage) | `HIGH` | `data_leak` | Token disimpan di localStorage — gunakan HttpOnly Cookie |
| [`permission-bypass`](#permission-bypass) | `CRITICAL` | `security_misconfiguration` | Permission bypass hardcoded ditemukan |
| [`weak-secret`](#weak-secret) | `CRITICAL` | `secret_leak` | Default atau weak secret ditemukan |
| [`frontend-sensitive-log`](#frontend-sensitive-log) | `MEDIUM` | `data_leak` | Log frontend mungkin menampilkan data sensitif |
| [`backend-sensitive-log`](#backend-sensitive-log) | `MEDIUM` | `data_leak` | Log backend mungkin menampilkan data sensitif |
| [`sql-string-format`](#sql-string-format) | `HIGH` | `injection` | Potensi SQL injection — gunakan parameterized query |
| [`hardcoded-credential`](#hardcoded-credential) | `HIGH` | `secret_leak` | Credential hardcoded ditemukan |
| [`unsafe-inner-html`](#unsafe-inner-html) | `HIGH` | `xss` | dangerouslySetInnerHTML ditemukan — pastikan input disanitasi |
| [`dynamic-order`](#dynamic-order) | `HIGH` | `injection` | ORDER BY dinamis harus memakai whitelist |
| [`api-struct-response`](#api-struct-response) | `HIGH` | `data_leak` | Struct sensitif mungkin dikirim langsung ke response |
| [`sensitive-json-field`](#sensitive-json-field) | `HIGH` | `data_leak` | Field sensitif mungkin terekspos dalam JSON |
| [`go-shell-command`](#go-shell-command) | `HIGH` | `command_injection` | Shell command interpreter digunakan melalui os/exec |
| [`go-weak-cryptographic-hash`](#go-weak-cryptographic-hash) | `MEDIUM` | `weak_cryptography` | Algoritma hash kriptografi yang lemah ditemukan |
| [`go-tainted-file-path`](#go-tainted-file-path) | `HIGH` | `path_traversal` | Input request mungkin digunakan langsung sebagai path file |
| [`go-weak-random-secret`](#go-weak-random-secret) | `HIGH` | `insecure_randomness` | Nilai keamanan mungkin dibuat menggunakan math/rand |
| [`javascript-dynamic-eval`](#javascript-dynamic-eval) | `HIGH` | `unsafe_deserialization` | Dynamic eval mungkin mengeksekusi data sebagai kode |

### Details & Guidance

#### `mock-token`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Hardcoded mock token ditemukan — hapus sebelum production

---

#### `browser-token-storage`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Token disimpan di localStorage — gunakan HttpOnly Cookie

---

#### `permission-bypass`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `security_misconfiguration`

**Description**: Permission bypass hardcoded ditemukan

---

#### `weak-secret`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Default atau weak secret ditemukan

---

#### `frontend-sensitive-log`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Log frontend mungkin menampilkan data sensitif

---

#### `backend-sensitive-log`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Log backend mungkin menampilkan data sensitif

---

#### `sql-string-format`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Potensi SQL injection — gunakan parameterized query

---

#### `hardcoded-credential`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `secret_leak`

**Description**: Credential hardcoded ditemukan

---

#### `unsafe-inner-html`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `xss`

**Description**: dangerouslySetInnerHTML ditemukan — pastikan input disanitasi

---

#### `dynamic-order`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: ORDER BY dinamis harus memakai whitelist

---

#### `api-struct-response`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Struct sensitif mungkin dikirim langsung ke response

---

#### `sensitive-json-field`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Field sensitif mungkin terekspos dalam JSON

---

#### `go-shell-command`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `command_injection`

**Description**: Shell command interpreter digunakan melalui os/exec

**Recommendation**: Jalankan executable secara langsung dengan argument array dan validasi input yang tidak dipercaya

---

#### `go-weak-cryptographic-hash`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `weak_cryptography`

**Description**: Algoritma hash kriptografi yang lemah ditemukan

**Recommendation**: Gunakan SHA-256 atau algoritma yang sesuai; gunakan password KDF untuk password

---

#### `go-tainted-file-path`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `path_traversal`

**Description**: Input request mungkin digunakan langsung sebagai path file

**Recommendation**: Normalisasi path, enforce base directory, dan gunakan allowlist identifier

---

#### `go-weak-random-secret`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `insecure_randomness`

**Description**: Nilai keamanan mungkin dibuat menggunakan math/rand

**Recommendation**: Gunakan crypto/rand untuk token, nonce, session identifier, dan secret

---

#### `javascript-dynamic-eval`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `unsafe_deserialization`

**Description**: Dynamic eval mungkin mengeksekusi data sebagai kode

**Recommendation**: Gunakan parser data terstruktur dan validasi schema tanpa evaluasi kode

---

## 🛡️ Hardening Rules

Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`hardcoded-api-url`](#hardcoded-api-url) | `MEDIUM` | `configuration_leak` | URL API hardcoded — gunakan environment variable |
| [`tls-insecure-skip-verify`](#tls-insecure-skip-verify) | `HIGH` | `transport_security` | Verifikasi sertifikat TLS dinonaktifkan |
| [`wildcard-cors-origin`](#wildcard-cors-origin) | `HIGH` | `cors` | Wildcard CORS origin ditemukan |
| [`go-permissive-file-mode`](#go-permissive-file-mode) | `MEDIUM` | `file_permission` | File atau directory dibuat dengan permission world-writable |
| [`debug-mode-enabled`](#debug-mode-enabled) | `MEDIUM` | `debug_configuration` | Debug mode tampak diaktifkan secara eksplisit |
| [`go-insecure-cookie-attribute`](#go-insecure-cookie-attribute) | `HIGH` | `cookie_security` | Cookie memiliki atribut keamanan yang secara eksplisit tidak aman |

### Details & Guidance

#### `hardcoded-api-url`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `configuration_leak`

**Description**: URL API hardcoded — gunakan environment variable

---

#### `tls-insecure-skip-verify`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `transport_security`

**Description**: Verifikasi sertifikat TLS dinonaktifkan

**Recommendation**: Aktifkan certificate verification dan konfigurasi trust store yang sesuai

---

#### `wildcard-cors-origin`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cors`

**Description**: Wildcard CORS origin ditemukan

**Recommendation**: Gunakan allowlist origin yang eksplisit untuk environment terkait

---

#### `go-permissive-file-mode`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `file_permission`

**Description**: File atau directory dibuat dengan permission world-writable

**Recommendation**: Gunakan permission minimum yang diperlukan, misalnya 0600 atau 0750

---

#### `debug-mode-enabled`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `debug_configuration`

**Description**: Debug mode tampak diaktifkan secara eksplisit

**Recommendation**: Nonaktifkan debug mode pada konfigurasi deployment production

---

#### `go-insecure-cookie-attribute`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cookie_security`

**Description**: Cookie memiliki atribut keamanan yang secara eksplisit tidak aman

**Recommendation**: Aktifkan Secure dan HttpOnly serta gunakan kebijakan SameSite yang sesuai

---

## ⚡ Reliability Rules

Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`go-multipart-memory`](#go-multipart-memory) | `MEDIUM` | `resource_exhaustion` | Pastikan request multipart memiliki batas ukuran |
| [`go-http-default-server`](#go-http-default-server) | `MEDIUM` | `missing_timeout` | Default HTTP server tidak mengonfigurasi timeout defensif |
| [`go-unbounded-request-read`](#go-unbounded-request-read) | `MEDIUM` | `resource_exhaustion` | Request body mungkin dibaca tanpa batas ukuran |
| [`go-discarded-error`](#go-discarded-error) | `MEDIUM` | `error_handling` | Return value error mungkin dibuang secara eksplisit |
| [`go-process-termination`](#go-process-termination) | `MEDIUM` | `process_termination` | Application path mungkin menghentikan seluruh process |
| [`go-http-client-without-timeout`](#go-http-client-without-timeout) | `MEDIUM` | `missing_timeout` | HTTP client literal tidak menetapkan timeout keseluruhan |

### Details & Guidance

#### `go-multipart-memory`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Pastikan request multipart memiliki batas ukuran

---

#### `go-http-default-server`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: Default HTTP server tidak mengonfigurasi timeout defensif

**Recommendation**: Gunakan http.Server dengan ReadHeaderTimeout, ReadTimeout, WriteTimeout, dan IdleTimeout

---

#### `go-unbounded-request-read`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Request body mungkin dibaca tanpa batas ukuran

**Recommendation**: Batasi body dengan http.MaxBytesReader atau io.LimitReader sebelum membacanya

---

#### `go-discarded-error`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `error_handling`

**Description**: Return value error mungkin dibuang secara eksplisit

**Recommendation**: Periksa dan tangani error, atau dokumentasikan alasan aman untuk mengabaikannya

---

#### `go-process-termination`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `process_termination`

**Description**: Application path mungkin menghentikan seluruh process

**Recommendation**: Propagasikan error ke boundary dan lakukan shutdown terkontrol

---

#### `go-http-client-without-timeout`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: HTTP client literal tidak menetapkan timeout keseluruhan

**Recommendation**: Tetapkan http.Client.Timeout dan timeout transport yang sesuai

---

## 🧹 Quality Rules

Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`merge-conflict-marker`](#merge-conflict-marker) | `HIGH` | `repository_hygiene` | Unresolved merge-conflict marker ditemukan |
| [`javascript-debugger`](#javascript-debugger) | `MEDIUM` | `debug_code` | JavaScript debugger statement ditemukan |
| [`trailing-whitespace`](#trailing-whitespace) | `LOW` | `formatting` | Trailing whitespace ditemukan |
| [`mixed-indentation`](#mixed-indentation) | `LOW` | `formatting` | Tab dan spasi tercampur pada indentation baris yang sama |
| [`javascript-console-debug`](#javascript-console-debug) | `LOW` | `debug_code` | Console debug statement mungkin tertinggal |

### Details & Guidance

#### `merge-conflict-marker`

- **Domain**: `quality`
- **Severity**: `HIGH`
- **Category**: `repository_hygiene`

**Description**: Unresolved merge-conflict marker ditemukan

**Recommendation**: Selesaikan conflict dan hapus seluruh marker sebelum commit

---

#### `javascript-debugger`

- **Domain**: `quality`
- **Severity**: `MEDIUM`
- **Category**: `debug_code`

**Description**: JavaScript debugger statement ditemukan

**Recommendation**: Hapus debugger statement sebelum commit

---

#### `trailing-whitespace`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Trailing whitespace ditemukan

**Recommendation**: Hapus whitespace pada akhir baris

---

#### `mixed-indentation`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Tab dan spasi tercampur pada indentation baris yang sama

**Recommendation**: Gunakan satu gaya indentation yang konsisten

---

#### `javascript-console-debug`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `debug_code`

**Description**: Console debug statement mungkin tertinggal

**Recommendation**: Hapus statement debug atau gunakan logger aplikasi dengan level yang sesuai

---

## 📜 Governance Rules

Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`privacy-pii-log`](#privacy-pii-log) | `HIGH` | `privacy_log` | Logging statement may expose personally identifiable information |
| [`privacy-pii-url`](#privacy-pii-url) | `HIGH` | `privacy_url` | Personally identifiable information may be placed in a URL query string |
| [`privacy-pii-fixture`](#privacy-pii-fixture) | `MEDIUM` | `privacy_fixture` | Fixture may contain a literal personally identifiable value |
| [`privacy-sensitive-response`](#privacy-sensitive-response) | `HIGH` | `privacy_response` | Response construction may expose a sensitive personal field |

### Details & Guidance

#### `privacy-pii-log`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_log`

**Description**: Logging statement may expose personally identifiable information

**Recommendation**: Remove the PII field or log a non-reversible, access-controlled reference identifier

---

#### `privacy-pii-url`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_url`

**Description**: Personally identifiable information may be placed in a URL query string

**Recommendation**: Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs

---

#### `privacy-pii-fixture`

- **Domain**: `governance`
- **Severity**: `MEDIUM`
- **Category**: `privacy_fixture`

**Description**: Fixture may contain a literal personally identifiable value

**Recommendation**: Use clearly synthetic, reserved test data and keep production-derived records out of the repository

---

#### `privacy-sensitive-response`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_response`

**Description**: Response construction may expose a sensitive personal field

**Recommendation**: Map the response through an explicit allowlisted DTO and omit sensitive fields

---

