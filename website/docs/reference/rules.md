---
title: Rule Catalog
description: Complete catalog of default built-in security, secret, governance, and quality rules.
---

# Built-In Rule Catalog

Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog is automatically generated from Go rule registries.

| Rule ID | Domain | Severity | Category | Description |
| :--- | :--- | :--- | :--- | :--- |
| `merge-conflict-marker` | `quality` | `HIGH` | `repository_hygiene` | Unresolved merge-conflict marker ditemukan |
| `javascript-debugger` | `quality` | `MEDIUM` | `debug_code` | JavaScript debugger statement ditemukan |
| `trailing-whitespace` | `quality` | `LOW` | `formatting` | Trailing whitespace ditemukan |
| `mixed-indentation` | `quality` | `LOW` | `formatting` | Tab dan spasi tercampur pada indentation baris yang sama |
| `javascript-console-debug` | `quality` | `LOW` | `debug_code` | Console debug statement mungkin tertinggal |
| `mock-token` | `security` | `CRITICAL` | `secret_leak` | Hardcoded mock token ditemukan — hapus sebelum production |
| `browser-token-storage` | `security` | `HIGH` | `data_leak` | Token disimpan di localStorage — gunakan HttpOnly Cookie |
| `permission-bypass` | `security` | `CRITICAL` | `security_misconfiguration` | Permission bypass hardcoded ditemukan |
| `weak-secret` | `security` | `CRITICAL` | `secret_leak` | Default atau weak secret ditemukan |
| `frontend-sensitive-log` | `security` | `MEDIUM` | `data_leak` | Log frontend mungkin menampilkan data sensitif |
| `backend-sensitive-log` | `security` | `MEDIUM` | `data_leak` | Log backend mungkin menampilkan data sensitif |
| `sql-string-format` | `security` | `HIGH` | `injection` | Potensi SQL injection — gunakan parameterized query |
| `hardcoded-credential` | `security` | `HIGH` | `secret_leak` | Credential hardcoded ditemukan |
| `unsafe-inner-html` | `security` | `HIGH` | `xss` | dangerouslySetInnerHTML ditemukan — pastikan input disanitasi |
| `dynamic-order` | `security` | `HIGH` | `injection` | ORDER BY dinamis harus memakai whitelist |
| `api-struct-response` | `security` | `HIGH` | `data_leak` | Struct sensitif mungkin dikirim langsung ke response |
| `sensitive-json-field` | `security` | `HIGH` | `data_leak` | Field sensitif mungkin terekspos dalam JSON |
| `go-shell-command` | `security` | `HIGH` | `command_injection` | Shell command interpreter digunakan melalui os/exec |
| `go-weak-cryptographic-hash` | `security` | `MEDIUM` | `weak_cryptography` | Algoritma hash kriptografi yang lemah ditemukan |
| `go-tainted-file-path` | `security` | `HIGH` | `path_traversal` | Input request mungkin digunakan langsung sebagai path file |
| `go-weak-random-secret` | `security` | `HIGH` | `insecure_randomness` | Nilai keamanan mungkin dibuat menggunakan math/rand |
| `javascript-dynamic-eval` | `security` | `HIGH` | `unsafe_deserialization` | Dynamic eval mungkin mengeksekusi data sebagai kode |
| `hardcoded-api-url` | `hardening` | `MEDIUM` | `configuration_leak` | URL API hardcoded — gunakan environment variable |
| `tls-insecure-skip-verify` | `hardening` | `HIGH` | `transport_security` | Verifikasi sertifikat TLS dinonaktifkan |
| `wildcard-cors-origin` | `hardening` | `HIGH` | `cors` | Wildcard CORS origin ditemukan |
| `go-permissive-file-mode` | `hardening` | `MEDIUM` | `file_permission` | File atau directory dibuat dengan permission world-writable |
| `debug-mode-enabled` | `hardening` | `MEDIUM` | `debug_configuration` | Debug mode tampak diaktifkan secara eksplisit |
| `go-insecure-cookie-attribute` | `hardening` | `HIGH` | `cookie_security` | Cookie memiliki atribut keamanan yang secara eksplisit tidak aman |
| `go-multipart-memory` | `reliability` | `MEDIUM` | `resource_exhaustion` | Pastikan request multipart memiliki batas ukuran |
| `go-http-default-server` | `reliability` | `MEDIUM` | `missing_timeout` | Default HTTP server tidak mengonfigurasi timeout defensif |
| `go-unbounded-request-read` | `reliability` | `MEDIUM` | `resource_exhaustion` | Request body mungkin dibaca tanpa batas ukuran |
| `go-discarded-error` | `reliability` | `MEDIUM` | `error_handling` | Return value error mungkin dibuang secara eksplisit |
| `go-process-termination` | `reliability` | `MEDIUM` | `process_termination` | Application path mungkin menghentikan seluruh process |
| `go-http-client-without-timeout` | `reliability` | `MEDIUM` | `missing_timeout` | HTTP client literal tidak menetapkan timeout keseluruhan |
| `privacy-pii-log` | `governance` | `HIGH` | `privacy_log` | Logging statement may expose personally identifiable information |
| `privacy-pii-url` | `governance` | `HIGH` | `privacy_url` | Personally identifiable information may be placed in a URL query string |
| `privacy-pii-fixture` | `governance` | `MEDIUM` | `privacy_fixture` | Fixture may contain a literal personally identifiable value |
| `privacy-sensitive-response` | `governance` | `HIGH` | `privacy_response` | Response construction may expose a sensitive personal field |


## Rule Details & Guidance

### `merge-conflict-marker`

- **Domain**: `quality`
- **Severity**: `HIGH`
- **Category**: `repository_hygiene`

**Description**: Unresolved merge-conflict marker ditemukan

**Recommendation**: Selesaikan conflict dan hapus seluruh marker sebelum commit

---

### `javascript-debugger`

- **Domain**: `quality`
- **Severity**: `MEDIUM`
- **Category**: `debug_code`

**Description**: JavaScript debugger statement ditemukan

**Recommendation**: Hapus debugger statement sebelum commit

---

### `trailing-whitespace`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Trailing whitespace ditemukan

**Recommendation**: Hapus whitespace pada akhir baris

---

### `mixed-indentation`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Tab dan spasi tercampur pada indentation baris yang sama

**Recommendation**: Gunakan satu gaya indentation yang konsisten

---

### `javascript-console-debug`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `debug_code`

**Description**: Console debug statement mungkin tertinggal

**Recommendation**: Hapus statement debug atau gunakan logger aplikasi dengan level yang sesuai

---

### `mock-token`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Hardcoded mock token ditemukan — hapus sebelum production

**Recommendation**: 

---

### `browser-token-storage`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Token disimpan di localStorage — gunakan HttpOnly Cookie

**Recommendation**: 

---

### `permission-bypass`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `security_misconfiguration`

**Description**: Permission bypass hardcoded ditemukan

**Recommendation**: 

---

### `weak-secret`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Default atau weak secret ditemukan

**Recommendation**: 

---

### `frontend-sensitive-log`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Log frontend mungkin menampilkan data sensitif

**Recommendation**: 

---

### `backend-sensitive-log`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Log backend mungkin menampilkan data sensitif

**Recommendation**: 

---

### `sql-string-format`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Potensi SQL injection — gunakan parameterized query

**Recommendation**: 

---

### `hardcoded-credential`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `secret_leak`

**Description**: Credential hardcoded ditemukan

**Recommendation**: 

---

### `unsafe-inner-html`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `xss`

**Description**: dangerouslySetInnerHTML ditemukan — pastikan input disanitasi

**Recommendation**: 

---

### `dynamic-order`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: ORDER BY dinamis harus memakai whitelist

**Recommendation**: 

---

### `api-struct-response`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Struct sensitif mungkin dikirim langsung ke response

**Recommendation**: 

---

### `sensitive-json-field`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Field sensitif mungkin terekspos dalam JSON

**Recommendation**: 

---

### `go-shell-command`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `command_injection`

**Description**: Shell command interpreter digunakan melalui os/exec

**Recommendation**: Jalankan executable secara langsung dengan argument array dan validasi input yang tidak dipercaya

---

### `go-weak-cryptographic-hash`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `weak_cryptography`

**Description**: Algoritma hash kriptografi yang lemah ditemukan

**Recommendation**: Gunakan SHA-256 atau algoritma yang sesuai; gunakan password KDF untuk password

---

### `go-tainted-file-path`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `path_traversal`

**Description**: Input request mungkin digunakan langsung sebagai path file

**Recommendation**: Normalisasi path, enforce base directory, dan gunakan allowlist identifier

---

### `go-weak-random-secret`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `insecure_randomness`

**Description**: Nilai keamanan mungkin dibuat menggunakan math/rand

**Recommendation**: Gunakan crypto/rand untuk token, nonce, session identifier, dan secret

---

### `javascript-dynamic-eval`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `unsafe_deserialization`

**Description**: Dynamic eval mungkin mengeksekusi data sebagai kode

**Recommendation**: Gunakan parser data terstruktur dan validasi schema tanpa evaluasi kode

---

### `hardcoded-api-url`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `configuration_leak`

**Description**: URL API hardcoded — gunakan environment variable

**Recommendation**: 

---

### `tls-insecure-skip-verify`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `transport_security`

**Description**: Verifikasi sertifikat TLS dinonaktifkan

**Recommendation**: Aktifkan certificate verification dan konfigurasi trust store yang sesuai

---

### `wildcard-cors-origin`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cors`

**Description**: Wildcard CORS origin ditemukan

**Recommendation**: Gunakan allowlist origin yang eksplisit untuk environment terkait

---

### `go-permissive-file-mode`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `file_permission`

**Description**: File atau directory dibuat dengan permission world-writable

**Recommendation**: Gunakan permission minimum yang diperlukan, misalnya 0600 atau 0750

---

### `debug-mode-enabled`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `debug_configuration`

**Description**: Debug mode tampak diaktifkan secara eksplisit

**Recommendation**: Nonaktifkan debug mode pada konfigurasi deployment production

---

### `go-insecure-cookie-attribute`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cookie_security`

**Description**: Cookie memiliki atribut keamanan yang secara eksplisit tidak aman

**Recommendation**: Aktifkan Secure dan HttpOnly serta gunakan kebijakan SameSite yang sesuai

---

### `go-multipart-memory`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Pastikan request multipart memiliki batas ukuran

**Recommendation**: 

---

### `go-http-default-server`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: Default HTTP server tidak mengonfigurasi timeout defensif

**Recommendation**: Gunakan http.Server dengan ReadHeaderTimeout, ReadTimeout, WriteTimeout, dan IdleTimeout

---

### `go-unbounded-request-read`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Request body mungkin dibaca tanpa batas ukuran

**Recommendation**: Batasi body dengan http.MaxBytesReader atau io.LimitReader sebelum membacanya

---

### `go-discarded-error`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `error_handling`

**Description**: Return value error mungkin dibuang secara eksplisit

**Recommendation**: Periksa dan tangani error, atau dokumentasikan alasan aman untuk mengabaikannya

---

### `go-process-termination`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `process_termination`

**Description**: Application path mungkin menghentikan seluruh process

**Recommendation**: Propagasikan error ke boundary dan lakukan shutdown terkontrol

---

### `go-http-client-without-timeout`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: HTTP client literal tidak menetapkan timeout keseluruhan

**Recommendation**: Tetapkan http.Client.Timeout dan timeout transport yang sesuai

---

### `privacy-pii-log`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_log`

**Description**: Logging statement may expose personally identifiable information

**Recommendation**: Remove the PII field or log a non-reversible, access-controlled reference identifier

---

### `privacy-pii-url`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_url`

**Description**: Personally identifiable information may be placed in a URL query string

**Recommendation**: Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs

---

### `privacy-pii-fixture`

- **Domain**: `governance`
- **Severity**: `MEDIUM`
- **Category**: `privacy_fixture`

**Description**: Fixture may contain a literal personally identifiable value

**Recommendation**: Use clearly synthetic, reserved test data and keep production-derived records out of the repository

---

### `privacy-sensitive-response`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_response`

**Description**: Response construction may expose a sensitive personal field

**Recommendation**: Map the response through an explicit allowlisted DTO and omit sensitive fields

---

