package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

// DefaultQuality returns fast repository-hygiene and source-quality rules.
func DefaultQuality() []Rule {
	return []Rule{
		{
			ID: "merge-conflict-marker", Pattern: `^(<<<<<<< .+|=======|>>>>>>> .+)$`,
			Severity: finding.High, Domain: finding.Quality, Category: "repository_hygiene",
			Description:    "Unresolved merge-conflict marker ditemukan",
			Recommendation: "Selesaikan conflict dan hapus seluruh marker sebelum commit",
		},
		{
			ID: "javascript-debugger", Pattern: `^\s*debugger\s*;?\s*$`,
			Severity: finding.Medium, Domain: finding.Quality, Category: "debug_code",
			Description:    "JavaScript debugger statement ditemukan",
			Recommendation: "Hapus debugger statement sebelum commit",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
	}
}
