package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

func TestExecutionAndMessagingDetection(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewExecutionMessagingChecker(cfg)

	tests := []struct {
		name          string
		path          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:           "Unsafe eval variable",
			path:           "/project/src/eval.js",
			source:         `eval(dynamicCode);`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-execution",
		},
		{
			name:           "Unsafe Function constructor variable",
			path:           "/project/src/func.js",
			source:         `const fn = new Function(userCode);`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-execution",
		},
		{
			name:           "Unsafe string-based timer",
			path:           "/project/src/timer.js",
			source:         `setTimeout("doSomething()", 1000);`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-execution",
		},
		{
			name:           "Safe function timer closure",
			path:           "/project/src/timer_safe.js",
			source:         `setTimeout(() => { doSomething(); }, 1000);`,
			wantFinding:    false,
		},
		{
			name:           "Unsafe wildcard postMessage",
			path:           "/project/src/msg.js",
			source:         `window.postMessage(payload, "*");`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-messaging",
		},
		{
			name:           "Safe target origin postMessage",
			path:           "/project/src/msg_safe.js",
			source:         `window.postMessage(payload, "https://app.example.com");`,
			wantFinding:    false,
		},
		{
			name:           "Unsafe message listener without origin check",
			path:           "/project/src/listener.js",
			source:         `window.addEventListener('message', (e) => { handleData(e.data); });`,
			wantFinding:    true,
			expectedRuleID: "frontend/unsafe-messaging",
		},
		{
			name:           "Safe message listener with origin check",
			path:           "/project/src/listener_safe.js",
			source:         `window.addEventListener('message', (e) => { if (e.origin !== 'https://example.com') return; handleData(e.data); });`,
			wantFinding:    false,
		},
		{
			name:        "Ignored test file",
			path:        "/project/src/app.test.js",
			source:      `eval(dynamicCode);`,
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
				if findings[0].Domain != finding.Security {
					t.Errorf("expected Security domain, got %v", findings[0].Domain)
				}
			}
		})
	}
}
