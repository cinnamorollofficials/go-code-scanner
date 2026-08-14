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
			UnsafeExample:  `const AUTH_HEADER = "Bearer google-mock-jwt-token-12345";`,
			SafeExample:    `const AUTH_HEADER = ` + "`" + `Bearer ${process.env.AUTH_TOKEN}` + "`" + `;`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `const authHeader = "Bearer google-mock-jwt-token-12345"`,
					Safe:   `authHeader := fmt.Sprintf("Bearer %s", os.Getenv("AUTH_TOKEN"))`,
				},
				{
					Language: "ts", Label: "TypeScript / JavaScript",
					Unsafe: `const AUTH_HEADER = "Bearer google-mock-jwt-token-12345";`,
					Safe:   `const AUTH_HEADER = ` + "`" + `Bearer ${process.env.AUTH_TOKEN}` + "`" + `;`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `AUTH_HEADER = "Bearer google-mock-jwt-token-12345"`,
					Safe:   `auth_header = f"Bearer {os.environ.get('AUTH_TOKEN')}"`,
				},
			},
		},
		{
			ID: "browser-token-storage", Pattern: `localStorage\.(setItem|getItem)\(['\"]?(access_token|refresh_token|token)`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Token stored in localStorage — vulnerable to XSS token theft",
			Recommendation: "Store authentication tokens in HttpOnly, Secure, SameSite cookies instead of localStorage",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `localStorage.setItem("access_token", response.token);`,
			SafeExample:    `await fetch("/api/login", { credentials: "include", method: "POST", body });`,
		},
		{
			ID: "permission-bypass", Pattern: `SHOP_.*return true|bypass.*permission|permission.*bypass`,
			Severity: finding.Critical, Domain: finding.Security, Category: "security_misconfiguration",
			Description:    "Hardcoded permission bypass found in application logic",
			Recommendation: "Remove permission bypass conditions and enforce strict authorization checks",
			UnsafeExample: `func CheckPermission(user User) bool {
    if user.Role == "admin" || bypassPermission {
        return true
    }
    return false
}`,
			SafeExample: `func CheckPermission(ctx context.Context, user User, resource string) bool {
    return authzService.CanAccess(ctx, user.ID, resource)
}`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `func CheckPermission(user User) bool {
    if user.Role == "admin" || bypassPermission {
        return true
    }
    return false
}`,
					Safe: `func CheckPermission(ctx context.Context, user User, resource string) bool {
    return authzService.CanAccess(ctx, user.ID, resource)
}`,
				},
				{
					Language: "ts", Label: "TypeScript / JavaScript",
					Unsafe: `function checkPermission(user: User): boolean {
    if (user.role === 'admin' || process.env.BYPASS_PERMISSIONS === 'true') {
        return true;
    }
    return false;
}`,
					Safe: `async function checkPermission(user: User, resource: string): Promise<boolean> {
    return await authzService.canAccess(user.id, resource);
}`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `def check_permission(user):
    if user.role == "admin" or bypass_permission:
        return True
    return False`,
					Safe: `def check_permission(user, resource):
    return authz_service.can_access(user.id, resource)`,
				},
			},
		},
		{
			ID: "weak-secret", Pattern: `change-me-in-production|your_super_secret|your_secret_key_here`,
			Severity: finding.Critical, Domain: finding.Security, Category: "secret_leak",
			Description:    "Default or weak secret value found",
			Recommendation: "Replace default/placeholder secrets with cryptographically strong random values from secure configuration",
			UnsafeExample:  `jwtSecret := []byte("change-me-in-production")`,
			SafeExample:    `jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `jwtSecret := []byte("change-me-in-production")`,
					Safe:   `jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))`,
				},
				{
					Language: "ts", Label: "TypeScript / JavaScript",
					Unsafe: `const jwtSecret = "change-me-in-production";`,
					Safe:   `const jwtSecret = process.env.JWT_SECRET_KEY;`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `JWT_SECRET = "change-me-in-production"`,
					Safe:   `JWT_SECRET = os.environ.get("JWT_SECRET_KEY")`,
				},
			},
		},
		{
			ID: "frontend-sensitive-log", Pattern: `console\.(log|debug|info|error).*\b(token|password|secret|permission|user_id|tenant)`,
			Severity: finding.Medium, Domain: finding.Security, Category: "data_leak",
			Description:    "Frontend log statement may expose sensitive credentials or PII",
			Recommendation: "Sanitize log parameters and remove sensitive tokens or user identifiers from console logs",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `console.log("User auth failed for password:", password);`,
			SafeExample:    `console.error("User authentication failed", { username });`,
		},
		{
			ID: "backend-sensitive-log", Pattern: `fmt\.Print.*(token|password|secret|key|DatabaseURL)|log\.(Info|Debug|Warn).*\b(password|secret|token|key)\b`,
			Severity: finding.Medium, Domain: finding.Security, Category: "data_leak",
			Description:    "Backend log statement may expose sensitive credentials or keys",
			Recommendation: "Redact sensitive parameters before writing to application log streams",
			Extensions:     []string{".go"},
			UnsafeExample:  `log.Printf("Connecting to DB with secret: %s", dbSecret)`,
			SafeExample:    `log.Printf("Connecting to DB host: %s", dbHost)`,
		},
		{
			ID: "sql-string-format", Pattern: `fmt\.Sprintf.*\b(SELECT|INSERT|UPDATE|DELETE|WHERE)\b`,
			Severity: finding.High, Domain: finding.Security, Category: "injection",
			Description:    "Potential SQL injection using formatted strings",
			Recommendation: "Use parameterized queries or prepared statements instead of string formatting",
			Extensions:     []string{".go"},
			UnsafeExample:  `query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)`,
			SafeExample:    `db.Query("SELECT * FROM users WHERE email = $1", userEmail)`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)
rows, err := db.Query(query)`,
					Safe: `query := "SELECT * FROM users WHERE email = $1"
rows, err := db.Query(query, userEmail)`,
				},
				{
					Language: "ts", Label: "TypeScript / JavaScript",
					Unsafe: `const query = ` + "`SELECT * FROM users WHERE email = '${userEmail}'`" + `;
const result = await client.query(query);`,
					Safe: `const query = "SELECT * FROM users WHERE email = $1";
const result = await client.query(query, [userEmail]);`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `query = f"SELECT * FROM users WHERE email = '{user_email}'"
cursor.execute(query)`,
					Safe: `query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_email,))`,
				},
			},
		},
		{
			ID: "hardcoded-credential", Pattern: `(password|passwd|pwd|secret|api_key)\s*[:=]\s*['\"][^'\"]{6,}['\"]`,
			Severity: finding.High, Domain: finding.Security, Category: "secret_leak",
			Description:    "Hardcoded credential or API secret key found",
			Recommendation: "Extract credentials to environment variables or secret management services",
			UnsafeExample:  `const apiKey = "synthetic_secret_api_key_12345"`,
			SafeExample:    `apiKey := os.Getenv("STRIPE_API_KEY")`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `apiKey := "synthetic_secret_api_key_12345"`,
					Safe:   `apiKey := os.Getenv("STRIPE_API_KEY")`,
				},
				{
					Language: "ts", Label: "TypeScript / JavaScript",
					Unsafe: `const apiKey = "synthetic_secret_api_key_12345";`,
					Safe:   `const apiKey = process.env.STRIPE_API_KEY;`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `api_key = "synthetic_secret_api_key_12345"`,
					Safe:   `api_key = os.environ.get("STRIPE_API_KEY")`,
				},
				{
					Language: "java", Label: "Java",
					Unsafe: `String apiKey = "synthetic_secret_api_key_12345";`,
					Safe:   `String apiKey = System.getenv("STRIPE_API_KEY");`,
				},
			},
		},
		{
			ID: "unsafe-inner-html", Pattern: `dangerouslySetInnerHTML\s*=\s*\{`,
			Severity: finding.High, Domain: finding.Security, Category: "xss",
			Description:    "dangerouslySetInnerHTML used — potential DOM XSS vulnerability",
			Recommendation: "Sanitize raw HTML using DOMPurify before injecting into the DOM",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `<div dangerouslySetInnerHTML={{ __html: userInput }} />`,
			SafeExample:    `<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(userInput) }} />`,
		},
		{
			ID: "dynamic-order", Pattern: `\.Order\(.*fmt\.Sprintf`,
			Severity: finding.High, Domain: finding.Security, Category: "injection",
			Description:    "Dynamic ORDER BY clause built via string formatting",
			Recommendation: "Validate dynamic column names against an explicit allowlist before building queries",
			Extensions:     []string{".go"},
			UnsafeExample:  `db.Order(fmt.Sprintf("%s ASC", sortColumn))`,
			SafeExample: `allowedColumns := map[string]bool{"created_at": true, "name": true}
if allowedColumns[sortColumn] {
    db.Order(sortColumn + " ASC")
}`,
		},
		{
			ID: "api-struct-response", Pattern: `c\.JSON\(.*,\s*\*?(user|account|member|staff|karyawan)\b`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Internal domain struct may be serialized directly into HTTP response",
			Recommendation: "Map internal domain entities to explicit response DTOs to avoid leaking sensitive fields",
			Extensions:     []string{".go"},
			UnsafeExample: `var user User // Contains HashedPassword, SecretToken
c.JSON(http.StatusOK, user)`,
			SafeExample: `response := UserResponse{ID: user.ID, Email: user.Email}
c.JSON(http.StatusOK, response)`,
		},
		{
			ID: "sensitive-json-field", Pattern: `(Password|PasswordHash|Hash|SecretKey|ApiKey)\s+\w+.*json:\"[^-]`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Sensitive struct field may be exposed in JSON serialization",
			Recommendation: "Use json:\"-\" struct tag or custom serializer to exclude sensitive attributes",
			Extensions:     []string{".go"},
			UnsafeExample: `type Account struct {
    ID           string ` + "`json:\"id\"`" + `
    PasswordHash string ` + "`json:\"password_hash\"`" + `
}`,
			SafeExample: `type Account struct {
    ID           string ` + "`json:\"id\"`" + `
    PasswordHash string ` + "`json:\"-\"`" + `
}`,
		},
		{
			ID: "go-shell-command", Pattern: `exec\.Command(Context)?\([^)]*['\"](sh|bash)['\"]\s*,\s*['\"]-c['\"]`,
			Severity: finding.High, Domain: finding.Security, Category: "command_injection",
			Description:    "Shell command interpreter executed via os/exec",
			Recommendation: "Execute binary commands directly with argument arrays and sanitize untrusted input",
			Extensions:     []string{".go"},
			UnsafeExample:  `exec.Command("sh", "-c", "ls " + userInput)`,
			SafeExample:    `exec.Command("ls", "--", validatedPath)`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `cmd := exec.Command("sh", "-c", "ls " + userInput)`,
					Safe:   `cmd := exec.Command("ls", "--", validatedPath)`,
				},
				{
					Language: "ts", Label: "TypeScript / Node.js",
					Unsafe: `child_process.exec("ls " + userInput);`,
					Safe:   `child_process.execFile("ls", ["--", validatedPath]);`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `subprocess.Popen("ls " + user_input, shell=True)`,
					Safe:   `subprocess.Popen(["ls", "--", validated_path], shell=False)`,
				},
			},
		},
		{
			ID: "go-weak-cryptographic-hash", Pattern: `(md5|sha1)\.(New|Sum)\(`,
			Severity: finding.Medium, Domain: finding.Security, Category: "weak_cryptography",
			Description:    "Weak cryptographic hash algorithm (MD5/SHA1) detected",
			Recommendation: "Use SHA-256 or stronger algorithms; use bcrypt/argon2 for password hashing",
			Extensions:     []string{".go"},
			UnsafeExample: `hasher := md5.New()
hasher.Write([]byte(password))`,
			SafeExample: `hasher := sha256.New()
hasher.Write([]byte(password))`,
			Examples: []CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `hasher := md5.New()
hasher.Write([]byte(password))`,
					Safe: `hasher := sha256.New()
hasher.Write([]byte(password))`,
				},
				{
					Language: "ts", Label: "TypeScript / Node.js",
					Unsafe: `const hash = crypto.createHash("md5").update(password).digest("hex");`,
					Safe:   `const hash = crypto.createHash("sha256").update(password).digest("hex");`,
				},
				{
					Language: "python", Label: "Python",
					Unsafe: `hash_val = hashlib.md5(password.encode()).hexdigest()`,
					Safe:   `hash_val = hashlib.sha256(password.encode()).hexdigest()`,
				},
			},
		},
		{
			ID: "go-tainted-file-path", Pattern: `os\.(Open|OpenFile|ReadFile|WriteFile|Remove)\([^)]*(r\.URL\.Query|r\.FormValue|c\.Param)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "path_traversal",
			Description:    "Untrusted request parameter used directly in file system operation",
			Recommendation: "Normalize paths, enforce base directory boundaries, and use allowlisted identifiers",
			Extensions:     []string{".go"},
			UnsafeExample: `filePath := r.URL.Query().Get("file")
data, _ := os.ReadFile(filePath)`,
			SafeExample: `filename := filepath.Base(r.URL.Query().Get("file"))
safePath := filepath.Join("/var/app/storage", filename)
data, _ := os.ReadFile(safePath)`,
		},
		{
			ID: "go-weak-random-secret", Pattern: `(?i)(token|secret|nonce|session).{0,40}\brand\.(Int|Intn|Read|Uint32|Uint64)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "insecure_randomness",
			Description:    "Security-sensitive value generated using pseudo-random math/rand package",
			Recommendation: "Use crypto/rand for generating tokens, nonces, session identifiers, and secret keys",
			Extensions:     []string{".go"},
			UnsafeExample:  `sessionToken := fmt.Sprintf("%d", rand.Intn(1000000))`,
			SafeExample: `tokenBytes := make([]byte, 32)
cryptoRand.Read(tokenBytes)
sessionToken := hex.EncodeToString(tokenBytes)`,
		},
		{
			ID: "javascript-dynamic-eval", Pattern: `\beval\s*\([^)]*[A-Za-z_$]`,
			Severity: finding.High, Domain: finding.Security, Category: "unsafe_deserialization",
			Description:    "Dynamic eval execution of untrusted input detected",
			Recommendation: "Use structured data parsers (JSON.parse) and schema validators instead of code evaluation",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `const config = eval("(" + jsonString + ")");`,
			SafeExample:    `const config = JSON.parse(jsonString);`,
		},
		{
			ID: "node-prisma-raw-query", Pattern: `prisma\.\$(queryRawUnsafe|executeRawUnsafe)\(`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Prisma raw unsafe query executed with potentially untrusted dynamic string",
			Recommendation: "Use prisma.$queryRaw with tagged template literals (parameterized) instead of unsafe variants",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `const users = await prisma.$queryRawUnsafe(` + "`SELECT * FROM users WHERE id = '${id}'`" + `);`,
			SafeExample:    `const users = await prisma.$queryRaw` + "`SELECT * FROM users WHERE id = ${id}`" + `;`,
		},
		{
			ID: "node-typeorm-raw-query", Pattern: `(manager|connection|repository|queryRunner)\.query\([` + "`" + `\"'].*\$\{`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "TypeORM raw query with dynamic string interpolation",
			Recommendation: "Pass parameters as the second argument array to query() rather than template interpolation",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `await connection.query(` + "`SELECT * FROM users WHERE email = '${email}'`" + `);`,
			SafeExample:    `await connection.query("SELECT * FROM users WHERE email = $1", [email]);`,
		},
		{
			ID: "node-sequelize-raw-query", Pattern: `sequelize\.query\([` + "`" + `\"'].*\$\{`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Sequelize raw query executed with template string interpolation",
			Recommendation: "Use replacements or bind options in sequelize.query for safe parameter binding",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `await sequelize.query(` + "`SELECT * FROM users WHERE status = '${status}'`" + `);`,
			SafeExample:    `await sequelize.query("SELECT * FROM users WHERE status = :status", { replacements: { status } });`,
		},
		{
			ID: "node-pg-dynamic-query", Pattern: `(client|pool|db)\.query\([` + "`" + `\"'].*\$\{`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "node-postgres query executed with template string interpolation",
			Recommendation: "Use parameterized query format ($1, $2) and pass values in the values parameter array",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `await client.query(` + "`SELECT * FROM accounts WHERE id = '${id}'`" + `);`,
			SafeExample:    `await client.query("SELECT * FROM accounts WHERE id = $1", [id]);`,
		},
		{
			ID: "node-mysql-dynamic-query", Pattern: `(connection|pool)\.query\([` + "`" + `\"'].*\$\{`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "mysql2 query executed with dynamic template string interpolation",
			Recommendation: "Use query placeholders (?) and pass arguments in the parameter array",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample:  `await pool.query(` + "`SELECT * FROM products WHERE category = '${category}'`" + `);`,
			SafeExample:    `await pool.query("SELECT * FROM products WHERE category = ?", [category]);`,
		},
		{
			ID: "python-sqlalchemy-raw-sql", Pattern: `(text|session\.execute)\(f['\"]|format\(|%\s*\(|session\.execute\(["'].*\{`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "SQLAlchemy raw text expression formatted with dynamic Python f-string or format()",
			Recommendation: "Use bound parameters (:param_name) with session.execute(text(\"...\"), {\"param_name\": val})",
			Extensions:     []string{".py"},
			UnsafeExample:  `session.execute(text(f"SELECT * FROM users WHERE username = '{username}'"))`,
			SafeExample:    `session.execute(text("SELECT * FROM users WHERE username = :u"), {"u": username})`,
		},
		{
			ID: "python-django-raw-sql", Pattern: `(\.raw\(f['\"]|\.extra\(.*select=|\.extra\(.*where=|RawSQL\(f['\"])`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Django raw SQL query constructed with f-string or unsafe .extra() clause",
			Recommendation: "Pass parameters as params list to Model.objects.raw(query, [params]) or use standard ORM filters",
			Extensions:     []string{".py"},
			UnsafeExample:  `User.objects.raw(f"SELECT * FROM auth_user WHERE username = '{username}'")`,
			SafeExample:    `User.objects.raw("SELECT * FROM auth_user WHERE username = %s", [username])`,
		},
		{
			ID: "python-psycopg-format-query", Pattern: `cursor\.execute\(f['\"]|cursor\.execute\(['\"].*%s.*%\s*`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "psycopg database cursor executed with Python string formatting instead of query parameters",
			Recommendation: "Pass query parameters as the second tuple argument to cursor.execute(query, (param,))",
			Extensions:     []string{".py"},
			UnsafeExample:  `cursor.execute(f"SELECT * FROM items WHERE owner_id = '{owner_id}'")`,
			SafeExample:    `cursor.execute("SELECT * FROM items WHERE owner_id = %s", (owner_id,))`,
		},
		{
			ID: "java-spring-jpa-native-query", Pattern: `@Query\([^)]*\+`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Spring Data JPA native query built via string concatenation",
			Recommendation: "Use named parameters (:param) or positional parameters (?1) in native @Query annotations",
			Extensions:     []string{".java", ".kt"},
			UnsafeExample:  `@Query(value = "SELECT * FROM users WHERE role = '" + ROLE + "'", nativeQuery = true)`,
			SafeExample:    `@Query(value = "SELECT * FROM users WHERE role = :role", nativeQuery = true)`,
		},
		{
			ID: "java-hibernate-native-query", Pattern: `createNativeQuery\([^)]*\+`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Hibernate createNativeQuery executed with dynamic string concatenation",
			Recommendation: "Use parameterized placeholders and bind parameters via query.setParameter()",
			Extensions:     []string{".java", ".kt"},
			UnsafeExample:  `session.createNativeQuery("SELECT * FROM orders WHERE status = '" + status + "'")`,
			SafeExample:    `session.createNativeQuery("SELECT * FROM orders WHERE status = :status").setParameter("status", status)`,
		},
		{
			ID: "java-jdbc-dynamic-query", Pattern: `jdbcTemplate\.(query|update|execute)\([^)]*\+`,
			Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Spring JdbcTemplate executed with concatenated SQL string",
			Recommendation: "Pass query parameters as separate Object[] or varargs to jdbcTemplate",
			Extensions:     []string{".java", ".kt"},
			UnsafeExample:  `jdbcTemplate.query("SELECT * FROM users WHERE id = " + id, rowMapper)`,
			SafeExample:    `jdbcTemplate.query("SELECT * FROM users WHERE id = ?", rowMapper, id)`,
		},
		{
			ID: "DBSEC-002", Pattern: `(log\.(Print|Info|Debug|Warn)|logger\.(info|debug|warn)|console\.(log|debug))\s*\([^\n]*(password|token|secret_key|api_key|credit_card|cvv|auth_header)`,
			Severity: finding.High, Domain: finding.Security, Category: "data_leak",
			Description:    "Sensitive credentials or PII fields logged to application tracing stream",
			Recommendation: "Redact credentials, tokens, and payment card details before writing to log sinks",
			Extensions:     []string{".go", ".ts", ".js", ".py", ".java"},
			UnsafeExample:  `logger.info("Processing payment for card:", cardToken, secretKey);`,
			SafeExample:    `logger.info("Processing payment for transaction ID:", transactionId);`,
		},
		{
			ID: "DBSEC-003", Pattern: `(c\.JSON|c\.String|http\.Error|res\.send|res\.json)\s*\([^\n]*(err\.Error\(\)|error\.message|driverErr|sqlErr)`,
			Severity: finding.High, Domain: finding.Security, Category: "information_exposure",
			Description:    "Internal database driver error exposed directly in HTTP client response",
			Recommendation: "Log the internal database error securely on the server and return a sanitized, generic error message to the client",
			Extensions:     []string{".go", ".ts", ".js", ".py", ".java"},
			UnsafeExample:  `c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})`,
			SafeExample:    `c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})`,
		},
	}
}
