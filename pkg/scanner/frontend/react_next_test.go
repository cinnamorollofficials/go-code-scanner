package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestReactAndNextScanning(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewReactNextChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "React dangerouslySetInnerHTML with dynamic expr",
			path:           "/project/src/components/Card.client.jsx",
			source:         `"use client"; export default function Card({ content }) { return <div dangerouslySetInnerHTML={{ __html: content }} />; }`,
			wantFinding:    true,
			expectedRuleID: "frontend/react-dangerously-set-inner-html",
		},
		{
			name:           "Client component importing fs node module",
			path:           "/project/src/components/FileLoader.client.js",
			source:         `"use client"; import fs from 'fs'; export default function Loader() {}`,
			wantFinding:    true,
			expectedRuleID: "frontend/next-server-module-in-client",
		},
		{
			name:           "Client component reading private server env var",
			path:           "/project/src/components/DB.client.js",
			source:         `"use client"; const db = process.env.DATABASE_URL;`,
			wantFinding:    true,
			expectedRuleID: "frontend/next-private-env-in-client",
		},
		{
			name:        "Server component reading private env var is safe",
			path:        "/project/src/app/api/route.ts",
			source:      `export async function GET() { const db = process.env.DATABASE_URL; return Response.json({ db }); }`,
			wantFinding: false,
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
