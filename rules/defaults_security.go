package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

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
	}
}
