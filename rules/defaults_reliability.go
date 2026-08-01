package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

// DefaultReliability returns the built-in runtime-safety rules.
func DefaultReliability() []Rule {
	return []Rule{
		{ID: "go-multipart-memory", Pattern: `c\.(FormFile|MultipartForm)\(`, Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion", Description: "Pastikan request multipart memiliki batas ukuran", Extensions: []string{".go"}},
	}
}
