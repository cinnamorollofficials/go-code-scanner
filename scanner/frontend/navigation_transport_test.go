package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

func TestNavigationAndTransportDetection(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewNavigationTransportChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "Unsafe location.href redirect variable",
			path:           "/project/src/nav.js",
			source:         `window.location.href = redirectUrl;`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-navigation",
		},
		{
			name:           "Unsafe javascript: pseudo-protocol URL",
			path:           "/project/src/link.jsx",
			source:         `<a href="javascript:doBadThing()">Click</a>`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-navigation",
		},
		{
			name:           "Insecure HTTP remote endpoint",
			path:           "/project/src/api.js",
			source:         `const api = "http://api.production-server.com/v1/data";`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-transport",
		},
		{
			name:           "Safe localhost development HTTP endpoint",
			path:           "/project/src/api_dev.js",
			source:         `const api = "http://localhost:8080/v1/data";`,
			wantFinding:    false,
		},
		{
			name:           "Safe 127.0.0.1 development HTTP endpoint",
			path:           "/project/src/api_dev2.js",
			source:         `const api = "http://127.0.0.1:3000/v1/data";`,
			wantFinding:    false,
		},
		{
			name:           "Unsafe window.open target _blank without noopener",
			path:           "/project/src/win.js",
			source:         `window.open(url, "_blank");`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-navigation",
		},
		{
			name:           "Safe window.open target _blank with noopener",
			path:           "/project/src/win_safe.js",
			source:         `window.open(url, "_blank", "noopener,noreferrer");`,
			wantFinding:    false,
		},
		{
			name:           "Sensitive credential in URL query parameter string",
			path:           "/project/src/query.js",
			source:         `const fullUrl = "https://example.com/api?token=secret123";`,
			wantFinding:    true,
			expectedRuleID: "frontend/sensitive-query-param",
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
			}
		})
	}
}
