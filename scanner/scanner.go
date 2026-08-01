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
	Root            string
	Mode            string
	Sources         []Source
	Files           []Source
	RepositoryFiles []Source
}

type Result struct {
	Findings []finding.Finding
	State    finding.ScannerState
	Message  string
	Version  string
	Duration time.Duration
	Failure  FailureKind
}

type FailureKind string

const (
	FailureCanceled  FailureKind = "canceled"
	FailureExecution FailureKind = "execution"
	FailureMissing   FailureKind = "missing_dependency"
	FailurePanic     FailureKind = "panic"
	FailurePartial   FailureKind = "partial"
	FailureTimeout   FailureKind = "timeout"
)

func (k FailureKind) Valid() bool {
	return k == FailureCanceled || k == FailureExecution || k == FailureMissing || k == FailurePanic || k == FailurePartial || k == FailureTimeout
}

type Descriptor struct {
	Domain          finding.Domain
	Version         string
	Capabilities    []string
	SupportedModes  []string
	RequiresNetwork bool
}

type Described interface {
	Describe() Descriptor
}

type Scanner interface {
	ID() string
	Scan(context.Context, Request) Result
}
