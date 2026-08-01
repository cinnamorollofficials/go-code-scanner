package adapters

import (
	"fmt"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
	"github.com/cinnamorollofficials/go-code-scanner/scanner/command"
)

const (
	Gofmt       = "gofmt"
	GoVet       = "go-vet"
	GoTest      = "go-test"
	Govulncheck = "govulncheck"
	Gosec       = "gosec"
	Gitleaks    = "gitleaks"
	Trivy       = "trivy"
	OSVScanner  = "osv-scanner"
	Semgrep     = "semgrep"
)

type Options struct {
	Args             []string
	Workspace        string
	OnMissing        string
	Environment      []string
	MaxOutputBytes   int
	SnapshotMaxFiles int64
	SnapshotMaxBytes int64
}

func New(id, name string, options Options) (scanner.Scanner, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("adapter scanner id is required")
	}
	spec, err := Spec(id, name, options)
	if err != nil {
		return nil, err
	}
	return command.New(spec)
}

func Spec(id, name string, options Options) (command.Spec, error) {
	var spec command.Spec
	switch name {
	case Gofmt:
		spec = command.Spec{ID: id, Domain: finding.Quality, Command: []string{"gofmt", "-l", "."},
			Severity: finding.Low, Category: "formatting", Description: "Go source is not formatted with gofmt",
			OutputFormat: command.OutputPaths, FindingsOnOutput: true}
	case GoVet:
		spec = command.Spec{ID: id, Domain: finding.Reliability, Command: []string{"go", "vet", "./..."},
			Severity: finding.High, Category: "static_analysis", Description: "go vet reported a possible correctness issue"}
	case GoTest:
		spec = command.Spec{ID: id, Domain: finding.Reliability, Command: []string{"go", "test", "./..."},
			Severity: finding.High, Category: "test_failure", Description: "Go test suite failed"}
	case Govulncheck:
		spec = command.Spec{ID: id, Domain: finding.SupplyChain, Command: []string{"govulncheck", "./..."},
			FindingExitCodes: []int{3}, Severity: finding.High, Category: "vulnerability", Description: "govulncheck reported a reachable vulnerability"}
	case Gosec:
		spec = command.Spec{ID: id, Domain: finding.Security, Command: []string{"gosec", "./..."},
			FindingExitCodes: []int{1}, Severity: finding.High, Category: "static_analysis", Description: "gosec reported a security issue"}
	case Gitleaks:
		spec = command.Spec{ID: id, Domain: finding.Security, Command: []string{"gitleaks", "dir", "--redact", "--no-banner", "."},
			FindingExitCodes: []int{1}, Severity: finding.Critical, Category: "secret_leak", Description: "Gitleaks reported a potential secret"}
	case Trivy:
		spec = command.Spec{ID: id, Domain: finding.SupplyChain, Command: []string{"trivy", "fs", "--exit-code", "1", "--quiet", "."},
			FindingExitCodes: []int{1}, Severity: finding.High, Category: "vulnerability", Description: "Trivy reported a filesystem vulnerability"}
	case OSVScanner:
		spec = command.Spec{ID: id, Domain: finding.SupplyChain, Command: []string{"osv-scanner", "scan", "source", "--verbosity=error", "--recursive", "."},
			FindingExitCodes: []int{1}, Severity: finding.High, Category: "vulnerability", Description: "OSV-Scanner reported a dependency vulnerability"}
	case Semgrep:
		spec = command.Spec{ID: id, Domain: finding.Security, Command: []string{"semgrep", "scan", "--error", "--quiet", "."},
			FindingExitCodes: []int{1}, Severity: finding.High, Category: "static_analysis", Description: "Semgrep reported a security finding"}
	default:
		return command.Spec{}, fmt.Errorf("unknown adapter %q", name)
	}
	if len(options.Args) > 0 {
		spec.Command = append(spec.Command[:1], options.Args...)
	}
	spec.Workspace, spec.OnMissing, spec.Environment = options.Workspace, options.OnMissing, options.Environment
	spec.MaxOutputBytes = options.MaxOutputBytes
	spec.SnapshotMaxFiles, spec.SnapshotMaxBytes = options.SnapshotMaxFiles, options.SnapshotMaxBytes
	return spec, nil
}
