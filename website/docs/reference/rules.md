---
title: Legacy Rule Reference
description: "For maintainers preserving legacy anchors: inspect the generated monolithic rule reference and its remediation examples."
---

# Legacy Rule Reference

This generated compatibility page preserves historical rule anchors. Use the [Rule Catalog](/reference/rule-catalog) to search rules and open focused remediation pages.

## Domain Overview

| Domain | Icon | Total Rules | Scope & Focus |
| :--- | :---: | :---: | :--- |
| **[Security Rules](#security-rules)** | 🔒 | 40 | Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws. |
| **[Hardening Rules](#hardening-rules)** | 🛡️ | 6 | Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings. |
| **[Reliability Rules](#reliability-rules)** | ⚡ | 16 | Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes. |
| **[Quality Rules](#quality-rules)** | 🧹 | 5 | Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements. |
| **[Supply Chain Rules](#supply-chain-rules)** | 📦 | 0 | Rules auditing third-party dependencies, version pins, package vulnerabilities, and license restrictions. |
| **[Governance Rules](#governance-rules)** | 📜 | 4 | Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints. |

---

## 🔒 Security Rules {#security-rules}

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
| [`node-prisma-raw-query`](#node-prisma-raw-query) | `HIGH` | `sql-injection` | Prisma raw unsafe query executed with potentially untrusted dynamic string |
| [`node-typeorm-raw-query`](#node-typeorm-raw-query) | `HIGH` | `sql-injection` | TypeORM raw query with dynamic string interpolation |
| [`node-sequelize-raw-query`](#node-sequelize-raw-query) | `HIGH` | `sql-injection` | Sequelize raw query executed with template string interpolation |
| [`node-pg-dynamic-query`](#node-pg-dynamic-query) | `HIGH` | `sql-injection` | node-postgres query executed with template string interpolation |
| [`node-mysql-dynamic-query`](#node-mysql-dynamic-query) | `HIGH` | `sql-injection` | mysql2 query executed with dynamic template string interpolation |
| [`python-sqlalchemy-raw-sql`](#python-sqlalchemy-raw-sql) | `HIGH` | `sql-injection` | SQLAlchemy raw text expression formatted with dynamic Python f-string or format() |
| [`python-django-raw-sql`](#python-django-raw-sql) | `HIGH` | `sql-injection` | Django raw SQL query constructed with f-string or unsafe .extra() clause |
| [`python-psycopg-format-query`](#python-psycopg-format-query) | `HIGH` | `sql-injection` | psycopg database cursor executed with Python string formatting instead of query parameters |
| [`java-spring-jpa-native-query`](#java-spring-jpa-native-query) | `HIGH` | `sql-injection` | Spring Data JPA native query built via string concatenation |
| [`java-hibernate-native-query`](#java-hibernate-native-query) | `HIGH` | `sql-injection` | Hibernate createNativeQuery executed with dynamic string concatenation |
| [`java-jdbc-dynamic-query`](#java-jdbc-dynamic-query) | `HIGH` | `sql-injection` | Spring JdbcTemplate executed with concatenated SQL string |
| [`DBSEC-002`](#dbsec-002) | `HIGH` | `data_leak` | Sensitive credentials or PII fields logged to application tracing stream |
| [`DBSEC-003`](#dbsec-003) | `HIGH` | `information_exposure` | Internal database driver error exposed directly in HTTP client response |
| [`SQLI-001`](#sqli-001) | `HIGH` | `sql-injection` | Untrusted value concatenated or formatted into executable SQL at database driver sink |
| [`SQLI-002`](#sqli-002) | `HIGH` | `sql-injection` | Untrusted table, column, or identifier dynamically interpolated into SQL |
| [`SQLI-004`](#sqli-004) | `HIGH` | `orm-escape-hatch` | Unsafe raw ORM escape hatch called with dynamic or concatenated string |
| [`SQLI-008`](#sqli-008) | `MEDIUM` | `bind-mismatch` | SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed |
| [`SQLI-011`](#sqli-011) | `HIGH` | `list-expansion` | Unsafe list or IN clause expansion using strings.Join or manual string interpolation |
| [`SQLI-012`](#sqli-012) | `HIGH` | `prepared-statement` | Tainted SQL query template passed into statement preparation method db.Prepare() |
| [`SQLAUTH-001`](#sqlauth-001) | `HIGH` | `multi-tenant-isolation` | Multi-tenant entity queried without tenant_id or organization_id scoping constraint |
| [`SQLAUTH-002`](#sqlauth-002) | `HIGH` | `authorization-idor` | Sensitive resource queried solely by object ID without user ownership scoping (IDOR risk) |
| [`SQLAUTH-003`](#sqlauth-003) | `HIGH` | `raw-query-bypass` | Raw query bypasses standard ORM authorization scopes and permission filters |
| [`SQLAUTH-004`](#sqlauth-004) | `HIGH` | `rls-misconfiguration` | Database query assumes Row-Level Security but explicitly switches to superuser or bypass role |

### Details and Guidance

#### `mock-token` {#mock-token}

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Hardcoded mock token found — remove before production deployment

**Recommendation**: Remove hardcoded mock tokens and load credentials from environment variables or key vaults

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
const authHeader = "Bearer google-mock-jwt-token-12345"

// Safer example
authHeader := fmt.Sprintf("Bearer %s", os.Getenv("AUTH_TOKEN"))
```

```ts [TypeScript / JavaScript]
// Unsafe example
const AUTH_HEADER = "Bearer google-mock-jwt-token-12345";

// Safer example
const AUTH_HEADER = `Bearer ${process.env.AUTH_TOKEN}`;
```

```python [Python]
# Unsafe example
AUTH_HEADER = "Bearer google-mock-jwt-token-12345"

# Safer example
auth_header = f"Bearer {os.environ.get('AUTH_TOKEN')}"
```

:::

<p class="rule-nav">[↑ Back to Security Rules](#security-rules) | [`browser-token-storage`](#browser-token-storage) →</p>

---

#### `browser-token-storage` {#browser-token-storage}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Token stored in localStorage — vulnerable to XSS token theft

**Recommendation**: Store authentication tokens in HttpOnly, Secure, SameSite cookies instead of localStorage

##### Unsafe and Safer Example

```ts
// Unsafe example
localStorage.setItem("access_token", response.token);

// Safer example
await fetch("/api/login", { credentials: "include", method: "POST", body });
```

<p class="rule-nav">← [`mock-token`](#mock-token) | [↑ Back to Security Rules](#security-rules) | [`permission-bypass`](#permission-bypass) →</p>

---

#### `permission-bypass` {#permission-bypass}

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `security_misconfiguration`

**Description**: Hardcoded permission bypass found in application logic

**Recommendation**: Remove permission bypass conditions and enforce strict authorization checks

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func CheckPermission(user User) bool {
    if user.Role == "admin" || bypassPermission {
        return true
    }
    return false
}

// Safer example
func CheckPermission(ctx context.Context, user User, resource string) bool {
    return authzService.CanAccess(ctx, user.ID, resource)
}
```

```ts [TypeScript / JavaScript]
// Unsafe example
function checkPermission(user: User): boolean {
    if (user.role === 'admin' || process.env.BYPASS_PERMISSIONS === 'true') {
        return true;
    }
    return false;
}

// Safer example
async function checkPermission(user: User, resource: string): Promise<boolean> {
    return await authzService.canAccess(user.id, resource);
}
```

```python [Python]
# Unsafe example
def check_permission(user):
    if user.role == "admin" or bypass_permission:
        return True
    return False

# Safer example
def check_permission(user, resource):
    return authz_service.can_access(user.id, resource)
```

:::

<p class="rule-nav">← [`browser-token-storage`](#browser-token-storage) | [↑ Back to Security Rules](#security-rules) | [`weak-secret`](#weak-secret) →</p>

---

#### `weak-secret` {#weak-secret}

- **Domain**: `security`
- **Severity**: `CRITICAL`
- **Category**: `secret_leak`

**Description**: Default or weak secret value found

**Recommendation**: Replace default/placeholder secrets with cryptographically strong random values from secure configuration

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
jwtSecret := []byte("change-me-in-production")

// Safer example
jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))
```

```ts [TypeScript / JavaScript]
// Unsafe example
const jwtSecret = "change-me-in-production";

// Safer example
const jwtSecret = process.env.JWT_SECRET_KEY;
```

```python [Python]
# Unsafe example
JWT_SECRET = "change-me-in-production"

# Safer example
JWT_SECRET = os.environ.get("JWT_SECRET_KEY")
```

:::

<p class="rule-nav">← [`permission-bypass`](#permission-bypass) | [↑ Back to Security Rules](#security-rules) | [`frontend-sensitive-log`](#frontend-sensitive-log) →</p>

---

#### `frontend-sensitive-log` {#frontend-sensitive-log}

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Frontend log statement may expose sensitive credentials or PII

**Recommendation**: Sanitize log parameters and remove sensitive tokens or user identifiers from console logs

##### Unsafe and Safer Example

```ts
// Unsafe example
console.log("User auth failed for password:", password);

// Safer example
console.error("User authentication failed", { username });
```

<p class="rule-nav">← [`weak-secret`](#weak-secret) | [↑ Back to Security Rules](#security-rules) | [`backend-sensitive-log`](#backend-sensitive-log) →</p>

---

#### `backend-sensitive-log` {#backend-sensitive-log}

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `data_leak`

**Description**: Backend log statement may expose sensitive credentials or keys

**Recommendation**: Redact sensitive parameters before writing to application log streams

##### Unsafe and Safer Example

```go
// Unsafe example
log.Printf("Connecting to DB with secret: %s", dbSecret)

// Safer example
log.Printf("Connecting to DB host: %s", dbHost)
```

<p class="rule-nav">← [`frontend-sensitive-log`](#frontend-sensitive-log) | [↑ Back to Security Rules](#security-rules) | [`sql-string-format`](#sql-string-format) →</p>

---

#### `sql-string-format` {#sql-string-format}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Potential SQL injection using formatted strings

**Recommendation**: Use parameterized queries or prepared statements instead of string formatting

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)
rows, err := db.Query(query)

// Safer example
query := "SELECT * FROM users WHERE email = $1"
rows, err := db.Query(query, userEmail)
```

```ts [TypeScript / JavaScript]
// Unsafe example
const query = `SELECT * FROM users WHERE email = '${userEmail}'`;
const result = await client.query(query);

// Safer example
const query = "SELECT * FROM users WHERE email = $1";
const result = await client.query(query, [userEmail]);
```

```python [Python]
# Unsafe example
query = f"SELECT * FROM users WHERE email = '{user_email}'"
cursor.execute(query)

# Safer example
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_email,))
```

:::

<p class="rule-nav">← [`backend-sensitive-log`](#backend-sensitive-log) | [↑ Back to Security Rules](#security-rules) | [`hardcoded-credential`](#hardcoded-credential) →</p>

---

#### `hardcoded-credential` {#hardcoded-credential}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `secret_leak`

**Description**: Hardcoded credential or API secret key found

**Recommendation**: Extract credentials to environment variables or secret management services

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
apiKey := "synthetic_secret_api_key_12345"

// Safer example
apiKey := os.Getenv("STRIPE_API_KEY")
```

```ts [TypeScript / JavaScript]
// Unsafe example
const apiKey = "synthetic_secret_api_key_12345";

// Safer example
const apiKey = process.env.STRIPE_API_KEY;
```

```python [Python]
# Unsafe example
api_key = "synthetic_secret_api_key_12345"

# Safer example
api_key = os.environ.get("STRIPE_API_KEY")
```

```java [Java]
// Unsafe example
String apiKey = "synthetic_secret_api_key_12345";

// Safer example
String apiKey = System.getenv("STRIPE_API_KEY");
```

:::

<p class="rule-nav">← [`sql-string-format`](#sql-string-format) | [↑ Back to Security Rules](#security-rules) | [`unsafe-inner-html`](#unsafe-inner-html) →</p>

---

#### `unsafe-inner-html` {#unsafe-inner-html}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `xss`

**Description**: dangerouslySetInnerHTML used — potential DOM XSS vulnerability

**Recommendation**: Sanitize raw HTML using DOMPurify before injecting into the DOM

##### Unsafe and Safer Example

```ts
// Unsafe example
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// Safer example
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(userInput) }} />
```

<p class="rule-nav">← [`hardcoded-credential`](#hardcoded-credential) | [↑ Back to Security Rules](#security-rules) | [`dynamic-order`](#dynamic-order) →</p>

---

#### `dynamic-order` {#dynamic-order}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `injection`

**Description**: Dynamic ORDER BY clause built via string formatting

**Recommendation**: Validate dynamic column names against an explicit allowlist before building queries

##### Unsafe and Safer Example

```go
// Unsafe example
db.Order(fmt.Sprintf("%s ASC", sortColumn))

// Safer example
allowedColumns := map[string]bool{"created_at": true, "name": true}
if allowedColumns[sortColumn] {
    db.Order(sortColumn + " ASC")
}
```

<p class="rule-nav">← [`unsafe-inner-html`](#unsafe-inner-html) | [↑ Back to Security Rules](#security-rules) | [`api-struct-response`](#api-struct-response) →</p>

---

#### `api-struct-response` {#api-struct-response}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Internal domain struct may be serialized directly into HTTP response

**Recommendation**: Map internal domain entities to explicit response DTOs to avoid leaking sensitive fields

##### Unsafe and Safer Example

```go
// Unsafe example
var user User // Contains HashedPassword, SecretToken
c.JSON(http.StatusOK, user)

// Safer example
response := UserResponse{ID: user.ID, Email: user.Email}
c.JSON(http.StatusOK, response)
```

<p class="rule-nav">← [`dynamic-order`](#dynamic-order) | [↑ Back to Security Rules](#security-rules) | [`sensitive-json-field`](#sensitive-json-field) →</p>

---

#### `sensitive-json-field` {#sensitive-json-field}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive struct field may be exposed in JSON serialization

**Recommendation**: Use json:"-" struct tag or custom serializer to exclude sensitive attributes

##### Unsafe and Safer Example

```go
// Unsafe example
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"password_hash"`
}

// Safer example
type Account struct {
    ID           string `json:"id"`
    PasswordHash string `json:"-"`
}
```

<p class="rule-nav">← [`api-struct-response`](#api-struct-response) | [↑ Back to Security Rules](#security-rules) | [`go-shell-command`](#go-shell-command) →</p>

---

#### `go-shell-command` {#go-shell-command}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `command_injection`

**Description**: Shell command interpreter executed via os/exec

**Recommendation**: Execute binary commands directly with argument arrays and sanitize untrusted input

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
cmd := exec.Command("sh", "-c", "ls " + userInput)

// Safer example
cmd := exec.Command("ls", "--", validatedPath)
```

```ts [TypeScript / Node.js]
// Unsafe example
child_process.exec("ls " + userInput);

// Safer example
child_process.execFile("ls", ["--", validatedPath]);
```

```python [Python]
# Unsafe example
subprocess.Popen("ls " + user_input, shell=True)

# Safer example
subprocess.Popen(["ls", "--", validated_path], shell=False)
```

:::

<p class="rule-nav">← [`sensitive-json-field`](#sensitive-json-field) | [↑ Back to Security Rules](#security-rules) | [`go-weak-cryptographic-hash`](#go-weak-cryptographic-hash) →</p>

---

#### `go-weak-cryptographic-hash` {#go-weak-cryptographic-hash}

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `weak_cryptography`

**Description**: Weak cryptographic hash algorithm (MD5/SHA1) detected

**Recommendation**: Use SHA-256 or stronger algorithms; use bcrypt/argon2 for password hashing

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
hasher := md5.New()
hasher.Write([]byte(password))

// Safer example
hasher := sha256.New()
hasher.Write([]byte(password))
```

```ts [TypeScript / Node.js]
// Unsafe example
const hash = crypto.createHash("md5").update(password).digest("hex");

// Safer example
const hash = crypto.createHash("sha256").update(password).digest("hex");
```

```python [Python]
# Unsafe example
hash_val = hashlib.md5(password.encode()).hexdigest()

# Safer example
hash_val = hashlib.sha256(password.encode()).hexdigest()
```

:::

<p class="rule-nav">← [`go-shell-command`](#go-shell-command) | [↑ Back to Security Rules](#security-rules) | [`go-tainted-file-path`](#go-tainted-file-path) →</p>

---

#### `go-tainted-file-path` {#go-tainted-file-path}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `path_traversal`

**Description**: Untrusted request parameter used directly in file system operation

**Recommendation**: Normalize paths, enforce base directory boundaries, and use allowlisted identifiers

##### Unsafe and Safer Example

```go
// Unsafe example
filePath := r.URL.Query().Get("file")
data, _ := os.ReadFile(filePath)

// Safer example
filename := filepath.Base(r.URL.Query().Get("file"))
safePath := filepath.Join("/var/app/storage", filename)
data, _ := os.ReadFile(safePath)
```

<p class="rule-nav">← [`go-weak-cryptographic-hash`](#go-weak-cryptographic-hash) | [↑ Back to Security Rules](#security-rules) | [`go-weak-random-secret`](#go-weak-random-secret) →</p>

---

#### `go-weak-random-secret` {#go-weak-random-secret}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `insecure_randomness`

**Description**: Security-sensitive value generated using pseudo-random math/rand package

**Recommendation**: Use crypto/rand for generating tokens, nonces, session identifiers, and secret keys

##### Unsafe and Safer Example

```go
// Unsafe example
sessionToken := fmt.Sprintf("%d", rand.Intn(1000000))

// Safer example
tokenBytes := make([]byte, 32)
cryptoRand.Read(tokenBytes)
sessionToken := hex.EncodeToString(tokenBytes)
```

<p class="rule-nav">← [`go-tainted-file-path`](#go-tainted-file-path) | [↑ Back to Security Rules](#security-rules) | [`javascript-dynamic-eval`](#javascript-dynamic-eval) →</p>

---

#### `javascript-dynamic-eval` {#javascript-dynamic-eval}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `unsafe_deserialization`

**Description**: Dynamic eval execution of untrusted input detected

**Recommendation**: Use structured data parsers (JSON.parse) and schema validators instead of code evaluation

##### Unsafe and Safer Example

```ts
// Unsafe example
const config = eval("(" + jsonString + ")");

// Safer example
const config = JSON.parse(jsonString);
```

<p class="rule-nav">← [`go-weak-random-secret`](#go-weak-random-secret) | [↑ Back to Security Rules](#security-rules) | [`node-prisma-raw-query`](#node-prisma-raw-query) →</p>

---

#### `node-prisma-raw-query` {#node-prisma-raw-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Prisma raw unsafe query executed with potentially untrusted dynamic string

**Recommendation**: Use prisma.$queryRaw with tagged template literals (parameterized) instead of unsafe variants

##### Unsafe and Safer Example

```ts
// Unsafe example
const users = await prisma.$queryRawUnsafe(`SELECT * FROM users WHERE id = '${id}'`);

// Safer example
const users = await prisma.$queryRaw`SELECT * FROM users WHERE id = ${id}`;
```

<p class="rule-nav">← [`javascript-dynamic-eval`](#javascript-dynamic-eval) | [↑ Back to Security Rules](#security-rules) | [`node-typeorm-raw-query`](#node-typeorm-raw-query) →</p>

---

#### `node-typeorm-raw-query` {#node-typeorm-raw-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: TypeORM raw query with dynamic string interpolation

**Recommendation**: Pass parameters as the second argument array to query() rather than template interpolation

##### Unsafe and Safer Example

```ts
// Unsafe example
await connection.query(`SELECT * FROM users WHERE email = '${email}'`);

// Safer example
await connection.query("SELECT * FROM users WHERE email = $1", [email]);
```

<p class="rule-nav">← [`node-prisma-raw-query`](#node-prisma-raw-query) | [↑ Back to Security Rules](#security-rules) | [`node-sequelize-raw-query`](#node-sequelize-raw-query) →</p>

---

#### `node-sequelize-raw-query` {#node-sequelize-raw-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Sequelize raw query executed with template string interpolation

**Recommendation**: Use replacements or bind options in sequelize.query for safe parameter binding

##### Unsafe and Safer Example

```ts
// Unsafe example
await sequelize.query(`SELECT * FROM users WHERE status = '${status}'`);

// Safer example
await sequelize.query("SELECT * FROM users WHERE status = :status", { replacements: { status } });
```

<p class="rule-nav">← [`node-typeorm-raw-query`](#node-typeorm-raw-query) | [↑ Back to Security Rules](#security-rules) | [`node-pg-dynamic-query`](#node-pg-dynamic-query) →</p>

---

#### `node-pg-dynamic-query` {#node-pg-dynamic-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: node-postgres query executed with template string interpolation

**Recommendation**: Use parameterized query format ($1, $2) and pass values in the values parameter array

##### Unsafe and Safer Example

```ts
// Unsafe example
await client.query(`SELECT * FROM accounts WHERE id = '${id}'`);

// Safer example
await client.query("SELECT * FROM accounts WHERE id = $1", [id]);
```

<p class="rule-nav">← [`node-sequelize-raw-query`](#node-sequelize-raw-query) | [↑ Back to Security Rules](#security-rules) | [`node-mysql-dynamic-query`](#node-mysql-dynamic-query) →</p>

---

#### `node-mysql-dynamic-query` {#node-mysql-dynamic-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: mysql2 query executed with dynamic template string interpolation

**Recommendation**: Use query placeholders (?) and pass arguments in the parameter array

##### Unsafe and Safer Example

```ts
// Unsafe example
await pool.query(`SELECT * FROM products WHERE category = '${category}'`);

// Safer example
await pool.query("SELECT * FROM products WHERE category = ?", [category]);
```

<p class="rule-nav">← [`node-pg-dynamic-query`](#node-pg-dynamic-query) | [↑ Back to Security Rules](#security-rules) | [`python-sqlalchemy-raw-sql`](#python-sqlalchemy-raw-sql) →</p>

---

#### `python-sqlalchemy-raw-sql` {#python-sqlalchemy-raw-sql}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: SQLAlchemy raw text expression formatted with dynamic Python f-string or format()

**Recommendation**: Use bound parameters (:param_name) with session.execute(text("..."), {"param_name": val})

##### Unsafe and Safer Example

```text
// Unsafe example
session.execute(text(f"SELECT * FROM users WHERE username = '{username}'"))

// Safer example
session.execute(text("SELECT * FROM users WHERE username = :u"), {"u": username})
```

<p class="rule-nav">← [`node-mysql-dynamic-query`](#node-mysql-dynamic-query) | [↑ Back to Security Rules](#security-rules) | [`python-django-raw-sql`](#python-django-raw-sql) →</p>

---

#### `python-django-raw-sql` {#python-django-raw-sql}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Django raw SQL query constructed with f-string or unsafe .extra() clause

**Recommendation**: Pass parameters as params list to Model.objects.raw(query, [params]) or use standard ORM filters

##### Unsafe and Safer Example

```text
// Unsafe example
User.objects.raw(f"SELECT * FROM auth_user WHERE username = '{username}'")

// Safer example
User.objects.raw("SELECT * FROM auth_user WHERE username = %s", [username])
```

<p class="rule-nav">← [`python-sqlalchemy-raw-sql`](#python-sqlalchemy-raw-sql) | [↑ Back to Security Rules](#security-rules) | [`python-psycopg-format-query`](#python-psycopg-format-query) →</p>

---

#### `python-psycopg-format-query` {#python-psycopg-format-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: psycopg database cursor executed with Python string formatting instead of query parameters

**Recommendation**: Pass query parameters as the second tuple argument to cursor.execute(query, (param,))

##### Unsafe and Safer Example

```text
// Unsafe example
cursor.execute(f"SELECT * FROM items WHERE owner_id = '{owner_id}'")

// Safer example
cursor.execute("SELECT * FROM items WHERE owner_id = %s", (owner_id,))
```

<p class="rule-nav">← [`python-django-raw-sql`](#python-django-raw-sql) | [↑ Back to Security Rules](#security-rules) | [`java-spring-jpa-native-query`](#java-spring-jpa-native-query) →</p>

---

#### `java-spring-jpa-native-query` {#java-spring-jpa-native-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Spring Data JPA native query built via string concatenation

**Recommendation**: Use named parameters (:param) or positional parameters (?1) in native @Query annotations

##### Unsafe and Safer Example

```text
// Unsafe example
@Query(value = "SELECT * FROM users WHERE role = '" + ROLE + "'", nativeQuery = true)

// Safer example
@Query(value = "SELECT * FROM users WHERE role = :role", nativeQuery = true)
```

<p class="rule-nav">← [`python-psycopg-format-query`](#python-psycopg-format-query) | [↑ Back to Security Rules](#security-rules) | [`java-hibernate-native-query`](#java-hibernate-native-query) →</p>

---

#### `java-hibernate-native-query` {#java-hibernate-native-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Hibernate createNativeQuery executed with dynamic string concatenation

**Recommendation**: Use parameterized placeholders and bind parameters via query.setParameter()

##### Unsafe and Safer Example

```text
// Unsafe example
session.createNativeQuery("SELECT * FROM orders WHERE status = '" + status + "'")

// Safer example
session.createNativeQuery("SELECT * FROM orders WHERE status = :status").setParameter("status", status)
```

<p class="rule-nav">← [`java-spring-jpa-native-query`](#java-spring-jpa-native-query) | [↑ Back to Security Rules](#security-rules) | [`java-jdbc-dynamic-query`](#java-jdbc-dynamic-query) →</p>

---

#### `java-jdbc-dynamic-query` {#java-jdbc-dynamic-query}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Spring JdbcTemplate executed with concatenated SQL string

**Recommendation**: Pass query parameters as separate Object[] or varargs to jdbcTemplate

##### Unsafe and Safer Example

```text
// Unsafe example
jdbcTemplate.query("SELECT * FROM users WHERE id = " + id, rowMapper)

// Safer example
jdbcTemplate.query("SELECT * FROM users WHERE id = ?", rowMapper, id)
```

<p class="rule-nav">← [`java-hibernate-native-query`](#java-hibernate-native-query) | [↑ Back to Security Rules](#security-rules) | [`DBSEC-002`](#dbsec-002) →</p>

---

#### `DBSEC-002` {#dbsec-002}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `data_leak`

**Description**: Sensitive credentials or PII fields logged to application tracing stream

**Recommendation**: Redact credentials, tokens, and payment card details before writing to log sinks

##### Unsafe and Safer Example

```go
// Unsafe example
logger.info("Processing payment for card:", cardToken, secretKey);

// Safer example
logger.info("Processing payment for transaction ID:", transactionId);
```

<p class="rule-nav">← [`java-jdbc-dynamic-query`](#java-jdbc-dynamic-query) | [↑ Back to Security Rules](#security-rules) | [`DBSEC-003`](#dbsec-003) →</p>

---

#### `DBSEC-003` {#dbsec-003}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `information_exposure`

**Description**: Internal database driver error exposed directly in HTTP client response

**Recommendation**: Log the internal database error securely on the server and return a sanitized, generic error message to the client

##### Unsafe and Safer Example

```go
// Unsafe example
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

// Safer example
c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
```

<p class="rule-nav">← [`DBSEC-002`](#dbsec-002) | [↑ Back to Security Rules](#security-rules) | [`SQLI-001`](#sqli-001) →</p>

---

#### `SQLI-001` {#sqli-001}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Untrusted value concatenated or formatted into executable SQL at database driver sink

**Recommendation**: Use parameterized queries ($1, ?, :name) instead of string concatenation or fmt.Sprintf

##### Unsafe and Safer Examples

::: code-group

```go [Go (database/sql)]
// Unsafe example
query := "SELECT * FROM users WHERE id = " + id
row := db.QueryRow(query)

// Safer example
query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRow(query, id)
```

:::

<p class="rule-nav">← [`DBSEC-003`](#dbsec-003) | [↑ Back to Security Rules](#security-rules) | [`SQLI-002`](#sqli-002) →</p>

---

#### `SQLI-002` {#sqli-002}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `sql-injection`

**Description**: Untrusted table, column, or identifier dynamically interpolated into SQL

**Recommendation**: Validate SQL identifiers against an explicit allow-list of known safe column/table names before interpolation

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := fmt.Sprintf("SELECT * FROM %s WHERE active = 1", tableName)
rows, err := db.Query(query)

// Safer example
allowed := map[string]string{"users": "users", "admins": "admins"}
table, ok := allowed[tableName]
if !ok { return nil, errors.New("invalid table") }
query := fmt.Sprintf("SELECT * FROM %s WHERE active = 1", table)
rows, err := db.Query(query)
```

:::

<p class="rule-nav">← [`SQLI-001`](#sqli-001) | [↑ Back to Security Rules](#security-rules) | [`SQLI-004`](#sqli-004) →</p>

---

#### `SQLI-004` {#sqli-004}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `orm-escape-hatch`

**Description**: Unsafe raw ORM escape hatch called with dynamic or concatenated string

**Recommendation**: Pass parameters as separate arguments to ORM clauses (e.g. db.Where("name = ?", val)) rather than dynamic string formatting

##### Unsafe and Safer Examples

::: code-group

```go [Go (GORM)]
// Unsafe example
db.Where(fmt.Sprintf("role = '%s'", role)).Find(&users)

// Safer example
db.Where("role = ?", role).Find(&users)
```

:::

<p class="rule-nav">← [`SQLI-002`](#sqli-002) | [↑ Back to Security Rules](#security-rules) | [`SQLI-008`](#sqli-008) →</p>

---

#### `SQLI-008` {#sqli-008}

- **Domain**: `security`
- **Severity**: `MEDIUM`
- **Category**: `bind-mismatch`

**Description**: SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed

**Recommendation**: Ensure the number of bind placeholders ($1, ?) matches the count of passed query arguments exactly

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id)

// Safer example
db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id, tenantID)
```

:::

<p class="rule-nav">← [`SQLI-004`](#sqli-004) | [↑ Back to Security Rules](#security-rules) | [`SQLI-011`](#sqli-011) →</p>

---

#### `SQLI-011` {#sqli-011}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `list-expansion`

**Description**: Unsafe list or IN clause expansion using strings.Join or manual string interpolation

**Recommendation**: Use sqlx.In or generate parameterized bind variable lists (?, ?, ...) for slice queries

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", strings.Join(ids, ","))
rows, err := db.Query(query)

// Safer example
query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
query = db.Rebind(query)
rows, err := db.Query(query, args...)
```

:::

<p class="rule-nav">← [`SQLI-008`](#sqli-008) | [↑ Back to Security Rules](#security-rules) | [`SQLI-012`](#sqli-012) →</p>

---

#### `SQLI-012` {#sqli-012}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `prepared-statement`

**Description**: Tainted SQL query template passed into statement preparation method db.Prepare()

**Recommendation**: Keep the SQL query string passed to db.Prepare strictly constant and bind dynamic values via stmt.Query / stmt.Exec

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
stmt, err := db.Prepare("SELECT * FROM users WHERE status = " + filter)

// Safer example
stmt, err := db.Prepare("SELECT * FROM users WHERE status = $1")
rows, err := stmt.Query(filter)
```

:::

<p class="rule-nav">← [`SQLI-011`](#sqli-011) | [↑ Back to Security Rules](#security-rules) | [`SQLAUTH-001`](#sqlauth-001) →</p>

---

#### `SQLAUTH-001` {#sqlauth-001}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `multi-tenant-isolation`

**Description**: Multi-tenant entity queried without tenant_id or organization_id scoping constraint

**Recommendation**: Enforce explicit tenant_id or organization_id filtering on all multi-tenant queries to prevent cross-tenant data access

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func getAccounts(db *sql.DB) (*sql.Rows, error) {
    return db.Query("SELECT * FROM accounts WHERE status = 'active'")
}

// Safer example
func getAccounts(db *sql.DB, tenantID string) (*sql.Rows, error) {
    return db.Query("SELECT * FROM accounts WHERE tenant_id = $1 AND status = 'active'", tenantID)
}
```

:::

<p class="rule-nav">← [`SQLI-012`](#sqli-012) | [↑ Back to Security Rules](#security-rules) | [`SQLAUTH-002`](#sqlauth-002) →</p>

---

#### `SQLAUTH-002` {#sqlauth-002}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `authorization-idor`

**Description**: Sensitive resource queried solely by object ID without user ownership scoping (IDOR risk)

**Recommendation**: Scope entity lookups by both the object ID and authenticated user/account ID

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func getOrder(db *sql.DB, orderID string) (*sql.Row, error) {
    return db.QueryRow("SELECT * FROM orders WHERE id = $1", orderID), nil
}

// Safer example
func getOrder(db *sql.DB, orderID, userID string) (*sql.Row, error) {
    return db.QueryRow("SELECT * FROM orders WHERE id = $1 AND user_id = $2", orderID, userID), nil
}
```

:::

<p class="rule-nav">← [`SQLAUTH-001`](#sqlauth-001) | [↑ Back to Security Rules](#security-rules) | [`SQLAUTH-003`](#sqlauth-003) →</p>

---

#### `SQLAUTH-003` {#sqlauth-003}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `raw-query-bypass`

**Description**: Raw query bypasses standard ORM authorization scopes and permission filters

**Recommendation**: Ensure raw queries replicate all security barriers, role restrictions, and tenant scopes provided by ORM repositories

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Raw("SELECT * FROM users")

// Safer example
db.Raw("SELECT * FROM users WHERE organization_id = ? AND role <= ?", orgID, maxRole)
```

:::

<p class="rule-nav">← [`SQLAUTH-002`](#sqlauth-002) | [↑ Back to Security Rules](#security-rules) | [`SQLAUTH-004`](#sqlauth-004) →</p>

---

#### `SQLAUTH-004` {#sqlauth-004}

- **Domain**: `security`
- **Severity**: `HIGH`
- **Category**: `rls-misconfiguration`

**Description**: Database query assumes Row-Level Security but explicitly switches to superuser or bypass role

**Recommendation**: Connect and execute application queries using least-privilege non-superuser roles to enforce database Row-Level Security

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Exec("SET ROLE postgres")
db.Query("SELECT * FROM sensitive_documents")

// Safer example
db.Exec("SET LOCAL app.current_tenant_id = $1", tenantID)
db.Query("SELECT * FROM sensitive_documents")
```

:::

<p class="rule-nav">← [`SQLAUTH-003`](#sqlauth-003) | [↑ Back to Security Rules](#security-rules)</p>

---

## 🛡️ Hardening Rules {#hardening-rules}

Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`hardcoded-api-url`](#hardcoded-api-url) | `MEDIUM` | `configuration_leak` | Hardcoded localhost API URL found — load dynamically from environment variable |
| [`tls-insecure-skip-verify`](#tls-insecure-skip-verify) | `HIGH` | `transport_security` | TLS certificate verification is explicitly disabled |
| [`wildcard-cors-origin`](#wildcard-cors-origin) | `HIGH` | `cors` | Wildcard CORS origin header found in configuration |
| [`go-permissive-file-mode`](#go-permissive-file-mode) | `MEDIUM` | `file_permission` | File or directory created with permissive world-writable file permissions (0777) |
| [`debug-mode-enabled`](#debug-mode-enabled) | `MEDIUM` | `debug_configuration` | Debug mode appears to be explicitly enabled in configuration |
| [`go-insecure-cookie-attribute`](#go-insecure-cookie-attribute) | `HIGH` | `cookie_security` | Cookie configured with explicitly insecure security attributes |

### Details and Guidance

#### `hardcoded-api-url` {#hardcoded-api-url}

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `configuration_leak`

**Description**: Hardcoded localhost API URL found — load dynamically from environment variable

**Recommendation**: Configure API endpoints dynamically via environment variables for different environments

##### Unsafe and Safer Example

```go
// Unsafe example
const API_URL = "http://localhost:8080/api/v1";

// Safer example
const API_URL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";
```

<p class="rule-nav">[↑ Back to Hardening Rules](#hardening-rules) | [`tls-insecure-skip-verify`](#tls-insecure-skip-verify) →</p>

---

#### `tls-insecure-skip-verify` {#tls-insecure-skip-verify}

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `transport_security`

**Description**: TLS certificate verification is explicitly disabled

**Recommendation**: Enable certificate verification and configure valid trust stores

##### Unsafe and Safer Example

```go
// Unsafe example
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// Safer example
tr := &http.Transport{
    TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
}
```

<p class="rule-nav">← [`hardcoded-api-url`](#hardcoded-api-url) | [↑ Back to Hardening Rules](#hardening-rules) | [`wildcard-cors-origin`](#wildcard-cors-origin) →</p>

---

#### `wildcard-cors-origin` {#wildcard-cors-origin}

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cors`

**Description**: Wildcard CORS origin header found in configuration

**Recommendation**: Use an explicit CORS origin allowlist tailored for each deployment environment

##### Unsafe and Safer Example

```go
// Unsafe example
c.Header("Access-Control-Allow-Origin", "*")

// Safer example
c.Header("Access-Control-Allow-Origin", "https://app.example.com")
```

<p class="rule-nav">← [`tls-insecure-skip-verify`](#tls-insecure-skip-verify) | [↑ Back to Hardening Rules](#hardening-rules) | [`go-permissive-file-mode`](#go-permissive-file-mode) →</p>

---

#### `go-permissive-file-mode` {#go-permissive-file-mode}

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `file_permission`

**Description**: File or directory created with permissive world-writable file permissions (0777)

**Recommendation**: Use minimum required file permissions such as 0600 for files or 0750 for directories

##### Unsafe and Safer Example

```go
// Unsafe example
os.WriteFile("config.json", data, 0777)

// Safer example
os.WriteFile("config.json", data, 0600)
```

<p class="rule-nav">← [`wildcard-cors-origin`](#wildcard-cors-origin) | [↑ Back to Hardening Rules](#hardening-rules) | [`debug-mode-enabled`](#debug-mode-enabled) →</p>

---

#### `debug-mode-enabled` {#debug-mode-enabled}

- **Domain**: `hardening`
- **Severity**: `MEDIUM`
- **Category**: `debug_configuration`

**Description**: Debug mode appears to be explicitly enabled in configuration

**Recommendation**: Disable debug mode in production deployment configurations to prevent information disclosure

##### Unsafe and Safer Example

```go
// Unsafe example
debug := true

// Safer example
debug := os.Getenv("APP_ENV") == "development"
```

<p class="rule-nav">← [`go-permissive-file-mode`](#go-permissive-file-mode) | [↑ Back to Hardening Rules](#hardening-rules) | [`go-insecure-cookie-attribute`](#go-insecure-cookie-attribute) →</p>

---

#### `go-insecure-cookie-attribute` {#go-insecure-cookie-attribute}

- **Domain**: `hardening`
- **Severity**: `HIGH`
- **Category**: `cookie_security`

**Description**: Cookie configured with explicitly insecure security attributes

**Recommendation**: Enable Secure and HttpOnly flags and set an appropriate SameSite policy for session cookies

##### Unsafe and Safer Example

```go
// Unsafe example
cookie := &http.Cookie{Name: "session", Value: token, Secure: false}

// Safer example
cookie := &http.Cookie{Name: "session", Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
```

<p class="rule-nav">← [`debug-mode-enabled`](#debug-mode-enabled) | [↑ Back to Hardening Rules](#hardening-rules)</p>

---

## ⚡ Reliability Rules {#reliability-rules}

Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`go-multipart-memory`](#go-multipart-memory) | `MEDIUM` | `resource_exhaustion` | Ensure multipart request processing configures explicit memory limits |
| [`go-http-default-server`](#go-http-default-server) | `MEDIUM` | `missing_timeout` | Default HTTP server does not configure defensive timeouts |
| [`go-unbounded-request-read`](#go-unbounded-request-read) | `MEDIUM` | `resource_exhaustion` | Request body may be read without explicit size limits |
| [`go-discarded-error`](#go-discarded-error) | `MEDIUM` | `error_handling` | Returned error value is explicitly ignored with blank identifier |
| [`go-process-termination`](#go-process-termination) | `MEDIUM` | `process_termination` | Application path may terminate entire process unexpectedly |
| [`go-http-client-without-timeout`](#go-http-client-without-timeout) | `MEDIUM` | `missing_timeout` | HTTP client struct literal does not set an overall request timeout |
| [`DBMIG-001`](#dbmig-001) | `HIGH` | `destructive-migration` | Destructive schema migration detected without guarded rollout or deprecation phase |
| [`DBMIG-002`](#dbmig-002) | `MEDIUM` | `migration-safety` | Database migration file lacks reversible rollback instructions |
| [`DBMIG-003`](#dbmig-003) | `MEDIUM` | `schema-integrity` | Security-sensitive key column defined in table definition |
| [`DBPERF-001`](#dbperf-001) | `MEDIUM` | `query-performance` | Public dataset queried without an explicit LIMIT or pagination boundary |
| [`DBPERF-002`](#dbperf-002) | `HIGH` | `n-plus-one` | Database query executed inside loop (N+1 query anti-pattern) |
| [`SQLSAFE-001`](#sqlsafe-001) | `HIGH` | `destructive-query` | Unbounded UPDATE or DELETE query without a WHERE clause |
| [`SQLSAFE-003`](#sqlsafe-003) | `HIGH` | `concurrency-hazard` | Non-atomic read-modify-write pattern detected on balance/inventory field without row locking |
| [`SQLSAFE-004`](#sqlsafe-004) | `HIGH` | `transaction-loss` | Database operation executes on global connection pool escaping active transaction boundary |
| [`SQLSAFE-005`](#sqlsafe-005) | `HIGH` | `logic-operator-precedence` | Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence |
| [`SQLSAFE-006`](#sqlsafe-006) | `MEDIUM` | `soft-delete-bypass` | Raw query omits deleted_at IS NULL condition on soft-deletable entity table |

### Details and Guidance

#### `go-multipart-memory` {#go-multipart-memory}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Ensure multipart request processing configures explicit memory limits

**Recommendation**: Set explicit memory limit with ParseMultipartForm or MaxBytesReader to prevent memory exhaustion

##### Unsafe and Safer Example

```go
// Unsafe example
c.Request.ParseMultipartForm(100 << 20) // Unbounded 100MB buffer

// Safer example
c.Request.ParseMultipartForm(10 << 20) // Controlled 10MB memory limit
```

<p class="rule-nav">[↑ Back to Reliability Rules](#reliability-rules) | [`go-http-default-server`](#go-http-default-server) →</p>

---

#### `go-http-default-server` {#go-http-default-server}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: Default HTTP server does not configure defensive timeouts

**Recommendation**: Use custom http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout, and IdleTimeout

##### Unsafe and Safer Example

```go
// Unsafe example
http.ListenAndServe(":8080", handler)

// Safer example
server := &http.Server{
    Addr: ":8080", Handler: handler,
    ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
}
server.ListenAndServe()
```

<p class="rule-nav">← [`go-multipart-memory`](#go-multipart-memory) | [↑ Back to Reliability Rules](#reliability-rules) | [`go-unbounded-request-read`](#go-unbounded-request-read) →</p>

---

#### `go-unbounded-request-read` {#go-unbounded-request-read}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `resource_exhaustion`

**Description**: Request body may be read without explicit size limits

**Recommendation**: Limit request body with http.MaxBytesReader or io.LimitReader before reading into memory

##### Unsafe and Safer Example

```go
// Unsafe example
body, err := io.ReadAll(r.Body)

// Safer example
body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB size limit
```

<p class="rule-nav">← [`go-http-default-server`](#go-http-default-server) | [↑ Back to Reliability Rules](#reliability-rules) | [`go-discarded-error`](#go-discarded-error) →</p>

---

#### `go-discarded-error` {#go-discarded-error}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `error_handling`

**Description**: Returned error value is explicitly ignored with blank identifier

**Recommendation**: Check and handle returned errors or document valid reason for ignoring

##### Unsafe and Safer Example

```go
// Unsafe example
_ = db.Close()

// Safer example
if err := db.Close(); err != nil {
    log.Printf("Failed to close DB connection: %v", err)
}
```

<p class="rule-nav">← [`go-unbounded-request-read`](#go-unbounded-request-read) | [↑ Back to Reliability Rules](#reliability-rules) | [`go-process-termination`](#go-process-termination) →</p>

---

#### `go-process-termination` {#go-process-termination}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `process_termination`

**Description**: Application path may terminate entire process unexpectedly

**Recommendation**: Propagate errors to request boundaries and perform controlled shutdown instead of calling panic/log.Fatal

##### Unsafe and Safer Example

```go
// Unsafe example
if err != nil {
    panic(err)
}

// Safer example
if err != nil {
    return fmt.Errorf("process request: %w", err)
}
```

<p class="rule-nav">← [`go-discarded-error`](#go-discarded-error) | [↑ Back to Reliability Rules](#reliability-rules) | [`go-http-client-without-timeout`](#go-http-client-without-timeout) →</p>

---

#### `go-http-client-without-timeout` {#go-http-client-without-timeout}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `missing_timeout`

**Description**: HTTP client struct literal does not set an overall request timeout

**Recommendation**: Configure explicit http.Client.Timeout and appropriate transport timeouts

##### Unsafe and Safer Example

```go
// Unsafe example
client := &http.Client{}

// Safer example
client := &http.Client{Timeout: 10 * time.Second}
```

<p class="rule-nav">← [`go-process-termination`](#go-process-termination) | [↑ Back to Reliability Rules](#reliability-rules) | [`DBMIG-001`](#dbmig-001) →</p>

---

#### `DBMIG-001` {#dbmig-001}

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `destructive-migration`

**Description**: Destructive schema migration detected without guarded rollout or deprecation phase

**Recommendation**: Follow the expand-contract migration pattern and avoid immediate column/table drops in live environments

##### Unsafe and Safer Example

```go
// Unsafe example
ALTER TABLE users DROP COLUMN phone_number;

// Safer example
-- Phase 1: Mark column deprecated in application code; Phase 2: Drop after code deployment
```

<p class="rule-nav">← [`go-http-client-without-timeout`](#go-http-client-without-timeout) | [↑ Back to Reliability Rules](#reliability-rules) | [`DBMIG-002`](#dbmig-002) →</p>

---

#### `DBMIG-002` {#dbmig-002}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `migration-safety`

**Description**: Database migration file lacks reversible rollback instructions

**Recommendation**: Always provide corresponding down migrations or automated rollback scripts for disaster recovery

##### Unsafe and Safer Example

```go
// Unsafe example
-- no-down: Irreversible migration

// Safer example
-- Provide matching down.sql migration with schema restore steps
```

<p class="rule-nav">← [`DBMIG-001`](#dbmig-001) | [↑ Back to Reliability Rules](#reliability-rules) | [`DBMIG-003`](#dbmig-003) →</p>

---

#### `DBMIG-003` {#dbmig-003}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `schema-integrity`

**Description**: Security-sensitive key column defined in table definition

**Recommendation**: Enforce explicit FOREIGN KEY, UNIQUE, or CHECK constraints on tenant and account scoping columns

##### Unsafe and Safer Example

```text
// Unsafe example
CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID);

// Safer example
CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE);
```

<p class="rule-nav">← [`DBMIG-002`](#dbmig-002) | [↑ Back to Reliability Rules](#reliability-rules) | [`DBPERF-001`](#dbperf-001) →</p>

---

#### `DBPERF-001` {#dbperf-001}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `query-performance`

**Description**: Public dataset queried without an explicit LIMIT or pagination boundary

**Recommendation**: Always enforce LIMIT and OFFSET / cursor pagination to prevent unbounded memory allocation and DB stalls

##### Unsafe and Safer Example

```go
// Unsafe example
db.Query("SELECT * FROM events WHERE created_at > $1", startTime)

// Safer example
db.Query("SELECT * FROM events WHERE created_at > $1 ORDER BY id ASC LIMIT 100", startTime)
```

<p class="rule-nav">← [`DBMIG-003`](#dbmig-003) | [↑ Back to Reliability Rules](#reliability-rules) | [`DBPERF-002`](#dbperf-002) →</p>

---

#### `DBPERF-002` {#dbperf-002}

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `n-plus-one`

**Description**: Database query executed inside loop (N+1 query anti-pattern)

**Recommendation**: Batch queries using WHERE id IN (...) or JOINs to fetch data in a single roundtrip

##### Unsafe and Safer Example

```go
// Unsafe example
for _, userID := range userIDs {
    db.QueryRow("SELECT * FROM profiles WHERE user_id = $1", userID)
}

// Safer example
db.Query("SELECT * FROM profiles WHERE user_id IN ($1, $2, ...)", userIDs)
```

<p class="rule-nav">← [`DBPERF-001`](#dbperf-001) | [↑ Back to Reliability Rules](#reliability-rules) | [`SQLSAFE-001`](#sqlsafe-001) →</p>

---

#### `SQLSAFE-001` {#sqlsafe-001}

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `destructive-query`

**Description**: Unbounded UPDATE or DELETE query without a WHERE clause

**Recommendation**: Always specify a WHERE clause or explicit target filter to prevent accidental table-wide mutation

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Exec("DELETE FROM users")

// Safer example
db.Exec("DELETE FROM users WHERE expires_at < $1", cutoffTime)
```

:::

<p class="rule-nav">← [`DBPERF-002`](#dbperf-002) | [↑ Back to Reliability Rules](#reliability-rules) | [`SQLSAFE-003`](#sqlsafe-003) →</p>

---

#### `SQLSAFE-003` {#sqlsafe-003}

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `concurrency-hazard`

**Description**: Non-atomic read-modify-write pattern detected on balance/inventory field without row locking

**Recommendation**: Use SELECT ... FOR UPDATE within a transaction or perform atomic SQL mutations

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
var balance int
db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&balance)
balance += 100
db.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", balance, id)

// Safer example
tx, _ := db.Begin()
var balance int
tx.QueryRow("SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", id).Scan(&balance)
balance += 100
tx.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", balance, id)
tx.Commit()
```

:::

<p class="rule-nav">← [`SQLSAFE-001`](#sqlsafe-001) | [↑ Back to Reliability Rules](#reliability-rules) | [`SQLSAFE-004`](#sqlsafe-004) →</p>

---

#### `SQLSAFE-004` {#sqlsafe-004}

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `transaction-loss`

**Description**: Database operation executes on global connection pool escaping active transaction boundary

**Recommendation**: Execute queries using the active transaction handle (tx) to guarantee atomic rollback on error

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
func Transfer(tx *sql.Tx, from, to string, amount int) error {
    db.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}

// Safer example
func Transfer(tx *sql.Tx, from, to string, amount int) error {
    tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}
```

:::

<p class="rule-nav">← [`SQLSAFE-003`](#sqlsafe-003) | [↑ Back to Reliability Rules](#reliability-rules) | [`SQLSAFE-005`](#sqlsafe-005) →</p>

---

#### `SQLSAFE-005` {#sqlsafe-005}

- **Domain**: `reliability`
- **Severity**: `HIGH`
- **Category**: `logic-operator-precedence`

**Description**: Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence

**Recommendation**: Explicitly group logical expressions with parentheses to avoid inadvertent filter bypass or tenant leakage

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
query := "SELECT * FROM orders WHERE tenant_id = $1 AND status = 'active' OR is_admin = true"

// Safer example
query := "SELECT * FROM orders WHERE tenant_id = $1 AND (status = 'active' OR is_admin = true)"
```

:::

<p class="rule-nav">← [`SQLSAFE-004`](#sqlsafe-004) | [↑ Back to Reliability Rules](#reliability-rules) | [`SQLSAFE-006`](#sqlsafe-006) →</p>

---

#### `SQLSAFE-006` {#sqlsafe-006}

- **Domain**: `reliability`
- **Severity**: `MEDIUM`
- **Category**: `soft-delete-bypass`

**Description**: Raw query omits deleted_at IS NULL condition on soft-deletable entity table

**Recommendation**: Include 'deleted_at IS NULL' in WHERE clauses when querying tables that use soft deletion

##### Unsafe and Safer Examples

::: code-group

```go [Go]
// Unsafe example
db.Query("SELECT * FROM users WHERE email = $1", email)

// Safer example
db.Query("SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", email)
```

:::

<p class="rule-nav">← [`SQLSAFE-005`](#sqlsafe-005) | [↑ Back to Reliability Rules](#reliability-rules)</p>

---

## 🧹 Quality Rules {#quality-rules}

Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`merge-conflict-marker`](#merge-conflict-marker) | `HIGH` | `repository_hygiene` | Unresolved merge-conflict marker found |
| [`javascript-debugger`](#javascript-debugger) | `MEDIUM` | `debug_code` | JavaScript debugger statement found |
| [`trailing-whitespace`](#trailing-whitespace) | `LOW` | `formatting` | Trailing whitespace found at end of line |
| [`mixed-indentation`](#mixed-indentation) | `LOW` | `formatting` | Mixed tabs and spaces used for indentation on the same line |
| [`javascript-console-debug`](#javascript-console-debug) | `LOW` | `debug_code` | Console debug statement left in code |

### Details and Guidance

#### `merge-conflict-marker` {#merge-conflict-marker}

- **Domain**: `quality`
- **Severity**: `HIGH`
- **Category**: `repository_hygiene`

**Description**: Unresolved merge-conflict marker found

**Recommendation**: Resolve merge conflict and remove all markers before committing

##### Unsafe and Safer Example

```go
// Unsafe example
<<<<<<< HEAD
const apiURL = "http://localhost:8080";
=======
const apiURL = "https://api.production.com";
>>>>>>> main

// Safer example
const apiURL = process.env.API_URL || "https://api.production.com";
```

<p class="rule-nav">[↑ Back to Quality Rules](#quality-rules) | [`javascript-debugger`](#javascript-debugger) →</p>

---

#### `javascript-debugger` {#javascript-debugger}

- **Domain**: `quality`
- **Severity**: `MEDIUM`
- **Category**: `debug_code`

**Description**: JavaScript debugger statement found

**Recommendation**: Remove debugger statement before committing

##### Unsafe and Safer Example

```ts
// Unsafe example
function calculateTotal(items: Item[]) {
  debugger; // Leftover debug statement
  return items.reduce((acc, item) => acc + item.price, 0);
}

// Safer example
function calculateTotal(items: Item[]) {
  return items.reduce((acc, item) => acc + item.price, 0);
}
```

<p class="rule-nav">← [`merge-conflict-marker`](#merge-conflict-marker) | [↑ Back to Quality Rules](#quality-rules) | [`trailing-whitespace`](#trailing-whitespace) →</p>

---

#### `trailing-whitespace` {#trailing-whitespace}

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Trailing whitespace found at end of line

**Recommendation**: Remove trailing whitespace at line end

##### Unsafe and Safer Example

```go
// Unsafe example
const username = "john_doe";

// Safer example
const username = "john_doe";
```

<p class="rule-nav">← [`javascript-debugger`](#javascript-debugger) | [↑ Back to Quality Rules](#quality-rules) | [`mixed-indentation`](#mixed-indentation) →</p>

---

#### `mixed-indentation` {#mixed-indentation}

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `formatting`

**Description**: Mixed tabs and spaces used for indentation on the same line

**Recommendation**: Use a consistent indentation style throughout the project

##### Unsafe and Safer Example

```go
// Unsafe example
func process() {
	  var x = 10 // Mixed tabs and spaces
}

// Safer example
func process() {
	var x = 10 // Consistent tab indentation
}
```

<p class="rule-nav">← [`trailing-whitespace`](#trailing-whitespace) | [↑ Back to Quality Rules](#quality-rules) | [`javascript-console-debug`](#javascript-console-debug) →</p>

---

#### `javascript-console-debug` {#javascript-console-debug}

- **Domain**: `quality`
- **Severity**: `LOW`
- **Category**: `debug_code`

**Description**: Console debug statement left in code

**Recommendation**: Remove debug statements or use an application logger with proper log level

##### Unsafe and Safer Example

```ts
// Unsafe example
function handleLogin(user: User) {
  console.log("User logged in:", user);
}

// Safer example
function handleLogin(user: User) {
  logger.info("User logged in", { userId: user.id });
}
```

<p class="rule-nav">← [`mixed-indentation`](#mixed-indentation) | [↑ Back to Quality Rules](#quality-rules)</p>

---

## 📜 Governance Rules {#governance-rules}

Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints.

| Rule ID | Severity | Category | Description |
| :--- | :--- | :--- | :--- |
| [`privacy-pii-log`](#privacy-pii-log) | `HIGH` | `privacy_log` | Logging statement may expose personally identifiable information |
| [`privacy-pii-url`](#privacy-pii-url) | `HIGH` | `privacy_url` | Personally identifiable information may be placed in a URL query string |
| [`privacy-pii-fixture`](#privacy-pii-fixture) | `MEDIUM` | `privacy_fixture` | Fixture may contain a literal personally identifiable value |
| [`privacy-sensitive-response`](#privacy-sensitive-response) | `HIGH` | `privacy_response` | Response construction may expose a sensitive personal field |

### Details and Guidance

#### `privacy-pii-log` {#privacy-pii-log}

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_log`

**Description**: Logging statement may expose personally identifiable information

**Recommendation**: Remove the PII field or log a non-reversible, access-controlled reference identifier

##### Unsafe and Safer Example

```go
// Unsafe example
log.Printf("User registered with email: %s, phone: %s", email, phone)

// Safer example
log.Printf("User registered with ID: %s", userID)
```

<p class="rule-nav">[↑ Back to Governance Rules](#governance-rules) | [`privacy-pii-url`](#privacy-pii-url) →</p>

---

#### `privacy-pii-url` {#privacy-pii-url}

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_url`

**Description**: Personally identifiable information may be placed in a URL query string

**Recommendation**: Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs

##### Unsafe and Safer Example

```go
// Unsafe example
urlParams.append("email", userEmail);

// Safer example
// Transmit sensitive parameters in authenticated POST request body
```

<p class="rule-nav">← [`privacy-pii-log`](#privacy-pii-log) | [↑ Back to Governance Rules](#governance-rules) | [`privacy-pii-fixture`](#privacy-pii-fixture) →</p>

---

#### `privacy-pii-fixture` {#privacy-pii-fixture}

- **Domain**: `governance`
- **Severity**: `MEDIUM`
- **Category**: `privacy_fixture`

**Description**: Fixture may contain a literal personally identifiable value

**Recommendation**: Use clearly synthetic, reserved test data and keep production-derived records out of the repository

##### Unsafe and Safer Example

```json
// Unsafe example
{"email": "real_person_1985@gmail.com", "ssn": "123-45-6789"}

// Safer example
{"email": "user@example.com", "ssn": "000-00-0000"}
```

<p class="rule-nav">← [`privacy-pii-url`](#privacy-pii-url) | [↑ Back to Governance Rules](#governance-rules) | [`privacy-sensitive-response`](#privacy-sensitive-response) →</p>

---

#### `privacy-sensitive-response` {#privacy-sensitive-response}

- **Domain**: `governance`
- **Severity**: `HIGH`
- **Category**: `privacy_response`

**Description**: Response construction may expose a sensitive personal field

**Recommendation**: Map the response through an explicit allowlisted DTO and omit sensitive fields

##### Unsafe and Safer Example

```go
// Unsafe example
res.json({ id: user.id, email: user.email, ssn: user.ssn });

// Safer example
res.json({ id: user.id, email: user.email });
```

<p class="rule-nav">← [`privacy-pii-fixture`](#privacy-pii-fixture) | [↑ Back to Governance Rules](#governance-rules)</p>

---

