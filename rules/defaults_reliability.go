package rules

import "github.com/cinnamorollofficials/go-code-scanner/finding"

// DefaultReliability returns the built-in runtime-safety rules.
func DefaultReliability() []Rule {
	return []Rule{
		{ID: "go-multipart-memory", Pattern: `c\.(FormFile|MultipartForm)\(`, Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion", Description: "Pastikan request multipart memiliki batas ukuran", Extensions: []string{".go"}},
		{
			ID: "go-http-default-server", Pattern: `http\.(ListenAndServe|ListenAndServeTLS)\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "missing_timeout",
			Description:    "Default HTTP server tidak mengonfigurasi timeout defensif",
			Recommendation: "Gunakan http.Server dengan ReadHeaderTimeout, ReadTimeout, WriteTimeout, dan IdleTimeout",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-unbounded-request-read", Pattern: `io\.ReadAll\((r|req|request|c\.Request)\.Body\)`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion",
			Description:    "Request body mungkin dibaca tanpa batas ukuran",
			Recommendation: "Batasi body dengan http.MaxBytesReader atau io.LimitReader sebelum membacanya",
			Extensions:     []string{".go"},
		},
	}
}
