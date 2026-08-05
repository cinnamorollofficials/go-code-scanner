package rules

import "testing"

func TestDefaultHardeningRuleExamples(t *testing.T) {
	testRuleExamples(t, DefaultHardening(), map[string]ruleExample{
		"hardcoded-api-url": {
			positive: `const API_URL = "http://localhost:8080"`,
			negative: `const API_URL = os.Getenv("API_URL")`,
		},
		"tls-insecure-skip-verify": {
			positive: `TLSClientConfig: &tls.Config{InsecureSkipVerify: true}`,
			negative: `TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS13}`,
		},
		"wildcard-cors-origin": {
			positive: `Access-Control-Allow-Origin: "*"`,
			negative: `Access-Control-Allow-Origin: "https://app.example.com"`,
		},
		"go-permissive-file-mode": {
			positive: `os.WriteFile(path, content, 0777)`,
			negative: `os.WriteFile(path, content, 0600)`,
		},
		"debug-mode-enabled": {
			positive: `debug_mode: true`,
			negative: `debug_mode: false`,
		},
		"go-insecure-cookie-attribute": {
			positive: `cookie := http.Cookie{Name: "session", HttpOnly: false}`,
			negative: `cookie := http.Cookie{Name: "session", Secure: true, HttpOnly: true}`,
		},
	})
}
