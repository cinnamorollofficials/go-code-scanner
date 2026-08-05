package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestVueAndNuxtScanning(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewVueNuxtChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "Vue v-html directive with dynamic expression",
			path:           "/project/src/components/Notice.vue",
			source:         `<template><div v-html="rawNoticeHtml"></div></template>`,
			wantFinding:    true,
			expectedRuleID: "frontend/vue-v-html",
		},
		{
			name:           "Client code accessing private Nuxt runtimeConfig",
			path:           "/project/src/components/Header.client.vue",
			source:         `<script setup> const config = useRuntimeConfig(); const secret = config.secretKey; </script>`,
			wantFinding:    true,
			expectedRuleID: "frontend/nuxt-private-runtime-config",
		},
		{
			name:           "Client code importing Nuxt server directory module",
			path:           "/project/src/components/Auth.client.vue",
			source:         `<script setup> import db from '~/server/utils/db'; </script>`,
			wantFinding:    true,
			expectedRuleID: "frontend/nuxt-server-import-in-client",
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
