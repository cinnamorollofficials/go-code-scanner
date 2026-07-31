package scanner

import (
	"context"
	"io"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
)

type Source struct {
	Path string
	Open func(context.Context) (io.ReadCloser, error)
}

type Request struct {
	Root    string
	Mode    string
	Sources []Source
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
