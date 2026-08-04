package frontend

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cachepkg "github.com/cinnamorollofficials/go-code-scanner/cache"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/reporter"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func TestSecretAndPIIRedactionGates(t *testing.T) {
	canarySecret := "CANARY-SECRET-TOKEN-DO-NOT-LEAK-12345"

	// Raw scanner result with snippet
	rawResult := scanner.Result{
		State: finding.ScannerFindings,
		Findings: []finding.Finding{
			{
				ID: "F-RED-1", Fingerprint: "fp-red-1", RuleID: "frontend/secret-exposure",
				Tool: "frontend", Domain: finding.Security, Category: "secret", Severity: finding.Critical,
				Description: "API secret token hardcoded in client build",
				Recommendation: "Move secret to backend environment variable",
				Snippet: canarySecret, Location: finding.Location{File: "config.ts", Line: 5},
				Metadata: map[string]string{"type": "api_token"},
			},
		},
	}

	// 1. Verify cache.Sanitize redacts snippets from cache and report outputs
	sanitizedResult := cachepkg.Sanitize(rawResult)
	if sanitizedResult.Findings[0].Snippet != "" {
		t.Errorf("Sanitize failed to redact snippet from finding")
	}

	report := &finding.Report{
		Project: "redaction-test", SchemaVersion: "1.0", FingerprintVersion: "3",
		Findings: sanitizedResult.Findings,
		Scanners: []finding.ScannerStatus{
			{ID: "frontend", State: finding.ScannerFindings, Required: true},
		},
	}

	// 2. Terminal report redaction check
	var termBuf bytes.Buffer
	if err := reporter.WriteTerminal(&termBuf, report); err != nil {
		t.Fatalf("WriteTerminal failed: %v", err)
	}
	if strings.Contains(termBuf.String(), canarySecret) {
		t.Errorf("Terminal report leaked secret canary")
	}

	// 3. JSON, SARIF, JUnit artifact redaction checks
	formats := map[string]func(string, *finding.Report) error{
		"json":  reporter.WriteJSON,
		"sarif": reporter.WriteSARIF,
		"junit": reporter.WriteJUnit,
	}

	for name, write := range formats {
		path := filepath.Join(t.TempDir(), "report."+name)
		if err := write(path, report); err != nil {
			t.Fatalf("%s writer failed: %v", name, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), canarySecret) {
			t.Errorf("%s report format leaked secret canary", name)
		}
	}
}
