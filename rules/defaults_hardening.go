package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

// DefaultHardening returns the built-in secure-configuration rules.
func DefaultHardening() []Rule {
	return []Rule{
		{ID: "hardcoded-api-url", Pattern: `API_URL\s*=\s*['\"]https?://localhost`, Severity: finding.Medium, Domain: finding.Hardening, Category: "configuration_leak", Description: "URL API hardcoded — gunakan environment variable"},
		{
			ID: "tls-insecure-skip-verify", Pattern: `InsecureSkipVerify\s*:\s*true`,
			Severity: finding.High, Domain: finding.Hardening, Category: "transport_security",
			Description:    "Verifikasi sertifikat TLS dinonaktifkan",
			Recommendation: "Aktifkan certificate verification dan konfigurasi trust store yang sesuai",
			Extensions:     []string{".go"},
		},
		{
			ID: "wildcard-cors-origin", Pattern: `(AllowOrigins|Access-Control-Allow-Origin).{0,40}['\"]\*['\"]`,
			Severity: finding.High, Domain: finding.Hardening, Category: "cors",
			Description:    "Wildcard CORS origin ditemukan",
			Recommendation: "Gunakan allowlist origin yang eksplisit untuk environment terkait",
			Extensions:     []string{".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"},
		},
		{
			ID: "go-permissive-file-mode", Pattern: `os\.(WriteFile|OpenFile|Mkdir|MkdirAll)\(.*,\s*0?777\s*\)`,
			Severity: finding.Medium, Domain: finding.Hardening, Category: "file_permission",
			Description:    "File atau directory dibuat dengan permission world-writable",
			Recommendation: "Gunakan permission minimum yang diperlukan, misalnya 0600 atau 0750",
			Extensions:     []string{".go"},
		},
		{
			ID: "debug-mode-enabled", Pattern: `(?i)\b(debug|debug_mode)\s*[:=]\s*(true|1|['"]true['"])`,
			Severity: finding.Medium, Domain: finding.Hardening, Category: "debug_configuration",
			Description:    "Debug mode tampak diaktifkan secara eksplisit",
			Recommendation: "Nonaktifkan debug mode pada konfigurasi deployment production",
			Extensions:     []string{".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"},
		},
		{
			ID: "go-insecure-cookie-attribute", Pattern: `(Secure|HttpOnly)\s*:\s*false|SameSite\s*:\s*http\.SameSiteDefaultMode`,
			Severity: finding.High, Domain: finding.Hardening, Category: "cookie_security",
			Description:    "Cookie memiliki atribut keamanan yang secara eksplisit tidak aman",
			Recommendation: "Aktifkan Secure dan HttpOnly serta gunakan kebijakan SameSite yang sesuai",
			Extensions:     []string{".go"},
		},
	}
}
