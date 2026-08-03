package frontend

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func mockSource(path, content string) scanner.Source {
	return scanner.Source{
		Path: path,
		Open: func(context.Context) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte(content))), nil
		},
	}
}

func mockErrorSource(path string) scanner.Source {
	return scanner.Source{
		Path: path,
		Open: func(context.Context) (io.ReadCloser, error) {
			return nil, errors.New("read error")
		},
	}
}

func TestClassifierVanillaAndFileConventions(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	classifier := NewClassifier(cfg)
	ctx := context.Background()

	tests := []struct {
		name     string
		source   scanner.Source
		expected Scope
	}{
		{
			name:     "HTML entry point",
			source:   mockSource("/project/index.html", "<html></html>"),
			expected: ScopeClient,
		},
		{
			name:     "Vanilla JS",
			source:   mockSource("/project/src/main.js", "console.log('hello');"),
			expected: ScopeClient,
		},
		{
			name:     "Explicit .client.ts suffix",
			source:   mockSource("/project/lib/api.client.ts", "export const x = 1;"),
			expected: ScopeClient,
		},
		{
			name:     "Explicit .server.ts suffix",
			source:   mockSource("/project/lib/db.server.ts", "export const db = {};"),
			expected: ScopeServer,
		},
		{
			name:     "Next.js API route",
			source:   mockSource("/project/pages/api/users.ts", "export default function handler() {}"),
			expected: ScopeServer,
		},
		{
			name:     "Next.js App router API route",
			source:   mockSource("/project/app/api/route.ts", "export async function GET() {}"),
			expected: ScopeServer,
		},
		{
			name:     "Nuxt server API route",
			source:   mockSource("/project/server/api/login.ts", "export default defineEventHandler(() => {})"),
			expected: ScopeServer,
		},
		{
			name:     "SvelteKit +page.svelte",
			source:   mockSource("/project/src/routes/+page.svelte", "<script></script>"),
			expected: ScopeClient,
		},
		{
			name:     "SvelteKit +page.server.ts loader",
			source:   mockSource("/project/src/routes/+page.server.ts", "export const load = () => {};"),
			expected: ScopeServer,
		},
		{
			name:     "SvelteKit +server.ts endpoint",
			source:   mockSource("/project/src/routes/api/+server.ts", "export const GET = () => {};"),
			expected: ScopeServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.Classify(ctx, tt.source)
			if got != tt.expected {
				t.Errorf("Classify(%s) = %s, want %s", tt.source.Path, got, tt.expected)
			}
		})
	}
}

func TestClassifierDirectives(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true

	classifier := NewClassifier(cfg)
	ctx := context.Background()

	clientSrc := mockSource("/project/components/Button.tsx", `"use client";\nimport React from 'react';`)
	if got := classifier.Classify(ctx, clientSrc); got != ScopeClient {
		t.Errorf("expected ScopeClient for use client directive, got %s", got)
	}

	serverSrc := mockSource("/project/actions/user.ts", `'use server';\nexport async function createUser() {}`)
	if got := classifier.Classify(ctx, serverSrc); got != ScopeServer {
		t.Errorf("expected ScopeServer for use server directive, got %s", got)
	}
}

func TestClassifierConfiguredRootsPrecedence(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true
	cfg.Frontend.ClientRoots = []string{"src/client"}
	cfg.Frontend.ServerRoots = []string{"src/server"}
	cfg.Frontend.SharedRoots = []string{"src/shared"}

	classifier := NewClassifier(cfg)
	ctx := context.Background()

	// File under client root even if named .server.ts (explicit root takes precedence)
	srcClient := mockSource("/project/src/client/helper.server.ts", "export const a = 1;")
	if got := classifier.Classify(ctx, srcClient); got != ScopeClient {
		t.Errorf("configured client_root precedence failed: got %s, want ScopeClient", got)
	}

	srcServer := mockSource("/project/src/server/db.ts", "export const db = 1;")
	if got := classifier.Classify(ctx, srcServer); got != ScopeServer {
		t.Errorf("configured server_root precedence failed: got %s, want ScopeServer", got)
	}

	srcShared := mockSource("/project/src/shared/types.ts", "export type User = {};")
	if got := classifier.Classify(ctx, srcShared); got != ScopeShared {
		t.Errorf("configured shared_root precedence failed: got %s, want ScopeShared", got)
	}
}

func TestClassifierMalformedAndCancelledInput(t *testing.T) {
	cfg := config.Default()
	cfg.Root = "/project"
	cfg.Frontend.Enabled = true
	classifier := NewClassifier(cfg)

	// Cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := classifier.Classify(ctx, mockSource("/project/app.js", "content")); got != ScopeUnknown {
		t.Errorf("expected ScopeUnknown for cancelled context, got %s", got)
	}

	// Read error source
	if got := classifier.Classify(context.Background(), mockErrorSource("/project/app.go")); got != ScopeUnknown {
		t.Errorf("expected ScopeUnknown for non-JS/TS error source, got %s", got)
	}
}
