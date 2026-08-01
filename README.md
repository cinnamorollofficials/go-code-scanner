# Go Code Scanner

`go-code-scanner` adalah library Go sekaligus CLI untuk melakukan pemeriksaan keamanan statis berbasis pola pada source code. Scanner dapat memeriksa seluruh proyek, hanya perubahan terhadap `HEAD`, atau isi Git staging area; hasilnya dinormalisasi, dideduplikasi, disanitasi, lalu ditulis sebagai laporan JSON yang stabil untuk dipakai secara lokal maupun di CI.

> Proyek ini sedang diekstrak dari HINT Core. Library ini menangani orkestrasi scanner dan format finding, sedangkan kebijakan, rule tambahan, suppression, path, serta konfigurasi CI tetap dimiliki oleh repository yang menggunakannya.

## Fitur utama

- 14 rule keamanan bawaan untuk mendeteksi indikasi secret leak, SQL injection, XSS, logging data sensitif, permission bypass, dan masalah lain.
- Mendukung file Go, TypeScript, JavaScript, YAML, dan JSON secara default.
- Tiga mode pemindaian: seluruh proyek (`full`), perubahan working tree (`changed`), dan isi staging area (`staged`).
- Rule regex tambahan melalui file JSON.
- Suppression yang audit-friendly, wajib memiliki alasan dan tanggal kedaluwarsa.
- Pemindaian paralel dengan jumlah worker yang dapat dikonfigurasi.
- Snippet yang berpotensi mengandung kredensial otomatis disensor.
- Output terminal ringkas dan laporan JSON berversi (`schema_version: "1.0"`).
- API ekstensi untuk menambahkan scanner lain ke dalam pipeline yang sama.
- Hanya menggunakan Go standard library.

## Persyaratan

- Go 1.25 atau lebih baru, sesuai deklarasi pada `go.mod`.
- Git tersedia di `PATH` jika menggunakan mode `--changed` atau `--staged`.

## Instalasi

Jalankan langsung dari source:

```sh
go run ./cmd/security-review scan --root /path/to/project
```

Atau build binary lokal:

```sh
go build -o security-review ./cmd/security-review
./security-review scan --root /path/to/project
```

Untuk menggunakan library dari module Go lain:

```sh
go get github.com/cinnamorollofficials/go-code-scanner
```

## Mulai cepat

Scan seluruh file yang didukung di direktori saat ini:

```sh
security-review scan
```

Scan proyek tertentu dan simpan hasil ke lokasi khusus:

```sh
security-review scan \
  --root /path/to/project \
  --output artifacts/security-findings.json
```

Scan perubahan working tree terhadap `HEAD`:

```sh
security-review scan --root /path/to/project --changed
```

Scan konten yang benar-benar sudah masuk staging area dan aktifkan policy CI:

```sh
security-review scan --root /path/to/project --staged --ci --fail-on high
```

`--changed` dan `--staged` tidak dapat digunakan bersamaan. Mode `changed` membaca file dari working tree, sedangkan mode `staged` membaca blob dari Git index; karena itu hasil staged tetap sesuai dengan konten yang akan di-commit meskipun working tree sudah berubah lagi.

## Perintah CLI

```text
security-review scan [options]
security-review config validate <path>
security-review version
security-review help
```

Tanpa subcommand, CLI menjalankan `scan` dengan konfigurasi default.

### Opsi `scan`

| Opsi | Default | Keterangan |
| --- | --- | --- |
| `--root <path>` | `.` | Root proyek yang dipindai tanpa file konfigurasi. |
| `--config <path>` | kosong | Muat konfigurasi JSON. Jika diberikan, root berasal dari konfigurasi. |
| `--output <path>` | `security_findings.json` | Path laporan JSON; path relatif dihitung dari root proyek. |
| `--changed` | `false` | Scan file yang berubah terhadap `HEAD` (`git diff HEAD`). |
| `--staged` | `false` | Scan file yang sudah di-stage (`git diff --cached`). |
| `--ci` | `false` | Kembalikan exit code `1` jika finding mencapai ambang batas. |
| `--fail-on <severity>` | `critical` | Ambang CI: `critical`, `high`, `medium`, atau `low`. |
| `--quiet` | `false` | Jangan cetak ringkasan terminal; laporan JSON tetap dibuat. |

Flag CLI untuk mode, output, dan ambang severity menimpa nilai yang berasal dari file konfigurasi.

### Exit code

| Kode | Arti |
| --- | --- |
| `0` | Scan selesai dan policy CI tidak dilanggar, atau `--ci` tidak diaktifkan. |
| `1` | Scan selesai, tetapi ada finding aktif pada/di atas ambang `--fail-on`. |
| `2` | Argumen atau konfigurasi tidak valid. |
| `3` | Kegagalan operasional, misalnya discovery, scanner wajib, atau penulisan laporan gagal. |

Finding tidak membuat command gagal kecuali `--ci` digunakan. Finding yang sudah disuppress juga tidak dihitung sebagai pelanggaran policy.

## Konfigurasi

Contoh `security-review.json`:

```json
{
  "version": 1,
  "project": "my-service",
  "root": ".",
  "mode": "full",
  "output": "artifacts/security_findings.json",
  "fail_on": "HIGH",
  "include_extensions": [".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"],
  "exclude_directories": [".git", "node_modules", "vendor", "dist", "build"],
  "exclude_files": ["security_findings.json", "package-lock.json"],
  "rule_files": ["security-rules.json"],
  "suppression_file": ".security-ignore",
  "workers": 4,
  "scanners": {
    "pattern": {
      "enabled": true,
      "required": true
    }
  }
}
```

Validasi konfigurasi sebelum menjalankan scan:

```sh
security-review config validate security-review.json
security-review scan --config security-review.json
```

Path `root` yang relatif dihitung dari direktori file konfigurasi. `rule_files`, `suppression_file`, dan `output` yang relatif kemudian dihitung dari root tersebut.

Nilai default penting:

- `mode`: `full`
- `fail_on`: `CRITICAL`
- `workers`: nilai `GOMAXPROCS`
- ekstensi: `.go`, `.ts`, `.tsx`, `.js`, `.jsx`, `.yaml`, `.yml`, `.json`
- direktori yang dilewati: `.git`, `node_modules`, `vendor`, `dist`, `build`, `.next`, `out`, `bin`
- file yang dilewati: `security_findings.json`, `package-lock.json`

Catatan: konfigurasi scanner saat ini digunakan untuk menentukan apakah kegagalan scanner bersifat wajib (`required`). Pattern scanner bawaan selalu didaftarkan oleh library.

## Rule tambahan

Rule tambahan tidak menggantikan rule bawaan; semuanya digabung dan setiap ID harus unik. Regex menggunakan sintaks [RE2 Go](https://pkg.go.dev/regexp) dan dikompilasi secara case-insensitive. Pencocokan dilakukan per baris.

Contoh `security-rules.json`:

```json
{
  "version": 1,
  "rules": [
    {
      "id": "private-key-header",
      "pattern": "-----BEGIN (RSA |EC )?PRIVATE KEY-----",
      "severity": "CRITICAL",
      "category": "secret_leak",
      "description": "Private key ditemukan di source code",
      "recommendation": "Pindahkan key ke secret manager dan lakukan rotasi",
      "extensions": [".pem", ".key"]
    }
  ]
}
```

Pastikan ekstensi rule juga ada di `include_extensions`; jika tidak, file tersebut tidak akan ditemukan pada tahap discovery. Sebuah rule dapat dinonaktifkan dengan menambahkan `"enabled": false` pada definisinya.

Severity yang valid adalah `CRITICAL`, `HIGH`, `MEDIUM`, dan `LOW`.

## Suppression

Suppression disimpan sebagai JSON (secara default `.security-ignore`). Setiap entri wajib memiliki `file`, `reason`, dan `expires`. Gunakan `rule_id`, `fingerprint`, dan `line` untuk mempersempit kecocokan; nilai `line: -1` berlaku untuk semua baris pada file yang cocok.

```json
{
  "version": 1,
  "suppressions": [
    {
      "rule_id": "hardcoded-api-url",
      "file": "internal/example/config.go",
      "line": 18,
      "reason": "Endpoint lokal hanya digunakan oleh fixture pengujian",
      "approved_by": "security-team",
      "expires": "2026-12-31",
      "ticket": "SEC-123"
    }
  ]
}
```

Fingerprint tersedia pada laporan JSON dan lebih stabil daripada nomor baris. Suppression yang sudah kedaluwarsa tidak menyembunyikan finding; file terkait dicatat dalam `stale_suppression_files` dan ringkasan `stale_suppressions`.

## Format laporan

Laporan selalu ditulis, termasuk ketika policy CI gagal. Contoh ringkas:

```json
{
  "schema_version": "1.0",
  "timestamp": "2026-08-01T10:00:00Z",
  "scan_mode": "full",
  "project": "my-service",
  "summary": {
    "total": 1,
    "critical": 0,
    "high": 1,
    "medium": 0,
    "low": 0,
    "suppressed": 0,
    "stale_suppressions": 0
  },
  "findings": [
    {
      "id": "F-0001",
      "fingerprint": "0123456789abcdef",
      "rule_id": "sql-string-format",
      "tool": "pattern",
      "category": "injection",
      "severity": "HIGH",
      "description": "Potensi SQL injection — gunakan parameterized query",
      "location": { "file": "internal/store.go", "line": 42 },
      "suppressed": false
    }
  ],
  "scanner_status": [
    { "id": "pattern", "state": "findings", "required": true, "duration_ns": 1200000 }
  ]
}
```

Finding diurutkan dari severity tertinggi, lalu path file dan nomor baris. Duplikat dihapus berdasarkan rule, file, baris, dan deskripsi. Baris dari kategori `secret_leak` selalu diganti dengan placeholder; baris lain yang tampak sensitif juga disensor, dan snippet biasa dibatasi hingga 200 karakter.

Penulisan laporan dilakukan melalui temporary file lalu rename. Jika target sudah ada, implementasi menjaga file lama sebagai backup sementara selama proses penggantian.

## Penggunaan sebagai library

```go
package main

import (
    "context"
    "log"

    securityreview "github.com/cinnamorollofficials/go-code-scanner"
    "github.com/cinnamorollofficials/go-code-scanner/config"
)

func main() {
    cfg := config.Default()
    cfg.Project = "my-service"
    cfg.Root = "/path/to/project"

    reviewer, err := securityreview.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    report, err := reviewer.Run(context.Background())
    if err != nil {
        // Report dapat tetap tersedia ketika scanner mengalami kegagalan operasional.
        log.Printf("scan completed with an operational error: %v", err)
    }
    if report != nil {
        log.Printf("active findings: %d", report.Summary.Total)
    }
}
```

Scanner eksternal dapat ditambahkan dengan mengimplementasikan interface `scanner.Scanner` dan meneruskannya melalui `securityreview.WithScanner(...)` atau `securityreview.WithRequiredScanner(...)`. Scanner opsional yang gagal menghasilkan warning; scanner wajib yang gagal membuat `Run` mengembalikan error beserta report parsial.

## Alur kerja internal

```text
Config → file discovery → pattern/custom scanners → normalisasi & deduplikasi
       → suppression → summary/report → terminal + JSON → policy CI
```

Paket utama:

- `config`: default, loading, resolusi path, dan validasi konfigurasi.
- `discovery`: pemilihan file dari filesystem atau Git.
- `rules`: rule bawaan, loading rule JSON, dan kompilasi regex.
- `scanner/pattern`: eksekusi rule secara paralel dan redaksi snippet.
- `suppression`: pencocokan pengecualian dan deteksi suppression kedaluwarsa.
- `finding`: model finding, severity, status scanner, dan report.
- `reporter`: output terminal dan penulisan JSON atomik.
- `policy`: evaluasi ambang severity untuk CI.

## Integrasi CI

Contoh GitHub Actions yang memindai seluruh checkout dan mengunggah laporan walaupun policy gagal:

```yaml
name: security-review

on:
  pull_request:

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.25.x"

      - name: Run security review
        run: go run ./cmd/security-review scan --ci --fail-on high

      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: security-findings
          path: security_findings.json
```

Jika memakai `--changed` di CI, pastikan checkout memiliki riwayat dan referensi `HEAD` yang dibutuhkan oleh Git. Untuk pull request lintas commit, semantics bawaan tetap `git diff HEAD`; pipeline yang membutuhkan merge-base khusus sebaiknya menyiapkan staging/diff yang sesuai atau menggunakan mode penuh.

## Pengembangan

```sh
go test ./...
go vet ./...
```

Struktur proyek sengaja modular agar discovery, rules, suppression, reporter, dan policy dapat diuji secara terpisah. Saat menambahkan perilaku baru, sertakan test pada package terkait dan pertahankan kompatibilitas schema laporan.

Rencana pengembangan bertahap menuju commit gate untuk Quality, Reliability, Hardening, Security, Supply Chain, dan Governance tersedia di [TODO.md](TODO.md).

## Batasan saat ini

- Engine bawaan adalah pencocokan regex per baris, bukan parser AST atau taint analysis; hasil perlu ditinjau dan false positive mungkin terjadi.
- Mode changed/staged hanya menyertakan file berstatus added, copied, modified, atau renamed (`ACMR`), bukan file yang dihapus.
- File dengan satu baris di atas 1 MiB dapat menyebabkan pattern scanner berstatus partial.
- Hanya format konfigurasi, rule, dan suppression JSON yang didukung pada versi awal.
- Rule bawaan berfokus pada pola yang umum di Go dan ekosistem JavaScript/TypeScript, bukan cakupan lengkap seluruh kelas kerentanan.

## Kontribusi

Issue dan pull request dipersilakan. Sebelum mengirim perubahan, jalankan `go test ./...` dan `go vet ./...`, lalu jelaskan perubahan perilaku atau schema yang relevan di pull request.
