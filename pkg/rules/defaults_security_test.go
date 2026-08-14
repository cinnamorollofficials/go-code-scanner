package rules

import "testing"

func TestDefaultSecurityRuleExamples(t *testing.T) {
	testRuleExamples(t, DefaultSecurity(), map[string]ruleExample{
		"mock-token": {
			positive: `const token = "google-mock-jwt-token"`,
			negative: `const issuer = "accounts.example.com"`,
		},
		"browser-token-storage": {
			positive: `localStorage.setItem("access_token", token)`,
			negative: `sessionStorage.setItem("theme", theme)`,
		},
		"permission-bypass": {
			positive: `// bypass permission for this request`,
			negative: `return permissionService.Check(ctx, user)`,
		},
		"weak-secret": {
			positive: `JWT_SECRET=change-me-in-production`,
			negative: `JWT_SECRET=${JWT_SECRET}`,
		},
		"frontend-sensitive-log": {
			positive: `console.log("token", token)`,
			negative: `console.log("request completed")`,
		},
		"backend-sensitive-log": {
			positive: `fmt.Printf("token=%s", token)`,
			negative: `fmt.Printf("request=%s", requestID)`,
		},
		"sql-string-format": {
			positive: `query := fmt.Sprintf("SELECT * FROM users WHERE id=%s", id)`,
			negative: `query := "SELECT * FROM users WHERE id = ?"`,
		},
		"hardcoded-credential": {
			positive: `password = "supersecret"`,
			negative: `password = os.Getenv("PASSWORD")`,
		},
		"unsafe-inner-html": {
			positive: `return <div dangerouslySetInnerHTML={{__html: content}} />`,
			negative: `return <div>{content}</div>`,
		},
		"dynamic-order": {
			positive: `db.Order(fmt.Sprintf("%s %s", field, direction))`,
			negative: `db.Order("created_at DESC")`,
		},
		"api-struct-response": {
			positive: `c.JSON(http.StatusOK, user)`,
			negative: `c.JSON(http.StatusOK, response)`,
		},
		"sensitive-json-field": {
			positive: "Password string `json:\"password\"`",
			negative: "Password string `json:\"-\"`",
		},
		"go-shell-command": {
			positive: `exec.CommandContext(ctx, "sh", "-c", input)`,
			negative: `exec.CommandContext(ctx, "git", "status", "--short")`,
		},
		"go-weak-cryptographic-hash": {
			positive: `digest := sha1.Sum(content)`,
			negative: `digest := sha256.Sum256(content)`,
		},
		"go-tainted-file-path": {
			positive: `content, err := os.ReadFile(c.Param("path"))`,
			negative: `content, err := os.ReadFile(filepath.Join(base, validatedName))`,
		},
		"go-weak-random-secret": {
			positive: `sessionToken := rand.Intn(999999)`,
			negative: `simulationValue := rand.Intn(999999)`,
		},
		"javascript-dynamic-eval": {
			positive: `const result = eval(payload)`,
			negative: `const result = JSON.parse(payload)`,
		},
		"node-prisma-raw-query": {
			positive: `const users = await prisma.$queryRawUnsafe("SELECT * FROM users")`,
			negative: `const users = await prisma.$queryRaw` + "`SELECT * FROM users WHERE id = ${id}`",
		},
		"node-typeorm-raw-query": {
			positive: `await connection.query(` + "`SELECT * FROM users WHERE id = '${id}'`" + `)`,
			negative: `await connection.query("SELECT * FROM users WHERE id = $1", [id])`,
		},
		"node-sequelize-raw-query": {
			positive: `await sequelize.query(` + "`SELECT * FROM users WHERE id = '${id}'`" + `)`,
			negative: `await sequelize.query("SELECT * FROM users WHERE id = :id", { replacements: { id } })`,
		},
		"node-pg-dynamic-query": {
			positive: `await client.query(` + "`SELECT * FROM users WHERE id = '${id}'`" + `)`,
			negative: `await client.query("SELECT * FROM users WHERE id = $1", [id])`,
		},
		"node-mysql-dynamic-query": {
			positive: `await pool.query(` + "`SELECT * FROM users WHERE id = '${id}'`" + `)`,
			negative: `await pool.query("SELECT * FROM users WHERE id = ?", [id])`,
		},
		"python-sqlalchemy-raw-sql": {
			positive: `session.execute(text(f"SELECT * FROM users WHERE id = '{id}'"))`,
			negative: `session.execute(text("SELECT * FROM users WHERE id = :id"), {"id": id})`,
		},
		"python-django-raw-sql": {
			positive: `User.objects.raw(f"SELECT * FROM users WHERE id = '{id}'")`,
			negative: `User.objects.raw("SELECT * FROM users WHERE id = %s", [id])`,
		},
		"python-psycopg-format-query": {
			positive: `cursor.execute(f"SELECT * FROM users WHERE id = '{id}'")`,
			negative: `cursor.execute("SELECT * FROM users WHERE id = %s", (id,))`,
		},
		"java-spring-jpa-native-query": {
			positive: `@Query(value = "SELECT * FROM users WHERE id = '" + id + "'", nativeQuery = true)`,
			negative: `@Query(value = "SELECT * FROM users WHERE id = :id", nativeQuery = true)`,
		},
		"java-hibernate-native-query": {
			positive: `session.createNativeQuery("SELECT * FROM users WHERE id = '" + id + "'")`,
			negative: `session.createNativeQuery("SELECT * FROM users WHERE id = :id").setParameter("id", id)`,
		},
		"java-jdbc-dynamic-query": {
			positive: `jdbcTemplate.query("SELECT * FROM users WHERE id = " + id, mapper)`,
			negative: `jdbcTemplate.query("SELECT * FROM users WHERE id = ?", mapper, id)`,
		},
		"DBSEC-002": {
			positive: `logger.info("payment secret_key=", secret_key)`,
			negative: `logger.info("payment processed successfully")`,
		},
		"DBSEC-003": {
			positive: `c.JSON(500, gin.H{"error": err.Error()})`,
			negative: `c.JSON(500, gin.H{"error": "internal error"})`,
		},
	})
}
