package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

// DefaultHardening returns the built-in secure-configuration rules.
func DefaultHardening() []Rule {
	return []Rule{
		{ID: "hardcoded-api-url", Pattern: `API_URL\s*=\s*['\"]https?://localhost`, Severity: finding.Medium, Domain: finding.Hardening, Category: "configuration_leak", Description: "URL API hardcoded — gunakan environment variable"},
	}
}
