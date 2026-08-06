package rules

import "github.com/cinnamorollofficials/go-code-scanner/pkg/finding"

// DefaultReliability returns the built-in runtime-safety rules.
func DefaultReliability() []Rule {
	return []Rule{
		{
			ID: "go-multipart-memory", Pattern: `c\.(FormFile|MultipartForm)\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion",
			Description:    "Ensure multipart request processing configures explicit memory limits",
			Recommendation: "Set explicit memory limit with ParseMultipartForm or MaxBytesReader to prevent memory exhaustion",
			Extensions:     []string{".go"},
			UnsafeExample:  `c.Request.ParseMultipartForm(100 << 20) // Unbounded 100MB buffer`,
			SafeExample:    `c.Request.ParseMultipartForm(10 << 20) // Controlled 10MB memory limit`,
		},
		{
			ID: "go-http-default-server", Pattern: `http\.(ListenAndServe|ListenAndServeTLS)\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "missing_timeout",
			Description:    "Default HTTP server does not configure defensive timeouts",
			Recommendation: "Use custom http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout, and IdleTimeout",
			Extensions:     []string{".go"},
			UnsafeExample:  `http.ListenAndServe(":8080", handler)`,
			SafeExample: `server := &http.Server{
    Addr: ":8080", Handler: handler,
    ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
}
server.ListenAndServe()`,
		},
		{
			ID: "go-unbounded-request-read", Pattern: `io\.ReadAll\((r|req|request|c\.Request)\.Body\)`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "resource_exhaustion",
			Description:    "Request body may be read without explicit size limits",
			Recommendation: "Limit request body with http.MaxBytesReader or io.LimitReader before reading into memory",
			Extensions:     []string{".go"},
			UnsafeExample:  `body, err := io.ReadAll(r.Body)`,
			SafeExample:    `body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB size limit`,
		},
		{
			ID: "go-discarded-error", Pattern: `^\s*_\s*=\s*[A-Za-z_][A-Za-z0-9_.]*\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "error_handling",
			Description:    "Returned error value is explicitly ignored with blank identifier",
			Recommendation: "Check and handle returned errors or document valid reason for ignoring",
			Extensions:     []string{".go"},
			UnsafeExample:  `_ = db.Close()`,
			SafeExample: `if err := db.Close(); err != nil {
    log.Printf("Failed to close DB connection: %v", err)
}`,
		},
		{
			ID: "go-process-termination", Pattern: `\b(panic|log\.Fatal|log\.Fatalf|log\.Fatalln)\s*\(`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "process_termination",
			Description:    "Application path may terminate entire process unexpectedly",
			Recommendation: "Propagate errors to request boundaries and perform controlled shutdown instead of calling panic/log.Fatal",
			Extensions:     []string{".go"},
			UnsafeExample: `if err != nil {
    panic(err)
}`,
			SafeExample: `if err != nil {
    return fmt.Errorf("process request: %w", err)
}`,
		},
		{
			ID: "go-http-client-without-timeout", Pattern: `http\.Client\s*\{\s*\}`,
			Severity: finding.Medium, Domain: finding.Reliability, Category: "missing_timeout",
			Description:    "HTTP client struct literal does not set an overall request timeout",
			Recommendation: "Configure explicit http.Client.Timeout and appropriate transport timeouts",
			Extensions:     []string{".go"},
			UnsafeExample:  `client := &http.Client{}`,
			SafeExample:    `client := &http.Client{Timeout: 10 * time.Second}`,
		},
	}
}
