package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
)

func TestRemoteResourceIntegrityChecks(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantRuleID    string
		wantFindings  int
	}{
		{
			name: "insecure HTTP script",
			html: `<html><head>
<script src="http://cdn.example.com/jquery.js"></script>
</head></html>`,
			wantRuleID:   "frontend/insecure-resource-url",
			wantFindings: 1,
		},
		{
			name: "cross-origin script missing SRI",
			html: `<html><head>
<script src="https://cdn.example.com/jquery@3.7.1/dist/jquery.min.js"></script>
</head></html>`,
			wantRuleID:   "frontend/missing-subresource-integrity",
			wantFindings: 1,
		},
		{
			name: "cross-origin script with SRI passes",
			html: `<html><head>
<script src="https://cdn.jsdelivr.net/npm/jquery@3.7.1/dist/jquery.min.js"
  integrity="sha256-abc123" crossorigin="anonymous"></script>
</head></html>`,
			wantRuleID:   "",
			wantFindings: 0,
		},
		{
			name: "unversioned remote resource",
			html: `<html><head><script src="https://cdn.example.com/react.js" integrity="sha256-abc" crossorigin="anonymous"></script></head></html>`,
			wantRuleID:   "frontend/unversioned-remote-resource",
			wantFindings: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := mockSource("/project/index.html", tc.html)
			cfg := config.Default()
			cfg.Root = "/project"

			checker := NewRemoteResourceChecker(cfg)
			findings, err := checker.Check(context.Background(), src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(findings) != tc.wantFindings {
				t.Fatalf("expected %d findings, got %d: %+v", tc.wantFindings, len(findings), findings)
			}
			if tc.wantFindings > 0 && findings[0].RuleID != tc.wantRuleID {
				t.Errorf("expected ruleID %s, got %s", tc.wantRuleID, findings[0].RuleID)
			}
		})
	}
}
