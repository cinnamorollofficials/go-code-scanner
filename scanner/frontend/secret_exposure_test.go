package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestSecretExposureDetection(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewSecretExposureChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "Unsafe secret in NEXT_PUBLIC_ env var",
			path:           "/project/src/config.ts",
			source:         `const key = process.env.NEXT_PUBLIC_STRIPE_SECRET_KEY;`,
			wantFinding:    true,
			expectedRuleID: "frontend/client-credential-exposure",
		},
		{
			name:           "Unsafe private key in VITE_ env var",
			path:           "/project/src/config.ts",
			source:         `const key = import.meta.env.VITE_AUTH_PRIVATE_KEY;`,
			wantFinding:    true,
			expectedRuleID: "frontend/client-credential-exposure",
		},
		{
			name:           "Safe public API URL in NEXT_PUBLIC_ env var",
			path:           "/project/src/config.ts",
			source:         `const url = process.env.NEXT_PUBLIC_API_URL;`,
			wantFinding:    false,
		},
		{
			name:           "Unsafe localStorage token storage",
			path:           "/project/src/auth.js",
			source:         `localStorage.setItem('user_auth_token', jwtToken);`,
			wantFinding:    true,
			expectedRuleID: "frontend/client-credential-exposure",
		},
		{
			name:           "Embedded private key block",
			path:           "/project/src/keys.js",
			source:         "const key = `-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC...\n-----END PRIVATE KEY-----`;",
			wantFinding:    true,
			expectedRuleID: "frontend/client-credential-exposure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := mockSource(tt.path, tt.source)
			findings, err := checker.Check(context.Background(), src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantFinding && len(findings) == 0 {
				t.Errorf("expected finding for %q, got 0", tt.source)
			}
			if !tt.wantFinding && len(findings) > 0 {
				t.Errorf("expected no finding for %q, got %d findings", tt.source, len(findings))
			}
			if tt.wantFinding && len(findings) > 0 {
				if findings[0].RuleID != tt.expectedRuleID {
					t.Errorf("expected ruleID %s, got %s", tt.expectedRuleID, findings[0].RuleID)
				}
				if findings[0].Domain != finding.Security {
					t.Errorf("expected Security domain, got %v", findings[0].Domain)
				}
			}
		})
	}
}
