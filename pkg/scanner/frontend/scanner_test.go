package frontend

import (
	"context"
	"io"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

func TestScannerDescriptorAndID(t *testing.T) {
	cfg := config.Default()
	s := New(cfg)

	if s.ID() != "frontend" {
		t.Fatalf("expected ID 'frontend', got %q", s.ID())
	}

	desc := s.Describe()
	if desc.Domain != finding.Security {
		t.Fatalf("expected Security domain, got %v", desc.Domain)
	}
	if len(desc.Capabilities) == 0 {
		t.Fatal("expected capabilities to be non-empty")
	}
}

func TestScannerLifecycleCleanAndEmpty(t *testing.T) {
	cfg := config.Default()
	cfg.Frontend.Enabled = true
	s := New(cfg)

	req := scanner.Request{
		Root: t.TempDir(),
		Mode: "full",
		Sources: []scanner.Source{
			mockSource("/project/src/App.tsx", "import React from 'react';"),
		},
	}

	res := s.Scan(context.Background(), req)
	if res.State != finding.ScannerClean {
		t.Fatalf("expected clean result state, got %v", res.State)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(res.Findings))
	}
}

func TestScannerLifecycleCancellation(t *testing.T) {
	cfg := config.Default()
	cfg.Frontend.Enabled = true
	s := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := scanner.Request{
		Root: t.TempDir(),
		Mode: "full",
		Sources: []scanner.Source{
			mockSource("/project/src/App.tsx", "import React from 'react';"),
		},
	}

	res := s.Scan(ctx, req)
	if res.State != finding.ScannerFailed || res.Failure != scanner.FailureCanceled {
		t.Fatalf("expected failure canceled, got state=%v failure=%v", res.State, res.Failure)
	}
}

func TestScannerLifecyclePanicRecovery(t *testing.T) {
	cfg := config.Default()
	cfg.Frontend.Enabled = true
	s := New(cfg)

	panicSource := scanner.Source{
		Path: "/project/src/Panic.tsx",
		Open: func(context.Context) (io.ReadCloser, error) {
			panic("simulated panic in worker")
		},
	}

	req := scanner.Request{
		Root:    t.TempDir(),
		Mode:    "full",
		Sources: []scanner.Source{panicSource},
	}

	res := s.Scan(context.Background(), req)
	if res.State != finding.ScannerPartial || res.Failure != scanner.FailurePanic {
		t.Fatalf("expected partial panic state, got state=%v failure=%v msg=%q", res.State, res.Failure, res.Message)
	}
}

func TestScannerLifecycleNilContext(t *testing.T) {
	cfg := config.Default()
	s := New(cfg)

	// Context nil test
	res := s.Scan(nil, scanner.Request{})
	if res.State != finding.ScannerFailed {
		t.Fatalf("expected failed state for nil context, got %v", res.State)
	}
}
