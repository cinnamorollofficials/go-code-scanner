package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

// DefaultGovernance returns built-in privacy and repository governance rules.
func DefaultGovernance() []Rule {
	return []Rule{
		{
			ID: "privacy-pii-log", Pattern: `(console\.(log|debug|info|warn|error)|log\.(Print|Printf|Info|Debug|Warn|Error)|fmt\.Print(f|ln)?).*\b(email|phone|ssn|national_?id|date_?of_?birth)\b`,
			Severity: finding.High, Domain: finding.Governance, Category: "privacy_log",
			Description:    "Logging statement may expose personally identifiable information",
			Recommendation: "Remove the PII field or log a non-reversible, access-controlled reference identifier",
			Tags:           []string{"pii", "privacy", "sensitive"},
		},
		{
			ID: "privacy-pii-url", Pattern: `(Query\(\)\.(Add|Set)|URLSearchParams.*\.(append|set))\s*\(\s*['"](email|phone|ssn|national_?id|date_?of_?birth)['"]`,
			Severity: finding.High, Domain: finding.Governance, Category: "privacy_url",
			Description:    "Personally identifiable information may be placed in a URL query string",
			Recommendation: "Transmit sensitive fields in an authenticated request body and avoid retaining them in URLs or access logs",
			Tags:           []string{"pii", "privacy", "sensitive"},
		},
		{
			ID: "privacy-pii-fixture", Pattern: `['"]?(email|phone|ssn|national_?id|date_?of_?birth)['"]?\s*:\s*['"][^'"$<{]{3,}['"]`,
			Severity: finding.Medium, Domain: finding.Governance, Category: "privacy_fixture",
			Description:    "Fixture may contain a literal personally identifiable value",
			Recommendation: "Use clearly synthetic, reserved test data and keep production-derived records out of the repository",
			Tags:           []string{"pii", "privacy", "sensitive"}, Extensions: []string{".json", ".yaml", ".yml", ".js", ".ts"},
		},
		{
			ID: "privacy-sensitive-response", Pattern: `(c\.JSON|json\.NewEncoder\([^)]*\)\.Encode|res\.json)\s*\([^\n]*(password|ssn|national_?id|date_?of_?birth)`,
			Severity: finding.High, Domain: finding.Governance, Category: "privacy_response",
			Description:    "Response construction may expose a sensitive personal field",
			Recommendation: "Map the response through an explicit allowlisted DTO and omit sensitive fields",
			Tags:           []string{"pii", "privacy", "sensitive"}, Extensions: []string{".go", ".js", ".ts", ".tsx", ".jsx"},
		},
	}
}
