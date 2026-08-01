package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
	OutputExitCode   = "exit-code"
	OutputJSONLines  = "json-lines"
	OutputPaths      = "paths"
)

var defaultEnvironment = []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "SYSTEMROOT", "USERPROFILE", "LANG", "LC_ALL"}

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
	SnapshotMaxFiles int64
	SnapshotMaxBytes int64
	OutputFormat     string
	Environment      []string
	FindingsOnOutput bool
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
	if spec.SnapshotMaxFiles == 0 {
		spec.SnapshotMaxFiles = workspace.DefaultMaxFiles
	}
	if spec.SnapshotMaxBytes == 0 {
		spec.SnapshotMaxBytes = workspace.DefaultMaxBytes
	}
	if spec.SnapshotMaxFiles < 1 || spec.SnapshotMaxBytes < 1 {
		return nil, fmt.Errorf("command scanner %s: snapshot limits must be at least 1", spec.ID)
	}
	if spec.OutputFormat == "" {
		spec.OutputFormat = OutputExitCode
	}
	if spec.OutputFormat != OutputExitCode && spec.OutputFormat != OutputJSONLines && spec.OutputFormat != OutputPaths {
		return nil, fmt.Errorf("command scanner %s: invalid output format %q", spec.ID, spec.OutputFormat)
	}
	if spec.FindingsOnOutput && spec.OutputFormat != OutputPaths {
		return nil, fmt.Errorf("command scanner %s: findings_on_output requires paths output format", spec.ID)
	}
	environmentNames := make(map[string]struct{}, len(spec.Environment))
	for _, name := range spec.Environment {
		if !validEnvironmentName(name) {
			return nil, fmt.Errorf("command scanner %s: invalid environment name %q", spec.ID, name)
		}
		if _, ok := environmentNames[name]; ok {
			return nil, fmt.Errorf("command scanner %s: duplicate environment name %q", spec.ID, name)
		}
		environmentNames[name] = struct{}{}
	}
	return &Scanner{spec: spec}, nil
}

func (s *Scanner) ID() string { return s.spec.ID }

func (s *Scanner) Describe() scanner.Descriptor {
	modes := []string{"full", "changed", "staged"}
	if s.spec.Workspace == WorkspaceStaged {
		modes = []string{"staged"}
	}
	return scanner.Descriptor{
		Domain: s.spec.Domain, Version: s.spec.Version,
		Capabilities: []string{"external-command", s.spec.OutputFormat}, SupportedModes: modes,
	}
}

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
			result.Failure = scanner.FailureMissing
		}
		return finish()
	}

	root := request.Root
	var snapshot *workspace.Snapshot
	if s.spec.Workspace == WorkspaceStaged {
		if request.Mode != "staged" {
			result.State = finding.ScannerFailed
			result.Failure = scanner.FailureExecution
			result.Message = "staged workspace requires staged scan mode"
			return finish()
		}
		repository, openErr := gitrepo.Open(ctx, request.Root)
		if openErr != nil {
			result.State = finding.ScannerFailed
			result.Failure = scanner.FailureExecution
			result.Message = openErr.Error()
			return finish()
		}
		snapshot, err = workspace.MaterializeIndex(ctx, repository, workspace.Limits{
			MaxFiles: s.spec.SnapshotMaxFiles,
			MaxBytes: s.spec.SnapshotMaxBytes,
		})
		if err != nil {
			result.State = finding.ScannerFailed
			result.Failure = scanner.FailureExecution
			result.Message = err.Error()
			return finish()
		}
		defer snapshot.Close()
		root = snapshot.Root()
	}

	command := exec.CommandContext(ctx, executable, s.spec.Command[1:]...)
	configureProcessGroup(command)
	command.Cancel = func() error { return terminateProcessGroup(command.Process) }
	command.WaitDelay = 2 * time.Second
	command.Dir = root
	command.Env = allowedEnvironment(s.spec.Environment)
	stdout := &limitedBuffer{limit: s.spec.MaxOutputBytes}
	stderr := &limitedBuffer{limit: s.spec.MaxOutputBytes}
	command.Stdout = stdout
	command.Stderr = stderr
	err = command.Run()
	if err == nil {
		if s.spec.FindingsOnOutput && stdout.buffer.Len() > 0 {
			if stdout.truncated {
				result.State, result.Failure, result.Message = finding.ScannerFailed, scanner.FailureExecution, "path command output exceeded configured limit"
				return finish()
			}
			result.Findings, err = parsePathLines(stdout.buffer.Bytes(), root, s.spec)
			if err != nil {
				result.State, result.Failure, result.Message = finding.ScannerFailed, scanner.FailureExecution, err.Error()
				return finish()
			}
			if len(result.Findings) > 0 {
				result.State = finding.ScannerFindings
				result.Message = "command output reported findings"
			}
		}
		return finish()
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		result.State = finding.ScannerFailed
		result.Failure = scanner.FailureCanceled
		result.Message = ctxErr.Error()
		return finish()
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		result.State = finding.ScannerFailed
		result.Failure = scanner.FailureExecution
		result.Message = fmt.Sprintf("execute command: %v", err)
		return finish()
	}
	exitCode := exitErr.ExitCode()
	if containsExitCode(s.spec.FindingExitCodes, exitCode) {
		result.State = finding.ScannerFindings
		result.Message = fmt.Sprintf("command reported findings with exit code %d", exitCode)
		if s.spec.OutputFormat == OutputJSONLines {
			if stdout.truncated {
				result.State = finding.ScannerFailed
				result.Failure = scanner.FailureExecution
				result.Message = "structured command output exceeded configured limit"
				return finish()
			}
			result.Findings, err = parseJSONLines(stdout.buffer.Bytes(), root, s.spec)
			if err != nil {
				result.State = finding.ScannerFailed
				result.Failure = scanner.FailureExecution
				result.Message = fmt.Sprintf("decode structured command output: %v", err)
				return finish()
			}
		} else {
			result.Findings = []finding.Finding{{
				RuleID: s.spec.ID, Tool: s.spec.ID, Domain: s.spec.Domain,
				Category: s.spec.Category, Severity: s.spec.Severity,
				Description: s.spec.Description,
				Location:    finding.Location{File: filepath.ToSlash("."), Line: 1},
				Metadata:    map[string]string{"exit_code": fmt.Sprintf("%d", exitCode)},
			}}
		}
		return finish()
	}
	result.State = finding.ScannerFailed
	result.Failure = scanner.FailureExecution
	result.Message = fmt.Sprintf("command failed with exit code %d", exitCode)
	if stdout.truncated || stderr.truncated {
		result.Message += " (output truncated)"
	}
	return finish()
}

func parsePathLines(data []byte, root string, spec Spec) ([]finding.Finding, error) {
	var findings []finding.Finding
	seen := make(map[string]struct{})
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path, err := normalizeOutputPath(root, line)
		if err != nil {
			return nil, fmt.Errorf("decode path output: %w", err)
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		findings = append(findings, finding.Finding{
			RuleID: spec.ID, Tool: spec.ID, Domain: spec.Domain, Category: spec.Category,
			Severity: spec.Severity, Description: spec.Description,
			Location: finding.Location{File: path, Line: 1}, Fixable: spec.ID == "gofmt",
		})
	}
	return findings, nil
}

type outputFinding struct {
	RuleID         string           `json:"rule_id"`
	Category       string           `json:"category"`
	Severity       finding.Severity `json:"severity"`
	Description    string           `json:"description"`
	Recommendation string           `json:"recommendation,omitempty"`
	File           string           `json:"file"`
	Line           int              `json:"line"`
}

func parseJSONLines(data []byte, root string, spec Spec) ([]finding.Finding, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var result []finding.Finding
	for {
		var item outputFinding
		if err := decoder.Decode(&item); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		if item.RuleID == "" {
			item.RuleID = spec.ID
		}
		if item.Category == "" {
			item.Category = spec.Category
		}
		if item.Severity == "" {
			item.Severity = spec.Severity
		}
		if item.Description == "" {
			item.Description = spec.Description
		}
		if !item.Severity.Valid() {
			return nil, fmt.Errorf("finding %s has invalid severity %q", item.RuleID, item.Severity)
		}
		file, err := normalizeOutputPath(root, item.File)
		if err != nil {
			return nil, fmt.Errorf("finding %s: %w", item.RuleID, err)
		}
		if item.Line < 0 {
			return nil, fmt.Errorf("finding %s has negative line %d", item.RuleID, item.Line)
		}
		result = append(result, finding.Finding{
			RuleID: item.RuleID, Tool: spec.ID, Domain: spec.Domain,
			Category: item.Category, Severity: item.Severity,
			Description: item.Description, Recommendation: item.Recommendation,
			Location: finding.Location{File: file, Line: item.Line},
		})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("command returned a finding exit code without findings")
	}
	return result, nil
}

func normalizeOutputPath(root, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return ".", nil
	}
	path := filepath.Clean(value)
	if filepath.IsAbs(path) {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		path = relative
	}
	if path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes scanner workspace", value)
	}
	return filepath.ToSlash(path), nil
}

func validEnvironmentName(value string) bool {
	if value == "" {
		return false
	}
	for index, character := range value {
		if (character >= 'A' && character <= 'Z') || character == '_' || (index > 0 && character >= '0' && character <= '9') {
			continue
		}
		return false
	}
	return true
}

func allowedEnvironment(additional []string) []string {
	names := append(append([]string(nil), defaultEnvironment...), additional...)
	seen := make(map[string]struct{}, len(names))
	result := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		if value, ok := os.LookupEnv(name); ok {
			result = append(result, name+"="+value)
		}
	}
	return result
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
