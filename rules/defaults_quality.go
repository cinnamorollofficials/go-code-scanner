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
		{
			ID: "trailing-whitespace", Pattern: `[ \t]+$`,
			Severity: finding.Low, Domain: finding.Quality, Category: "formatting",
			Description:    "Trailing whitespace ditemukan",
			Recommendation: "Hapus whitespace pada akhir baris",
			Tags:           []string{"formatting"}, Fixable: true,
		},
		{
			ID: "mixed-indentation", Pattern: `^( +\t|\t+ )`,
			Severity: finding.Low, Domain: finding.Quality, Category: "formatting",
			Description:    "Tab dan spasi tercampur pada indentation baris yang sama",
			Recommendation: "Gunakan satu gaya indentation yang konsisten",
			Tags:           []string{"formatting"},
		},
		{
			ID: "javascript-console-debug", Pattern: `^\s*console\.(log|debug)\s*\(`,
			Severity: finding.Low, Domain: finding.Quality, Category: "debug_code",
			Description:    "Console debug statement mungkin tertinggal",
			Recommendation: "Hapus statement debug atau gunakan logger aplikasi dengan level yang sesuai",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
		},
	}
}
