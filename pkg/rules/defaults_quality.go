package rules

import "github.com/cinnamorollofficials/go-code-scanner/pkg/finding"

// DefaultQuality returns fast repository-hygiene and source-quality rules.
func DefaultQuality() []Rule {
	return []Rule{
		{
			ID: "merge-conflict-marker", Pattern: `^(<<<<<<< .+|=======|>>>>>>> .+)$`,
			Severity: finding.High, Domain: finding.Quality, Category: "repository_hygiene",
			Description:    "Unresolved merge-conflict marker found",
			Recommendation: "Resolve merge conflict and remove all markers before committing",
			UnsafeExample: `<<<<<<< HEAD
const apiURL = "http://localhost:8080";
=======
const apiURL = "https://api.production.com";
>>>>>>> main`,
			SafeExample: `const apiURL = process.env.API_URL || "https://api.production.com";`,
		},
		{
			ID: "javascript-debugger", Pattern: `^\s*debugger\s*;?\s*$`,
			Severity: finding.Medium, Domain: finding.Quality, Category: "debug_code",
			Description:    "JavaScript debugger statement found",
			Recommendation: "Remove debugger statement before committing",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample: `function calculateTotal(items: Item[]) {
  debugger; // Leftover debug statement
  return items.reduce((acc, item) => acc + item.price, 0);
}`,
			SafeExample: `function calculateTotal(items: Item[]) {
  return items.reduce((acc, item) => acc + item.price, 0);
}`,
		},
		{
			ID: "trailing-whitespace", Pattern: `[ \t]+$`,
			Severity: finding.Low, Domain: finding.Quality, Category: "formatting",
			Description:    "Trailing whitespace found at end of line",
			Recommendation: "Remove trailing whitespace at line end",
			Tags:           []string{"formatting"}, Fixable: true,
			UnsafeExample:  `const username = "john_doe";   `,
			SafeExample:    `const username = "john_doe";`,
		},
		{
			ID: "mixed-indentation", Pattern: `^( +\t|\t+ )`,
			Severity: finding.Low, Domain: finding.Quality, Category: "formatting",
			Description:    "Mixed tabs and spaces used for indentation on the same line",
			Recommendation: "Use a consistent indentation style throughout the project",
			Tags:           []string{"formatting"},
			UnsafeExample: `func process() {
	  var x = 10 // Mixed tabs and spaces
}`,
			SafeExample: `func process() {
	var x = 10 // Consistent tab indentation
}`,
		},
		{
			ID: "javascript-console-debug", Pattern: `^\s*console\.(log|debug)\s*\(`,
			Severity: finding.Low, Domain: finding.Quality, Category: "debug_code",
			Description:    "Console debug statement left in code",
			Recommendation: "Remove debug statements or use an application logger with proper log level",
			Extensions:     []string{".ts", ".tsx", ".js", ".jsx"},
			UnsafeExample: `function handleLogin(user: User) {
  console.log("User logged in:", user);
}`,
			SafeExample: `function handleLogin(user: User) {
  logger.info("User logged in", { userId: user.id });
}`,
		},
	}
}
