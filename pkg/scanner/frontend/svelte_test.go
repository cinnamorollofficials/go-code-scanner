package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestSvelteAndSvelteKitScanning(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewSvelteChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "Svelte {@html} tag with dynamic expression",
			path:           "/project/src/routes/+page.svelte",
			source:         `<script> export int data; </script> <div> {@html data.body} </div>`,
			wantFinding:    true,
			expectedRuleID: "frontend/svelte-html",
		},
		{
			name:           "Client component importing SvelteKit private env module",
			path:           "/project/src/routes/profile.client.svelte",
			source:         `<script> import { SECRET_KEY } from '$env/static/private'; </script>`,
			wantFinding:    true,
			expectedRuleID: "frontend/sveltekit-private-env-in-client",
		},
		{
			name:           "Client component importing SvelteKit .server module",
			path:           "/project/src/routes/dashboard.client.svelte",
			source:         `<script> import { loadData } from './+page.server'; </script>`,
			wantFinding:    true,
			expectedRuleID: "frontend/sveltekit-server-module-in-client",
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
				if findings[0].Domain != finding.Security && findings[0].Domain != finding.Hardening {
					t.Errorf("unexpected domain: %v", findings[0].Domain)
				}
			}
		})
	}
}
