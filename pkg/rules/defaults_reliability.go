package rules

import "github.com/cinnamorollofficials/go-code-scanner/pkg/finding"

// DefaultReliability returns the built-in runtime-safety rules.
func DefaultReliability() []Rule {
	return []Rule{
		{
			ID: "go-multipart-memory", Pattern: `c\.(FormFile|MultipartForm)\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion",
			Description:    "Ensure multipart request processing configures explicit memory limits",
			Recommendation: "Set explicit memory limit with ParseMultipartForm or MaxBytesReader to prevent memory exhaustion",
			Extensions:     []string{".go"},
			UnsafeExample:  `c.Request.ParseMultipartForm(100 << 20) // Unbounded 100MB buffer`,
			SafeExample:    `c.Request.ParseMultipartForm(10 << 20) // Controlled 10MB memory limit`,
		},
		{
			ID: "go-http-default-server", Pattern: `http\.(ListenAndServe|ListenAndServeTLS)\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "missing_timeout",
			Description:    "Default HTTP server does not configure defensive timeouts",
			Recommendation: "Use custom http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout, and IdleTimeout",
			Extensions:     []string{".go"},
			UnsafeExample:  `http.ListenAndServe(":8080", handler)`,
			SafeExample: `server := &http.Server{
    Addr: ":8080", Handler: handler,
    ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
}
server.ListenAndServe()`,
		},
		{
			ID: "go-unbounded-request-read", Pattern: `io\.ReadAll\((r|req|request|c\.Request)\.Body\)`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion",
			Description:    "Request body may be read without explicit size limits",
			Recommendation: "Limit request body with http.MaxBytesReader or io.LimitReader before reading into memory",
			Extensions:     []string{".go"},
			UnsafeExample:  `body, err := io.ReadAll(r.Body)`,
			SafeExample:    `body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB size limit`,
		},
		{
			ID: "go-discarded-error", Pattern: `^\s*_\s*=\s*[A-Za-z_][A-Za-z0-9_.]*\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "error_handling",
			Description:    "Returned error value is explicitly ignored with blank identifier",
			Recommendation: "Check and handle returned errors or document valid reason for ignoring",
			Extensions:     []string{".go"},
			UnsafeExample:  `_ = db.Close()`,
			SafeExample: `if err := db.Close(); err != nil {
    log.Printf("Failed to close DB connection: %v", err)
}`,
		},
		{
			ID: "go-process-termination", Pattern: `\b(panic|log\.Fatal|log\.Fatalf|log\.Fatalln)\s*\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "process_termination",
			Description:    "Application path may terminate entire process unexpectedly",
			Recommendation: "Propagate errors to request boundaries and perform controlled shutdown instead of calling panic/log.Fatal",
			Extensions:     []string{".go"},
			UnsafeExample: `if err != nil {
    panic(err)
}`,
			SafeExample: `if err != nil {
    return fmt.Errorf("process request: %w", err)
}`,
		},
		{
			ID: "go-http-client-without-timeout", Pattern: `http\.Client\s*\{\s*\}`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "missing_timeout",
			Description:    "HTTP client struct literal does not set an overall request timeout",
			Recommendation: "Configure explicit http.Client.Timeout and appropriate transport timeouts",
			Extensions:     []string{".go"},
			UnsafeExample:  `client := &http.Client{}`,
			SafeExample:    `client := &http.Client{Timeout: 10 * time.Second}`,
		},
		{
			ID: "DBMIG-001", Pattern: `(?i)(DROP\s+TABLE|DROP\s+COLUMN|ALTER\s+TABLE.*DROP\s+COLUMN|DROP\s+DATABASE)\b`,
			Severity: finding.High, Domain: finding.Reliability, Category: "destructive-migration",
			Description:    "Destructive schema migration detected without guarded rollout or deprecation phase",
			Recommendation: "Follow the expand-contract migration pattern and avoid immediate column/table drops in live environments",
			Extensions:     []string{".sql", ".go", ".ts", ".js", ".py"},
			UnsafeExample:  `ALTER TABLE users DROP COLUMN phone_number;`,
			SafeExample:    `-- Phase 1: Mark column deprecated in application code; Phase 2: Drop after code deployment`,
		},
		{
			ID: "DBMIG-002", Pattern: `(?i)(--\s*no-down|--\s*irreversible|cannot\s+be\s+undone|irreversible\s+migration)`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "migration-safety",
			Description:    "Database migration file lacks reversible rollback instructions",
			Recommendation: "Always provide corresponding down migrations or automated rollback scripts for disaster recovery",
			Extensions:     []string{".sql", ".go", ".ts", ".py"},
			UnsafeExample:  `-- no-down: Irreversible migration`,
			SafeExample:    `-- Provide matching down.sql migration with schema restore steps`,
		},
		{
			ID: "DBMIG-003", Pattern: `(?i)CREATE\s+TABLE[^(]+\([^)]*\b(tenant_id|org_id|account_id)\s+(VARCHAR|INT|BIGINT|UUID)\b`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "schema-integrity",
			Description:    "Security-sensitive key column defined in table definition",
			Recommendation: "Enforce explicit FOREIGN KEY, UNIQUE, or CHECK constraints on tenant and account scoping columns",
			Extensions:     []string{".sql"},
			UnsafeExample:  `CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID);`,
			SafeExample:    `CREATE TABLE documents (id UUID PRIMARY KEY, tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE);`,
		},
		{
			ID: "DBPERF-001", Pattern: `(?i)SELECT\s+[^;]+FROM\s+(users|orders|accounts|logs|events|transactions|messages)\s+WHERE\b`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "query-performance",
			Description:    "Public dataset queried without an explicit LIMIT or pagination boundary",
			Recommendation: "Always enforce LIMIT and OFFSET / cursor pagination to prevent unbounded memory allocation and DB stalls",
			Extensions:     []string{".go", ".ts", ".js", ".py", ".java"},
			UnsafeExample:  `db.Query("SELECT * FROM events WHERE created_at > $1", startTime)`,
			SafeExample:    `db.Query("SELECT * FROM events WHERE created_at > $1 ORDER BY id ASC LIMIT 100", startTime)`,
		},
		{
			ID: "DBPERF-002", Pattern: `(?i)for\s+.*\{\s*.*\b(db\.(Query|QueryRow|Exec|Get|Select)|tx\.(Query|QueryRow|Exec))\b`,
			Severity: finding.High, Domain: finding.Reliability, Category: "n-plus-one",
			Description:    "Database query executed inside loop (N+1 query anti-pattern)",
			Recommendation: "Batch queries using WHERE id IN (...) or JOINs to fetch data in a single roundtrip",
			Extensions:     []string{".go"},
			UnsafeExample: `for _, userID := range userIDs {
    db.QueryRow("SELECT * FROM profiles WHERE user_id = $1", userID)
}`,
			SafeExample: `db.Query("SELECT * FROM profiles WHERE user_id IN ($1, $2, ...)", userIDs)`,
		},
	}
}
