package rules

import "github.com/cinnamorollofficials/go-code-scanner/pkg/finding"

// DefaultSecurity returns the built-in application security rules.
func DefaultSecurity() []Rule {
	return []Rule{
		{ID: "mock-token", Pattern: `google-mock-jwt-token|mock.*jwt|dummy.*token`, Severity: finding.Critical, Domain: finding.Security, Category: "secret_leak", Description: "Hardcoded mock token ditemukan — hapus sebelum production"},
		{ID: "browser-token-storage", Pattern: `localStorage\.(setItem|getItem)\(['\"]?(access_token|refresh_token|token)`, Severity: finding.High, Domain: finding.Security, Category: "data_leak", Description: "Token disimpan di localStorage — gunakan HttpOnly Cookie", Extensions: []string{".ts", ".tsx", ".js", ".jsx"}},
		{ID: "permission-bypass", Pattern: `SHOP_.*return true|bypass.*permission|permission.*bypass`, Severity: finding.Critical, Domain: finding.Security, Category: "security_misconfiguration", Description: "Permission bypass hardcoded ditemukan"},
		{ID: "weak-secret", Pattern: `change-me-in-production|your_super_secret|your_secret_key_here`, Severity: finding.Critical, Domain: finding.Security, Category: "secret_leak", Description: "Default atau weak secret ditemukan"},
		{ID: "frontend-sensitive-log", Pattern: `console\.(log|debug|info|error).*\b(token|password|secret|permission|user_id|tenant)`, Severity: finding.Medium, Domain: finding.Security, Category: "data_leak", Description: "Log frontend mungkin menampilkan data sensitif", Extensions: []string{".ts", ".tsx", ".js", ".jsx"}},
		{ID: "backend-sensitive-log", Pattern: `fmt\.Print.*(token|password|secret|key|DatabaseURL)|log\.(Info|Debug|Warn).*\b(password|secret|token|key)\b`, Severity: finding.Medium, Domain: finding.Security, Category: "data_leak", Description: "Log backend mungkin menampilkan data sensitif", Extensions: []string{".go"}},
		{ID: "sql-string-format", Pattern: `fmt\.Sprintf.*\b(SELECT|INSERT|UPDATE|DELETE|WHERE)\b`, Severity: finding.High, Domain: finding.Security, Category: "injection", Description: "Potensi SQL injection — gunakan parameterized query", Extensions: []string{".go"}},
		{ID: "hardcoded-credential", Pattern: `(password|passwd|pwd|secret|api_key)\s*[:=]\s*['\"][^'\"]{6,}['\"]`, Severity: finding.High, Domain: finding.Security, Category: "secret_leak", Description: "Credential hardcoded ditemukan"},
		{ID: "unsafe-inner-html", Pattern: `dangerouslySetInnerHTML\s*=\s*\{`, Severity: finding.High, Domain: finding.Security, Category: "xss", Description: "dangerouslySetInnerHTML ditemukan — pastikan input disanitasi", Extensions: []string{".ts", ".tsx", ".js", ".jsx"}},
		{ID: "dynamic-order", Pattern: `\.Order\(.*fmt\.Sprintf`, Severity: finding.High, Domain: finding.Security, Category: "injection", Description: "ORDER BY dinamis harus memakai whitelist", Extensions: []string{".go"}},
		{ID: "api-struct-response", Pattern: `c\.JSON\(.*,\s*\*?(user|account|member|staff|karyawan)\b`, Severity: finding.High, Domain: finding.Security, Category: "data_leak", Description: "Struct sensitif mungkin dikirim langsung ke response", Extensions: []string{".go"}},
		{ID: "sensitive-json-field", Pattern: `(Password|PasswordHash|Hash|SecretKey|ApiKey)\s+\w+.*json:\"[^-]`, Severity: finding.High, Domain: finding.Security, Category: "data_leak", Description: "Field sensitif mungkin terekspos dalam JSON", Extensions: []string{".go"}},
		{
			ID: "go-shell-command", Pattern: `exec\.Command(Context)?\([^)]*['\"](sh|bash)['\"]\s*,\s*['\"]-c['\"]`,
			Severity: finding.High, Domain: finding.Security, Category: "command_injection",
			Description:    "Shell command interpreter digunakan melalui os/exec",
			Recommendation: "Jalankan executable secara langsung dengan argument array dan validasi input yang tidak dipercaya",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-weak-cryptographic-hash", Pattern: `(md5|sha1)\.(New|Sum)\(`,
			Severity: finding.Medium, Domain: finding.Security, Category: "weak_cryptography",
			Description:    "Algoritma hash kriptografi yang lemah ditemukan",
			Recommendation: "Gunakan SHA-256 atau algoritma yang sesuai; gunakan password KDF untuk password",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-tainted-file-path", Pattern: `os\.(Open|OpenFile|ReadFile|WriteFile|Remove)\([^)]*(r\.URL\.Query|r\.FormValue|c\.Param)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "path_traversal",
			Description:    "Input request mungkin digunakan langsung sebagai path file",
			Recommendation: "Normalisasi path, enforce base directory, dan gunakan allowlist identifier",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-weak-random-secret", Pattern: `(?i)(token|secret|nonce|session).{0,40}\brand\.(Int|Intn|Read|Uint32|Uint64)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "insecure_randomness",
			Description:    "Nilai keamanan mungkin dibuat menggunakan math/rand",
			Recommendation: "Gunakan crypto/rand untuk token, nonce, session identifier, dan secret",
			Extensions:     []string{".go"},
		},
		{
			ID: "javascript-dynamic-eval", Pattern: `\beval\s*\([^)]*[A-Za-z_$]`,
			Severity: finding.High, Domain: finding.Security, Category: "unsafe_deserialization",
			Description:    "Dynamic eval mungkin mengeksekusi data sebagai kode",
			Recommendation: "Gunakan parser data terstruktur dan validasi schema tanpa evaluasi kode",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
	}
}
