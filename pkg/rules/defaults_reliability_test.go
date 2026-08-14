package rules

import "testing"

func TestDefaultReliabilityRuleExamples(t *testing.T) {
	testRuleExamples(t, DefaultReliability(), map[string]ruleExample{
		"go-multipart-memory": {
			positive: `file, err := c.FormFile("upload")`,
			negative: `file, err := form.File["upload"]`,
		},
		"go-http-default-server": {
			positive: `log.Fatal(http.ListenAndServe(":8080", handler))`,
			negative: `log.Fatal(server.ListenAndServe())`,
		},
		"go-unbounded-request-read": {
			positive: `body, err := io.ReadAll(r.Body)`,
			negative: `body, err := io.ReadAll(io.LimitReader(r.Body, maxBody))`,
		},
		"go-discarded-error": {
			positive: `_ = response.Body.Close()`,
			negative: `if err := response.Body.Close(); err != nil { return err }`,
		},
		"go-process-termination": {
			positive: `log.Fatalf("serve: %v", err)`,
			negative: `return fmt.Errorf("serve: %w", err)`,
		},
		"go-http-client-without-timeout": {
			positive: `client := &http.Client{}`,
			negative: `client := &http.Client{Timeout: 10 * time.Second}`,
		},
		"DBMIG-001": {
			positive: `ALTER TABLE users DROP COLUMN phone;`,
			negative: `ALTER TABLE users ADD COLUMN phone VARCHAR(20);`,
		},
		"DBMIG-002": {
			positive: `-- no-down: irreversible migration`,
			negative: `-- reversible migration down steps below`,
		},
		"DBMIG-003": {
			positive: `CREATE TABLE docs (id UUID, tenant_id UUID NOT NULL);`,
			negative: `CREATE TABLE docs (id UUID, name TEXT);`,
		},
		"DBPERF-001": {
			positive: `SELECT id, name FROM users WHERE active = 1;`,
			negative: `SELECT id, name FROM items ORDER BY id ASC;`,
		},
		"DBPERF-002": {
			positive: `for _, id := range ids { db.QueryRow("SELECT * FROM profiles WHERE id = $1", id) }`,
			negative: `db.Query("SELECT * FROM profiles WHERE id IN ($1, $2)", ids)`,
		},
	})
}
