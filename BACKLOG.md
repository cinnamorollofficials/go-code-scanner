# Go Code Scanner - Implementation Backlog & Roadmap

Dokumen ini melacak daftar tugas, fitur, dan rule engine lanjutan yang belum diimplementasikan untuk pengembangan `go-code-scanner` berikutnya.

---

## 📌 Prioritas 2 (P2 - Mid-Term Engine & Rule Enhancements)

### 1. 🔍 Analisis Aliran Data Lintas Fungsi (*Interprocedural Taint & Call Graph*)
- [ ] **Interprocedural Call Graph Engine**:
  - Bangun graph panggilan fungsi lintas fungsi dan lintas package dalam workspace Go.
  - Lacak aliran parameter yang melalui helper functions, constructor, dan DTO/struct fields menuju sink database.
- [ ] **Model Jangkauan Router HTTP Otomatis**:
  - Deteksi otomatis handler entry-point dari framework Go populer (`gin.Context`, `chi.URLParam`, `fiber.Ctx`, `echo.Context`, `gorilla/mux`).
  - Tandai parameter rute/query/body sebagai `SourceStep` terverifikasi.

### 2. 🛡️ Taksonomi Rule Otorisasi & Multi-Tenant (`SQLAUTH`)
- [ ] **`SQLAUTH-001` (Missing Tenant Constraint)**:
  - Deteksi query pada model data multi-tenant yang lupa menyertakan filter `tenant_id` atau scoping repository.
- [ ] **`SQLAUTH-002` (Insecure Direct Object Reference / IDOR)**:
  - Deteksi pemanggilan data sensitif melalui ID input pengguna tanpa validasi kepemilikan akun (*ownership scope*).
- [ ] **`SQLAUTH-003` (Auth Filter Dropped in Raw Queries)**:
  - Deteksi query raw SQL pengganti ORM yang menghilangkan filter izin atau batasan role pengguna.
- [ ] **`SQLAUTH-004` (Row-Level Security Assumption Mismatch)**:
  - Deteksi kode yang berasumsi database memberlakukan RLS (Row-Level Security) padahal koneksi menggunakan role superuser/bypass.

### 3. ⚡ Taksonomi Rule Integritas & Transaksi (`SQLSAFE`)
- [ ] **`SQLSAFE-003` (Non-Atomic Read-Modify-Write)**:
  - Deteksi perubahan status kritis (keuangan, poin, kuota, inventaris) yang dijalankan tanpa transaksi atau row lock (`SELECT FOR UPDATE`).
- [ ] **`SQLSAFE-004` (Transaction Boundary Loss)**:
  - Deteksi goroutine asinkron atau koneksi database sekunder yang lepas dari batas transaksi aktif.
- [ ] **`SQLSAFE-005` (Incorrect AND/OR Precedence)**:
  - Deteksi kesalahan pengelompokan operator logika `AND`/`OR` pada query builder yang dapat membocorkan baris data antar-tenant.
- [ ] **`SQLSAFE-006` (Soft-Delete Bypass)**:
  - Deteksi query raw SQL yang lupa menyertakan kondisi `deleted_at IS NULL`.

### 4. 🛠️ CLI Automated Remediation (`--fix`)
- [ ] **`security-review scan --fix`**:
  - Implementasi safe AST code rewriting untuk mengubah konkatenasi string SQL rentan (`SQLI-001`) menjadi parameterized query secara otomatis.

---

## 📌 Prioritas 3 (P3 - Long-Term & Ecosystem Expansion)

### 1. 🌐 Framework Models Eksternal (Multi-Bahasa)
- [ ] **Node.js / TypeScript Ecosystem**:
  - Model analisis taint untuk Prisma ORM, TypeORM, Sequelize, `pg` (node-postgres), dan `mysql2`.
- [ ] **Python Ecosystem**:
  - Model analisis taint untuk SQLAlchemy, Django ORM, psycopg2/psycopg3.
- [ ] **Java / Kotlin Ecosystem**:
  - Model analisis taint untuk Spring Data JPA, Hibernate, JDBC templates.

### 2. 🗄️ Database Migration & Schema Safety (`DBMIG`)
- [ ] **`DBMIG-001` (Destructive Migration Without Guarded Rollout)**:
  - Deteksi operasi `DROP TABLE`, `DROP COLUMN`, atau pengecilan tipe data tanpa pola expand-contract.
- [ ] **`DBMIG-002` (Irreversible Migration)**:
  - Peringatan jika file migrasi database tidak menyediakan skrip rollback `down`.
- [ ] **`DBMIG-003` (Constraint Gap on Security Keys)**:
  - Deteksi ketiadaan constraint unique/foreign key pada kolom ID tenant atau kunci relasi penting.

### 3. 📊 Performance & Privacy Guardrails (`DBPERF` & `DBSEC`)
- [ ] **`DBPERF-001` (Unbounded Result Set)**:
  - Deteksi query endpoint publik yang tidak membatasi jumlah baris (`LIMIT` / pagination).
- [ ] **`DBPERF-002` (N+1 Query in Externally Controlled Loops)**:
  - Deteksi eksekusi query database di dalam perulangan loop slice/array.
- [ ] **`DBSEC-002` (Sensitive Data Logged)**:
  - Penelusuran data kredensial, token, atau kolom PII ke dalam sink logger / tracing.
- [ ] **`DBSEC-003` (Database Error Exposed to Untrusted Client)**:
  - Penelusuran error internal driver database yang diteruskan mentah-mentah ke respons HTTP.

---

## 📋 Status Ringkasan

| Tahap | Fokus Utama | Status |
| :--- | :--- | :---: |
| **Phase 1 (MVP)** | AST Go Parser, Intraprocedural Taint, `SQLI-001`, `SQLI-002`, `SQLI-004`, `SQLI-008`, `SQLSAFE-001`, AI Agent Skill |  **Completed** |
| **Phase 2 (P2)** | Interprocedural Call Graph, `SQLAUTH-001..004`, `SQLSAFE-003..006`, `--fix` Auto-Remediation | ⏳ **Next Up** |
| **Phase 3 (P3)** | Multi-Language Adapters (Node/Python/Java), `DBMIG`, `DBPERF`, `DBSEC` | 📅 **Planned** |
