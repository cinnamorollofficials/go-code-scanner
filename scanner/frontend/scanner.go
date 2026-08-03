package frontend

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

type Scanner struct {
	cfg        config.Config
	classifier *Classifier
	workers    int
}

func New(cfg config.Config) *Scanner {
	workers := cfg.Workers
	if workers < 1 {
		workers = runtime.GOMAXPROCS(0)
	}
	return &Scanner{
		cfg:        cfg,
		classifier: NewClassifier(cfg),
		workers:    workers,
	}
}

func (s *Scanner) ID() string {
	return "frontend"
}

func (s *Scanner) Describe() scanner.Descriptor {
	return scanner.Descriptor{
		Domain:         finding.Security,
		Version:        "1.0",
		Capabilities:   []string{"built-in-rules", "classifier", "lexical-analysis"},
		SupportedModes: []string{"full", "changed", "staged"},
	}
}

func (s *Scanner) Scan(ctx context.Context, req scanner.Request) scanner.Result {
	started := time.Now()
	res := scanner.Result{
		State:    finding.ScannerClean,
		Version:  "1.0",
		Findings: []finding.Finding{},
	}

	if ctx == nil {
		res.State = finding.ScannerFailed
		res.Failure = scanner.FailureCanceled
		res.Message = "nil context"
		res.Duration = time.Since(started)
		return res
	}

	if err := ctx.Err(); err != nil {
		res.State = finding.ScannerFailed
		res.Failure = scanner.FailureCanceled
		res.Message = err.Error()
		res.Duration = time.Since(started)
		return res
	}

	sources := req.Sources
	if len(sources) == 0 {
		res.Duration = time.Since(started)
		return res
	}

	work := make(chan scanner.Source, len(sources))
	for _, src := range sources {
		work <- src
	}
	close(work)

	type workerResult struct {
		findings []finding.Finding
		err      error
		partial  bool
		panicErr error
	}

	results := make(chan workerResult, len(sources))
	var wg sync.WaitGroup

	numWorkers := s.workers
	if numWorkers > len(sources) {
		numWorkers = len(sources)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for src := range work {
				func() {
					defer func() {
						if r := recover(); r != nil {
							results <- workerResult{
								partial:  true,
								panicErr: fmt.Errorf("panic processing %s: %v", src.Path, r),
							}
						}
					}()

					if ctx.Err() != nil {
						return
					}

					scope := s.classifier.Classify(ctx, src)
					if scope != ScopeClient && scope != ScopeShared {
						return
					}
				}()
			}
		}()
	}

	wg.Wait()
	close(results)

	var allFindings []finding.Finding
	var partialErrors []string
	var panics []string

	for r := range results {
		if r.panicErr != nil {
			panics = append(panics, r.panicErr.Error())
		}
		if r.err != nil {
			partialErrors = append(partialErrors, r.err.Error())
		}
		if len(r.findings) > 0 {
			allFindings = append(allFindings, r.findings...)
		}
	}

	if err := ctx.Err(); err != nil {
		res.State = finding.ScannerFailed
		res.Failure = scanner.FailureCanceled
		res.Message = err.Error()
		res.Duration = time.Since(started)
		return res
	}

	if len(panics) > 0 {
		res.State = finding.ScannerPartial
		res.Failure = scanner.FailurePanic
		res.Message = strings.Join(panics, "; ")
	} else if len(partialErrors) > 0 {
		res.State = finding.ScannerPartial
		res.Failure = scanner.FailurePartial
		res.Message = strings.Join(partialErrors, "; ")
	} else if len(allFindings) > 0 {
		res.State = finding.ScannerFindings
	} else {
		res.State = finding.ScannerClean
	}

	res.Findings = allFindings
	res.Duration = time.Since(started)
	return res
}
