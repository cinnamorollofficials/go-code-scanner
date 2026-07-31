package scanner

import (
	"context"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type Request struct {
	Root  string
	Mode  string
	Files []string
}

type Result struct {
	Findings []finding.Finding
	State    finding.ScannerState
	Message  string
	Version  string
	Duration time.Duration
}

type Scanner interface {
	ID() string
	Scan(context.Context, Request) Result
}
