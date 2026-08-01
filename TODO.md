# Implementation Roadmap

Roadmap ini mengarahkan `go-code-scanner` menjadi **policy-driven commit gate** untuk enam domain pemeriksaan:

1. Quality
2. Reliability
3. Hardening
4. Security
5. Supply Chain
6. Governance

Urutan di bawah sengaja dibuat bertahap. Tahap berikutnya boleh dimulai setelah acceptance criteria tahap sebelumnya terpenuhi. Checklist kode menggunakan path package yang ada atau path baru yang direncanakan.

## Sasaran produk

```text
git commit
    -> pre-commit (fast profile, staged content)
    -> quality + reliability + hardening + security
    -> allow/block commit

git push
    -> pre-push (standard profile)
    -> tests + architecture + supply chain
    -> allow/block push

CI
    -> full profile
    -> seluruh domain + report JSON/SARIF/JUnit
```

Target nonfungsional:

- Pemeriksaan pre-commit normal selesai kurang dari 2 detik; maksimal 5 detik saat memakai tool eksternal.
- Pemeriksaan staged selalu membaca konten Git index, bukan perubahan unstaged.
- Eksekusi offline dan deterministik secara default.
- Tidak ada source snippet sensitif yang bocor ke terminal atau report.
- Instalasi dan penghapusan hook bersifat idempotent serta tidak merusak hook milik pengguna.
- Semua perilaku publik memiliki unit test atau integration test.

## Milestone dan dependensi

| Milestone | Hasil | Bergantung pada |
| --- | --- | --- |
| M0 | Kontrak domain dan konfigurasi stabil | - |
| M1 | Orchestrator scanner benar | M0 |
| M2 | Git hooks dan staged workspace | M1 |
| M3 | Fast built-in checks | M2 |
| M4 | External command scanners | M1, M2 |
| M5 | Baseline dan policy finding baru | M3 |
| M6 | Reporter dan CI formats | M5 |
| M7 | Supply Chain dan Governance | M4, M6 |
| M8 | Cache, distribusi, dan v1.0 | Semua milestone |

---

## M0 — Kontrak domain dan konfigurasi

Tujuan: menentukan model data sebelum menambah hook dan scanner baru.

### Context: `finding/`

- [ ] Tambahkan tipe `Domain` dengan nilai `quality`, `reliability`, `hardening`, `security`, `supply_chain`, dan `governance`.
- [ ] Tambahkan `Domain` ke `finding.Finding`.
- [ ] Pertahankan `Category` sebagai subkategori, misalnya `formatting`, `error_handling`, `secret_leak`, atau `dependency`.
- [ ] Tambahkan lifecycle finding: `new`, `existing`, `resolved`, dan `suppressed` tanpa mengubah arti field suppression yang sudah ada.
- [ ] Tentukan cara policy menilai severity non-security; semua domain tetap menggunakan `CRITICAL`, `HIGH`, `MEDIUM`, dan `LOW` agar reporter konsisten.
- [ ] Tambahkan optional metadata: dokumentasi rule, tags, dan apakah finding dapat diperbaiki otomatis.
- [ ] Perbarui summary agar memiliki total per domain dan per lifecycle.
- [ ] Tambahkan test JSON backward compatibility untuk report schema `1.0`.

### Context: `rules/`

- [ ] Tambahkan `domain` pada `rules.Rule`.
- [ ] Jadikan `security` sebagai default sementara untuk custom rule lama yang tidak memiliki domain.
- [ ] Validasi kombinasi domain, category, severity, ID, dan regex.
- [ ] Dokumentasikan konvensi ID: `<domain>/<rule-name>` atau pertahankan ID lama melalui alias/migration map.
- [ ] Pisahkan default rules ke `rules/defaults_quality.go`, `defaults_reliability.go`, `defaults_hardening.go`, dan `defaults_security.go`.
- [ ] Tambahkan unit test untuk duplicate ID lintas file dan rule disabled.

### Context: `config/`

- [ ] Definisikan `Profile` (`fast`, `standard`, `full`).
- [ ] Tambahkan policy threshold per domain, bukan hanya satu `fail_on` global.
- [ ] Tambahkan konfigurasi hook `pre_commit`, `commit_msg`, dan `pre_push`.
- [ ] Tentukan strategi schema: perluas version 1 secara backward-compatible sebelum menaikkan ke version 2.
- [ ] Tolak unknown field dengan pesan error yang menunjukkan lokasi field.
- [ ] Validasi duration menggunakan `time.ParseDuration`.
- [ ] Tambahkan contoh konfigurasi minimal dan lengkap pada `testdata/`.

Contoh target konfigurasi:

```json
{
  "version": 1,
  "project": "my-service",
  "profiles": {
    "fast": ["pattern", "gofmt"],
    "standard": ["pattern", "gofmt", "go-vet", "go-test"]
  },
  "policy": {
    "quality": "HIGH",
    "reliability": "MEDIUM",
    "hardening": "HIGH",
    "security": "HIGH",
    "supply_chain": "HIGH",
    "governance": "HIGH"
  },
  "hooks": {
    "pre_commit": { "enabled": true, "profile": "fast", "staged_only": true },
    "pre_push": { "enabled": false, "profile": "standard" }
  }
}
```

### Acceptance criteria M0

- [ ] Konfigurasi lama dan seluruh test lama tetap valid.
- [ ] Setiap finding baru selalu memiliki domain yang valid.
- [ ] Policy dapat memakai threshold berbeda untuk setiap domain.
- [ ] Report masih dapat dibaca konsumen schema lama atau perubahan schema telah diberi versi baru dan migration note.

---

## M1 — Orchestrator dan lifecycle scanner

Tujuan: membuat konfigurasi scanner benar-benar dijalankan oleh runtime.

### Context: `scanner/scanner.go`

- [ ] Tambahkan metadata scanner: domain, version, capabilities, dan supported modes.
- [ ] Definisikan state transition yang valid untuk `clean`, `findings`, `partial`, `failed`, dan `skipped`.
- [ ] Tambahkan error type terstruktur agar reporter tidak perlu parsing string.
- [ ] Putuskan kontrak scanner ketika context dibatalkan atau timeout.

### Context: `securityreview.go`

- [ ] Hormati `scanners.<id>.enabled`; scanner disabled menghasilkan status `skipped`.
- [ ] Terapkan `scanners.<id>.timeout` menggunakan child context.
- [ ] Jalankan scanner independen secara paralel dengan batas concurrency.
- [ ] Pastikan urutan status dan finding tetap deterministik meskipun scanner paralel.
- [ ] Terapkan `required`: kegagalan optional menjadi warning, kegagalan required menjadi operational error.
- [ ] Recover panic dari plugin scanner dan ubah menjadi failure terstruktur.
- [ ] Hindari goroutine leak ketika scanner timeout atau context dibatalkan.
- [ ] Tambahkan test fake scanner untuk success, skip, timeout, partial, failure, panic, dan cancellation.

### Context: `policy/`

- [ ] Evaluasi finding berdasarkan threshold per domain.
- [ ] Pisahkan operational failure dari policy violation.
- [ ] Tambahkan hasil policy terstruktur berisi alasan block/allow.
- [ ] Pastikan suppressed dan existing finding dapat dikecualikan sesuai profile.

### Acceptance criteria M1

- [ ] `enabled`, `required`, dan `timeout` memiliki efek nyata.
- [ ] Satu scanner macet tidak menggantung seluruh proses.
- [ ] Output identik pada repeated run dengan input sama.
- [ ] Race detector lulus: `go test -race ./...`.

---

## M2 — Git hooks dan staged workspace

Tujuan: menyediakan commit hook yang aman dan benar terhadap Git index.

### Context baru: `gitrepo/`

- [ ] Buat wrapper command Git tanpa melalui shell.
- [ ] Implementasikan deteksi repository root.
- [ ] Implementasikan pembacaan `core.hooksPath` dan lokasi hooks efektif.
- [ ] Pisahkan operasi read-only dan write agar mudah diuji.
- [ ] Tambahkan parser path berbasis NUL (`-z`) untuk mendukung spasi dan karakter khusus.
- [ ] Tambahkan integration test menggunakan temporary Git repository.

### Context: `discovery/`

- [ ] Ubah changed/staged discovery agar menggunakan output NUL-delimited.
- [ ] Test added, copied, modified, renamed, deleted, path dengan spasi, dan repository tanpa `HEAD`.
- [ ] Pertahankan aturan bahwa deleted file tidak dipindai, tetapi laporkan statusnya bila dibutuhkan oleh check governance.
- [ ] Pastikan symlink tidak menyebabkan scanner keluar dari root.

### Context baru: `workspace/`

- [ ] Implementasikan materialisasi snapshot Git index ke temporary directory.
- [ ] Gunakan snapshot ini untuk external scanner yang membutuhkan struktur proyek lengkap.
- [ ] Jangan menyalin `.git`, credential, atau file unstaged ke snapshot.
- [ ] Bersihkan snapshot setelah proses, termasuk ketika context dibatalkan.
- [ ] Tambahkan batas ukuran dan jumlah file untuk mencegah resource exhaustion.

### Context baru: `hook/`

- [ ] Implementasikan `Install`, `Uninstall`, `Status`, dan `Run`.
- [ ] Gunakan marker/version pada hook yang dikelola tool.
- [ ] Instalasi tidak boleh diam-diam menimpa hook yang sudah ada.
- [ ] Definisikan strategi chaining atau fail-safe ketika existing hook ditemukan.
- [ ] Pastikan uninstall hanya menghapus artefak milik `go-code-scanner`.
- [ ] Jaga permission executable pada Unix.
- [ ] Dokumentasikan dukungan Windows dan format hook yang dipakai.

### Context: `cmd/security-review/main.go`

- [ ] Tambahkan `hook install`.
- [ ] Tambahkan `hook uninstall`.
- [ ] Tambahkan `hook status`.
- [ ] Tambahkan `hook run pre-commit`.
- [ ] Tambahkan `hook run commit-msg --file <path>`.
- [ ] Tambahkan `hook run pre-push`.
- [ ] Stabilkan exit code untuk allow, policy block, configuration error, dan operational error.

### Acceptance criteria M2

- [ ] Install dan uninstall idempotent.
- [ ] Existing hook tidak hilang atau berubah tanpa persetujuan eksplisit.
- [ ] Pre-commit memindai versi staged meskipun working tree berisi versi berbeda.
- [ ] Hook bekerja pada repository baru yang belum memiliki commit.
- [ ] Seluruh skenario Git memiliki end-to-end test.

---

## M3 — Built-in fast checks

Tujuan: menyediakan nilai langsung tanpa mewajibkan dependency eksternal.

### Context: `scanner/pattern/`

- [ ] Pertahankan scanner regex sebagai pemeriksaan cepat per baris.
- [ ] Tambahkan batas file, batas baris, dan pesan partial yang actionable.
- [ ] Pastikan redaction mengikuti category dan tags, bukan hanya pencarian kata sederhana.
- [ ] Tambahkan test untuk token, authorization header, multiline truncation, dan false-positive redaction.

### Context: rule pack Quality

- [ ] Merge-conflict marker.
- [ ] Trailing whitespace dan mixed indentation opsional.
- [ ] Debug statement yang tersisa.
- [ ] File generated yang diedit manual berdasarkan marker.
- [ ] File temporary, dump, atau artefak build yang ikut di-stage.
- [ ] Batas ukuran file dan line length sebagai policy opsional.

### Context: rule pack Reliability

- [ ] Error return Go yang jelas-jelas diabaikan.
- [ ] `panic` atau `log.Fatal` pada application path configurable.
- [ ] HTTP client/server tanpa timeout.
- [ ] Context tidak diteruskan pada boundary yang dapat dikenali.
- [ ] Retry loop tanpa limit/backoff.
- [ ] Pembacaan request/file tanpa limit yang jelas.

### Context: rule pack Hardening

- [ ] Debug mode aktif pada konfigurasi production.
- [ ] Wildcard CORS.
- [ ] Cookie tanpa `Secure`, `HttpOnly`, atau `SameSite` sesuai konteks.
- [ ] TLS verification dimatikan.
- [ ] Permission file/directory terlalu longgar.
- [ ] Docker container berjalan sebagai root.
- [ ] Default credential dan localhost production endpoint.

### Context: rule pack Security

- [ ] Pertahankan rule bawaan yang ada dengan ID kompatibel.
- [ ] Perluas deteksi secret tanpa menampilkan nilai secret.
- [ ] Tambahkan command execution dan path traversal patterns.
- [ ] Tambahkan weak cryptography dan insecure random patterns.
- [ ] Tambahkan unsafe deserialization patterns.
- [ ] Buat test fixture positive dan negative untuk setiap rule.
- [ ] Ukur false-positive rate pada fixture repository yang representatif.

### Acceptance criteria M3

- [ ] Setiap rule memiliki minimal satu positive dan satu negative test.
- [ ] Rule tidak mengklaim kepastian kerentanan; deskripsi menggunakan bahasa yang sesuai tingkat keyakinan.
- [ ] Fast profile tidak memerlukan network atau binary tambahan.
- [ ] Fast profile memenuhi target waktu pre-commit.

---

## M4 — External command scanners

Tujuan: menjadikan proyek sebagai orchestrator, bukan membangun ulang semua analyzer matang.

### Context baru: `scanner/command/`

- [ ] Jalankan executable menggunakan argument array, bukan shell string.
- [ ] Dukung working directory berupa root atau staged snapshot.
- [ ] Terapkan timeout, cancellation, dan process-group termination.
- [ ] Batasi ukuran stdout/stderr.
- [ ] Sanitasi environment; hanya teruskan environment yang diizinkan.
- [ ] Validasi executable dan tampilkan status `skipped` jika optional dependency tidak tersedia.
- [ ] Dukung parser `exit-code`, `json`, `json-lines`, dan parser adapter khusus.
- [ ] Jangan menganggap semua non-zero exit code sebagai operational failure; petakan exit code tool ke findings/failure.

### Context baru: `scanner/adapters/`

- [ ] Adapter `gofmt` untuk Quality.
- [ ] Adapter `go vet` untuk Quality/Reliability.
- [ ] Adapter `go test` untuk Quality/Reliability.
- [ ] Adapter `govulncheck` untuk Supply Chain/Security.
- [ ] Adapter `gosec` untuk Security.
- [ ] Adapter Gitleaks untuk Security.
- [ ] Adapter Trivy/OSV-Scanner untuk Supply Chain.
- [ ] Pertimbangkan Semgrep sebagai optional adapter, bukan dependency wajib.

### Context: `config/`

- [ ] Tambahkan scanner type `command` dan `adapter`.
- [ ] Dukung `args` sebagai array.
- [ ] Dukung allowlisted environment variables.
- [ ] Dukung `workspace: staged|root`.
- [ ] Dukung `on_missing: skip|fail`.
- [ ] Tolak konfigurasi shell command ambigu secara default.

### Acceptance criteria M4

- [ ] Command injection tidak mungkin melalui konfigurasi argument normal.
- [ ] Tool yang macet dihentikan bersama child processes.
- [ ] Findings eksternal memakai path relatif dan format yang sama dengan built-in scanner.
- [ ] Adapter memiliki fixture output dari beberapa versi tool yang didukung.

---

## M5 — Baseline, fingerprint, dan incremental policy

Tujuan: adopsi bertahap tanpa langsung memblokir seluruh technical debt lama.

### Context baru: `baseline/`

- [ ] Definisikan schema baseline berversi.
- [ ] Implementasikan create, load, compare, dan update.
- [ ] Tandai finding sebagai new, existing, atau resolved.
- [ ] Tolak baseline rusak dengan error yang jelas.
- [ ] Tulis baseline secara atomik seperti report JSON.

### Context: normalisasi di `securityreview.go`

- [ ] Ganti fingerprint yang saat ini bergantung langsung pada nomor baris.
- [ ] Bangun fingerprint dari rule ID, normalized path, dan normalized code context.
- [ ] Pertimbangkan symbol/function sebagai identity ketika scanner menyediakannya.
- [ ] Simpan fingerprint version agar algoritma dapat dimigrasikan.
- [ ] Tambahkan fuzzy relocation terbatas untuk finding yang hanya berpindah baris.
- [ ] Pertahankan deduplikasi deterministik.

### Context: `cmd/security-review/`

- [ ] Tambahkan `baseline create`.
- [ ] Tambahkan `baseline update`.
- [ ] Tambahkan `baseline status`.
- [ ] Tambahkan `scan --new-only`.
- [ ] Tambahkan konfirmasi atau dry-run sebelum update baseline menghapus finding.

### Context: `policy/`

- [ ] Pre-commit default hanya memblokir finding baru.
- [ ] CI dapat memilih `new-only` atau seluruh finding.
- [ ] Finding resolved tidak memengaruhi exit code.
- [ ] Suppression tetap terpisah dari baseline dan wajib memiliki reason + expiry.

### Acceptance criteria M5

- [ ] Penambahan baris di atas finding tidak membuatnya terlihat baru.
- [ ] Perubahan substansial pada baris bermasalah menghasilkan finding baru.
- [ ] Baseline update memiliki diff yang mudah direview.
- [ ] Suppression kedaluwarsa kembali menjadi finding aktif.

---

## M6 — Developer experience dan reporting

Tujuan: membuat kegagalan hook dapat ditindaklanjuti tanpa membuka JSON manual.

### Context: `reporter/terminal.go`

- [ ] Cetak finding aktif dengan domain, severity, rule ID, path, line, dan recommendation.
- [ ] Kelompokkan hasil menjadi new, existing, suppressed, dan operational warnings.
- [ ] Tambahkan warna hanya ketika output adalah TTY dan hormati `NO_COLOR`.
- [ ] Tambahkan `--quiet`, `--verbose`, dan batas jumlah finding terminal.
- [ ] Jangan pernah mencetak snippet yang sudah ditandai sensitif.
- [ ] Tampilkan command berikutnya: explain, suppress, atau fix.

### Context: `reporter/json.go`

- [ ] Tambahkan schema contract/golden test.
- [ ] Pastikan backup/replace bekerja lintas platform.
- [ ] Tambahkan file permission aman untuk report yang mungkin sensitif.
- [ ] Sertakan tool version, config hash, rule-set hash, dan fingerprint version.

### Context baru: `reporter/sarif/`

- [ ] Implementasikan SARIF 2.1.0.
- [ ] Petakan rule metadata, severity, lokasi, dan remediation.
- [ ] Tambahkan golden test yang dapat dibaca GitHub Code Scanning.

### Context baru: `reporter/junit/`

- [ ] Implementasikan JUnit XML untuk integrasi CI.
- [ ] Bedakan policy finding dan operational failure.

### Context: CLI

- [ ] Tambahkan `--format terminal|json|sarif|junit`.
- [ ] Tambahkan `--explain <rule-id>`.
- [ ] Tambahkan `--fix` hanya untuk fixer yang deterministik dan aman.
- [ ] Tambahkan `--dry-run` untuk hook install, baseline update, dan suppression helper.

### Acceptance criteria M6

- [ ] Developer dapat mengetahui penyebab block dan cara memperbaikinya dari terminal.
- [ ] JSON, SARIF, dan JUnit lolos validasi format masing-masing.
- [ ] Snapshot/golden test menjaga output tetap stabil.

---

## M7 — Supply Chain dan Governance

Tujuan: memperluas pemeriksaan di luar source pattern tanpa memperlambat pre-commit.

### Context: Supply Chain checks

- [ ] Integrasikan `govulncheck` untuk module Go.
- [ ] Integrasikan OSV-Scanner atau Trivy sebagai adapter opsional.
- [ ] Deteksi dependency version yang tidak dikunci.
- [ ] Deteksi Docker base image memakai tag `latest`.
- [ ] Deteksi GitHub Actions yang hanya dikunci ke mutable tag.
- [ ] Tambahkan allowlist/denylist dependency dan license.
- [ ] Pisahkan hasil yang memerlukan network dari checks offline.
- [ ] Jalankan Supply Chain pada pre-push/CI secara default, bukan pre-commit.

### Context: Governance checks

- [ ] Commit message policy: format, ticket ID, dan ukuran pesan.
- [ ] Required file/header/license policy.
- [ ] Ownership rule untuk path sensitif.
- [ ] Architecture dependency boundaries.
- [ ] Privacy rules untuk PII pada log, URL, fixture, dan response.
- [ ] Wajibkan ticket, approver, dan expiry untuk suppression tertentu.
- [ ] Buat rule governance configurable agar tidak hardcoded ke satu organisasi.

### Context baru: `scanner/architecture/`

- [ ] Definisikan layer/module rules melalui config.
- [ ] Bangun import graph Go menggunakan standard library parser bila cukup.
- [ ] Laporkan forbidden dependency dengan source dan target package.
- [ ] Deteksi cycle bila belum ditangani toolchain yang digunakan.

### Acceptance criteria M7

- [ ] Network checks tidak berjalan diam-diam saat offline profile.
- [ ] Dependency database version dicatat dalam report.
- [ ] Governance rules dapat dinonaktifkan atau disesuaikan per repository.
- [ ] Architecture violation memiliki dependency path yang actionable.

---

## M8 — Cache, hardening tool, dan distribusi

Tujuan: membuat tool cepat, aman, dan mudah dipasang oleh tim.

### Context baru: `cache/`

- [ ] Cache hasil berdasarkan content hash, scanner version, config hash, dan rule-set hash.
- [ ] Invalidasi cache ketika salah satu input berubah.
- [ ] Gunakan atomic write dan lock lintas proses.
- [ ] Batasi ukuran serta umur cache.
- [ ] Jangan menyimpan raw secret atau source snippet sensitif.
- [ ] Tambahkan `cache clean` dan `cache stats`.

### Context: self-hardening

- [ ] Audit path traversal pada output, config, rules, suppression, baseline, dan cache.
- [ ] Audit symlink handling.
- [ ] Audit command environment dan executable resolution.
- [ ] Pastikan report permission mengikuti least privilege.
- [ ] Fuzz parser config, rules, suppression, baseline, dan external output.
- [ ] Jalankan `go test -race`, `go vet`, govulncheck, dan scanner ini terhadap dirinya sendiri.

### Context: release engineering

- [ ] Tambahkan semantic versioning dan changelog.
- [ ] Build binary Linux, macOS, dan Windows.
- [ ] Publikasikan checksum dan signed provenance.
- [ ] Dokumentasikan `go install` dan binary install.
- [ ] Tambahkan compatibility policy untuk config, report, rule, suppression, baseline, dan hook marker.
- [ ] Siapkan upgrade command atau migration guide bila schema berubah.

### Acceptance criteria M8

- [ ] Cached dan uncached run menghasilkan finding identik.
- [ ] Binary release dapat memasang dan menjalankan hook pada platform yang didukung.
- [ ] Repository menggunakan tool ini pada pre-commit atau CI sebagai dogfooding.
- [ ] Release candidate lulus seluruh unit, integration, end-to-end, race, fuzz-smoke, dan golden tests.

---

## Testing matrix

Checklist ini berlaku pada setiap milestone, bukan hanya menjelang rilis:

- [ ] Unit test untuk happy path, invalid input, cancellation, dan boundary condition.
- [ ] Integration test dengan temporary Git repository.
- [ ] Golden test untuk config dan report yang menjadi kontrak publik.
- [ ] Test staged-vs-unstaged untuk setiap scanner yang mengklaim staged support.
- [ ] Test redaction agar secret tidak muncul pada stdout, stderr, JSON, SARIF, atau JUnit.
- [ ] Cross-platform test untuk path separator, executable hook, rename, dan atomic write.
- [ ] Benchmark discovery, pattern scan, baseline compare, dan cache.
- [ ] `go test ./...`.
- [ ] `go test -race ./...`.
- [ ] `go vet ./...`.
- [ ] `govulncheck ./...` ketika tersedia.

## Definition of Done per task

Sebuah checkbox baru boleh ditandai selesai jika:

- [ ] Implementasi dan error handling selesai.
- [ ] Test relevan ditambahkan dan lulus.
- [ ] Dokumentasi CLI/config/API diperbarui.
- [ ] Tidak membocorkan data sensitif.
- [ ] Perilaku staged tidak membaca konten unstaged tanpa dokumentasi eksplisit.
- [ ] Backward compatibility diperiksa.
- [ ] Perubahan format publik memiliki versioning atau migration note.
- [ ] `go test ./...`, `go test -race ./...`, `go vet ./...`, dan `git diff --check` lulus.

## Prioritas implementasi terdekat

Sprint pertama sebaiknya hanya mengambil pekerjaan berikut:

1. [ ] Tambahkan domain pada finding dan rules dengan backward compatibility.
2. [ ] Terapkan scanner `enabled`, `required`, dan `timeout`.
3. [ ] Paralelkan scanner secara deterministik.
4. [ ] Perbaiki staged discovery menjadi NUL-safe.
5. [ ] Buat integration test temporary Git repository.
6. [ ] Implementasikan `hook status` dan safe `hook install` sebagai vertical slice pertama.

Sprint kedua:

1. [ ] Selesaikan `hook run` dan `hook uninstall`.
2. [ ] Implementasikan staged workspace snapshot.
3. [ ] Pisahkan built-in rules ke Quality, Reliability, Hardening, dan Security.
4. [ ] Tambahkan terminal output yang menampilkan finding actionable.
5. [ ] Dogfood fast profile pada repository ini.

Setelah dua sprint tersebut, proyek sudah memiliki fondasi commit gate yang dapat dipakai. External adapters dan baseline menjadi fokus berikutnya sebelum rollout ke banyak repository.
