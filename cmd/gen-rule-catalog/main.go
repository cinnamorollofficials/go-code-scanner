package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
)

type DomainInfo struct {
	Domain      finding.Domain
	Title       string
	Description string
}

func detectLanguage(exts []string) string {
	if len(exts) == 0 {
		return "go"
	}
	for _, ext := range exts {
		switch ext {
		case ".go":
			return "go"
		case ".ts", ".tsx", ".js", ".jsx":
			return "ts"
		case ".json":
			return "json"
		case ".yaml", ".yml":
			return "yaml"
		}
	}
	return "text"
}

func commentPrefix(lang string) string {
	switch strings.ToLower(lang) {
	case "python", "py", "yaml", "yml", "sh", "bash", "ps1", "powershell":
		return "#"
	default:
		return "//"
	}
}

func main() {
	allRules := rules.Default()

	// Append Native AST SQL Taint Rules
	allRules = append(allRules, []rules.Rule{
		{
			ID: "SQLI-001", Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Untrusted value concatenated or formatted into executable SQL at database driver sink",
			Recommendation: "Use parameterized queries ($1, ?, :name) instead of string concatenation or fmt.Sprintf",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go (database/sql)",
					Unsafe: `query := "SELECT * FROM users WHERE id = " + id
row := db.QueryRow(query)`,
					Safe: `query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRow(query, id)`,
				},
			},
		},
		{
			ID: "SQLI-002", Severity: finding.High, Domain: finding.Security, Category: "sql-injection",
			Description:    "Untrusted table, column, or identifier dynamically interpolated into SQL",
			Recommendation: "Validate SQL identifiers against an explicit allow-list of known safe column/table names before interpolation",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `query := fmt.Sprintf("SELECT * FROM %s WHERE active = 1", tableName)
rows, err := db.Query(query)`,
					Safe: `allowed := map[string]string{"users": "users", "admins": "admins"}
table, ok := allowed[tableName]
if !ok { return nil, errors.New("invalid table") }
query := fmt.Sprintf("SELECT * FROM %s WHERE active = 1", table)
rows, err := db.Query(query)`,
				},
			},
		},
		{
			ID: "SQLI-004", Severity: finding.High, Domain: finding.Security, Category: "orm-escape-hatch",
			Description:    "Unsafe raw ORM escape hatch called with dynamic or concatenated string",
			Recommendation: "Pass parameters as separate arguments to ORM clauses (e.g. db.Where(\"name = ?\", val)) rather than dynamic string formatting",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go (GORM)",
					Unsafe: `db.Where(fmt.Sprintf("role = '%s'", role)).Find(&users)`,
					Safe: `db.Where("role = ?", role).Find(&users)`,
				},
			},
		},
		{
			ID: "SQLI-008", Severity: finding.Medium, Domain: finding.Security, Category: "bind-mismatch",
			Description:    "SQL placeholder count mismatch: query specifies N placeholders but different number of parameters were passed",
			Recommendation: "Ensure the number of bind placeholders ($1, ?) matches the count of passed query arguments exactly",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id)`,
					Safe: `db.Query("SELECT * FROM users WHERE id = ? AND tenant_id = ?", id, tenantID)`,
				},
			},
		},
		{
			ID: "SQLI-011", Severity: finding.High, Domain: finding.Security, Category: "list-expansion",
			Description:    "Unsafe list or IN clause expansion using strings.Join or manual string interpolation",
			Recommendation: "Use sqlx.In or generate parameterized bind variable lists (?, ?, ...) for slice queries",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `query := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s)", strings.Join(ids, ","))
rows, err := db.Query(query)`,
					Safe: `query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
query = db.Rebind(query)
rows, err := db.Query(query, args...)`,
				},
			},
		},
		{
			ID: "SQLI-012", Severity: finding.High, Domain: finding.Security, Category: "prepared-statement",
			Description:    "Tainted SQL query template passed into statement preparation method db.Prepare()",
			Recommendation: "Keep the SQL query string passed to db.Prepare strictly constant and bind dynamic values via stmt.Query / stmt.Exec",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `stmt, err := db.Prepare("SELECT * FROM users WHERE status = " + filter)`,
					Safe: `stmt, err := db.Prepare("SELECT * FROM users WHERE status = $1")
rows, err := stmt.Query(filter)`,
				},
			},
		},
		{
			ID: "SQLSAFE-001", Severity: finding.High, Domain: finding.Reliability, Category: "destructive-query",
			Description:    "Unbounded UPDATE or DELETE query without a WHERE clause",
			Recommendation: "Always specify a WHERE clause or explicit target filter to prevent accidental table-wide mutation",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `db.Exec("DELETE FROM users")`,
					Safe:   `db.Exec("DELETE FROM users WHERE expires_at < $1", cutoffTime)`,
				},
			},
		},
		{
			ID: "SQLAUTH-001", Severity: finding.High, Domain: finding.Security, Category: "multi-tenant-isolation",
			Description:    "Multi-tenant entity queried without tenant_id or organization_id scoping constraint",
			Recommendation: "Enforce explicit tenant_id or organization_id filtering on all multi-tenant queries to prevent cross-tenant data access",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `func getAccounts(db *sql.DB) (*sql.Rows, error) {
    return db.Query("SELECT * FROM accounts WHERE status = 'active'")
}`,
					Safe: `func getAccounts(db *sql.DB, tenantID string) (*sql.Rows, error) {
    return db.Query("SELECT * FROM accounts WHERE tenant_id = $1 AND status = 'active'", tenantID)
}`,
				},
			},
		},
		{
			ID: "SQLAUTH-002", Severity: finding.High, Domain: finding.Security, Category: "authorization-idor",
			Description:    "Sensitive resource queried solely by object ID without user ownership scoping (IDOR risk)",
			Recommendation: "Scope entity lookups by both the object ID and authenticated user/account ID",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `func getOrder(db *sql.DB, orderID string) (*sql.Row, error) {
    return db.QueryRow("SELECT * FROM orders WHERE id = $1", orderID), nil
}`,
					Safe: `func getOrder(db *sql.DB, orderID, userID string) (*sql.Row, error) {
    return db.QueryRow("SELECT * FROM orders WHERE id = $1 AND user_id = $2", orderID, userID), nil
}`,
				},
			},
		},
		{
			ID: "SQLAUTH-003", Severity: finding.High, Domain: finding.Security, Category: "raw-query-bypass",
			Description:    "Raw query bypasses standard ORM authorization scopes and permission filters",
			Recommendation: "Ensure raw queries replicate all security barriers, role restrictions, and tenant scopes provided by ORM repositories",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `db.Raw("SELECT * FROM users")`,
					Safe:   `db.Raw("SELECT * FROM users WHERE organization_id = ? AND role <= ?", orgID, maxRole)`,
				},
			},
		},
		{
			ID: "SQLAUTH-004", Severity: finding.High, Domain: finding.Security, Category: "rls-misconfiguration",
			Description:    "Database query assumes Row-Level Security but explicitly switches to superuser or bypass role",
			Recommendation: "Connect and execute application queries using least-privilege non-superuser roles to enforce database Row-Level Security",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `db.Exec("SET ROLE postgres")
db.Query("SELECT * FROM sensitive_documents")`,
					Safe: `db.Exec("SET LOCAL app.current_tenant_id = $1", tenantID)
db.Query("SELECT * FROM sensitive_documents")`,
				},
			},
		},
		{
			ID: "SQLSAFE-003", Severity: finding.High, Domain: finding.Reliability, Category: "concurrency-hazard",
			Description:    "Non-atomic read-modify-write pattern detected on balance/inventory field without row locking",
			Recommendation: "Use SELECT ... FOR UPDATE within a transaction or perform atomic SQL mutations",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `var balance int
db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&balance)
balance += 100
db.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", balance, id)`,
					Safe: `tx, _ := db.Begin()
var balance int
tx.QueryRow("SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", id).Scan(&balance)
balance += 100
tx.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", balance, id)
tx.Commit()`,
				},
			},
		},
		{
			ID: "SQLSAFE-004", Severity: finding.High, Domain: finding.Reliability, Category: "transaction-loss",
			Description:    "Database operation executes on global connection pool escaping active transaction boundary",
			Recommendation: "Execute queries using the active transaction handle (tx) to guarantee atomic rollback on error",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `func Transfer(tx *sql.Tx, from, to string, amount int) error {
    db.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}`,
					Safe: `func Transfer(tx *sql.Tx, from, to string, amount int) error {
    tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
    tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
    return nil
}`,
				},
			},
		},
		{
			ID: "SQLSAFE-005", Severity: finding.High, Domain: finding.Reliability, Category: "logic-operator-precedence",
			Description:    "Query contains unparenthesized mixed AND / OR operators in WHERE clause, altering logical precedence",
			Recommendation: "Explicitly group logical expressions with parentheses to avoid inadvertent filter bypass or tenant leakage",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `query := "SELECT * FROM orders WHERE tenant_id = $1 AND status = 'active' OR is_admin = true"`,
					Safe:   `query := "SELECT * FROM orders WHERE tenant_id = $1 AND (status = 'active' OR is_admin = true)"`,
				},
			},
		},
		{
			ID: "SQLSAFE-006", Severity: finding.Medium, Domain: finding.Reliability, Category: "soft-delete-bypass",
			Description:    "Raw query omits deleted_at IS NULL condition on soft-deletable entity table",
			Recommendation: "Include 'deleted_at IS NULL' in WHERE clauses when querying tables that use soft deletion",
			Examples: []rules.CodeExample{
				{
					Language: "go", Label: "Go",
					Unsafe: `db.Query("SELECT * FROM users WHERE email = $1", email)`,
					Safe:   `db.Query("SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", email)`,
				},
			},
		},
	}...)

	domains := []DomainInfo{
		{
			Domain:      finding.Security,
			Title:       "🔒 Security Rules",
			Description: "Rules targeting vulnerability patterns, secret leaks, unsafe DOM injections, and authentication/authorization flaws.",
		},
		{
			Domain:      finding.Hardening,
			Title:       "🛡️ Hardening Rules",
			Description: "Rules enforcing defensive configurations, TLS verification, CORS allowlists, and secure environment settings.",
		},
		{
			Domain:      finding.Reliability,
			Title:       "⚡ Reliability Rules",
			Description: "Rules mitigating resource exhaustion, unhandled errors, missing HTTP timeouts, and unexpected process crashes.",
		},
		{
			Domain:      finding.Quality,
			Title:       "🧹 Quality Rules",
			Description: "Rules maintaining repository hygiene, formatting consistency, and flagging left-over debug statements.",
		},
		{
			Domain:      finding.SupplyChain,
			Title:       "📦 Supply Chain Rules",
			Description: "Rules auditing third-party dependencies, version pins, package vulnerabilities, and license restrictions.",
		},
		{
			Domain:      finding.Governance,
			Title:       "📜 Governance Rules",
			Description: "Rules enforcing data privacy, PII protection, fixture sanitization, and compliance policy constraints.",
		},
	}

	// Group rules by domain
	grouped := make(map[finding.Domain][]rules.Rule)
	for _, r := range allRules {
		grouped[r.Domain] = append(grouped[r.Domain], r)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("title: Rule Catalog\n")
	buf.WriteString("description: Complete catalog of default built-in security, secret, governance, and quality rules with Do's and Don'ts code examples.\n")
	buf.WriteString("---\n\n")

	buf.WriteString("# Built-In Rule Catalog\n\n")
	buf.WriteString("Below is the complete catalog of built-in detection rules provided by `security-review`. This catalog includes detailed guidance, recommendations, and **Do's and Don'ts** code examples for each rule.\n\n")

	// Domain Overview Table
	buf.WriteString("## Domain Overview\n\n")
	buf.WriteString("| Domain | Icon | Total Rules | Scope & Focus |\n")
	buf.WriteString("| :--- | :---: | :---: | :--- |\n")
	for _, info := range domains {
		count := len(grouped[info.Domain])
		icon := strings.Fields(info.Title)[0]
		titleWithoutIcon := strings.TrimSpace(strings.TrimPrefix(info.Title, icon))
		buf.WriteString(fmt.Sprintf("| **[%s](#%s)** | %s | %d | %s |\n", titleWithoutIcon, strings.ToLower(strings.ReplaceAll(titleWithoutIcon, " ", "-")), count, info.Description))
	}
	buf.WriteString("\n---\n\n")

	// Render each Domain Section
	for _, info := range domains {
		domainRules := grouped[info.Domain]
		if len(domainRules) == 0 {
			continue
		}

		icon := strings.Fields(info.Title)[0]
		titleWithoutIcon := strings.TrimSpace(strings.TrimPrefix(info.Title, icon))
		sectionID := strings.ToLower(strings.ReplaceAll(titleWithoutIcon, " ", "-"))

		buf.WriteString(fmt.Sprintf("## %s {#%s}\n\n", info.Title, sectionID))
		buf.WriteString(fmt.Sprintf("%s\n\n", info.Description))

		// Matrix table for this domain
		buf.WriteString("| Rule ID | Severity | Category | Description |\n")
		buf.WriteString("| :--- | :--- | :--- | :--- |\n")

		for _, r := range domainRules {
			desc := strings.ReplaceAll(r.Description, "\n", " ")
			buf.WriteString(fmt.Sprintf("| [`%s`](#%s) | `%s` | `%s` | %s |\n", r.ID, strings.ToLower(r.ID), r.Severity, r.Category, desc))
		}

		buf.WriteString("\n### Details & Guidance\n\n")
		for idx, r := range domainRules {
			anchorID := strings.ToLower(r.ID)
			buf.WriteString(fmt.Sprintf("#### `%s` {#%s}\n\n", r.ID, anchorID))
			buf.WriteString(fmt.Sprintf("- **Domain**: `%s`\n", r.Domain))
			buf.WriteString(fmt.Sprintf("- **Severity**: `%s`\n", r.Severity))
			buf.WriteString(fmt.Sprintf("- **Category**: `%s`\n\n", r.Category))
			buf.WriteString(fmt.Sprintf("**Description**: %s\n\n", r.Description))
			if r.Recommendation != "" {
				buf.WriteString(fmt.Sprintf("**Recommendation**: %s\n\n", r.Recommendation))
			}

			if len(r.Examples) > 0 {
				buf.WriteString("##### Code Examples (Don't vs Do)\n\n")
				buf.WriteString("::: code-group\n\n")
				for _, ex := range r.Examples {
					label := ex.Label
					if label == "" {
						label = strings.ToUpper(ex.Language)
					}
					comment := commentPrefix(ex.Language)
					buf.WriteString(fmt.Sprintf("```%s [%s]\n%s ❌ Don't (Unsafe)\n%s\n\n%s ✅ Do (Recommended)\n%s\n```\n\n",
						ex.Language,
						label,
						comment,
						strings.TrimSpace(ex.Unsafe),
						comment,
						strings.TrimSpace(ex.Safe),
					))
				}
				buf.WriteString(":::\n\n")
			} else if r.UnsafeExample != "" || r.SafeExample != "" {
				lang := detectLanguage(r.Extensions)
				comment := commentPrefix(lang)
				buf.WriteString("##### Code Example (Don't vs Do)\n\n")
				buf.WriteString(fmt.Sprintf("```%s\n%s ❌ Don't (Unsafe)\n%s\n\n%s ✅ Do (Recommended)\n%s\n```\n\n",
					lang,
					comment,
					strings.TrimSpace(r.UnsafeExample),
					comment,
					strings.TrimSpace(r.SafeExample),
				))
			}

			// Navigation helpers between rules
			var navLinks []string
			if idx > 0 {
				prevID := domainRules[idx-1].ID
				navLinks = append(navLinks, fmt.Sprintf("← [`%s`](#%s)", prevID, strings.ToLower(prevID)))
			}
			navLinks = append(navLinks, fmt.Sprintf("[↑ Back to %s](#%s)", titleWithoutIcon, sectionID))
			if idx < len(domainRules)-1 {
				nextID := domainRules[idx+1].ID
				navLinks = append(navLinks, fmt.Sprintf("[`%s`](#%s) →", nextID, strings.ToLower(nextID)))
			}
			buf.WriteString(fmt.Sprintf("<p class=\"rule-nav\">%s</p>\n\n", strings.Join(navLinks, " | ")))
			buf.WriteString("---\n\n")
		}
	}

	targetPath := filepath.Join("website", "docs", "reference", "rules.md")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing rule catalog: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated grouped rule catalog with code examples at %s (%d rules across %d domains)\n", targetPath, len(allRules), len(domains))
}
