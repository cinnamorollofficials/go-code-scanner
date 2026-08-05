package frontend

import (
	"context"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
)

func TestDOMInjectionSinkDetection(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	checker := NewDOMInjectionChecker(cfg)

	tests := []struct {
		name          string
		source        string
		wantFinding   bool
		expectedRuleID string
	}{
		{
			name:        "Unsafe innerHTML assignment",
			source:      `element.innerHTML = userInput;`,
			wantFinding: true,
		},
		{
			name:        "Unsafe outerHTML assignment",
			source:      `target.outerHTML = payload;`,
			wantFinding: true,
		},
		{
			name:        "Unsafe insertAdjacentHTML call",
			source:      `div.insertAdjacentHTML('beforeend', dynamicHTML);`,
			wantFinding: true,
		},
		{
			name:        "Unsafe document.write call",
			source:      `document.write(untrustedParam);`,
			wantFinding: true,
		},
		{
			name:        "Safe static innerHTML literal",
			source:      `element.innerHTML = "<div>Static Content</div>";`,
			wantFinding: false,
		},
		{
			name:        "Safe innerHTML sanitized with DOMPurify",
			source:      `element.innerHTML = DOMPurify.sanitize(userInput);`,
			wantFinding: false,
		},
		{
			name:        "Safe innerHTML with Trusted Types",
			source:      `element.innerHTML = trustedTypes.createPolicy('default').createHTML(input);`,
			wantFinding: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := mockSource("/project/src/component.js", tt.source)
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
				if findings[0].RuleID != "frontend/dom-injection" {
					t.Errorf("expected ruleID frontend/dom-injection, got %s", findings[0].RuleID)
				}
				if findings[0].Domain != finding.Security {
					t.Errorf("expected Security domain, got %v", findings[0].Domain)
				}
			}
		})
	}
}

func TestDOMInjectionRecognizesCustomConfiguredSanitizers(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true
	cfg.Frontend.RecognizeSanitizers = []string{"myCustomSanitizer"}

	checker := NewDOMInjectionChecker(cfg)
	src := mockSource("/project/src/custom.js", `element.innerHTML = myCustomSanitizer(rawInput);`)

	findings, err := checker.Check(context.Background(), src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected custom sanitizer to neutralize DOM injection sink, got %d findings", len(findings))
	}
}
