---
title: Rule Catalog
description: Complete catalog of default built-in security, secret, governance, and quality rules with Do's and Don'ts code examples.
---

# Built-In Rule Catalog

Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog includes detailed guidance, recommendations, and **Do's and Don'ts** code examples for each rule.

## Domain Overview

| Domain | Icon | Total Rules | Scope & Focus |
| :--- | :---: | :---: | :--- |
| **Security Rules** | 🔒 | 17 | Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws. |
| **Hardening Rules** | 🛡️ | 6 | Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings. |
| **Reliability Rules** | ⚡ | 6 | Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes. |
| **Quality Rules** | 🧹 | 5 | Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements. |
| **Governance Rules** | 📜 | 4 | Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints. |

---

## 🔒 Security Rules

Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`mock-token`](#mock-token) | `CRITICAL` | `secret_leak` | Hardcoded mock token found — remove before production deployment |
| [`browser-token-storage`](#browser-token-storage) | `HIGH` | `data_leak` | Token stored in localStorage — vulnerable to XSS token theft |
| [`permission-bypass`](#permission-bypass) | `CRITICAL` | `security_misconfiguration` | Hardcoded permission bypass found in application logic |
| [`weak-secret`](#weak-secret) | `CRITICAL` | `secret_leak` | Default or weak secret value found |
| [`frontend-sensitive-log`](#frontend-sensitive-log) | `MEDIUM` | `data_leak` | Frontend log statement may expose sensitive credentials or PII |
| [`backend-sensitive-log`](#backend-sensitive-log) | `MEDIUM` | `data_leak` | Backend log statement may expose sensitive credentials or keys |
| [`sql-string-format`](#sql-string-format) | `HIGH` | `injection` | Potential SQL injection using formatted strings |
| [`hardcoded-credential`](#hardcoded-credential) | `HIGH` | `secret_leak` | Hardcoded credential or API secret key found |
| [`unsafe-inner-html`](#unsafe-inner-html) | `HIGH` | `xss` | dangerouslySetInnerHTML used — potential DOM XSS vulnerability |
| [`dynamic-order`](#dynamic-order) | `HIGH` | `injection` | Dynamic ORDER BY clause built via string formatting |
| [`api-struct-response`](#api-struct-response) | `HIGH` | `data_leak` | Internal domain struct may be serialized directly into HTTP response |
| [`sensitive-json-field`](#sensitive-json-field) | `HIGH` | `data_leak` | Sensitive struct field may be exposed in JSON serialization |
| [`go-shell-command`](#go-shell-command) | `HIGH` | `command_injection` | Shell command interpreter executed via os/exec |
| [`go-weak-cryptographic-hash`](#go-weak-cryptographic-hash) | `MEDIUM` | `weak_cryptography` | Weak cryptographic hash algorithm (MD5/SHA1) detected |
| [`go-tainted-file-path`](#go-tainted-file-path) | `HIGH` | `path_traversal` | Untrusted request parameter used directly in file system operation |
| [`go-weak-random-secret`](#go-weak-random-secret) | `HIGH` | `insecure_randomness` | Security-sensitive value generated using pseudo-random math/rand package |
| [`javascript-dynamic-eval`](#javascript-dynamic-eval) | `HIGH` | `unsafe_deserialization` | Dynamic eval execution of untrusted input detected |

### Details & Guidance

#### `mock-token`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Hardcoded mock token found — remove before production deployment

**Recommendation**: Remove hardcoded mock tokens and load credentials from environment variables or key vaults

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
const authHeader = "Bearer google-mock-jwt-token-12345"
```

```ts [TypeScript / JavaScript]
const AUTH_HEADER = "Bearer google-mock-jwt-token-12345";
```

```python [Python]
AUTH_HEADER = "Bearer google-mock-jwt-token-12345"
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
authHeader := fmt.Sprintf("Bearer %s", os.Getenv("AUTH_TOKEN"))
```

```ts [TypeScript / JavaScript]
const AUTH_HEADER = `Bearer ${process.env.AUTH_TOKEN}`;
```

```python [Python]
auth_header = f"Bearer {os.environ.get('AUTH_TOKEN')}"
```

:::

---

#### `browser-token-storage`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Token stored in localStorage — vulnerable to XSS token theft

**Recommendation**: Store authentication tokens in HttpOnly, Secure, SameSite cookies instead of localStorage

##### ❌ Don't (Unsafe)

```ts
localStorage.setItem("access_token", response.token);
```

##### ✅ Do (Recommended)

```ts
await fetch("/api/login", { credentials: "include", method: "POST", body });
```

---

#### `permission-bypass`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `security_misconfiguration`

**Description**: Hardcoded permission bypass found in application logic

**Recommendation**: Remove permission bypass conditions and enforce strict authorization checks

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
func CheckPermission(user User) bool {
    if user.Role == "admin" || bypassPermission {
        return true
    }
    return false
}
```

```ts [TypeScript / JavaScript]
function checkPermission(user: User): boolean {
    if (user.role === 'admin' || process.env.BYPASS_PERMISSIONS === 'true') {
        return true;
    }
    return false;
}
```

```python [Python]
def check_permission(user):
    if user.role == "admin" or bypass_permission:
        return True
    return False
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
func CheckPermission(ctx context.Context, user User, resource string) bool {
    return authzService.CanAccess(ctx, user.ID, resource)
}
```

```ts [TypeScript / JavaScript]
async function checkPermission(user: User, resource: string): Promise<boolean> {
    return await authzService.canAccess(user.id, resource);
}
```

```python [Python]
def check_permission(user, resource):
    return authz_service.can_access(user.id, resource)
```

:::

---

#### `weak-secret`

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Default or weak secret value found

**Recommendation**: Replace default/placeholder secrets with cryptographically strong random values from secure configuration

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
jwtSecret := []byte("change-me-in-production")
```

```ts [TypeScript / JavaScript]
const jwtSecret = "change-me-in-production";
```

```python [Python]
JWT_SECRET = "change-me-in-production"
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))
```

```ts [TypeScript / JavaScript]
const jwtSecret = process.env.JWT_SECRET_KEY;
```

```python [Python]
JWT_SECRET = os.environ.get("JWT_SECRET_KEY")
```

:::

---

#### `frontend-sensitive-log`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Frontend log statement may expose sensitive credentials or PII

**Recommendation**: Sanitize log parameters and remove sensitive tokens or user identifiers from console logs

##### ❌ Don't (Unsafe)

```ts
console.log("User auth failed for password:", password);
```

##### ✅ Do (Recommended)

```ts
console.error("User authentication failed", { username });
```

---

#### `backend-sensitive-log`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Backend log statement may expose sensitive credentials or keys

**Recommendation**: Redact sensitive parameters before writing to application log streams

##### ❌ Don't (Unsafe)

```go
log.Printf("Connecting to DB with secret: %s", dbSecret)
```

##### ✅ Do (Recommended)

```go
log.Printf("Connecting to DB host: %s", dbHost)
```

---

#### `sql-string-format`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Potential SQL injection using formatted strings

**Recommendation**: Use parameterized queries or prepared statements instead of string formatting

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)
rows, err := db.Query(query)
```

```ts [TypeScript / JavaScript]
const query = `SELECT * FROM users WHERE email = '${userEmail}'`;
const result = await client.query(query);
```

```python [Python]
query = f"SELECT * FROM users WHERE email = '{user_email}'"
cursor.execute(query)
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
query := "SELECT * FROM users WHERE email = $1"
rows, err := db.Query(query, userEmail)
```

```ts [TypeScript / JavaScript]
const query = "SELECT * FROM users WHERE email = $1";
const result = await client.query(query, [userEmail]);
```

```python [Python]
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_email,))
```

:::

---

#### `hardcoded-credential`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `secret_leak`

**Description**: Hardcoded credential or API secret key found

**Recommendation**: Extract credentials to environment variables or secret management services

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
apiKey := "synthetic_secret_api_key_12345"
```

```ts [TypeScript / JavaScript]
const apiKey = "synthetic_secret_api_key_12345";
```

```python [Python]
api_key = "synthetic_secret_api_key_12345"
```

```java [Java]
String apiKey = "synthetic_secret_api_key_12345";
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
apiKey := os.Getenv("STRIPE_API_KEY")
```

```ts [TypeScript / JavaScript]
const apiKey = process.env.STRIPE_API_KEY;
```

```python [Python]
api_key = os.environ.get("STRIPE_API_KEY")
```

```java [Java]
String apiKey = System.getenv("STRIPE_API_KEY");
```

:::

---

#### `unsafe-inner-html`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `xss`

**Description**: dangerouslySetInnerHTML used — potential DOM XSS vulnerability

**Recommendation**: Sanitize raw HTML using DOMPurify before injecting into the DOM

##### ❌ Don't (Unsafe)

```ts
<div dangerouslySetInnerHTML={{ __html: userInput }} />
```

##### ✅ Do (Recommended)

```ts
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(userInput) }} />
```

---

#### `dynamic-order`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Dynamic ORDER BY clause built via string formatting

**Recommendation**: Validate dynamic column names against an explicit allowlist before building queries

##### ❌ Don't (Unsafe)

```go
db.Order(fmt.Sprintf("%s ASC", sortColumn))
```

##### ✅ Do (Recommended)

```go
allowedColumns := map[string]bool{"created_at": true, "name": true}
if allowedColumns[sortColumn] {
    db.Order(sortColumn + " ASC")
}
```

---

#### `api-struct-response`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Internal domain struct may be serialized directly into HTTP response

**Recommendation**: Map internal domain entities to explicit response DTOs to avoid leaking sensitive fields

##### ❌ Don't (Unsafe)

```go
var user User // Contains HashedPassword, SecretToken
c.JSON(http.StatusOK, user)
```

##### ✅ Do (Recommended)

```go
response := UserResponse{ID: user.ID, Email: user.Email}
c.JSON(http.StatusOK, response)
```

---

#### `sensitive-json-field`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive struct field may be exposed in JSON serialization

**Recommendation**: Use json:"-" struct tag or custom serializer to exclude sensitive attributes

##### ❌ Don't (Unsafe)

```go
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"password_hash"`
}
```

##### ✅ Do (Recommended)

```go
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"-"`
}
```

---

#### `go-shell-command`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `command_injection`

**Description**: Shell command interpreter executed via os/exec

**Recommendation**: Execute binary commands directly with argument arrays and sanitize untrusted input

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
cmd := exec.Command("sh", "-c", "ls " + userInput)
```

```ts [TypeScript / Node.js]
child_process.exec("ls " + userInput);
```

```python [Python]
subprocess.Popen("ls " + user_input, shell=True)
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
cmd := exec.Command("ls", "--", validatedPath)
```

```ts [TypeScript / Node.js]
child_process.execFile("ls", ["--", validatedPath]);
```

```python [Python]
subprocess.Popen(["ls", "--", validated_path], shell=False)
```

:::

---

#### `go-weak-cryptographic-hash`

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `weak_cryptography`

**Description**: Weak cryptographic hash algorithm (MD5/SHA1) detected

**Recommendation**: Use SHA-256 or stronger algorithms; use bcrypt/argon2 for password hashing

##### ❌ Don't (Unsafe Examples)

::: code-group

```go [Go]
hasher := md5.New()
hasher.Write([]byte(password))
```

```ts [TypeScript / Node.js]
const hash = crypto.createHash("md5").update(password).digest("hex");
```

```python [Python]
hash_val = hashlib.md5(password.encode()).hexdigest()
```

:::

##### ✅ Do (Recommended Solutions)

::: code-group

```go [Go]
hasher := sha256.New()
hasher.Write([]byte(password))
```

```ts [TypeScript / Node.js]
const hash = crypto.createHash("sha256").update(password).digest("hex");
```

```python [Python]
hash_val = hashlib.sha256(password.encode()).hexdigest()
```

:::

---

#### `go-tainted-file-path`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `path_traversal`

**Description**: Untrusted request parameter used directly in file system operation

**Recommendation**: Normalize paths, enforce base directory boundaries, and use allowlisted identifiers

##### ❌ Don't (Unsafe)

```go
filePath := r.URL.Query().Get("file")
data, _ := os.ReadFile(filePath)
```

##### ✅ Do (Recommended)

```go
filename := filepath.Base(r.URL.Query().Get("file"))
safePath := filepath.Join("/var/app/storage", filename)
data, _ := os.ReadFile(safePath)
```

---

#### `go-weak-random-secret`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `insecure_randomness`

**Description**: Security-sensitive value generated using pseudo-random math/rand package

**Recommendation**: Use crypto/rand for generating tokens, nonces, session identifiers, and secret keys

##### ❌ Don't (Unsafe)

```go
sessionToken := fmt.Sprintf("%d", rand.Intn(1000000))
```

##### ✅ Do (Recommended)

```go
tokenBytes := make([]byte, 32)
cryptoRand.Read(tokenBytes)
sessionToken := hex.EncodeToString(tokenBytes)
```

---

#### `javascript-dynamic-eval`

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `unsafe_deserialization`

**Description**: Dynamic eval execution of untrusted input detected

**Recommendation**: Use structured data parsers (JSON.parse) and schema validators instead of code evaluation

##### ❌ Don't (Unsafe)

```ts
const config = eval("(" + jsonString + ")");
```

##### ✅ Do (Recommended)

```ts
const config = JSON.parse(jsonString);
```

---

## 🛡️ Hardening Rules

Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`hardcoded-api-url`](#hardcoded-api-url) | `MEDIUM` | `configuration_leak` | Hardcoded localhost API URL found — load dynamically from environment variable |
| [`tls-insecure-skip-verify`](#tls-insecure-skip-verify) | `HIGH` | `transport_security` | TLS certificate verification is explicitly disabled |
| [`wildcard-cors-origin`](#wildcard-cors-origin) | `HIGH` | `cors` | Wildcard CORS origin header found in configuration |
| [`go-permissive-file-mode`](#go-permissive-file-mode) | `MEDIUM` | `file_permission` | File or directory created with permissive world-writable file permissions (0777) |
| [`debug-mode-enabled`](#debug-mode-enabled) | `MEDIUM` | `debug_configuration` | Debug mode appears to be explicitly enabled in configuration |
| [`go-insecure-cookie-attribute`](#go-insecure-cookie-attribute) | `HIGH` | `cookie_security` | Cookie configured with explicitly insecure security attributes |

### Details & Guidance

#### `hardcoded-api-url`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `configuration_leak`

**Description**: Hardcoded localhost API URL found — load dynamically from environment variable

**Recommendation**: Configure API endpoints dynamically via environment variables for different environments

##### ❌ Don't (Unsafe)

```go
const API_URL = "http://localhost:8080/api/v1";
```

##### ✅ Do (Recommended)

```go
const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";
```

---

#### `tls-insecure-skip-verify`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `transport_security`

**Description**: TLS certificate verification is explicitly disabled

**Recommendation**: Enable certificate verification and configure valid trust stores

##### ❌ Don't (Unsafe)

```go
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
```

##### ✅ Do (Recommended)

```go
tr := &http.Transport{
    TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
}
```

---

#### `wildcard-cors-origin`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cors`

**Description**: Wildcard CORS origin header found in configuration

**Recommendation**: Use an explicit CORS origin allowlist tailored for each deployment environment

##### ❌ Don't (Unsafe)

```go
c.Header("Access-Control-Allow-Origin", "*")
```

##### ✅ Do (Recommended)

```go
c.Header("Access-Control-Allow-Origin", "https://app.example.com")
```

---

#### `go-permissive-file-mode`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `file_permission`

**Description**: File or directory created with permissive world-writable file permissions (0777)

**Recommendation**: Use minimum required file permissions such as 0600 for files or 0750 for directories

##### ❌ Don't (Unsafe)

```go
os.WriteFile("config.json", data, 0777)
```

##### ✅ Do (Recommended)

```go
os.WriteFile("config.json", data, 0600)
```

---

#### `debug-mode-enabled`

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `debug_configuration`

**Description**: Debug mode appears to be explicitly enabled in configuration

**Recommendation**: Disable debug mode in production deployment configurations to prevent information disclosure

##### ❌ Don't (Unsafe)

```go
debug := true
```

##### ✅ Do (Recommended)

```go
debug := os.Getenv("APP_ENV") == "development"
```

---

#### `go-insecure-cookie-attribute`

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cookie_security`

**Description**: Cookie configured with explicitly insecure security attributes

**Recommendation**: Enable Secure and HttpOnly flags and set an appropriate SameSite policy for session cookies

##### ❌ Don't (Unsafe)

```go
cookie := &http.Cookie{Name: "session", Value: token, Secure: false}
```

##### ✅ Do (Recommended)

```go
cookie := &http.Cookie{Name: "session", Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
```

---

## ⚡ Reliability Rules

Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`go-multipart-memory`](#go-multipart-memory) | `MEDIUM` | `resource_exhaustion` | Ensure multipart request processing configures explicit memory limits |
| [`go-http-default-server`](#go-http-default-server) | `MEDIUM` | `missing_timeout` | Default HTTP server does not configure defensive timeouts |
| [`go-unbounded-request-read`](#go-unbounded-request-read) | `MEDIUM` | `resource_exhaustion` | Request body may be read without explicit size limits |
| [`go-discarded-error`](#go-discarded-error) | `MEDIUM` | `error_handling` | Returned error value is explicitly ignored with blank identifier |
| [`go-process-termination`](#go-process-termination) | `MEDIUM` | `process_termination` | Application path may terminate entire process unexpectedly |
| [`go-http-client-without-timeout`](#go-http-client-without-timeout) | `MEDIUM` | `missing_timeout` | HTTP client struct literal does not set an overall request timeout |

### Details & Guidance

#### `go-multipart-memory`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Ensure multipart request processing configures explicit memory limits

**Recommendation**: Set explicit memory limit with ParseMultipartForm or MaxBytesReader to prevent memory exhaustion

##### ❌ Don't (Unsafe)

```go
c.Request.ParseMultipartForm(100 << 20) // Unbounded 100MB buffer
```

##### ✅ Do (Recommended)

```go
c.Request.ParseMultipartForm(10 << 20) // Controlled 10MB memory limit
```

---

#### `go-http-default-server`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: Default HTTP server does not configure defensive timeouts

**Recommendation**: Use custom http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout, and IdleTimeout

##### ❌ Don't (Unsafe)

```go
http.ListenAndServe(":8080", handler)
```

##### ✅ Do (Recommended)

```go
server := &http.Server{
    Addr: ":8080", Handler: handler,
    ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
}
server.ListenAndServe()
```

---

#### `go-unbounded-request-read`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Request body may be read without explicit size limits

**Recommendation**: Limit request body with http.MaxBytesReader or io.LimitReader before reading into memory

##### ❌ Don't (Unsafe)

```go
body, err := io.ReadAll(r.Body)
```

##### ✅ Do (Recommended)

```go
body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB size limit
```

---

#### `go-discarded-error`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `error_handling`

**Description**: Returned error value is explicitly ignored with blank identifier

**Recommendation**: Check and handle returned errors or document valid reason for ignoring

##### ❌ Don't (Unsafe)

```go
_ = db.Close()
```

##### ✅ Do (Recommended)

```go
if err := db.Close(); err != nil {
    log.Printf("Failed to close DB connection: %v", err)
}
```

---

#### `go-process-termination`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `process_termination`

**Description**: Application path may terminate entire process unexpectedly

**Recommendation**: Propagate errors to request boundaries and perform controlled shutdown instead of calling panic/log.Fatal

##### ❌ Don't (Unsafe)

```go
if err != nil {
    panic(err)
}
```

##### ✅ Do (Recommended)

```go
if err != nil {
    return fmt.Errorf("process request: %w", err)
}
```

---

#### `go-http-client-without-timeout`

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: HTTP client struct literal does not set an overall request timeout

**Recommendation**: Configure explicit http.Client.Timeout and appropriate transport timeouts

##### ❌ Don't (Unsafe)

```go
client := &http.Client{}
```

##### ✅ Do (Recommended)

```go
client := &http.Client{Timeout: 10 * time.Second}
```

---

## 🧹 Quality Rules

Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`merge-conflict-marker`](#merge-conflict-marker) | `HIGH` | `repository_hygiene` | Unresolved merge-conflict marker found |
| [`javascript-debugger`](#javascript-debugger) | `MEDIUM` | `debug_code` | JavaScript debugger statement found |
| [`trailing-whitespace`](#trailing-whitespace) | `LOW` | `formatting` | Trailing whitespace found at end of line |
| [`mixed-indentation`](#mixed-indentation) | `LOW` | `formatting` | Mixed tabs and spaces used for indentation on the same line |
| [`javascript-console-debug`](#javascript-console-debug) | `LOW` | `debug_code` | Console debug statement left in code |

### Details & Guidance

#### `merge-conflict-marker`

- **Domain**: `quality`
- **Severity**: `HIGH`
- **Category**: `repository_hygiene`

**Description**: Unresolved merge-conflict marker found

**Recommendation**: Resolve merge conflict and remove all markers before committing

##### ❌ Don't (Unsafe)

```go
<<<<<<< HEAD
const apiURL = "http://localhost:8080";
=======
const apiURL = "https://api.production.com";
>>>>>>> main
```

##### ✅ Do (Recommended)

```go
const apiURL = process.env.API_URL || "https://api.production.com";
```

---

#### `javascript-debugger`

- **Domain**: `quality`
- **Severity**: `MEDIUM`
- **Category**: `debug_code`

**Description**: JavaScript debugger statement found

**Recommendation**: Remove debugger statement before committing

##### ❌ Don't (Unsafe)

```ts
function calculateTotal(items: Item[]) {
  debugger; // Leftover debug statement
  return items.reduce((acc, item) => acc + item.price, 0);
}
```

##### ✅ Do (Recommended)

```ts
function calculateTotal(items: Item[]) {
  return items.reduce((acc, item) => acc + item.price, 0);
}
```

---

#### `trailing-whitespace`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Trailing whitespace found at end of line

**Recommendation**: Remove trailing whitespace at line end

##### ❌ Don't (Unsafe)

```go
const username = "john_doe";
```

##### ✅ Do (Recommended)

```go
const username = "john_doe";
```

---

#### `mixed-indentation`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Mixed tabs and spaces used for indentation on the same line

**Recommendation**: Use a consistent indentation style throughout the project

##### ❌ Don't (Unsafe)

```go
func process() {
	  var x = 10 // Mixed tabs and spaces
}
```

##### ✅ Do (Recommended)

```go
func process() {
	var x = 10 // Consistent tab indentation
}
```

---

#### `javascript-console-debug`

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `debug_code`

**Description**: Console debug statement left in code

**Recommendation**: Remove debug statements or use an application logger with proper log level

##### ❌ Don't (Unsafe)

```ts
function handleLogin(user: User) {
  console.log("User logged in:", user);
}
```

##### ✅ Do (Recommended)

```ts
function handleLogin(user: User) {
  logger.info("User logged in", { userId: user.id });
}
```

---

## 📜 Governance Rules

Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`privacy-pii-log`](#privacy-pii-log) | `HIGH` | `privacy_log` | Logging statement may expose personally identifiable information |
| [`privacy-pii-url`](#privacy-pii-url) | `HIGH` | `privacy_url` | Personally identifiable information may be placed in a URL query string |
| [`privacy-pii-fixture`](#privacy-pii-fixture) | `MEDIUM` | `privacy_fixture` | Fixture may contain a literal personally identifiable value |
| [`privacy-sensitive-response`](#privacy-sensitive-response) | `HIGH` | `privacy_response` | Response construction may expose a sensitive personal field |

### Details & Guidance

#### `privacy-pii-log`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_log`

**Description**: Logging statement may expose personally identifiable information

**Recommendation**: Remove the PII field or log a non-reversible, access-controlled reference identifier

##### ❌ Don't (Unsafe)

```go
log.Printf("User registered with email: %s, phone: %s", email, phone)
```

##### ✅ Do (Recommended)

```go
log.Printf("User registered with ID: %s", userID)
```

---

#### `privacy-pii-url`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_url`

**Description**: Personally identifiable information may be placed in a URL query string

**Recommendation**: Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs

##### ❌ Don't (Unsafe)

```go
urlParams.append("email", userEmail);
```

##### ✅ Do (Recommended)

```go
// Transmit sensitive parameters in authenticated POST request body
```

---

#### `privacy-pii-fixture`

- **Domain**: `governance`
- **Severity**: `MEDIUM`
- **Category**: `privacy_fixture`

**Description**: Fixture may contain a literal personally identifiable value

**Recommendation**: Use clearly synthetic, reserved test data and keep production-derived records out of the repository

##### ❌ Don't (Unsafe)

```json
{"email": "real_person_1985@gmail.com", "ssn": "123-45-6789"}
```

##### ✅ Do (Recommended)

```json
{"email": "user@example.com", "ssn": "000-00-0000"}
```

---

#### `privacy-sensitive-response`

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_response`

**Description**: Response construction may expose a sensitive personal field

**Recommendation**: Map the response through an explicit allowlisted DTO and omit sensitive fields

##### ❌ Don't (Unsafe)

```go
res.json({ id: user.id, email: user.email, ssn: user.ssn });
```

##### ✅ Do (Recommended)

```go
res.json({ id: user.id, email: user.email });
```

---

