package frontend

import (
	"context"
	"strings"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type CorpusFixture struct {
	Name           string
	Path           string
	Content        string
	ExpectedRuleID string // empty string if safe (should produce 0 findings)
	Sanitizers     []string
}

func TestCrossFrameworkFalsePositiveCorpus(t *testing.T) {
	fixtures := []CorpusFixture{
		// Vanilla JS/TS
		{
			Name:           "Vanilla safe textContent assignment",
			Path:           "src/vanilla/safe.ts",
			Content:        `const el = document.getElementById("app"); if (el) { el.textContent = userInput; }`,
			ExpectedRuleID: "",
		},
		{
			Name:           "Vanilla unsafe innerHTML assignment",
			Path:           "src/vanilla/unsafe.ts",
			Content:        `const el = document.getElementById("app"); if (el) { el.innerHTML = userInput; }`,
			ExpectedRuleID: "frontend/dom-injection",
		},

		// React / Next.js
		{
			Name:           "React safe text binding",
			Path:           "src/components/SafeReact.tsx",
			Content:        `import React from 'react'; export const SafeComp = ({ text }: { text: string }) => <div>{text}</div>;`,
			ExpectedRuleID: "",
		},
		{
			Name:           "React safe sanitized dangerouslySetInnerHTML",
			Path:           "src/components/SanitizedReact.tsx",
			Content:        `import React from 'react'; import DOMPurify from 'dompurify'; export const Sanitized = ({ html }: { html: string }) => <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(html) }} />;`,
			Sanitizers:     []string{"dompurify"},
			ExpectedRuleID: "",
		},
		{
			Name:           "React unsafe dangerouslySetInnerHTML",
			Path:           "src/components/UnsafeReact.tsx",
			Content:        `import React from 'react'; export const Unsafe = ({ html }: { html: string }) => <div dangerouslySetInnerHTML={{ __html: html }} />;`,
			ExpectedRuleID: "frontend/react-dangerously-set-inner-html",
		},
		{
			Name:           "Next.js server component boundary violation",
			Path:           "src/app/ServerComp.tsx",
			Content:        `import 'server-only'; import { useState } from 'react'; export default function ServerComp() { const [s] = useState(0); return <div>{s}</div>; }`,
			ExpectedRuleID: "frontend/next-server-module-in-client",
		},

		// Vue / Nuxt
		{
			Name:           "Vue safe interpolation",
			Path:           "src/components/SafeVue.vue",
			Content:        `<template><div>{{ message }}</div></template>`,
			ExpectedRuleID: "",
		},
		{
			Name:           "Vue unsafe v-html directive",
			Path:           "src/components/UnsafeVue.vue",
			Content:        `<template><div v-html="rawUserHTML"></div></template>`,
			ExpectedRuleID: "frontend/vue-v-html",
		},

		// Svelte / SvelteKit
		{
			Name:           "Svelte safe expression",
			Path:           "src/lib/SafeSvelte.svelte",
			Content:        `<h1>{title}</h1>`,
			ExpectedRuleID: "",
		},
		{
			Name:           "Svelte unsafe {@html} directive",
			Path:           "src/lib/UnsafeSvelte.svelte",
			Content:        `<div>{@html rawHTML}</div>`,
			ExpectedRuleID: "frontend/svelte-html",
		},

		// Hardened & Supply Chain
		{
			Name:           "Safe HTTPS script with SRI",
			Path:           "public/index.html",
			Content:        `<!DOCTYPE html><html><head><script src="https://cdn.example.com/lib@1.0.0/index.js" integrity="sha256-abc" crossorigin="anonymous"></script></head></html>`,
			ExpectedRuleID: "",
		},
		{
			Name:           "Unsafe HTTP script",
			Path:           "public/insecure.html",
			Content:        `<!DOCTYPE html><html><head><script src="http://cdn.example.com/lib.js"></script></head></html>`,
			ExpectedRuleID: "frontend/insecure-resource-url",
		},

		// Comments & Test / Mock files (noise budget check)
		{
			Name:           "Commented out unsafe code",
			Path:           "src/utils/legacy.ts",
			Content:        `// el.innerHTML = oldCode; /* dangerouslySetInnerHTML={{__html: x}} */`,
			ExpectedRuleID: "",
		},

		// Malformed syntax (should gracefully handle without panic/crash)
		{
			Name:           "Malformed TSX file without sinks",
			Path:           "src/malformed.tsx",
			Content:        `const x = <div class=>>> syntax error <<<`,
			ExpectedRuleID: "",
		},
	}

	for _, tc := range fixtures {
		t.Run(tc.Name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Frontend.Enabled = true
			cfg.Frontend.DetectClientServerBoundaries = true
			if len(tc.Sanitizers) > 0 {
				cfg.Frontend.RecognizeSanitizers = tc.Sanitizers
			}

			s := New(cfg)
			req := scanner.Request{
				Root: "/project",
				Mode: "full",
				Sources: []scanner.Source{
					mockSource("/project/"+tc.Path, tc.Content),
				},
			}

			res := s.Scan(context.Background(), req)

			if tc.ExpectedRuleID == "" {
				// Safe fixture: enforce noise budget of ZERO false positives
				if res.State == finding.ScannerFindings || len(res.Findings) > 0 {
					var ruleIDs []string
					for _, f := range res.Findings {
						ruleIDs = append(ruleIDs, f.RuleID)
					}
					t.Fatalf("False positive detected for safe fixture %q: findings=%v", tc.Name, strings.Join(ruleIDs, ", "))
				}
			} else {
				// Unsafe fixture: expect finding with matching RuleID
				if len(res.Findings) == 0 {
					t.Fatalf("Expected finding %q for unsafe fixture %q, got 0 findings", tc.ExpectedRuleID, tc.Name)
				}
				matched := false
				for _, f := range res.Findings {
					if f.RuleID == tc.ExpectedRuleID {
						matched = true
						break
					}
				}
				if !matched {
					t.Fatalf("Expected RuleID %q for fixture %q, got findings: %+v", tc.ExpectedRuleID, tc.Name, res.Findings)
				}
			}
		})
	}
}
