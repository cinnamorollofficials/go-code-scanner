package rules

import "github.com/cinnamorollofficials/go-code-scanner/pkg/finding"

// DefaultSecurity returns the built-in application security rules.
func DefaultSecurity() []Rule {
	return []Rule{
		{
			ID: "mock-token", Pattern: `google-mock-jwt-token|mock.*jwt|dummy.*token`,
			Severity: finding.Critical, Domain: finding.Security, Category: "secret_leak",
			Description:    "Hardcoded mock token found — remove before production deployment",
			Recommendation: "Remove hardcoded mock tokens and load credentials from environment variables or key vaults",
		},
		{
			ID: "browser-token-storage", Pattern: `localStorage\.(setItem|getItem)\(['\"]?(access_token|refresh_token|token)`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Token stored in localStorage — vulnerable to XSS token theft",
			Recommendation: "Store authentication tokens in HttpOnly, Secure, SameSite cookies instead of localStorage",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
		{
			ID: "permission-bypass", Pattern: `SHOP_.*return true|bypass.*permission|permission.*bypass`,
			Severity: finding.Critical, Domain: finding.Security, Category: "security_misconfiguration",
			Description:    "Hardcoded permission bypass found in application logic",
			Recommendation: "Remove permission bypass conditions and enforce strict authorization checks",
		},
		{
			ID: "weak-secret", Pattern: `change-me-in-production|your_super_secret|your_secret_key_here`,
			Severity: finding.Critical, Domain: finding.Security, Category: "secret_leak",
			Description:    "Default or weak secret value found",
			Recommendation: "Replace default/placeholder secrets with cryptographically strong random values from secure configuration",
		},
		{
			ID: "frontend-sensitive-log", Pattern: `console\.(log|debug|info|error).*\b(token|password|secret|permission|user_id|tenant)`,
			Severity: finding.Medium, Domain: finding.Security, Category: "data_leak",
			Description:    "Frontend log statement may expose sensitive credentials or PII",
			Recommendation: "Sanitize log parameters and remove sensitive tokens or user identifiers from console logs",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
		{
			ID: "backend-sensitive-log", Pattern: `fmt\.Print.*(token|password|secret|key|DatabaseURL)|log\.(Info|Debug|Warn).*\b(password|secret|token|key)\b`,
			Severity: finding.Medium, Domain: finding.Security, Category: "data_leak",
			Description:    "Backend log statement may expose sensitive credentials or keys",
			Recommendation: "Redact sensitive parameters before writing to application log streams",
			Extensions:     []string{".go"},
		},
		{
			ID: "sql-string-format", Pattern: `fmt\.Sprintf.*\b(SELECT|INSERT|UPDATE|DELETE|WHERE)\b`,
			Severity: finding.High, Domain: finding.Security, Category: "injection",
			Description:    "Potential SQL injection using formatted strings",
			Recommendation: "Use parameterized queries or prepared statements instead of string formatting",
			Extensions:     []string{".go"},
		},
		{
			ID: "hardcoded-credential", Pattern: `(password|passwd|pwd|secret|api_key)\s*[:=]\s*['\"][^'\"]{6,}['\"]`,
			Severity: finding.High, Domain: finding.Security, Category: "secret_leak",
			Description:    "Hardcoded credential or API secret key found",
			Recommendation: "Extract credentials to environment variables or secret management services",
		},
		{
			ID: "unsafe-inner-html", Pattern: `dangerouslySetInnerHTML\s*=\s*\{`,
			Severity: finding.High, Domain: finding.Security, Category: "xss",
			Description:    "dangerouslySetInnerHTML used — potential DOM XSS vulnerability",
			Recommendation: "Sanitize raw HTML using DOMPurify before injecting into the DOM",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
		{
			ID: "dynamic-order", Pattern: `\.Order\(.*fmt\.Sprintf`,
			Severity: finding.High, Domain: finding.Security, Category: "injection",
			Description:    "Dynamic ORDER BY clause built via string formatting",
			Recommendation: "Validate dynamic column names against an explicit allowlist before building queries",
			Extensions:     []string{".go"},
		},
		{
			ID: "api-struct-response", Pattern: `c\.JSON\(.*,\s*\*?(user|account|member|staff|karyawan)\b`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Internal domain struct may be serialized directly into HTTP response",
			Recommendation: "Map internal domain entities to explicit response DTOs to avoid leaking sensitive fields",
			Extensions:     []string{".go"},
		},
		{
			ID: "sensitive-json-field", Pattern: `(Password|PasswordHash|Hash|SecretKey|ApiKey)\s+\w+.*json:\"[^-]`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Sensitive struct field may be exposed in JSON serialization",
			Recommendation: "Use json:\"-\" struct tag or custom serializer to exclude sensitive attributes",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-shell-command", Pattern: `exec\.Command(Context)?\([^)]*['\"](sh|bash)['\"]\s*,\s*['\"]-c['\"]`,
			Severity: finding.High, Domain: finding.Security, Category: "command_injection",
			Description:    "Shell command interpreter executed via os/exec",
			Recommendation: "Execute binary commands directly with argument arrays and sanitize untrusted input",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-weak-cryptographic-hash", Pattern: `(md5|sha1)\.(New|Sum)\(`,
			Severity: finding.Medium, Domain: finding.Security, Category: "weak_cryptography",
			Description:    "Weak cryptographic hash algorithm (MD5/SHA1) detected",
			Recommendation: "Use SHA-256 or stronger algorithms; use bcrypt/argon2 for password hashing",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-tainted-file-path", Pattern: `os\.(Open|OpenFile|ReadFile|WriteFile|Remove)\([^)]*(r\.URL\.Query|r\.FormValue|c\.Param)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "path_traversal",
			Description:    "Untrusted request parameter used directly in file system operation",
			Recommendation: "Normalize paths, enforce base directory boundaries, and use allowlisted identifiers",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-weak-random-secret", Pattern: `(?i)(token|secret|nonce|session).{0,40}\brand\.(Int|Intn|Read|Uint32|Uint64)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "insecure_randomness",
			Description:    "Security-sensitive value generated using pseudo-random math/rand package",
			Recommendation: "Use crypto/rand for generating tokens, nonces, session identifiers, and secret keys",
			Extensions:     []string{".go"},
		},
		{
			ID: "javascript-dynamic-eval", Pattern: `\beval\s*\([^)]*[A-Za-z_$]`,
			Severity: finding.High, Domain: finding.Security, Category: "unsafe_deserialization",
			Description:    "Dynamic eval execution of untrusted input detected",
			Recommendation: "Use structured data parsers (JSON.parse) and schema validators instead of code evaluation",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
	}
}
