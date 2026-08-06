package rules

import "github.com/cinnamorollofficials/go-code-scanner/pkg/finding"

// DefaultHardening returns the built-in secure-configuration rules.
func DefaultHardening() []Rule {
	return []Rule{
		{
			ID: "hardcoded-api-url", Pattern: `API_URL\s*=\s*['\"]https?://localhost`,
			Severity: finding.Medium, Domain: finding.Hardening, Category: "configuration_leak",
			Description:    "Hardcoded localhost API URL found — load dynamically from environment variable",
			Recommendation: "Configure API endpoints dynamically via environment variables for different environments",
			UnsafeExample:  `const API_URL = "http://localhost:8080/api/v1";`,
			SafeExample:    `const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";`,
		},
		{
			ID: "tls-insecure-skip-verify", Pattern: `InsecureSkipVerify\s*:\s*true`,
			Severity: finding.High, Domain: finding.Hardening, Category: "transport_security",
			Description:    "TLS certificate verification is explicitly disabled",
			Recommendation: "Enable certificate verification and configure valid trust stores",
			Extensions:     []string{".go"},
			UnsafeExample: `tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}`,
			SafeExample: `tr := &http.Transport{
    TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
}`,
		},
		{
			ID: "wildcard-cors-origin", Pattern: `(AllowOrigins|Access-Control-Allow-Origin).{0,40}['\"]\*['\"]`,
			Severity: finding.High, Domain: finding.Hardening, Category: "cors",
			Description:    "Wildcard CORS origin header found in configuration",
			Recommendation: "Use an explicit CORS origin allowlist tailored for each deployment environment",
			Extensions:     []string{".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"},
			UnsafeExample:  `c.Header("Access-Control-Allow-Origin", "*")`,
			SafeExample:    `c.Header("Access-Control-Allow-Origin", "https://app.example.com")`,
		},
		{
			ID: "go-permissive-file-mode", Pattern: `os\.(WriteFile|OpenFile|Mkdir|MkdirAll)\(.*,\s*0?777\s*\)`,
			Severity: finding.Medium, Domain: finding.Hardening, Category: "file_permission",
			Description:    "File or directory created with permissive world-writable file permissions (0777)",
			Recommendation: "Use minimum required file permissions such as 0600 for files or 0750 for directories",
			Extensions:     []string{".go"},
			UnsafeExample:  `os.WriteFile("config.json", data, 0777)`,
			SafeExample:    `os.WriteFile("config.json", data, 0600)`,
		},
		{
			ID: "debug-mode-enabled", Pattern: `(?i)\b(debug|debug_mode)\s*[:=]\s*(true|1|['"]true['"])`,
			Severity: finding.Medium, Domain: finding.Hardening, Category: "debug_configuration",
			Description:    "Debug mode appears to be explicitly enabled in configuration",
			Recommendation: "Disable debug mode in production deployment configurations to prevent information disclosure",
			Extensions:     []string{".go", ".ts", ".tsx", ".js", ".jsx", ".yaml", ".yml", ".json"},
			UnsafeExample:  `debug := true`,
			SafeExample:    `debug := os.Getenv("APP_ENV") == "development"`,
		},
		{
			ID: "go-insecure-cookie-attribute", Pattern: `(Secure|HttpOnly)\s*:\s*false|SameSite\s*:\s*http\.SameSiteDefaultMode`,
			Severity: finding.High, Domain: finding.Hardening, Category: "cookie_security",
			Description:    "Cookie configured with explicitly insecure security attributes",
			Recommendation: "Enable Secure and HttpOnly flags and set an appropriate SameSite policy for session cookies",
			Extensions:     []string{".go"},
			UnsafeExample:  `cookie := &http.Cookie{Name: "session", Value: token, Secure: false}`,
			SafeExample:    `cookie := &http.Cookie{Name: "session", Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}`,
		},
	}
}
