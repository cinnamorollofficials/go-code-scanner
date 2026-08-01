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
		{
			ID: "go-discarded-error", Pattern: `^\s*_\s*=\s*[A-Za-z_][A-Za-z0-9_.]*\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "error_handling",
			Description:    "Return value error mungkin dibuang secara eksplisit",
			Recommendation: "Periksa dan tangani error, atau dokumentasikan alasan aman untuk mengabaikannya",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-process-termination", Pattern: `\b(panic|log\.Fatal|log\.Fatalf|log\.Fatalln)\s*\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "process_termination",
			Description:    "Application path mungkin menghentikan seluruh process",
			Recommendation: "Propagasikan error ke boundary dan lakukan shutdown terkontrol",
			Extensions:     []string{".go"},
		},
		{
			ID: "go-http-client-without-timeout", Pattern: `http\.Client\s*\{\s*\}`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "missing_timeout",
			Description:    "HTTP client literal tidak menetapkan timeout keseluruhan",
			Recommendation: "Tetapkan http.Client.Timeout dan timeout transport yang sesuai",
			Extensions:     []string{".go"},
		},
	}
}
