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
	})
}
