package frontend

import (
	"context"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestTelemetryPrivacyDetection(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewTelemetryPrivacyChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "Console log containing user email",
			path:           "/project/src/log.js",
			source:         `console.log("user_email", userEmail);`,
			wantFinding:    true,
			expectedRuleID: "frontend/telemetry-privacy-leak",
		},
		{
			name:           "Analytics track event with password field",
			path:           "/project/src/analytics.js",
			source:         `analytics.track("login_attempt", { password: inputPassword });`,
			wantFinding:    true,
			expectedRuleID: "frontend/telemetry-privacy-leak",
		},
		{
			name:           "Safe console log of non-PII metadata",
			path:           "/project/src/safe_log.js",
			source:         `console.log("Component mounted", { id: 123 });`,
			wantFinding:    false,
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
				if findings[0].Domain != finding.Governance {
					t.Errorf("expected Governance domain, got %v", findings[0].Domain)
				}
				// Verify description contains field name metadata but not raw literal content
				if !strings.Contains(findings[0].Description, "Potential PII field") {
					t.Errorf("expected description to report safe metadata field, got %s", findings[0].Description)
				}
			}
		})
	}
}
