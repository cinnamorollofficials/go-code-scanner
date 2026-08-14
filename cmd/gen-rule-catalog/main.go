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
					Safe: `db.Exec("DELETE FROM users WHERE expires_at < $1", cutoffTime)`,
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
		buf.WriteString(fmt.Sprintf("| **%s** | %s | %d | %s |\n", titleWithoutIcon, icon, count, info.Description))
	}
	buf.WriteString("\n---\n\n")

	// Render each Domain Section
	for _, info := range domains {
		domainRules := grouped[info.Domain]
		if len(domainRules) == 0 {
			continue
		}

		buf.WriteString(fmt.Sprintf("## %s\n\n", info.Title))
		buf.WriteString(fmt.Sprintf("%s\n\n", info.Description))

		// Table for this domain
		buf.WriteString("| Rule ID | Severity | Category | Description |\n")
		buf.WriteString("| :--- | :--- | :--- | :--- |\n")

		for _, r := range domainRules {
			desc := strings.ReplaceAll(r.Description, "\n", " ")
			buf.WriteString(fmt.Sprintf("| [`%s`](#%s) | `%s` | `%s` | %s |\n", r.ID, r.ID, r.Severity, r.Category, desc))
		}

		buf.WriteString("\n### Details & Guidance\n\n")
		for _, r := range domainRules {
			buf.WriteString(fmt.Sprintf("#### `%s`\n\n", r.ID))
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
