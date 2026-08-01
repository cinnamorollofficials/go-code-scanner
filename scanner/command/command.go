package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/gitrepo"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
	"github.com/cinnamorollofficials/go-code-scanner/workspace"
)

const (
	WorkspaceRoot    = "root"
	WorkspaceStaged  = "staged"
	OnMissingSkip    = "skip"
	OnMissingFail    = "fail"
	DefaultMaxOutput = 64 * 1024
)

type Spec struct {
	ID               string
	Domain           finding.Domain
	Command          []string
	Workspace        string
	OnMissing        string
	FindingExitCodes []int
	Severity         finding.Severity
	Category         string
	Description      string
	Version          string
	MaxOutputBytes   int
}

type Scanner struct {
	spec Spec
}

func New(spec Spec) (*Scanner, error) {
	if strings.TrimSpace(spec.ID) == "" {
		return nil, fmt.Errorf("command scanner id is required")
	}
	if len(spec.Command) == 0 || strings.TrimSpace(spec.Command[0]) == "" {
		return nil, fmt.Errorf("command scanner %s: command is required", spec.ID)
	}
	if !spec.Domain.Valid() {
		return nil, fmt.Errorf("command scanner %s: invalid domain %q", spec.ID, spec.Domain)
	}
	if !spec.Severity.Valid() {
		return nil, fmt.Errorf("command scanner %s: invalid severity %q", spec.ID, spec.Severity)
	}
	if strings.TrimSpace(spec.Category) == "" {
		return nil, fmt.Errorf("command scanner %s: category is required", spec.ID)
	}
	if strings.TrimSpace(spec.Description) == "" {
		return nil, fmt.Errorf("command scanner %s: description is required", spec.ID)
	}
	if spec.Workspace == "" {
		spec.Workspace = WorkspaceRoot
	}
	if spec.Workspace != WorkspaceRoot && spec.Workspace != WorkspaceStaged {
		return nil, fmt.Errorf("command scanner %s: invalid workspace %q", spec.ID, spec.Workspace)
	}
	if spec.OnMissing == "" {
		spec.OnMissing = OnMissingFail
	}
	if spec.OnMissing != OnMissingSkip && spec.OnMissing != OnMissingFail {
		return nil, fmt.Errorf("command scanner %s: invalid on_missing %q", spec.ID, spec.OnMissing)
	}
	if len(spec.FindingExitCodes) == 0 {
		spec.FindingExitCodes = []int{1}
	}
	seen := make(map[int]struct{}, len(spec.FindingExitCodes))
	for _, code := range spec.FindingExitCodes {
		if code == 0 {
			return nil, fmt.Errorf("command scanner %s: exit code 0 cannot represent findings", spec.ID)
		}
		if _, ok := seen[code]; ok {
			return nil, fmt.Errorf("command scanner %s: duplicate finding exit code %d", spec.ID, code)
		}
		seen[code] = struct{}{}
	}
	if spec.MaxOutputBytes == 0 {
		spec.MaxOutputBytes = DefaultMaxOutput
	}
	if spec.MaxOutputBytes < 1 {
		return nil, fmt.Errorf("command scanner %s: max output bytes must be at least 1", spec.ID)
	}
	return &Scanner{spec: spec}, nil
}

func (s *Scanner) ID() string { return s.spec.ID }

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean, Version: s.spec.Version}
	finish := func() scanner.Result {
		result.Duration = time.Since(started)
		return result
	}

	executable, err := exec.LookPath(s.spec.Command[0])
	if err != nil {
		result.Message = fmt.Sprintf("executable %q not found", s.spec.Command[0])
		if s.spec.OnMissing == OnMissingSkip {
			result.State = finding.ScannerSkipped
		} else {
			result.State = finding.ScannerFailed
		}
		return finish()
	}

	root := request.Root
	var snapshot *workspace.Snapshot
	if s.spec.Workspace == WorkspaceStaged {
		if request.Mode != "staged" {
			result.State = finding.ScannerFailed
			result.Message = "staged workspace requires staged scan mode"
			return finish()
		}
		repository, openErr := gitrepo.Open(ctx, request.Root)
		if openErr != nil {
			result.State = finding.ScannerFailed
			result.Message = openErr.Error()
			return finish()
		}
		snapshot, err = workspace.MaterializeIndex(ctx, repository, workspace.DefaultLimits())
		if err != nil {
			result.State = finding.ScannerFailed
			result.Message = err.Error()
			return finish()
		}
		defer snapshot.Close()
		root = snapshot.Root()
	}

	command := exec.CommandContext(ctx, executable, s.spec.Command[1:]...)
	command.Dir = root
	output := &limitedBuffer{limit: s.spec.MaxOutputBytes}
	command.Stdout = output
	command.Stderr = output
	err = command.Run()
	if err == nil {
		return finish()
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		result.State = finding.ScannerFailed
		result.Message = ctxErr.Error()
		return finish()
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		result.State = finding.ScannerFailed
		result.Message = fmt.Sprintf("execute command: %v", err)
		return finish()
	}
	exitCode := exitErr.ExitCode()
	if containsExitCode(s.spec.FindingExitCodes, exitCode) {
		result.State = finding.ScannerFindings
		result.Message = fmt.Sprintf("command reported findings with exit code %d", exitCode)
		result.Findings = []finding.Finding{{
			RuleID: s.spec.ID, Tool: s.spec.ID, Domain: s.spec.Domain,
			Category: s.spec.Category, Severity: s.spec.Severity,
			Description: s.spec.Description,
			Location:    finding.Location{File: filepath.ToSlash("."), Line: 1},
			Metadata:    map[string]string{"exit_code": fmt.Sprintf("%d", exitCode)},
		}}
		return finish()
	}
	result.State = finding.ScannerFailed
	result.Message = fmt.Sprintf("command failed with exit code %d", exitCode)
	if output.truncated {
		result.Message += " (output truncated)"
	}
	return finish()
}

func containsExitCode(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

type limitedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	originalLength := len(value)
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return originalLength, nil
	}
	if len(value) > remaining {
		value = value[:remaining]
		b.truncated = true
	}
	_, _ = b.buffer.Write(value)
	return originalLength, nil
}
